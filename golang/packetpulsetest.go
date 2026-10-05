// Package packetpulsetest holds the shared harness for PacketPulse's guard suite.
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
package packetpulsetest

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// BaseURL is where the suite expects the API.
func BaseURL() string {
	if url := os.Getenv("PACKETPULSE_TEST_URL"); url != "" {
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
		t.Skipf("PacketPulse is not running at %s: %v", BaseURL(), err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		t.Skipf("PacketPulse at %s answered /healthz with %d", BaseURL(), response.StatusCode)
	}
}

// Response is a decoded API reply.
type Response struct {
	Status int
	Body   map[string]any
	Raw    []byte
	Header http.Header
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

	return Response{Status: response.StatusCode, Body: decoded, Raw: raw, Header: response.Header}
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
		panic("packetpulsetest: no randomness available for a unique code: " + err.Error())
	}
	return strings.ToUpper(prefix) + hex.EncodeToString(buffer)
}

// SuperUserEmail is the console operator. The server compiles the address
// in; the suite writes it out rather than importing it, so that a change on
// one side fails here instead of being agreed with.
const SuperUserEmail = "superuser@rummaan53.com"

// Organisation is a licensed customer the suite created for one test: the
// console recorded it and issued its licence file, the file was installed
// where the server reads licences, and its owner signed up with the address
// the licence names.
type Organisation struct {
	Id            string
	Code          string
	LicenceId     string
	LicencePath   string
	OwnerEmail    string
	OwnerPassword string
	OwnerToken    string
}

// SignUpOrganisation creates a fresh, licensed organisation and returns its
// owner token.
//
// Each test gets its OWN organisation, named after the moment it ran, so tests
// cannot interfere with one another or with the demo data.
func SignUpOrganisation(t *testing.T, prefix string) (token, organisationId string) {
	t.Helper()
	return SignUpOrganisationWithSeats(t, prefix, 25)
}

// SignUpOrganisationWithSeats is SignUpOrganisation with an explicit number of
// people on the licence.
func SignUpOrganisationWithSeats(t *testing.T, prefix string, seats int) (token, organisationId string) {
	t.Helper()
	organisation := NewOrganisation(t, prefix, seats)
	return organisation.OwnerToken, organisation.Id
}

// NewOrganisation does what a customer's first day does, in order: the
// console records the organisation and issues its licence file, the file is
// installed on the server, and the owner the licence names signs up -
// finishing with the code emailed to them - and becomes its Administrator.
func NewOrganisation(t *testing.T, prefix string, seats int) Organisation {
	t.Helper()
	organisation := LicensedOrganisation(t, prefix, seats)
	organisation.OwnerToken = SignUpOwner(t, organisation.OwnerEmail, organisation.OwnerPassword, prefix+" owner")
	return organisation
}

// LicensedOrganisation is NewOrganisation stopped before the owner signs up:
// the licence is installed and names an owner who has no account yet.
func LicensedOrganisation(t *testing.T, prefix string, seats int) Organisation {
	t.Helper()

	stamp := time.Now().UnixNano()
	organisation := Organisation{
		Code:          UniqueCode("T"),
		OwnerEmail:    fmt.Sprintf("%s-%d@packetpulsetest.local", prefix, stamp),
		OwnerPassword: "PacketPulseTest2026x",
	}
	superUser := SuperUserToken(t)

	created := Call(t, http.MethodPost, "/platform/organisation/add", superUser, map[string]any{
		"org_code":      organisation.Code,
		"org_name":      fmt.Sprintf("%s %d", prefix, stamp),
		"country_code":  "IN",
		"contact_email": organisation.OwnerEmail,
	})
	if created.Status != http.StatusCreated {
		t.Fatalf("creating an organisation for %s: %d %s", prefix, created.Status, created.Raw)
	}
	organisation.Id = created.String("organisation_id")
	organisation.LicenceId = IssueLicence(t, superUser, organisation.Id, seats)
	filename, content := LicenceFile(t, superUser, organisation.LicenceId, organisation.OwnerEmail)
	organisation.LicencePath = InstallLicence(t, filename, content)
	return organisation
}

// IssueLicence records a licence for an organisation on the console and
// answers with its id. It licenses nothing by itself: a server honours only
// the signed file (see LicenceFile).
func IssueLicence(t *testing.T, superUserToken, organisationId string, seats int) string {
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
	return response.String("licence_id")
}

// LicenceFile has the console sign a recorded licence as the file a customer
// installs, naming the owner who may sign up.
func LicenceFile(t *testing.T, superUserToken, licenceId, ownerEmail string) (filename, content string) {
	t.Helper()

	response := Call(t, http.MethodPost, "/platform/licence/"+licenceId+"/file", superUserToken,
		map[string]any{"owner_email": ownerEmail})
	if response.Status != http.StatusOK || response.String("content") == "" {
		t.Fatalf("signing the licence file: %d %s", response.Status, response.Raw)
	}
	return response.String("filename"), response.String("content")
}

// InstallLicence puts a licence file where the server under test reads them,
// as a customer does, and answers with its path. The suite runs on the same
// machine as that server, locally and in CI.
//
// The file is removed when the test ends: the server re-reads and re-verifies
// every licence at each sign-in, and a directory that grew with every run
// would make each one slower.
func InstallLicence(t *testing.T, filename, content string) string {
	t.Helper()

	directory := setting("PACKETPULSE_TEST_LICENCE_DIR", "LICENCE_DIR")
	if directory == "" {
		t.Fatal("no licence directory: set LICENCE_DIR for the server and the suite " +
			"(or PACKETPULSE_TEST_LICENCE_DIR for the suite alone)")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatalf("the licence directory: %v", err)
	}
	path := filepath.Join(directory, filepath.Base(filename))
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("installing the licence: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })
	return path
}

// SignUpOwner creates the licensed owner's account and finishes it with the
// emailed code, answering with their first session.
func SignUpOwner(t *testing.T, email, password, displayName string) string {
	t.Helper()

	response := Call(t, http.MethodPost, "/user/signup", "", map[string]any{
		"email":        email,
		"password":     password,
		"display_name": displayName,
	})
	if response.Status != http.StatusCreated {
		t.Fatalf("signing up %s: %d %s", email, response.Status, response.Raw)
	}
	return completeSignIn(t, email, response, nil)
}

// SignIn exchanges credentials for a token: the password, then the code the
// server sent.
func SignIn(t *testing.T, email, password string) string {
	t.Helper()
	return SignInAt(t, email, password, nil)
}

// SignInAt is SignIn from a device that reports where it is: the location
// goes with the code, as the app sends it.
func SignInAt(t *testing.T, email, password string, location map[string]any) string {
	t.Helper()

	response := Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": email, "password": password,
	})
	if response.Status != http.StatusOK {
		t.Fatalf("signing in %s: %d %s", email, response.Status, response.Raw)
	}
	return completeSignIn(t, email, response, location)
}

// completeSignIn answers a challenge the way the person would: with the code
// that reached them.
func completeSignIn(t *testing.T, email string, challenge Response, location map[string]any) string {
	t.Helper()

	switch method := challenge.String("method"); method {
	case "email", "sms":
	default:
		t.Fatalf("signing in %s answered with no second step: %d %s", email, challenge.Status, challenge.Raw)
	}
	challengeId := challenge.String("challenge_id")
	step := map[string]any{"challenge_id": challengeId, "code": CodeFor(t, challengeId)}
	if location != nil {
		step["location"] = location
	}
	verified := Call(t, http.MethodPost, "/user/signin/verify", "", step)
	if verified.Status != http.StatusOK {
		t.Fatalf("the second step for %s: %d %s", email, verified.Status, verified.Raw)
	}
	return verified.String("token")
}

// CodeFor is the code the server sent for a sign-in, read from its outbox:
// the file a development or test server writes codes to instead of emailing
// or texting them. The server writes it before it answers, so it is there by
// the time the challenge is.
//
// Lines are matched on the challenge, never on the newest for an address:
// the suites sign in as the same person from several processes at once. The
// last match wins, because a resent code replaces the one before it.
func CodeFor(t *testing.T, challengeId string) string {
	t.Helper()

	path := setting("PACKETPULSE_TEST_OTP_OUTBOX", "OTP_OUTBOX_FILE")
	if path == "" {
		t.Fatal("no code outbox: set OTP_OUTBOX_FILE for the server and the suite " +
			"(or PACKETPULSE_TEST_OTP_OUTBOX for the suite alone)")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the code outbox: %v", err)
	}
	code := ""
	for _, line := range bytes.Split(content, []byte("\n")) {
		var entry struct {
			ChallengeId string `json:"challenge_id"`
			Code        string `json:"code"`
		}
		if json.Unmarshal(line, &entry) == nil && entry.ChallengeId == challengeId && challengeId != "" {
			code = entry.Code
		}
	}
	if code == "" {
		t.Fatalf("the server sent no code for that sign-in to %s", path)
	}
	return code
}

var superUserCache struct {
	sync.Mutex
	token   string
	checked time.Time
}

// SuperUserToken is a session for the console operator, signed in once and
// shared by every suite process.
//
// A sign-in costs an emailed code, and nobody is sent more than ten an hour.
// One sign-in per process - eight suites in a gate run - would lock the
// operator out by the second run, so the session is kept in a file of the
// temporary directory, readable only by its owner, as the spent
// authenticator steps once were. It is checked before it is reused, and
// replaced when it no longer works.
func SuperUserToken(t *testing.T) string {
	t.Helper()

	superUserCache.Lock()
	defer superUserCache.Unlock()
	if superUserCache.token != "" && time.Since(superUserCache.checked) < 10*time.Minute {
		return superUserCache.token
	}

	digest := sha256.Sum256([]byte(BaseURL()))
	path := filepath.Join(os.TempDir(), "packetpulsetest-superuser-"+hex.EncodeToString(digest[:8]))
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		t.Fatalf("the shared superuser session: %v", err)
	}
	defer file.Close()
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("locking the shared superuser session: %v", err)
	}
	defer func() { _ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN) }()

	stored, _ := io.ReadAll(file)
	token := strings.TrimSpace(string(stored))
	if token == "" || !isSuperUserSession(t, token) {
		token = superUserSignIn(t)
		if err := file.Truncate(0); err != nil {
			t.Fatalf("saving the superuser session: %v", err)
		}
		if _, err := file.WriteAt([]byte(token), 0); err != nil {
			t.Fatalf("saving the superuser session: %v", err)
		}
	}
	superUserCache.token, superUserCache.checked = token, time.Now()
	return token
}

// isSuperUserSession reports whether a token still opens the console.
func isSuperUserSession(t *testing.T, token string) bool {
	t.Helper()
	me := Call(t, http.MethodGet, "/user/me", token, nil)
	return me.Status == http.StatusOK && me.Body["is_superuser"] == true
}

func superUserSignIn(t *testing.T) string {
	t.Helper()

	password := setting("PACKETPULSE_TEST_SUPERUSER_PASSWORD", "OWNER_PASSWORD")
	response := Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": SuperUserEmail, "password": password,
	})
	if response.Status != http.StatusOK {
		// A skip here is dangerous: every suite that creates an organisation
		// needs the console, and a skipped suite is indistinguishable from a
		// passing suite on the board. Locally a skip is a convenience;
		// anywhere that claims to have verified the product it is a lie, so
		// CI sets PACKETPULSE_TEST_REQUIRE_SUPERUSER and gets a failure instead.
		if os.Getenv("PACKETPULSE_TEST_REQUIRE_SUPERUSER") != "" {
			t.Fatalf("no console superuser available (sign-in returned %d) and "+
				"PACKETPULSE_TEST_REQUIRE_SUPERUSER is set: run the server as the console "+
				"(LICENCE_SIGNING_KEY and OWNER_PASSWORD set), and give the suite the "+
				"password as OWNER_PASSWORD or PACKETPULSE_TEST_SUPERUSER_PASSWORD",
				response.Status)
		}
		t.Skipf("no console superuser available: %d", response.Status)
	}

	return completeSignIn(t, SuperUserEmail, response, nil)
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

// setting reads a value the suite shares with the server it tests: from the
// environment under the suite's own name or the server's, or - on a
// development machine, as DATABASE_URL is - from the repository's .env under
// the server's name.
func setting(own, servers string) string {
	if value := os.Getenv(own); value != "" {
		return value
	}
	if value := os.Getenv(servers); value != "" {
		return value
	}
	directory, err := os.Getwd()
	if err != nil {
		return ""
	}
	for range 6 {
		if content, err := os.ReadFile(directory + "/.env"); err == nil {
			for _, line := range strings.Split(string(content), "\n") {
				if value, ok := strings.CutPrefix(strings.TrimSpace(line), servers+"="); ok {
					return strings.Trim(strings.TrimSpace(value), `"`)
				}
			}
			return ""
		}
		directory = directory[:max(strings.LastIndex(directory, "/"), 0)]
	}
	return ""
}
