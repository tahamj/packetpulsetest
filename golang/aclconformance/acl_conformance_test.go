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

	pingletest "github.com/tahamj/pingletest"
)

type routeCheck struct {
	Name       string
	Method     string
	Path       string
	Payload    any
	Capability string
}

// gatedRoutes names the capability each route is supposed to require.
var gatedRoutes = []routeCheck{
	{"list sites", http.MethodGet, "/dnssite/list", nil, "dns_site_view"},
	{"add site", http.MethodPost, "/dnssite/add",
		map[string]any{"site_name": "acl probe", "ip_address": "203.0.113.200"}, "dns_site_manage"},
	{"list diagnostics", http.MethodGet, "/diagnostic/list", nil, "diagnostic_view_all"},
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
	{"list API keys", http.MethodGet, "/apikey/list", nil, "apikey_manage"},
	{"view licence", http.MethodGet, "/licence/my", nil, "licence_view"},
}

func TestEveryRoleGetsExactlyItsCapabilities(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "aclmatrix")

	roles := pingletest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	if roles.Status != http.StatusOK {
		t.Fatalf("listing roles: %d %s", roles.Status, roles.Raw)
	}

	for _, role := range pingletest.ListOf(t, roles, "roles") {
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
				granted, _ := grants[route.Capability].(bool)

				response := pingletest.Call(t, route.Method, route.Path, token, route.Payload)
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
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "aclrevoke")
	roles := pingletest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)

	viewerRoleId := ""
	for _, role := range pingletest.ListOf(t, roles, "roles") {
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

	before := pingletest.Call(t, http.MethodGet, "/dnssite/list", token, nil)
	if before.Status != http.StatusOK {
		t.Fatalf("a Viewer should be able to list sites, got %d", before.Status)
	}

	staffId := findStaffId(t, ownerToken, "Viewer")
	if staffId == "" {
		t.Skip("could not find the colleague's staff record")
	}

	// Deny a capability the role grants.
	pingletest.Call(t, http.MethodPut, "/staff/"+staffId+"/access", ownerToken,
		map[string]any{"overrides": map[string]any{"dns_site_view": false}})

	// The token is still signature-valid. It must stop working anyway: a
	// permission change that waits for expiry is not a revocation.
	after := pingletest.Call(t, http.MethodGet, "/dnssite/list", token, nil)
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
	email := fmt.Sprintf("acl-%s-%d@pingletest.local", slug, time.Now().UnixNano())
	password := "ColleaguePass2026"

	created := pingletest.Call(t, http.MethodPost, "/user/add", ownerToken, map[string]any{
		"email": email, "password": password,
		"full_name": roleName + " colleague", "role_id": roleId,
	})
	if created.Status != http.StatusCreated {
		return ""
	}
	return pingletest.SignIn(t, email, password)
}

// findStaffId locates the most recently added colleague holding a role.
func findStaffId(t *testing.T, ownerToken, roleName string) string {
	t.Helper()

	list := pingletest.Call(t, http.MethodGet, "/staff/list", ownerToken, nil)
	for _, staff := range pingletest.ListOf(t, list, "staff") {
		if staff["role_name"] == roleName {
			if id, ok := staff["staff_id"].(string); ok {
				return id
			}
		}
	}
	return ""
}
