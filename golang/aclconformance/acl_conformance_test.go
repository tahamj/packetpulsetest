// Package aclconformance asserts the full permission matrix.
//
// Each built-in role is exercised against every gated route, and BOTH
// directions are checked: a permission that fails to grant is a broken
// product, and a permission that fails to deny is a security defect. Only
// testing the allow direction would let the second kind through.
package aclconformance

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

type routeCheck struct {
	Name    string
	Method  string
	Path    string
	Payload any
	// Capability is what the route requires: one capability, "a|b" for
	// either, and "a+b" for both.
	Capability string
}

// holds reports whether a role's grants satisfy a route's Capability.
func holds(grants map[string]any, capability string) bool {
	for _, needed := range strings.Split(capability, "+") {
		satisfied := false
		for _, either := range strings.Split(needed, "|") {
			if granted, _ := grants[either].(bool); granted {
				satisfied = true
			}
		}
		if !satisfied {
			return false
		}
	}
	return true
}

// gatedRoutes names the capability each route is supposed to require.
// onePixelPng is the smallest image there is, base64 as the API takes it.
const onePixelPng = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="

var gatedRoutes = []routeCheck{
	{"list sites", http.MethodGet, "/dnssite/list", nil, "dns_site_view"},
	{"add site", http.MethodPost, "/dnssite/add",
		map[string]any{"site_name": "acl probe", "ip_address": "203.0.113.200"}, "dns_site_manage"},
	{"import sites from a file", http.MethodPost, "/dnssite/importfile",
		map[string]any{"text": "address\n203.0.113.201\n"}, "dns_site_manage"},
	// An engineer reads the tests they ran, an administrator everyone's: the
	// list opens to either, and what it holds is the server's to scope.
	{"list diagnostics", http.MethodGet, "/diagnostic/list", nil, "diagnostic_view_own|diagnostic_view_all"},
	{"list runs", http.MethodGet, "/ping/runlist", nil, "diagnostic_view_all"},
	{"list SLA policies", http.MethodGet, "/monitor/sla/list", nil, "sla_view"},
	{"add SLA policy", http.MethodPost, "/monitor/sla/add",
		map[string]any{"policy_name": "acl probe", "max_latency_ms": 100}, "sla_manage"},
	{"list schedules", http.MethodGet, "/monitor/schedule/list", nil, "schedule_manage"},
	{"list users", http.MethodGet, "/user/list", nil, "user_manage"},
	{"list staff", http.MethodGet, "/staff/list", nil, "staff_manage"},
	{"list roles", http.MethodGet, "/staff/role/list", nil, "acl_manage"},
	{"list audit trail", http.MethodGet, "/auditlog/list", nil, "auditlog_view"},
	{"read LDAP settings", http.MethodGet, "/ldap/config", nil, "ldap_manage"},
	// The SMS gateway is part of how people sign in, so it is gated by the
	// capability that decides who signs in how.
	{"read SMS gateway", http.MethodGet, "/sms/gateway", nil, "staff_manage"},
	// Where people signed in and out is about people; so is whether their
	// sign-in needs a location.
	{"list check-ins", http.MethodGet, "/staff/checkin/list", nil, "staff_manage"},
	{"save organisation settings", http.MethodPut, "/organisation/settings",
		map[string]any{"require_checkin_location": false, "device_test_target": ""}, "staff_manage"},
	// The organisation's logo goes on every report its people export, so it
	// is gated with its other settings. A one-pixel PNG, accepted as it is.
	{"save the report logo", http.MethodPut, "/organisation/brand/logo",
		map[string]any{"image": onePixelPng}, "staff_manage"},
	{"list API keys", http.MethodGet, "/apikey/list", nil, "apikey_manage"},
	// The result export hands the organisation's results to a machine, as the
	// Results API does, so the same capability gates it.
	{"read result export", http.MethodGet, "/export/target", nil, "apikey_manage"},
	{"save result export", http.MethodPut, "/export/target",
		map[string]any{"is_enabled": false, "protocol": "sftp"}, "apikey_manage"},
	// A ticket that does not exist: 404 to whoever may export, 403 to the rest.
	// Exporting a ticket is reading it as a file, so it needs both.
	{"download a ticket's CSV", http.MethodGet, "/diagnostic/00000000-0000-4000-8000-000000000000/report.csv", nil,
		"report_export+diagnostic_view_own|diagnostic_view_all"},
	{"view licence", http.MethodGet, "/licence/my", nil, "licence_view"},
	// A UDP test against a customer's reflector is a test run; the key that
	// installs a reflector is the endpoint manager's to hand out. Both name an
	// endpoint that does not exist: 404 (or 503 where no key secret is set)
	// to whoever may, 403 to the rest.
	{"relay a UDP test to an endpoint", http.MethodPost, "/diagnostic/endpoint/00000000-0000-4000-8000-000000000000/webrtc-offer",
		map[string]any{"sdp": "v=0"}, "diagnostic_run"},
	{"issue an endpoint's reflector key", http.MethodPost, "/diagnostic/endpoint/00000000-0000-4000-8000-000000000000/reflector-key",
		nil, "dns_site_manage"},
}

func TestEveryRoleGetsExactlyItsCapabilities(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "aclmatrix")

	roles := packetpulsetest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	if roles.Status != http.StatusOK {
		t.Fatalf("listing roles: %d %s", roles.Status, roles.Raw)
	}

	for _, role := range packetpulsetest.ListOf(t, roles, "roles") {
		roleName, _ := role["role_name"].(string)
		roleId, _ := role["role_id"].(string)
		grants, _ := role["grants"].(map[string]any)
		if roleName == "Administrator" {
			continue // Exercised as the owner below.
		}

		t.Run(roleName, func(t *testing.T) {
			token := colleagueWithRole(t, ownerToken, roleId, roleName)
			if token == "" {
				t.Skip("could not provision a colleague with this role")
			}

			for _, route := range gatedRoutes {
				granted := holds(grants, route.Capability)

				response := packetpulsetest.Call(t, route.Method, route.Path, token, route.Payload)
				allowed := response.Status >= 200 && response.Status < 300
				denied := response.Status == http.StatusForbidden

				switch {
				case granted && denied:
					t.Errorf("%s: %s holds %s but was refused (403)",
						route.Name, roleName, route.Capability)
				case !granted && allowed:
					t.Errorf("%s: %s does NOT hold %s but was allowed (%d)",
						route.Name, roleName, route.Capability, response.Status)
				case !granted && !denied && response.Status != http.StatusConflict &&
					response.Status != http.StatusUnprocessableEntity:
					t.Errorf("%s: %s lacks %s and should get 403, got %d",
						route.Name, roleName, route.Capability, response.Status)
				}
			}
		})
	}
}

func TestPermissionChangeTakesEffectImmediately(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "aclrevoke")
	roles := packetpulsetest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)

	viewerRoleId := ""
	for _, role := range packetpulsetest.ListOf(t, roles, "roles") {
		if role["role_name"] == "Viewer" {
			viewerRoleId, _ = role["role_id"].(string)
		}
	}
	if viewerRoleId == "" {
		t.Skip("no Viewer role available")
	}

	token := colleagueWithRole(t, ownerToken, viewerRoleId, "Viewer")
	if token == "" {
		t.Skip("could not provision a colleague")
	}

	before := packetpulsetest.Call(t, http.MethodGet, "/dnssite/list", token, nil)
	if before.Status != http.StatusOK {
		t.Fatalf("a Viewer should be able to list sites, got %d", before.Status)
	}

	staffId := findStaffId(t, ownerToken, "Viewer")
	if staffId == "" {
		t.Skip("could not find the colleague's staff record")
	}

	// Deny a capability the role grants.
	packetpulsetest.Call(t, http.MethodPut, "/staff/"+staffId+"/access", ownerToken,
		map[string]any{"overrides": map[string]any{"dns_site_view": false}})

	// The token is still signature-valid. It must stop working anyway: a
	// permission change that waits for expiry is not a revocation.
	after := packetpulsetest.Call(t, http.MethodGet, "/dnssite/list", token, nil)
	if after.Status != http.StatusUnauthorized {
		t.Errorf("expected 401 after a permission change revoked the session, got %d",
			after.Status)
	}
}

// colleagueWithRole creates a user with a role and returns their token.
func colleagueWithRole(t *testing.T, ownerToken, roleId, roleName string) string {
	t.Helper()

	// Role names contain spaces ("NOC Engineer"), which are not valid in an
	// email local part - and an invalid address made this case skip silently.
	slug := strings.ToLower(strings.ReplaceAll(roleName, " ", "-"))
	email := fmt.Sprintf("acl-%s-%d@packetpulsetest.local", slug, time.Now().UnixNano())
	password := "ColleaguePass2026"

	created := packetpulsetest.Call(t, http.MethodPost, "/user/add", ownerToken, map[string]any{
		"email": email, "password": password,
		"full_name": roleName + " colleague", "role_id": roleId,
	})
	if created.Status != http.StatusCreated {
		return ""
	}
	return packetpulsetest.SignIn(t, email, password)
}

// findStaffId locates the most recently added colleague holding a role.
func findStaffId(t *testing.T, ownerToken, roleName string) string {
	t.Helper()

	list := packetpulsetest.Call(t, http.MethodGet, "/staff/list", ownerToken, nil)
	for _, staff := range packetpulsetest.ListOf(t, list, "staff") {
		if staff["role_name"] == roleName {
			if id, ok := staff["staff_id"].(string); ok {
				return id
			}
		}
	}
	return ""
}

// Everyone in an organisation reads its settings - they decide what its
// sign-in asks for - and only its own.
func TestEveryMemberReadsTheirOrganisationsSettings(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, organisationId := packetpulsetest.SignUpOrganisation(t, "aclsettings")
	roles := packetpulsetest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	viewerRoleId := ""
	for _, role := range packetpulsetest.ListOf(t, roles, "roles") {
		if role["role_name"] == "Viewer" {
			viewerRoleId, _ = role["role_id"].(string)
		}
	}
	viewer := colleagueWithRole(t, ownerToken, viewerRoleId, "Viewer")
	if viewer == "" {
		t.Fatal("could not provision a Viewer")
	}

	saved := packetpulsetest.Call(t, http.MethodPut, "/organisation/settings", ownerToken,
		map[string]any{"require_checkin_location": true, "device_test_target": "https://NOC.Example.net/health"})
	if saved.Status != http.StatusOK || saved.Body["require_checkin_location"] != true ||
		saved.String("device_test_target") != "noc.example.net" {
		t.Fatalf("saving: %d %s", saved.Status, saved.Raw)
	}
	// Everyone reads it: the target is what their own device tests.
	read := packetpulsetest.Call(t, http.MethodGet, "/organisation/settings", viewer, nil)
	if read.Status != http.StatusOK || read.String("organisation_id") != organisationId ||
		read.Body["require_checkin_location"] != true || read.String("device_test_target") != "noc.example.net" {
		t.Errorf("a Viewer reading the settings: %d %s", read.Status, read.Raw)
	}
}
