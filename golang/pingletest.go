// Package pingletest holds the shared harness for Pingle's guard suite.
//
// These are INTEGRATION tests: they drive the real API over HTTP against a
// running server and a real database, because the properties they protect -
// tenant isolation, permission enforcement, audit integrity - are properties
// of the whole stack, not of any one function. A unit test with a mocked
// repository would pass while a missing WHERE clause leaked another tenant's
// rows.
//
// Every test skips, rather than fails, when the server is not running, so the
// suite is safe to run in a checkout that has not been started.
package pingletest

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// BaseURL is where the suite expects the API.
func BaseURL() string {
	if url := os.Getenv("PINGLE_TEST_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

var httpClient = &http.Client{Timeout: 180 * time.Second}

// RequireServer skips the test when the API is not reachable.
func RequireServer(t *testing.T) {
	t.Helper()

	response, err := httpClient.Get(BaseURL() + "/healthz")
	if err != nil {
		t.Skipf("Pingle is not running at %s: %v", BaseURL(), err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		t.Skipf("Pingle at %s answered /healthz with %d", BaseURL(), response.StatusCode)
	}
}

// Response is a decoded API reply.
type Response struct {
	Status int
	Body   map[string]any
	Raw    []byte
}

// String reads a top-level string field.
func (r Response) String(key string) string {
	if value, ok := r.Body[key].(string); ok {
		return value
	}
	return ""
}

// Float reads a top-level numeric field.
func (r Response) Float(key string) float64 {
	if value, ok := r.Body[key].(float64); ok {
		return value
	}
	return 0
}

// ErrorCode reads the code out of the error envelope.
func (r Response) ErrorCode() string {
	if envelope, ok := r.Body["error"].(map[string]any); ok {
		if code, ok := envelope["code"].(string); ok {
			return code
		}
	}
	return ""
}

// Call makes an authenticated API request. An empty token sends none.
func Call(t *testing.T, method, path, token string, payload any) Response {
	t.Helper()

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("encoding request: %v", err)
		}
		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, BaseURL()+"/api/v1"+path, body)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer func() { _ = response.Body.Close() }()

	raw, _ := io.ReadAll(response.Body)
	decoded := map[string]any{}
	_ = json.Unmarshal(raw, &decoded)

	return Response{Status: response.StatusCode, Body: decoded, Raw: raw}
}

// UniqueCode builds an organisation code that is actually unique.
//
// It replaced a clock-derived one, and the reason is worth recording. The old
// form was fmt.Sprintf("%s%d", prefix, time.Now().UnixNano()%100000), which
// looks like 100,000 possibilities and is not: this machine's clock advances in
// microseconds, so the low three digits of UnixNano are always zero and the
// expression yields exactly 100 distinct values. org_code is UNIQUE, the suite
// runs repeatedly against the same database, and by the twenty-fourth
// organisation sharing a prefix the suite was failing two runs in five with
// "An organisation with that code already exists" - a fixture collision that
// read, in the output, exactly like a tenancy bug.
//
// crypto/rand because the point is uniqueness against rows already stored, not
// unpredictability; math/rand seeded from the same clock would have reproduced
// the problem in a different shape.
func UniqueCode(prefix string) string {
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}
	buffer := make([]byte, 5)
	if _, err := rand.Read(buffer); err != nil {
		// Unreachable in practice, and a panic here is the honest outcome: a
		// fixture that cannot guarantee uniqueness would fail later, further
		// away, with a misleading message.
		panic("pingletest: no randomness available for a unique code: " + err.Error())
	}
	return strings.ToUpper(prefix) + hex.EncodeToString(buffer)
}

// SignUpOrganisation creates a fresh, licensed organisation and returns its
// owner token.
//
// Each test gets its OWN organisation, named after the moment it ran, so tests
// cannot interfere with one another or with the demo data.
//
// Since Addendum 1 this is no longer one call. Signing up lands an account in
// the holding organisation; only a superuser creates a tenant, assigns the
// account to it, and issues the licence that makes the product usable. The
// helper performs all four steps so that a test which only needs "an owner of a
// working organisation" still reads as one line.
func SignUpOrganisation(t *testing.T, prefix string) (token, organisationId string) {
	t.Helper()
	return SignUpOrganisationWithSeats(t, prefix, 25)
}

// SignUpOrganisationWithSeats is SignUpOrganisation with an explicit seat
// count, for tests that care about the concurrent-seat limit (Addendum 3).
func SignUpOrganisationWithSeats(t *testing.T, prefix string, seats int) (token, organisationId string) {
	t.Helper()

	stamp := time.Now().UnixNano()
	email := fmt.Sprintf("%s-%d@pingletest.local", prefix, stamp)
	const password = "PingleTest2026x"

	signUp := Call(t, http.MethodPost, "/user/signup", "", map[string]any{
		"email":        email,
		"password":     password,
		"display_name": prefix + " owner",
	})
	if signUp.Status != http.StatusCreated {
		t.Fatalf("signing up %s: %d %s", prefix, signUp.Status, signUp.Raw)
	}
	userId := signUp.String("user_id")

	superUser := SuperUserToken(t)

	created := Call(t, http.MethodPost, "/platform/organisation/add", superUser, map[string]any{
		"org_code":      UniqueCode("T"),
		"org_name":      fmt.Sprintf("%s %d", prefix, stamp),
		"country_code":  "IN",
		"contact_email": email,
	})
	if created.Status != http.StatusCreated {
		t.Fatalf("creating an organisation for %s: %d %s", prefix, created.Status, created.Raw)
	}
	organisationId = created.String("organisation_id")

	assigned := Call(t, http.MethodPost,
		"/platform/organisation/"+organisationId+"/assign", superUser, map[string]any{
			"user_id":      userId,
			"role_name":    "Administrator",
			"is_org_owner": true,
		})
	if assigned.Status != http.StatusOK {
		t.Fatalf("assigning %s to its organisation: %d %s", prefix, assigned.Status, assigned.Raw)
	}

	IssueLicence(t, superUser, organisationId, seats)

	// Assignment revokes every session minted against the holding organisation,
	// so the token from signup is deliberately dead by now: sign in again to get
	// one scoped to the real tenant.
	return SignIn(t, email, password), organisationId
}

// IssueLicence gives an organisation an active licence. Superuser-only
// (Addendum 2).
func IssueLicence(t *testing.T, superUserToken, organisationId string, seats int) {
	t.Helper()

	response := Call(t, http.MethodPost, "/platform/licence/issue", superUserToken, map[string]any{
		"organisation_id":  organisationId,
		"plan_code":        "trial",
		"currency_code":    "INR",
		"duration_months":  12,
		"seat_limit":       seats,
		"site_limit":       200,
		"is_complimentary": true,
	})
	if response.Status != http.StatusCreated && response.Status != http.StatusOK {
		t.Fatalf("issuing a licence: %d %s", response.Status, response.Raw)
	}
}

// SignUpUnassigned creates an account and leaves it in the holding
// organisation, which is where Addendum 1 says a self-signup belongs.
func SignUpUnassigned(t *testing.T, prefix string) (token, userId, email, password string) {
	t.Helper()

	stamp := time.Now().UnixNano()
	email = fmt.Sprintf("%s-%d@pingletest.local", prefix, stamp)
	password = "PingleTest2026x"

	response := Call(t, http.MethodPost, "/user/signup", "", map[string]any{
		"email":        email,
		"password":     password,
		"display_name": prefix,
	})
	if response.Status != http.StatusCreated {
		t.Fatalf("signing up %s: %d %s", prefix, response.Status, response.Raw)
	}
	return response.String("token"), response.String("user_id"), email, password
}

// SignIn exchanges credentials for a token.
func SignIn(t *testing.T, email, password string) string {
	t.Helper()

	response := Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": email, "password": password,
	})
	if response.Status != http.StatusOK {
		t.Fatalf("signing in %s: %d %s", email, response.Status, response.Raw)
	}
	return response.String("token")
}

// SuperUserToken signs in the platform operator, skipping when it is absent.
func SuperUserToken(t *testing.T) string {
	t.Helper()

	email := os.Getenv("PINGLE_TEST_SUPERUSER")
	password := os.Getenv("PINGLE_TEST_SUPERUSER_PASSWORD")
	if email == "" {
		email, password = "superuser@pingle.local", "PingleSuper2026!"
	}

	response := Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": email, "password": password,
	})
	if response.Status != http.StatusOK {
		// A skip here is dangerous: the suites that need a superuser are the
		// licensing and tenant-assignment ones, and a skipped suite is
		// indistinguishable from a passing suite on the board. Locally a skip
		// is a convenience; anywhere that claims to have verified the product
		// it is a lie, so CI sets PINGLE_TEST_REQUIRE_SUPERUSER and gets a
		// failure instead.
		if os.Getenv("PINGLE_TEST_REQUIRE_SUPERUSER") != "" {
			t.Fatalf("no platform superuser available (sign-in returned %d) and "+
				"PINGLE_TEST_REQUIRE_SUPERUSER is set: seed one with OWNER_EMAIL "+
				"and OWNER_PASSWORD, or point PINGLE_TEST_SUPERUSER at one",
				response.Status)
		}
		t.Skipf("no platform superuser available: %d", response.Status)
	}
	return response.String("token")
}

// ListOf reads a named array out of a list response.
func ListOf(t *testing.T, response Response, key string) []map[string]any {
	t.Helper()

	raw, ok := response.Body[key].([]any)
	if !ok {
		return nil
	}

	rows := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if row, ok := item.(map[string]any); ok {
			rows = append(rows, row)
		}
	}
	return rows
}
