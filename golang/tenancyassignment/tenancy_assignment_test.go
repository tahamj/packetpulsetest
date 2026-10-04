// Package tenancyassignment proves Addenda 1-3: who may create an organisation,
// who may licence one, and how many people may be signed in at once.
//
// These are integration tests against a running server and a real database,
// because each property is a property of the whole stack. A unit test with a
// mocked repository would pass while a missing WHERE clause let anyone create a
// tenant.
package tenancyassignment

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	pingletest "github.com/tahamj/pingletest"
)

// --------------------------------------------------- Addendum 1: tenancy ---

// Signing up must NOT create an organisation. It creates an account in the
// holding organisation, where it waits to be assigned.
func TestSignUpLandsInTheHoldingOrganisation(t *testing.T) {
	pingletest.RequireServer(t)

	token, _, _, _ := pingletest.SignUpUnassigned(t, "holding")

	me := pingletest.Call(t, http.MethodGet, "/user/me", token, nil)
	if me.Status != http.StatusOK {
		t.Fatalf("a freshly signed-up account cannot read /user/me: %d %s", me.Status, me.Raw)
	}

	// It must be scoped to SOMETHING: an unscoped account would 400 on nearly
	// every route and look like a broken server rather than one awaiting
	// assignment.
	if me.String("organisation_id") == "" {
		t.Error("the account has no organisation scope at all; it should be in the holding organisation")
	}
	if got := me.String("organisation_name"); got != "Unassigned" {
		t.Errorf("expected the holding organisation, got %q", got)
	}
	if me.Body["is_org_owner"] == true {
		t.Error("a self-signup became an owner; only a superuser confers ownership")
	}
}

// The organisation name is no longer part of signup, and sending one must be
// refused rather than quietly ignored - a silently dropped field would let a
// caller believe they had named their organisation.
func TestSignUpRejectsAnOrganisationName(t *testing.T) {
	pingletest.RequireServer(t)

	response := pingletest.Call(t, http.MethodPost, "/user/signup", "", map[string]any{
		"organisation_name": "Smuggled In",
		"email":             fmt.Sprintf("smuggle-%d@pingletest.local", randomStamp()),
		"password":          "PingleTest2026x",
		"display_name":      "Smuggler",
	})
	if response.Status == http.StatusCreated {
		t.Error("signup accepted an organisation name; only a superuser creates organisations")
	}
}

// Only a superuser creates a tenant.
func TestOnlySuperUserCreatesAnOrganisation(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "orgcreate")

	response := pingletest.Call(t, http.MethodPost, "/platform/organisation/add", ownerToken, map[string]any{
		"org_code": "SNEAKY", "org_name": "Sneaky Telecom",
	})
	if response.Status != http.StatusForbidden {
		t.Errorf("an organisation owner could reach organisation creation: %d %s",
			response.Status, response.Raw)
	}

	// And it is genuinely reachable by a superuser, so the test above is
	// measuring authority rather than a route that does not exist.
	superUser := pingletest.SuperUserToken(t)
	stamp := randomStamp()
	allowed := pingletest.Call(t, http.MethodPost, "/platform/organisation/add", superUser, map[string]any{
		"org_code": pingletest.UniqueCode("OK"),
		"org_name": fmt.Sprintf("Allowed %d", stamp),
	})
	if allowed.Status != http.StatusCreated {
		t.Errorf("a superuser could not create an organisation: %d %s", allowed.Status, allowed.Raw)
	}
}

// Assignment moves the account, gives it a role, and invalidates the token it
// held against the holding organisation.
func TestSuperUserAssignsAnAccountToAnOrganisation(t *testing.T) {
	pingletest.RequireServer(t)

	holdingToken, userId, email, password := pingletest.SignUpUnassigned(t, "assign")
	superUser := pingletest.SuperUserToken(t)

	stamp := randomStamp()
	created := pingletest.Call(t, http.MethodPost, "/platform/organisation/add", superUser, map[string]any{
		"org_code": pingletest.UniqueCode("AS"),
		"org_name": fmt.Sprintf("Assigned %d", stamp),
	})
	organisationId := created.String("organisation_id")

	assigned := pingletest.Call(t, http.MethodPost,
		"/platform/organisation/"+organisationId+"/assign", superUser, map[string]any{
			"user_id": userId, "role_name": "NOC Engineer",
		})
	if assigned.Status != http.StatusOK {
		t.Fatalf("assignment failed: %d %s", assigned.Status, assigned.Raw)
	}

	// The old token was minted against the holding organisation. Leaving it live
	// would leave a token scoped to a tenancy the account has left.
	stale := pingletest.Call(t, http.MethodGet, "/user/me", holdingToken, nil)
	if stale.Status == http.StatusOK {
		t.Error("the pre-assignment token still works; it is scoped to the organisation they left")
	}

	fresh := pingletest.SignIn(t, email, password)
	me := pingletest.Call(t, http.MethodGet, "/user/me", fresh, nil)
	if got := me.String("organisation_id"); got != organisationId {
		t.Errorf("after assignment the account is in %q, expected %q", got, organisationId)
	}
	if got := me.String("role_name"); got != "NOC Engineer" {
		t.Errorf("expected the assigned role, got %q", got)
	}
}

// The holding organisation is a waiting room, not a destination.
func TestAccountsCannotBeAssignedIntoTheHoldingOrganisation(t *testing.T) {
	pingletest.RequireServer(t)

	_, userId, _, _ := pingletest.SignUpUnassigned(t, "intoholding")
	superUser := pingletest.SuperUserToken(t)

	waiting := pingletest.Call(t, http.MethodGet, "/platform/user/unassigned", superUser, nil)
	holdingId := waiting.String("holding_organisation_id")
	if holdingId == "" {
		t.Fatal("the server did not report a holding organisation")
	}

	response := pingletest.Call(t, http.MethodPost,
		"/platform/organisation/"+holdingId+"/assign", superUser, map[string]any{
			"user_id": userId, "role_name": "Viewer",
		})
	if response.Status < 400 {
		t.Errorf("an account was assigned INTO the holding organisation: %d %s",
			response.Status, response.Raw)
	}
}

// An account already in a real tenant is not moved by this operation: doing so
// would carry its staff row across a tenant boundary while the previous
// organisation's audit trail kept pointing at it.
func TestAnAssignedAccountIsNotReassigned(t *testing.T) {
	pingletest.RequireServer(t)

	_, userId, _, _ := pingletest.SignUpUnassigned(t, "reassign")
	superUser := pingletest.SuperUserToken(t)

	first := newOrganisation(t, superUser, "first")
	second := newOrganisation(t, superUser, "second")

	if response := pingletest.Call(t, http.MethodPost,
		"/platform/organisation/"+first+"/assign", superUser, map[string]any{
			"user_id": userId, "role_name": "Viewer",
		}); response.Status != http.StatusOK {
		t.Fatalf("the first assignment failed: %d %s", response.Status, response.Raw)
	}

	response := pingletest.Call(t, http.MethodPost,
		"/platform/organisation/"+second+"/assign", superUser, map[string]any{
			"user_id": userId, "role_name": "Viewer",
		})
	if response.Status < 400 {
		t.Errorf("an assigned account was moved between tenants: %d %s",
			response.Status, response.Raw)
	}
}

// -------------------------------------------------- Addendum 2: licences ---

// Only a superuser issues a licence.
func TestOnlySuperUserIssuesALicence(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, organisationId := pingletest.SignUpOrganisation(t, "licauth")

	response := pingletest.Call(t, http.MethodPost, "/platform/licence/issue", ownerToken, map[string]any{
		"organisation_id": organisationId, "plan_code": "trial",
		"currency_code": "INR", "duration_months": 99,
		"seat_limit": 9999, "site_limit": 9999, "is_complimentary": true,
	})
	if response.Status != http.StatusForbidden {
		t.Errorf("an organisation owner could issue itself a licence: %d %s",
			response.Status, response.Raw)
	}
}

// Signing up must not grant its own trial: a tenant that can licence itself is
// not one a superuser controls.
func TestSignUpGrantsNoLicence(t *testing.T) {
	pingletest.RequireServer(t)

	token, _, _, _ := pingletest.SignUpUnassigned(t, "nolicence")

	response := pingletest.Call(t, http.MethodGet, "/licence/my", token, nil)
	if response.Status == http.StatusOK && response.Body["has_licence"] == true {
		t.Error("signing up granted a licence; only a superuser issues one")
	}
}

// The holding organisation is not a customer and cannot hold a licence.
func TestHoldingOrganisationCannotBeLicensed(t *testing.T) {
	pingletest.RequireServer(t)

	superUser := pingletest.SuperUserToken(t)
	waiting := pingletest.Call(t, http.MethodGet, "/platform/user/unassigned", superUser, nil)
	holdingId := waiting.String("holding_organisation_id")

	response := pingletest.Call(t, http.MethodPost, "/platform/licence/issue", superUser, map[string]any{
		"organisation_id": holdingId, "plan_code": "trial",
		"currency_code": "INR", "duration_months": 12,
		"seat_limit": 5, "site_limit": 5, "is_complimentary": true,
	})
	if response.Status < 400 {
		t.Errorf("the holding organisation was licensed: %d %s", response.Status, response.Raw)
	}
	// It must be refused as a RULE, not as a server failure: a 500 here would
	// mean the database caught it and nobody classified it.
	if response.Status >= 500 {
		t.Errorf("refused with %d; a rule violation is the caller being wrong, not the server failing",
			response.Status)
	}
}

// ----------------------------------------------------- Addendum 3: seats ---

// A licence for N people admits N signed in at once, and the (N+1)th is
// refused with an error that says what to do about it.
func TestSeatsAreCountedAsConcurrentSessions(t *testing.T) {
	pingletest.RequireServer(t)

	// One seat, two people.
	ownerToken, organisationId := pingletest.SignUpOrganisationWithSeats(t, "seats", 1)
	colleagueEmail, colleaguePassword := addColleague(t, ownerToken, organisationId)

	// The owner already holds the only seat from the sign-in inside the helper.
	response := pingletest.Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": colleagueEmail, "password": colleaguePassword,
	})
	if response.Status != http.StatusForbidden {
		t.Fatalf("a second person signed in against a one-seat licence: %d %s",
			response.Status, response.Raw)
	}
	if got := response.ErrorCode(); got != "seat_limit_reached" {
		t.Errorf("refused with %q; the client cannot tell seats from an expired licence", got)
	}
	// The remedy is a colleague signing out, not re-typing a password, so the
	// refusal has to carry an action the interface can act on.
	if action, _ := response.Body["error"].(map[string]any)["action"].(string); action == "" {
		t.Error("the refusal carries no action; the client cannot offer a way forward")
	}
}

// Signing in twice must not lock someone out of their own account.
func TestOwnSessionsDoNotConsumeASecondSeat(t *testing.T) {
	pingletest.RequireServer(t)

	stamp := randomStamp()
	email := fmt.Sprintf("selfseat-%d@pingletest.local", stamp)
	const password = "PingleTest2026x"

	signUp := pingletest.Call(t, http.MethodPost, "/user/signup", "", map[string]any{
		"email": email, "password": password, "display_name": "Self Seat",
	})
	userId := signUp.String("user_id")

	superUser := pingletest.SuperUserToken(t)
	organisationId := newOrganisation(t, superUser, "selfseat")
	pingletest.Call(t, http.MethodPost, "/platform/organisation/"+organisationId+"/assign",
		superUser, map[string]any{"user_id": userId, "role_name": "Administrator", "is_org_owner": true})
	pingletest.IssueLicence(t, superUser, organisationId, 1)

	// Three sign-ins by the same person against a one-seat licence. A seat is
	// held by a PERSON, not by a tab.
	// Each completes both steps: a session is what holds a seat, and only the
	// second step opens one. SignIn fails the test on any refusal.
	for attempt := 1; attempt <= 3; attempt++ {
		if token := pingletest.SignIn(t, email, password); token == "" {
			t.Fatalf("sign-in %d of the same person opened no session", attempt)
		}
	}
}

// A superuser is never seat-limited: they belong to no tenant.
func TestSuperUserIsNotSeatLimited(t *testing.T) {
	pingletest.RequireServer(t)

	for attempt := 1; attempt <= 3; attempt++ {
		if token := pingletest.SuperUserSignIn(t); token == "" {
			t.Fatalf("the superuser was refused on attempt %d", attempt)
		}
	}
}

// Seats may be set below the staff count: that is what a floating licence IS.
func TestSeatsMayBeSetBelowTheStaffCount(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, organisationId := pingletest.SignUpOrganisationWithSeats(t, "floating", 5)
	addColleague(t, ownerToken, organisationId)

	superUser := pingletest.SuperUserToken(t)
	licence := pingletest.Call(t, http.MethodGet,
		"/platform/organisation/"+organisationId+"/licence", superUser, nil)
	licenceId := licence.String("licence_id")
	if licenceId == "" {
		t.Fatalf("no licence found for the organisation: %s", licence.Raw)
	}

	response := pingletest.Call(t, http.MethodPut,
		"/platform/licence/"+licenceId+"/seats", superUser, map[string]any{
			"seat_limit": 1, "site_limit": 200,
		})
	if response.Status != http.StatusOK {
		t.Errorf("a floating licence below the staff count was refused: %d %s",
			response.Status, response.Raw)
	}
}

// ------------------------------------------------------------- helpers ---

func newOrganisation(t *testing.T, superUserToken, prefix string) string {
	t.Helper()

	stamp := randomStamp()
	created := pingletest.Call(t, http.MethodPost, "/platform/organisation/add", superUserToken,
		map[string]any{
			"org_code": pingletest.UniqueCode(prefix),
			"org_name": fmt.Sprintf("%s %d", prefix, stamp),
		})
	if created.Status != http.StatusCreated {
		t.Fatalf("creating %s: %d %s", prefix, created.Status, created.Raw)
	}
	return created.String("organisation_id")
}

// addColleague adds a second person to an organisation and returns their
// credentials.
func addColleague(t *testing.T, ownerToken, organisationId string) (email, password string) {
	t.Helper()

	roles := pingletest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	roleId := ""
	for _, role := range pingletest.ListOf(t, roles, "roles") {
		if role["role_name"] == "NOC Engineer" {
			roleId, _ = role["role_id"].(string)
		}
	}
	if roleId == "" {
		t.Skip("no NOC Engineer role available")
	}

	stamp := randomStamp()
	email = fmt.Sprintf("colleague-%d@pingletest.local", stamp)
	password = "PingleTest2026x"

	response := pingletest.Call(t, http.MethodPost, "/user/add", ownerToken, map[string]any{
		"email": email, "password": password,
		"display_name": "Colleague", "role_id": roleId,
	})
	if response.Status != http.StatusCreated {
		t.Fatalf("adding a colleague: %d %s", response.Status, response.Raw)
	}
	return email, password
}

// randomStamp gives each test its own names so a suite can run repeatedly
// against the same database without colliding.
//
// Use it whole. Truncating it with % does not do what it appears to: this
// clock advances in microseconds, so the low digits of UnixNano are always
// zero and "stamp%100000" has a hundred outcomes, not a hundred thousand. For
// anything that has to be unique against rows already stored - an org_code,
// which the database declares UNIQUE - use pingletest.UniqueCode instead.
func randomStamp() int64 { return time.Now().UnixNano() }

// A password alone opens nothing. The token comes from the second step, a
// wrong code says how many tries are left, and a challenge id that is not one
// is refused outright.
func TestSignInNeedsTheSecondFactor(t *testing.T) {
	pingletest.RequireServer(t)

	_, _, email, password := pingletest.SignUpUnassigned(t, "twostep")

	first := pingletest.Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": email, "password": password,
	})
	if first.Status != http.StatusOK || first.String("token") != "" || first.String("method") != "totp" {
		t.Fatalf("the password step = %d %s, want an authenticator challenge and no token", first.Status, first.Raw)
	}
	challengeId := first.String("challenge_id")

	wrong := pingletest.Call(t, http.MethodPost, "/user/signin/verify", "", map[string]any{
		"challenge_id": challengeId, "code": "000000",
	})
	if wrong.Status != http.StatusUnauthorized || wrong.ErrorCode() != "invalid_code" {
		t.Fatalf("a wrong code = %d %s", wrong.Status, wrong.Raw)
	}
	if details, _ := wrong.Body["error"].(map[string]any)["details"].(map[string]any); details["attempts_left"] != "4" {
		t.Errorf("a wrong code did not say four tries remain: %s", wrong.Raw)
	}

	malformed := pingletest.Call(t, http.MethodPost, "/user/signin/verify", "", map[string]any{
		"challenge_id": "not-a-challenge", "code": "123456",
	})
	if malformed.Status != http.StatusBadRequest {
		t.Errorf("a malformed challenge = %d, want 400", malformed.Status)
	}

	right := pingletest.Call(t, http.MethodPost, "/user/signin/verify", "", map[string]any{
		"challenge_id": challengeId, "code": pingletest.NextCode(t, email),
	})
	token := right.String("token")
	if right.Status != http.StatusOK || token == "" {
		t.Fatalf("the right code = %d %s", right.Status, right.Raw)
	}
	if me := pingletest.Call(t, http.MethodGet, "/user/me", token, nil); me.Status != http.StatusOK {
		t.Errorf("the token from the second step does not work: %d", me.Status)
	}

	// A finished sign-in cannot be finished again.
	again := pingletest.Call(t, http.MethodPost, "/user/signin/verify", "", map[string]any{
		"challenge_id": challengeId, "code": pingletest.NextCode(t, email),
	})
	if again.Status != http.StatusUnauthorized || again.ErrorCode() != "challenge_expired" {
		t.Errorf("finishing a sign-in twice = %d %s", again.Status, again.Raw)
	}
}
