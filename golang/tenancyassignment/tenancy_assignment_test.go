// Package tenancyassignment proves who may create an organisation, what
// licenses a server to admit its people, and how many people a licence
// covers.
//
// These are integration tests against a running server and a real database,
// because each property is a property of the whole stack. A unit test with a
// mocked repository would pass while a missing WHERE clause let anyone create a
// tenant.
//
// The server under test is the console as well as a customer's server: the
// suite records each organisation and issues its licence file through the
// console, then installs the file where the server reads licences - which is
// what a customer does with the file the vendor sends.
package tenancyassignment

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// ------------------------------------------- who creates an organisation ---

// Signing up creates nothing unless an installed licence names the address
// as its owner: a stranger cannot make themselves a tenant, and an
// organisation's people are added by its administrator, not by signing up.
func TestSignUpNeedsALicenceNamingTheAddress(t *testing.T) {
	packetpulsetest.RequireServer(t)

	licensed := packetpulsetest.LicensedOrganisation(t, "notowner", 5)
	for name, email := range map[string]string{
		"an address no licence names":  fmt.Sprintf("unlicensed-%d@packetpulsetest.local", randomStamp()),
		"someone other than the owner": "not-" + licensed.OwnerEmail,
	} {
		response := packetpulsetest.Call(t, http.MethodPost, "/user/signup", "", map[string]any{
			"email": email, "password": "PacketPulseTest2026x", "display_name": "Stranger",
		})
		if response.Status != http.StatusForbidden || response.String("challenge_id") != "" {
			t.Errorf("%s: sign-up = %d %s, want a refusal and no code", name, response.Status, response.Raw)
		}
		// Refused before anything was stored: the address has no account.
		signIn := packetpulsetest.Call(t, http.MethodPost, "/user/signin", "", map[string]any{
			"email": email, "password": "PacketPulseTest2026x",
		})
		if signIn.Status != http.StatusUnauthorized {
			t.Errorf("%s: a refused sign-up left an account behind (sign-in = %d)", name, signIn.Status)
		}
	}
}

// The organisation name is not part of sign-up - the licence names the
// organisation - and sending one must be refused rather than quietly ignored.
func TestSignUpRejectsAnOrganisationName(t *testing.T) {
	packetpulsetest.RequireServer(t)

	licensed := packetpulsetest.LicensedOrganisation(t, "smuggle", 5)
	response := packetpulsetest.Call(t, http.MethodPost, "/user/signup", "", map[string]any{
		"organisation_name": "Smuggled In",
		"email":             licensed.OwnerEmail,
		"password":          licensed.OwnerPassword,
		"display_name":      "Smuggler",
	})
	if response.Status == http.StatusCreated {
		t.Error("sign-up accepted an organisation name; the licence names the organisation")
	}
	// The address could sign up: it was the extra field that was refused.
	packetpulsetest.SignUpOwner(t, licensed.OwnerEmail, licensed.OwnerPassword, "Owner")
}

// The owner the licence names, once their emailed code is given, is the
// Administrator of the organisation the licence was issued to.
func TestTheLicensedOwnerBecomesTheAdministrator(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "owner", 5)
	me := packetpulsetest.Call(t, http.MethodGet, "/user/me", organisation.OwnerToken, nil)
	if me.Status != http.StatusOK {
		t.Fatalf("the owner cannot read /user/me: %d %s", me.Status, me.Raw)
	}
	if me.String("organisation_id") != organisation.Id || me.Body["is_org_owner"] != true ||
		me.String("role_name") != "Administrator" {
		t.Errorf("the owner is %s, want the Administrator and owner of %s", me.Raw, organisation.Id)
	}
}

// Once the owner has signed in, signing up with their address again is
// refused and changes nothing: otherwise anyone who learned the address could
// replace the owner's password.
func TestAFinishedOwnerSignUpCannotBeRepeated(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "again", 5)
	response := packetpulsetest.Call(t, http.MethodPost, "/user/signup", "", map[string]any{
		"email": organisation.OwnerEmail, "password": "Different2026xx", "display_name": "Usurper",
	})
	if response.Status != http.StatusConflict || response.String("challenge_id") != "" {
		t.Errorf("a second sign-up = %d %s, want 409 and no code", response.Status, response.Raw)
	}
	packetpulsetest.SignIn(t, organisation.OwnerEmail, organisation.OwnerPassword)
}

// Only a superuser creates a tenant.
func TestOnlySuperUserCreatesAnOrganisation(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "orgcreate")

	response := packetpulsetest.Call(t, http.MethodPost, "/platform/organisation/add", ownerToken, map[string]any{
		"org_code": "SNEAKY", "org_name": "Sneaky Telecom",
	})
	if response.Status != http.StatusForbidden {
		t.Errorf("an organisation owner could reach organisation creation: %d %s",
			response.Status, response.Raw)
	}

	// And it is genuinely reachable by a superuser, so the test above is
	// measuring authority rather than a route that does not exist.
	superUser := packetpulsetest.SuperUserToken(t)
	allowed := packetpulsetest.Call(t, http.MethodPost, "/platform/organisation/add", superUser, map[string]any{
		"org_code": packetpulsetest.UniqueCode("OK"),
		"org_name": fmt.Sprintf("Allowed %d", randomStamp()),
	})
	if allowed.Status != http.StatusCreated {
		t.Errorf("a superuser could not create an organisation: %d %s", allowed.Status, allowed.Raw)
	}
}

// --------------------------------------------------------- licence files ---

// Only a superuser records a licence or signs one as a file.
func TestOnlySuperUserIssuesALicence(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "licauth", 5)

	issued := packetpulsetest.Call(t, http.MethodPost, "/platform/licence/issue", organisation.OwnerToken, map[string]any{
		"organisation_id": organisation.Id, "plan_code": "trial",
		"currency_code": "INR", "duration_months": 99,
		"seat_limit": 9999, "site_limit": 9999, "is_complimentary": true,
	})
	if issued.Status != http.StatusForbidden {
		t.Errorf("an organisation owner could issue itself a licence: %d %s", issued.Status, issued.Raw)
	}
	file := packetpulsetest.Call(t, http.MethodPost, "/platform/licence/"+organisation.LicenceId+"/file",
		organisation.OwnerToken, map[string]any{"owner_email": organisation.OwnerEmail})
	if file.Status != http.StatusForbidden || file.String("content") != "" {
		t.Errorf("an organisation owner could sign a licence file: %d %s", file.Status, file.Raw)
	}
}

// Sign-in re-reads the licence file every time. Taken away or altered, it
// admits nobody - with the refusal that says to renew - and the genuine file
// put back admits them again, without a restart.
func TestSignInNeedsAGenuineLicenceFileOnDisk(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "licfile", 5)
	genuine, err := os.ReadFile(organisation.LicencePath)
	if err != nil {
		t.Fatalf("reading the installed licence: %v", err)
	}

	expectNoLicence := func(state string) {
		t.Helper()
		response := packetpulsetest.Call(t, http.MethodPost, "/user/signin", "", map[string]any{
			"email": organisation.OwnerEmail, "password": organisation.OwnerPassword,
		})
		if response.Status != http.StatusForbidden || response.ErrorCode() != "licence_expired" ||
			response.String("challenge_id") != "" {
			t.Errorf("%s: sign-in = %d %s, want licence_expired and no code", state, response.Status, response.Raw)
		}
		if action, _ := response.Body["error"].(map[string]any)["action"].(string); action != "renew" {
			t.Errorf("%s: action = %q, want renew", state, action)
		}
	}

	if err := os.Remove(organisation.LicencePath); err != nil {
		t.Fatal(err)
	}
	expectNoLicence("with the file removed")

	// Five people bought, five hundred claimed: the terms no longer match the
	// signature.
	seats := regexp.MustCompile(`"seat_limit":\s*5\b`)
	if !seats.Match(genuine) {
		t.Fatalf("the licence file carries no seat_limit of 5 to alter: %s", genuine)
	}
	if err := os.WriteFile(organisation.LicencePath, seats.ReplaceAll(genuine, []byte(`"seat_limit": 500`)), 0o600); err != nil {
		t.Fatal(err)
	}
	expectNoLicence("with the file altered")

	if err := os.WriteFile(organisation.LicencePath, genuine, 0o600); err != nil {
		t.Fatal(err)
	}
	packetpulsetest.SignIn(t, organisation.OwnerEmail, organisation.OwnerPassword)
}

// ------------------------------------------------- people on the licence ---

// A licence counts people, not sign-ins: the same person signed in on three
// devices takes one place, and adding a person beyond the licence is refused
// with a refusal that says what to do.
func TestTheLicenceCountsPeopleNotSignIns(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "people", 2)
	for attempt := 1; attempt <= 3; attempt++ {
		if token := packetpulsetest.SignIn(t, organisation.OwnerEmail, organisation.OwnerPassword); token == "" {
			t.Fatalf("sign-in %d of the same person opened no session", attempt)
		}
	}
	addColleague(t, organisation.OwnerToken)

	refused, _ := tryAddColleague(t, organisation.OwnerToken)
	expectFull(t, refused, "a third person on a licence for two")
}

// Someone who has left is switched off, not removed: their record stays, and
// so does their place on the licence. Deleting someone who never used the
// product takes them off it.
func TestAnInactivePersonStillCountsAndADeletedOneDoesNot(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "leaver", 2)
	leaver := addColleague(t, organisation.OwnerToken)

	setEnabled(t, organisation.OwnerToken, leaver, false)
	refused, _ := tryAddColleague(t, organisation.OwnerToken)
	expectFull(t, refused, "a switched-off person freed their place")

	deleted := packetpulsetest.Call(t, http.MethodDelete, "/staff/"+leaver.staffId, organisation.OwnerToken, nil)
	if deleted.Status != http.StatusNoContent {
		t.Fatalf("deleting a person with no records: %d %s", deleted.Status, deleted.Raw)
	}
	if added, _ := tryAddColleague(t, organisation.OwnerToken); added.Status != http.StatusCreated {
		t.Errorf("a deleted person still holds their place: adding another = %d %s", added.Status, added.Raw)
	}
}

// A person with something on record - here, a sign-in - can be switched off
// but not deleted: deleting them would orphan what they did.
func TestAPersonWithRecordsIsSwitchedOffNotDeleted(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "records", 5)
	colleague := addColleague(t, organisation.OwnerToken)
	packetpulsetest.SignIn(t, colleague.email, colleague.password)

	deleted := packetpulsetest.Call(t, http.MethodDelete, "/staff/"+colleague.staffId, organisation.OwnerToken, nil)
	if deleted.Status != http.StatusConflict {
		t.Errorf("deleting a person who has signed in = %d %s, want 409", deleted.Status, deleted.Raw)
	}
	listed := packetpulsetest.Call(t, http.MethodGet, "/staff/list", organisation.OwnerToken, nil)
	if !strings.Contains(string(listed.Raw), colleague.staffId) {
		t.Error("the refused deletion removed the person anyway")
	}
}

// Switching someone off signs them out everywhere at once - their session
// stops working on its next request, not when it expires - and refuses their
// next sign-in at the password, before a code is sent. Switched back on, they
// are admitted again.
func TestASwitchedOffPersonIsSignedOutAndCannotSignIn(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "switchedoff", 5)
	colleague := addColleague(t, organisation.OwnerToken)
	session := packetpulsetest.SignIn(t, colleague.email, colleague.password)

	setEnabled(t, organisation.OwnerToken, colleague, false)
	if me := packetpulsetest.Call(t, http.MethodGet, "/user/me", session, nil); me.Status != http.StatusUnauthorized {
		t.Errorf("a switched-off person's session still answers %d, want 401", me.Status)
	}
	refused := packetpulsetest.Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": colleague.email, "password": colleague.password,
	})
	if refused.Status != http.StatusForbidden || refused.String("challenge_id") != "" {
		t.Errorf("a switched-off person signing in = %d %s, want 403 and no code", refused.Status, refused.Raw)
	}

	setEnabled(t, organisation.OwnerToken, colleague, true)
	packetpulsetest.SignIn(t, colleague.email, colleague.password)
}

// A smaller licence installed over a bigger one - fewer places than people -
// keeps everyone but the owner out until people are deleted or a bigger
// licence arrives. The owner is let in because they are who can fix it.
func TestASmallerLicenceKeepsEveryoneButTheOwnerOut(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "smaller", 5)
	colleague := addColleague(t, organisation.OwnerToken)

	superUser := packetpulsetest.SuperUserToken(t)
	reduced := packetpulsetest.Call(t, http.MethodPut, "/platform/licence/"+organisation.LicenceId+"/seats",
		superUser, map[string]any{"seat_limit": 1, "site_limit": 200})
	if reduced.Status != http.StatusOK {
		t.Fatalf("the console refused a licence for fewer people than the organisation has: %d %s",
			reduced.Status, reduced.Raw)
	}
	_, content := packetpulsetest.LicenceFile(t, superUser, organisation.LicenceId, organisation.OwnerEmail)
	if err := os.WriteFile(organisation.LicencePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	refused := packetpulsetest.Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": colleague.email, "password": colleague.password,
	})
	expectFull(t, refused, "two people signing in on a licence for one")
	if refused.Status != http.StatusForbidden {
		t.Errorf("status = %d, want 403: the credentials were right", refused.Status)
	}
	packetpulsetest.SignIn(t, organisation.OwnerEmail, organisation.OwnerPassword)
}

// ---------------------------------------------------------- second step ---

// A password alone opens nothing. The token comes from the emailed code, a
// wrong code says how many tries are left, and a challenge id that is not one
// is refused outright.
func TestSignInNeedsTheSecondFactor(t *testing.T) {
	packetpulsetest.RequireServer(t)

	organisation := packetpulsetest.NewOrganisation(t, "twostep", 5)

	first := packetpulsetest.Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": organisation.OwnerEmail, "password": organisation.OwnerPassword,
	})
	if first.Status != http.StatusOK || first.String("token") != "" || first.String("method") != "email" {
		t.Fatalf("the password step = %d %s, want an emailed code and no token", first.Status, first.Raw)
	}
	// Enough to know which inbox to open; not the address itself.
	if hint := first.String("email_hint"); hint == "" || hint == organisation.OwnerEmail {
		t.Errorf("email_hint = %q, want the address partly hidden", hint)
	}
	challengeId := first.String("challenge_id")
	code := packetpulsetest.CodeFor(t, challengeId)
	wrongCode := "000000"
	if code == wrongCode {
		wrongCode = "111111"
	}

	wrong := packetpulsetest.Call(t, http.MethodPost, "/user/signin/verify", "", map[string]any{
		"challenge_id": challengeId, "code": wrongCode,
	})
	if wrong.Status != http.StatusUnauthorized || wrong.ErrorCode() != "invalid_code" {
		t.Fatalf("a wrong code = %d %s", wrong.Status, wrong.Raw)
	}
	if details, _ := wrong.Body["error"].(map[string]any)["details"].(map[string]any); details["attempts_left"] != "4" {
		t.Errorf("a wrong code did not say four tries remain: %s", wrong.Raw)
	}

	malformed := packetpulsetest.Call(t, http.MethodPost, "/user/signin/verify", "", map[string]any{
		"challenge_id": "not-a-challenge", "code": "123456",
	})
	if malformed.Status != http.StatusBadRequest {
		t.Errorf("a malformed challenge = %d, want 400", malformed.Status)
	}

	right := packetpulsetest.Call(t, http.MethodPost, "/user/signin/verify", "", map[string]any{
		"challenge_id": challengeId, "code": code,
	})
	token := right.String("token")
	if right.Status != http.StatusOK || token == "" {
		t.Fatalf("the right code = %d %s", right.Status, right.Raw)
	}
	if me := packetpulsetest.Call(t, http.MethodGet, "/user/me", token, nil); me.Status != http.StatusOK {
		t.Errorf("the token from the second step does not work: %d", me.Status)
	}

	// A finished sign-in cannot be finished again.
	again := packetpulsetest.Call(t, http.MethodPost, "/user/signin/verify", "", map[string]any{
		"challenge_id": challengeId, "code": code,
	})
	if again.Status != http.StatusUnauthorized || again.ErrorCode() != "challenge_expired" {
		t.Errorf("finishing a sign-in twice = %d %s", again.Status, again.Raw)
	}
}

// ------------------------------------------------------------- helpers ---

type colleague struct {
	email, password, staffId, userId, roleId string
}

// setEnabled switches a person on or off as the Staff screen does: the whole
// staff row, sent back with is_enabled changed.
func setEnabled(t *testing.T, ownerToken string, person colleague, enabled bool) {
	t.Helper()
	response := packetpulsetest.Call(t, http.MethodPut, "/staff/"+person.staffId, ownerToken, map[string]any{
		"staff_code": "", "full_name": "Colleague", "department": "", "designation": "",
		"role_id": person.roleId, "is_enabled": enabled,
	})
	if response.Status != http.StatusOK {
		t.Fatalf("setting is_enabled=%v: %d %s", enabled, response.Status, response.Raw)
	}
}

// addColleague adds an engineer to the owner's organisation.
func addColleague(t *testing.T, ownerToken string) colleague {
	t.Helper()
	response, added := tryAddColleague(t, ownerToken)
	if response.Status != http.StatusCreated {
		t.Fatalf("adding a colleague: %d %s", response.Status, response.Raw)
	}
	return added
}

// tryAddColleague asks to add an engineer and answers with the reply, for a
// test that expects the addition to be refused.
func tryAddColleague(t *testing.T, ownerToken string) (packetpulsetest.Response, colleague) {
	t.Helper()

	roles := packetpulsetest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	roleId := ""
	for _, role := range packetpulsetest.ListOf(t, roles, "roles") {
		if role["role_name"] == "NOC Engineer" {
			roleId, _ = role["role_id"].(string)
		}
	}
	if roleId == "" {
		t.Fatalf("no NOC Engineer role: %s", roles.Raw)
	}

	added := colleague{
		email:    fmt.Sprintf("colleague-%d@packetpulsetest.local", randomStamp()),
		password: "PacketPulseTest2026x",
		roleId:   roleId,
	}
	response := packetpulsetest.Call(t, http.MethodPost, "/user/add", ownerToken, map[string]any{
		"email": added.email, "password": added.password,
		"display_name": "Colleague", "role_id": roleId,
	})
	added.staffId, added.userId = response.String("staff_id"), response.String("user_id")
	return response, added
}

// expectFull asserts a refusal because the licence covers no more people,
// carrying an action the client can offer.
func expectFull(t *testing.T, response packetpulsetest.Response, what string) {
	t.Helper()
	if response.Status < 400 || response.ErrorCode() != "seat_limit_reached" {
		t.Errorf("%s: %d %s, want seat_limit_reached", what, response.Status, response.Raw)
		return
	}
	if action, _ := response.Body["error"].(map[string]any)["action"].(string); action == "" {
		t.Errorf("%s: the refusal carries no action; the client cannot offer a way forward", what)
	}
}

// randomStamp gives each test its own names so a suite can run repeatedly
// against the same database without colliding.
//
// Use it whole. Truncating it with % does not do what it appears to: this
// clock advances in microseconds, so the low digits of UnixNano are always
// zero and "stamp%100000" has a hundred outcomes, not a hundred thousand. For
// anything that has to be unique against rows already stored - an org_code,
// which the database declares UNIQUE - use packetpulsetest.UniqueCode instead.
func randomStamp() int64 { return time.Now().UnixNano() }
