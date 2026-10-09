// Package tenancyisolation guards the property that must never fail: one
// organisation can never read or change another's data.
//
// Everything else in PacketPulse is a feature. This is the promise the product is
// sold on, so it is tested by ATTACKING it: two real organisations are created,
// and every route that takes an id is given the other tenant's id.
package tenancyisolation

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

func testStamp() int64 { return time.Now().UnixNano() }

func TestOneTenantCannotReachAnother(t *testing.T) {
	packetpulsetest.RequireServer(t)

	alphaToken, alphaOrg := packetpulsetest.SignUpOrganisation(t, "alpha")
	betaToken, betaOrg := packetpulsetest.SignUpOrganisation(t, "beta")

	if alphaOrg == betaOrg {
		t.Fatal("two signups produced the same organisation")
	}

	// Alpha builds something worth stealing.
	site := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", alphaToken, map[string]any{
		"site_name": "Alpha core router", "ip_address": "203.0.113.11",
	})
	if site.Status != http.StatusCreated {
		t.Fatalf("alpha could not add a site: %d %s", site.Status, site.Raw)
	}
	alphaSiteId := site.String("dns_site_id")

	t.Run("beta cannot see alpha's inventory", func(t *testing.T) {
		list := packetpulsetest.Call(t, http.MethodGet, "/dnssite/list", betaToken, nil)
		for _, row := range packetpulsetest.ListOf(t, list, "dns_sites") {
			if row["dns_site_id"] == alphaSiteId {
				t.Fatal("beta can see a site belonging to alpha")
			}
		}
	})

	t.Run("beta cannot edit alpha's site", func(t *testing.T) {
		response := packetpulsetest.Call(t, http.MethodPut, "/dnssite/"+alphaSiteId, betaToken,
			map[string]any{"site_name": "taken over", "ip_address": "203.0.113.11"})
		if response.Status != http.StatusNotFound {
			t.Errorf("expected 404, got %d: a cross-tenant id must not resolve", response.Status)
		}
	})

	// The UDP relay and the reflector key both name an endpoint. Alpha's own
	// requests get past the tenant check (200, or 503 where the server has no
	// key secret), so beta's 404 is the predicate, not a missing endpoint.
	for _, route := range []struct{ name, path string }{
		{"relay a UDP test to", "/diagnostic/endpoint/" + alphaSiteId + "/webrtc-offer"},
		{"take the reflector key of", "/diagnostic/endpoint/" + alphaSiteId + "/reflector-key"},
	} {
		t.Run("beta cannot "+route.name+" alpha's endpoint", func(t *testing.T) {
			own := packetpulsetest.Call(t, http.MethodPost, route.path, alphaToken, map[string]any{"sdp": "v=0"})
			if own.Status == http.StatusNotFound || own.Status == http.StatusForbidden {
				t.Fatalf("alpha's own request was refused (%d %s); the test proves nothing", own.Status, own.Raw)
			}
			response := packetpulsetest.Call(t, http.MethodPost, route.path, betaToken, map[string]any{"sdp": "v=0"})
			if response.Status != http.StatusNotFound {
				t.Errorf("expected 404, got %d: a cross-tenant endpoint must not resolve", response.Status)
			}
		})
	}

	t.Run("beta cannot delete alpha's site", func(t *testing.T) {
		response := packetpulsetest.Call(t, http.MethodDelete, "/dnssite/"+alphaSiteId, betaToken, nil)
		if response.Status != http.StatusNotFound {
			t.Errorf("expected 404, got %d", response.Status)
		}
	})

	t.Run("alpha's site is untouched", func(t *testing.T) {
		list := packetpulsetest.Call(t, http.MethodGet, "/dnssite/list", alphaToken, nil)
		found := false
		for _, row := range packetpulsetest.ListOf(t, list, "dns_sites") {
			if row["dns_site_id"] == alphaSiteId {
				found = true
				if row["site_name"] != "Alpha core router" {
					t.Errorf("alpha's site was renamed to %v", row["site_name"])
				}
			}
		}
		if !found {
			t.Error("alpha's own site disappeared")
		}
	})
}

func TestSameAddressIsAllowedInTwoOrganisations(t *testing.T) {
	packetpulsetest.RequireServer(t)

	alphaToken, _ := packetpulsetest.SignUpOrganisation(t, "sharedaddr-a")
	betaToken, _ := packetpulsetest.SignUpOrganisation(t, "sharedaddr-b")

	// Two operators monitoring the same public resolver is normal, so the
	// uniqueness constraint must be per organisation and not global.
	for name, token := range map[string]string{"alpha": alphaToken, "beta": betaToken} {
		response := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", token, map[string]any{
			"site_name": name + " resolver", "ip_address": "9.9.9.9",
		})
		if response.Status != http.StatusCreated {
			t.Fatalf("%s could not monitor a shared address: %d %s",
				name, response.Status, response.Raw)
		}
	}

	// Within ONE organisation the same address twice is still a duplicate.
	response := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", alphaToken, map[string]any{
		"site_name": "duplicate", "ip_address": "9.9.9.9",
	})
	if response.Status != http.StatusConflict {
		t.Errorf("expected 409 for a duplicate inside one organisation, got %d", response.Status)
	}
}

func TestDiagnosticsAreTenantScoped(t *testing.T) {
	packetpulsetest.RequireServer(t)

	alphaToken, _ := packetpulsetest.SignUpOrganisation(t, "diag-a")
	betaToken, _ := packetpulsetest.SignUpOrganisation(t, "diag-b")

	packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", alphaToken, map[string]any{
		"site_name": "loopback", "ip_address": "127.0.0.1",
	})

	ttNumber := fmt.Sprintf("TT-ISOLATION-%d", testStamp())
	submit := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/submit", alphaToken, map[string]any{
		"customer_id": "CUST-ISO", "tt_number": ttNumber,
		"packet_count": 1, "timeout_ms": 900,
	})
	if submit.Status != http.StatusCreated {
		t.Skipf("alpha could not run a diagnostic (licence or sites): %d %s",
			submit.Status, submit.Raw)
	}

	t.Run("beta cannot read alpha's ticket", func(t *testing.T) {
		response := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/tt/"+ttNumber, betaToken, nil)
		if response.Status != http.StatusNotFound {
			t.Errorf("expected 404 for another tenant's ticket, got %d", response.Status)
		}
	})

	t.Run("beta's diagnostic list excludes alpha's", func(t *testing.T) {
		list := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/list", betaToken, nil)
		for _, row := range packetpulsetest.ListOf(t, list, "diagnostics") {
			if row["tt_number"] == ttNumber {
				t.Fatal("beta can see a diagnostic belonging to alpha")
			}
		}
	})
}

func TestSuperUserIsRefusedTenantRoutes(t *testing.T) {
	packetpulsetest.RequireServer(t)

	superToken := packetpulsetest.SuperUserToken(t)

	// A superuser belongs to no organisation. Letting it through a tenant
	// route would mean either reading across tenants or defaulting the scope,
	// and a defaulted scope is how one tenant reads another's rows.
	for _, path := range []string{"/dnssite/list", "/diagnostic/list", "/staff/list"} {
		response := packetpulsetest.Call(t, http.MethodGet, path, superToken, nil)
		if response.Status != http.StatusBadRequest {
			t.Errorf("%s: expected 400 for an unscoped superuser, got %d", path, response.Status)
		}
	}
}

// TestSeatsAreVisibleOnlyWithinTheOrganisation guards the seat screen's
// endpoints, which are a new tenant boundary.
//
// They are worth attacking specifically: the list names people, their email
// addresses and their addresses, and the revoke ends somebody's session. A
// leak here is both a data disclosure and a denial of service against another
// operator's engineers.
func TestSeatsAreVisibleOnlyWithinTheOrganisation(t *testing.T) {
	packetpulsetest.RequireServer(t)

	alpha := packetpulsetest.NewOrganisation(t, "seatalpha", 25)
	alphaToken := alpha.OwnerToken
	betaToken, _ := packetpulsetest.SignUpOrganisation(t, "seatbeta")

	// Each owner's own sign-in is a live session, so both organisations have
	// exactly one seat held and the lists must not overlap.
	alphaSessions := packetpulsetest.Call(t, http.MethodGet, "/staff/session/organisation", alphaToken, nil)
	if alphaSessions.Status != http.StatusOK {
		t.Fatalf("alpha could not list its own sessions: %d %s",
			alphaSessions.Status, alphaSessions.Raw)
	}

	alphaRows := packetpulsetest.ListOf(t, alphaSessions, "sessions")
	if len(alphaRows) == 0 {
		t.Fatal("alpha's own sign-in is not listed as a live session")
	}

	alphaSessionIds := make(map[string]bool, len(alphaRows))
	for _, row := range alphaRows {
		id, _ := row["session_id"].(string)
		alphaSessionIds[id] = true
	}

	betaSessions := packetpulsetest.Call(t, http.MethodGet, "/staff/session/organisation", betaToken, nil)
	if betaSessions.Status != http.StatusOK {
		t.Fatalf("beta could not list its own sessions: %d", betaSessions.Status)
	}

	t.Run("beta cannot see who is signed in at alpha", func(t *testing.T) {
		for _, row := range packetpulsetest.ListOf(t, betaSessions, "sessions") {
			id, _ := row["session_id"].(string)
			if alphaSessionIds[id] {
				t.Fatalf("beta can see a session belonging to alpha: %s", id)
			}
		}
	})

	t.Run("beta cannot sign alpha's people out", func(t *testing.T) {
		for alphaSessionId := range alphaSessionIds {
			response := packetpulsetest.Call(t, http.MethodDelete,
				"/staff/session/organisation/"+alphaSessionId, betaToken, nil)
			if response.Status != http.StatusNotFound {
				t.Fatalf("expected 404 revoking another tenant's session, got %d: "+
					"one operator could sign another operator's engineers out",
					response.Status)
			}
		}
	})

	t.Run("alpha is still signed in afterwards", func(t *testing.T) {
		// The proof that the refusals above refused rather than half-applied.
		after := packetpulsetest.Call(t, http.MethodGet, "/staff/session/organisation", alphaToken, nil)
		if after.Status != http.StatusOK {
			t.Fatalf("alpha lost access to its own sessions: %d", after.Status)
		}
		if len(packetpulsetest.ListOf(t, after, "sessions")) == 0 {
			t.Fatal("alpha's sessions were revoked by beta's attempt")
		}
	})

	t.Run("the licence counts people, not sign-ins", func(t *testing.T) {
		// A second device is a second session, not a second person.
		packetpulsetest.SignIn(t, alpha.OwnerEmail, alpha.OwnerPassword)
		after := packetpulsetest.Call(t, http.MethodGet, "/staff/session/organisation", alphaToken, nil)
		if sessions := len(packetpulsetest.ListOf(t, after, "sessions")); sessions < 2 {
			t.Fatalf("the owner signed in twice but %d sessions are listed", sessions)
		}
		if users := int(after.Float("users_on_licence")); users != 1 {
			t.Errorf("users_on_licence = %d for an organisation of one person, want 1", users)
		}
	})
}

// TestFreeingASeatRequiresStaffManage keeps the endpoint behind a capability
// rather than behind the client hiding a button.
func TestFreeingASeatRequiresStaffManage(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "seatacl")

	// A Viewer holds no staff_manage. The role is seeded, so a missing one is
	// a real failure rather than a reason to skip: skipping here would leave
	// the endpoint's capability gate unverified while the board stayed green.
	roles := packetpulsetest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	if roles.Status != http.StatusOK {
		t.Fatalf("listing roles: %d %s", roles.Status, roles.Raw)
	}

	viewerRoleId := ""
	for _, role := range packetpulsetest.ListOf(t, roles, "roles") {
		if role["role_name"] == "Viewer" {
			viewerRoleId, _ = role["role_id"].(string)
		}
	}
	if viewerRoleId == "" {
		t.Fatal("no seeded Viewer role: the capability gate cannot be tested")
	}

	memberEmail := fmt.Sprintf("member-%s@packetpulse.test", packetpulsetest.UniqueCode("seat"))
	const memberPassword = "MemberPass2026!"

	created := packetpulsetest.Call(t, http.MethodPost, "/user/add", ownerToken, map[string]any{
		"email": memberEmail, "password": memberPassword,
		"display_name": "Seat Member", "role_id": viewerRoleId,
	})
	if created.Status != http.StatusCreated {
		t.Fatalf("could not create a member: %d %s", created.Status, created.Raw)
	}

	memberToken := packetpulsetest.SignIn(t, memberEmail, memberPassword)

	list := packetpulsetest.Call(t, http.MethodGet, "/staff/session/organisation", memberToken, nil)
	if list.Status != http.StatusForbidden {
		t.Errorf("a member listing the organisation's sessions got %d, want 403", list.Status)
	}
}

// A field engineer's sign-in and sign-out are their organisation's check-ins,
// with where each happened. Another organisation never sees them, and an
// administrator's own sign-ins are in the activity trail instead.
func TestCheckinsAreTheOrganisationsOwnFieldSessions(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerA, _ := packetpulsetest.SignUpOrganisation(t, "checkina")
	ownerB, _ := packetpulsetest.SignUpOrganisation(t, "checkinb")

	roles := packetpulsetest.Call(t, http.MethodGet, "/staff/role/list", ownerA, nil)
	engineerRoleId := ""
	for _, role := range packetpulsetest.ListOf(t, roles, "roles") {
		if role["role_name"] == "NOC Engineer" {
			engineerRoleId, _ = role["role_id"].(string)
		}
	}
	email := fmt.Sprintf("engineer-%s@packetpulse.test", packetpulsetest.UniqueCode("field"))
	const password = "FieldPass2026!"
	created := packetpulsetest.Call(t, http.MethodPost, "/user/add", ownerA, map[string]any{
		"email": email, "password": password, "display_name": "Field Engineer", "role_id": engineerRoleId,
	})
	if created.Status != http.StatusCreated {
		t.Fatalf("adding an engineer: %d %s", created.Status, created.Raw)
	}

	token := packetpulsetest.SignInAt(t, email, password, map[string]any{
		"status": "captured", "latitude": 19.076090, "longitude": 72.877426, "accuracy_m": 9,
	})
	if out := packetpulsetest.Call(t, http.MethodPost, "/user/signout", token, map[string]any{
		"location": map[string]any{"status": "unavailable"},
	}); out.Status != http.StatusOK {
		t.Fatalf("signing out: %d %s", out.Status, out.Raw)
	}

	ours := packetpulsetest.ListOf(t, packetpulsetest.Call(t, http.MethodGet, "/staff/checkin/list", ownerA, nil), "checkins")
	if len(ours) != 1 {
		t.Fatalf("A lists %d check-ins, want the engineer's one - and not the owner's own sign-ins", len(ours))
	}
	row := ours[0]
	in, _ := row["in"].(map[string]any)
	out, _ := row["out"].(map[string]any)
	if row["email"] != strings.ToLower(email) || row["session_state"] != "signed_out" ||
		in["status"] != "captured" || in["latitude"] != 19.07609 || out["status"] != "unavailable" {
		t.Errorf("check-in = %v", row)
	}
	if _, leaked := row["token_id"]; leaked {
		t.Error("a check-in carries a token id")
	}

	theirs := packetpulsetest.Call(t, http.MethodGet, "/staff/checkin/list", ownerB, nil)
	if list := packetpulsetest.ListOf(t, theirs, "checkins"); theirs.Status != http.StatusOK || len(list) != 0 {
		t.Errorf("B lists %d check-ins (%d), want none of A's", len(list), theirs.Status)
	}
}

// Results leave an organisation three ways - a ticket's CSV, the API key
// pull, and the push to its own file server - and each carries only its own.
func TestExportsCarryOnlyTheOrganisationsOwnResults(t *testing.T) {
	packetpulsetest.RequireServer(t)
	alphaToken, _ := packetpulsetest.SignUpOrganisation(t, "export-a")
	betaToken, _ := packetpulsetest.SignUpOrganisation(t, "export-b")

	packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", alphaToken, map[string]any{
		"site_name": "loopback", "ip_address": "127.0.0.1",
	})
	ttNumber := fmt.Sprintf("TT-EXPORT-ISO-%d", testStamp())
	submit := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/submit", alphaToken, map[string]any{
		"customer_id": "CUST-ISO", "tt_number": ttNumber, "packet_count": 1, "timeout_ms": 900,
	})
	if submit.Status != http.StatusCreated {
		t.Skipf("alpha could not run a diagnostic: %d %s", submit.Status, submit.Raw)
	}
	request, _ := submit.Body["request"].(map[string]any)
	requestId, _ := request["request_id"].(string)

	t.Run("beta cannot download alpha's ticket", func(t *testing.T) {
		response := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId+"/report.csv", betaToken, nil)
		if response.Status != http.StatusNotFound || strings.Contains(string(response.Raw), ttNumber) {
			t.Errorf("beta's download of alpha's ticket = %d %s", response.Status, response.Raw)
		}
	})

	t.Run("beta's key pulls none of alpha's results", func(t *testing.T) {
		issued := packetpulsetest.Call(t, http.MethodPost, "/apikey/add", betaToken, map[string]any{"label": "iso"})
		if issued.Status != http.StatusCreated {
			t.Skipf("could not issue beta a key: %d", issued.Status)
		}
		pulled := packetpulsetest.Call(t, http.MethodGet, "/result/export.csv", issued.String("api_key"), nil)
		if pulled.Status != http.StatusOK || strings.Contains(string(pulled.Raw), ttNumber) {
			t.Errorf("beta's pull = %d, holding alpha's ticket: %v", pulled.Status, strings.Contains(string(pulled.Raw), ttNumber))
		}
	})

	t.Run("beta cannot see or change alpha's file server", func(t *testing.T) {
		saved := packetpulsetest.Call(t, http.MethodPut, "/export/target", alphaToken, map[string]any{
			"is_enabled": false, "protocol": "sftp", "host": "alpha-files.example.com",
			"username": "alpha", "password": "alpha-secret",
		})
		if saved.Status != http.StatusOK || saved.Body["has_password"] != true {
			t.Fatalf("alpha's save: %d %s", saved.Status, saved.Raw)
		}
		if strings.Contains(string(saved.Raw), "alpha-secret") {
			t.Error("the password came back in the response")
		}
		betaView := packetpulsetest.Call(t, http.MethodGet, "/export/target", betaToken, nil)
		if betaView.Status != http.StatusOK || betaView.String("host") != "" || betaView.Body["has_password"] != false {
			t.Errorf("beta sees %d %s, want an empty target of its own", betaView.Status, betaView.Raw)
		}
		packetpulsetest.Call(t, http.MethodPut, "/export/target", betaToken, map[string]any{
			"is_enabled": false, "protocol": "ftp", "host": "beta-files.example.com",
		})
		alphaView := packetpulsetest.Call(t, http.MethodGet, "/export/target", alphaToken, nil)
		if alphaView.String("host") != "alpha-files.example.com" || alphaView.Body["has_password"] != true {
			t.Errorf("beta's save changed alpha's target: %s", alphaView.Raw)
		}
	})
}
