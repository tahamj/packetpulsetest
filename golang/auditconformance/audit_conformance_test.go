// Package auditconformance guards the activity trail.
//
// It checks two things a log file cannot promise: that FAILED attempts leave no
// entry, and that the chain is tamper-evident. The second matters because an
// audit trail nobody can verify is a claim, not evidence.
package auditconformance

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	pingletest "github.com/tahamj/pingletest"
)

func TestOnlySuccessfulMutationsAreRecorded(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "audit")

	before := countEntries(t, ownerToken)

	// One that succeeds.
	ok := pingletest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"site_name": "audited site", "ip_address": "203.0.113.90",
	})
	if ok.Status != http.StatusCreated {
		t.Fatalf("the successful mutation did not succeed: %d %s", ok.Status, ok.Raw)
	}

	// Three that do not: invalid, duplicate, and unauthorised.
	pingletest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"site_name": "invalid", "ip_address": "not-an-address!!",
	})
	pingletest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"site_name": "duplicate", "ip_address": "203.0.113.90",
	})
	pingletest.Call(t, http.MethodPost, "/dnssite/add", "", map[string]any{
		"site_name": "anonymous", "ip_address": "203.0.113.91",
	})

	time.Sleep(2 * time.Second) // The trail is written off the response path.

	after := countEntries(t, ownerToken)
	if added := after - before; added != 1 {
		t.Errorf("expected exactly 1 audit entry for 1 successful and 3 failed "+
			"mutations, got %d", added)
	}
}

func TestAuditChainVerifies(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "auditchain")

	for i := 0; i < 3; i++ {
		pingletest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
			"site_name":  fmt.Sprintf("chain %d", i),
			"ip_address": fmt.Sprintf("203.0.113.%d", 120+i),
		})
	}
	time.Sleep(2 * time.Second)

	response := pingletest.Call(t, http.MethodGet, "/auditlog/verify", ownerToken, nil)
	if response.Status != http.StatusOK {
		t.Fatalf("verifying the trail: %d %s", response.Status, response.Raw)
	}

	intact, _ := response.Body["intact"].(bool)
	if !intact {
		t.Errorf("the audit chain is broken at entry %v: %v",
			response.Body["first_bad_activity_id"], response.Body["first_bad_note"])
	}
	if response.Float("entries_checked") == 0 {
		t.Error("the verifier checked no entries")
	}
}

func TestAuditEntriesNameTheActorAndOrigin(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "auditactor")

	pingletest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"site_name": "actor probe", "ip_address": "203.0.113.150",
	})
	time.Sleep(2 * time.Second)

	list := pingletest.Call(t, http.MethodGet, "/auditlog/list?limit=1", ownerToken, nil)
	entries := pingletest.ListOf(t, list, "entries")
	if len(entries) == 0 {
		t.Fatal("no audit entries were written")
	}

	entry := entries[0]
	for _, field := range []string{"actor_email", "actor_ip", "entity_type", "event_type", "entry_hash"} {
		if value, _ := entry[field].(string); value == "" {
			t.Errorf("audit entry has no %s: an entry nobody can attribute is not evidence", field)
		}
	}
}

// TestEveryMutationRouteIsRegistered is a static check: it reads the route
// handlers and the audit registry and reports mutations that are not audited.
//
// This is the check that stops the trail quietly falling behind the code. A new
// mutation route is easy to add and easy to forget; this makes forgetting
// visible in CI rather than in an audit.
func TestEveryMutationRouteIsRegistered(t *testing.T) {
	repoRoot := findRepoRoot(t)
	registryPath := filepath.Join(repoRoot,
		"pinglego/pkg/auditlogmicroservice/auditlogconstants/AuditLogRegistry.go")

	registry, err := os.ReadFile(registryPath)
	if err != nil {
		t.Skipf("audit registry not found: %v", err)
	}
	registryText := string(registry)

	routeFiles, err := filepath.Glob(filepath.Join(repoRoot, "pinglego/pkg/*/*/*RouteHandler.go"))
	if err != nil || len(routeFiles) == 0 {
		t.Skip("no route handlers found")
	}

	// Routes that mutate but are deliberately not audited, each with a reason.
	excluded := map[string]string{
		"RouteSignUp":        "creates an account in the holding organisation, which has no trail of its own",
		"RouteSignIn":        "recorded with its outcome by the auth path",
		"RouteRunSweep":      "superseded by /diagnostic/submit, which is audited",
		"RouteConfigTest":    "a read-only connection test",
		"RouteSetLanguage":   "a personal display preference, not authority",
		"RouteSetAppearance": "a personal theme preference, not authority",
	}

	// The registry matches on a path SUFFIX, so collect the suffixes it
	// declares and ask whether any of them ends the route's path.
	registeredSuffixes := regexp.MustCompile(`PathSuffix:\s*"([^"]+)"`).
		FindAllStringSubmatch(registryText, -1)

	mutationCall := regexp.MustCompile(`\.(Post|Put|Patch|Delete)\(([a-zA-Z0-9_.]+)`)

	var gaps []string
	for _, file := range routeFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		for _, match := range mutationCall.FindAllStringSubmatch(string(content), -1) {
			constant := match[2]
			shortName := constant
			if dot := strings.LastIndex(constant, "."); dot >= 0 {
				shortName = constant[dot+1:]
			}
			if _, skip := excluded[shortName]; skip {
				continue
			}
			// The registry matches on path SUFFIX, so compare against the
			// suffix the constant resolves to rather than its name.
			suffix := routeSuffix(t, repoRoot, shortName)
			if suffix == "" || isRegistered(suffix, registeredSuffixes) {
				continue
			}
			gaps = append(gaps, fmt.Sprintf("%s (%s) in %s",
				shortName, suffix, filepath.Base(file)))
		}
	}

	if len(gaps) > 0 {
		t.Errorf("these mutation routes are not in the audit registry:\n  %s\n"+
			"Add them to AuditLogRegistry.go, or to this test's exclusion list with a reason.",
			strings.Join(gaps, "\n  "))
	}
}

// isRegistered reports whether any registered suffix ends this route's path.
func isRegistered(routePath string, registered [][]string) bool {
	for _, entry := range registered {
		if strings.HasSuffix(routePath, entry[1]) {
			return true
		}
	}
	return false
}

// routeSuffix resolves a route constant to its path by scanning the constants.
func routeSuffix(t *testing.T, repoRoot, constantName string) string {
	t.Helper()

	constantFiles, _ := filepath.Glob(filepath.Join(repoRoot, "pinglego/pkg/*/*constants/*API.go"))
	pattern := regexp.MustCompile(regexp.QuoteMeta(constantName) + `\s*=\s*"([^"]+)"`)

	for _, file := range constantFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if match := pattern.FindStringSubmatch(string(content)); match != nil {
			// Registry entries name the full route pattern, so it is returned
			// unchanged - parameters included.
			return match[1]
		}
	}
	return ""
}

func countEntries(t *testing.T, token string) int {
	t.Helper()

	response := pingletest.Call(t, http.MethodGet, "/auditlog/list?limit=200", token, nil)
	if response.Status != http.StatusOK {
		return 0
	}
	return int(response.Float("count"))
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	directory, err := os.Getwd()
	if err != nil {
		t.Skip("cannot determine the working directory")
	}
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(directory, "pinglego", "go.mod")); err == nil {
			return directory
		}
		directory = filepath.Dir(directory)
	}
	t.Skip("cannot locate the repository root")
	return ""
}
