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

	// The owner's own sign-in is written to the trail off the response path.
	// Counting before it lands would count it as one of the mutations below.
	waitForSignIn(t, ownerToken)
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
		"RouteSignIn":        "the password step: public, and it opens no session",
		"RouteSignInVerify":  "public; an administrator's completed sign-in is recorded by UserMS with the principal it creates",
		"RouteSignInResend":  "public, before any principal exists; it sends a code and changes nothing",
		"RouteSignInEnrol":   "provisional: the authenticator is confirmed, and the sign-in recorded, at verify",
		"RouteSignOut":       "recorded by UserMS itself, administrators' only; a field engineer's is their check-out",
		"RouteRunSweep":      "superseded by /diagnostic/submit, which is audited",
		"RouteConfigTest":    "a read-only connection test",
		"RouteSetLanguage":   "a personal display preference, not authority",
		"RouteSetAppearance": "a personal theme preference, not authority",
	}

	// The registry matches on a path SUFFIX, so collect the suffixes it
	// declares and ask whether any of them ends the route's path.
	registeredSuffixes := regexp.MustCompile(`PathSuffix:\s*"([^"]+)"`).
		FindAllStringSubmatch(registryText, -1)

	// \s* between the dot and the verb: a route is usually registered as
	//   private.With(guards...).
	//       Post(constant, handler)
	// and a pattern requiring ".Post(" on one line saw 22 of 60 routes.
	mutationCall := regexp.MustCompile(`\.\s*(Post|Put|Patch|Delete)\(([a-zA-Z0-9_.]+)`)

	var gaps []string
	for _, file := range routeFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		for _, match := range mutationCall.FindAllStringSubmatch(string(content), -1) {
			constant := match[2]
			shortName, qualifier := constant, ""
			if dot := strings.LastIndex(constant, "."); dot >= 0 {
				shortName, qualifier = constant[dot+1:], constant[:dot]
			}
			if _, skip := excluded[shortName]; skip {
				continue
			}
			// The registry matches on path SUFFIX, so compare against the
			// suffix the constant resolves to rather than its name.
			suffix := routeSuffix(t, repoRoot, qualifier, shortName)
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

// routeSuffix resolves a route constant to its path by scanning the constants
// package the route handler named. Route constants share names across
// modules - RouteUpdate, RouteList - so resolving by bare name found whichever
// module sorted first, and an unaudited route could pass as another module's
// registered one.
func routeSuffix(t *testing.T, repoRoot, qualifier, constantName string) string {
	t.Helper()

	packageGlob := "*constants"
	if qualifier != "" {
		packageGlob = qualifier
	}
	constantFiles, _ := filepath.Glob(filepath.Join(repoRoot, "pinglego/pkg/*", packageGlob, "*API.go"))
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(constantName) + `\s*=\s*"([^"]+)"`)

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

// waitForSignIn waits for the caller's own sign-in to reach the trail - and
// fails if it never does, because an administrator's sign-in that is not
// recorded is the defect this trail exists to rule out.
func waitForSignIn(t *testing.T, token string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		response := pingletest.Call(t, http.MethodGet, "/auditlog/list?limit=200", token, nil)
		for _, entry := range pingletest.ListOf(t, response, "entries") {
			if entry["event_type"] == "sign_in" && entry["entity_type"] == "session" {
				return
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("the owner's sign-in never reached the activity trail")
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

// An administrator's sign-in is in the trail, naming who, from where, and how
// the second step was given - and the chain still verifies with it in.
func TestAnAdministratorsSignInIsRecordedWithItsActor(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "auditsignin")
	waitForSignIn(t, ownerToken)

	me := pingletest.Call(t, http.MethodGet, "/user/me", ownerToken, nil)
	response := pingletest.Call(t, http.MethodGet, "/auditlog/list?limit=200", ownerToken, nil)
	var signIn map[string]any
	for _, entry := range pingletest.ListOf(t, response, "entries") {
		if entry["event_type"] == "sign_in" {
			signIn = entry
		}
	}
	if signIn["actor_email"] != me.String("email") || signIn["actor_ip"] == "" || signIn["entity_id"] == "" {
		t.Errorf("sign-in entry = %v, want the owner, their address and their session", signIn)
	}
	if details, _ := signIn["details"].(map[string]any); details["second_factor"] != "totp" {
		t.Errorf("details = %v, want the second step named", signIn["details"])
	}

	verify := pingletest.Call(t, http.MethodGet, "/auditlog/verify", ownerToken, nil)
	if verify.Status != http.StatusOK || verify.Body["intact"] != true {
		t.Errorf("the chain does not verify with the sign-in in it: %d %s", verify.Status, verify.Raw)
	}
}

// An administrator's sign-in shows where it happened and their sign-out where
// it ended, beside entries whose hash never held the position - so the chain
// verifies and the position can still be erased.
func TestAnAdministratorsSignInAndOutShowWhereTheyHappened(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "auditplace")
	email := pingletest.Call(t, http.MethodGet, "/user/me", ownerToken, nil).String("email")
	token := pingletest.SignInAt(t, email, "PingleTest2026x", map[string]any{
		"status": "captured", "latitude": 18.520430, "longitude": 73.856743, "accuracy_m": 14,
	})
	signOut := pingletest.Call(t, http.MethodPost, "/user/signout", token, map[string]any{
		"location": map[string]any{"status": "denied"},
	})
	if signOut.Status != http.StatusOK {
		t.Fatalf("signing out: %d %s", signOut.Status, signOut.Raw)
	}

	var signIn, signedOut map[string]any
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline) && signedOut == nil; time.Sleep(200 * time.Millisecond) {
		response := pingletest.Call(t, http.MethodGet, "/auditlog/list?entity_type=session&limit=200", ownerToken, nil)
		entries := pingletest.ListOf(t, response, "entries")
		for _, entry := range entries {
			if entry["entity_type"] != "session" {
				t.Fatalf("the session filter listed %v", entry["entity_type"])
			}
			if location, _ := entry["location"].(map[string]any); entry["event_type"] == "sign_in" && location["status"] == "captured" {
				signIn = entry
			}
		}
		for _, entry := range entries {
			if signIn != nil && entry["event_type"] == "sign_out" && entry["entity_id"] == signIn["entity_id"] {
				signedOut = entry
			}
		}
	}
	if signIn == nil || signedOut == nil {
		t.Fatalf("sign-in %v, sign-out %v: want both, naming one session", signIn, signedOut)
	}
	in := signIn["location"].(map[string]any)
	if in["latitude"] != 18.52043 || in["longitude"] != 73.856743 || in["accuracy_m"] != float64(14) {
		t.Errorf("sign-in location = %v", in)
	}
	if out, _ := signedOut["location"].(map[string]any); out["status"] != "denied" || out["latitude"] != nil {
		t.Errorf("sign-out location = %v, want the refusal", out)
	}
	if details, _ := signIn["details"].(map[string]any); details["latitude"] != nil || details["location"] != "captured" {
		t.Errorf("the hashed details carry %v; want whether a position was given, never the position", details)
	}

	verify := pingletest.Call(t, http.MethodGet, "/auditlog/verify", ownerToken, nil)
	if verify.Status != http.StatusOK || verify.Body["intact"] != true {
		t.Errorf("the chain does not verify: %d %s", verify.Status, verify.Raw)
	}
}

// A wrong password on an administrator's account reaches the trail as a
// failed sign-in - the account, why, and from where - and the chain still
// verifies with it in. Asked for by the customer: administrator logins
// tracked in full, which includes the ones that did not get in.
func TestAFailedAdministratorSignInIsRecorded(t *testing.T) {
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "auditfailed")
	waitForSignIn(t, ownerToken)
	me := pingletest.Call(t, http.MethodGet, "/user/me", ownerToken, nil)

	refused := pingletest.Call(t, http.MethodPost, "/user/signin", "", map[string]any{
		"email": me.String("email"), "password": "not-the-password",
	})
	if refused.Status != http.StatusUnauthorized {
		t.Fatalf("a wrong password answered %d", refused.Status)
	}

	var failed map[string]any
	deadline := time.Now().Add(10 * time.Second)
	for failed == nil && time.Now().Before(deadline) {
		response := pingletest.Call(t, http.MethodGet, "/auditlog/list?entity_type=session&limit=200", ownerToken, nil)
		for _, entry := range pingletest.ListOf(t, response, "entries") {
			if entry["event_type"] == "sign_in_failed" {
				failed = entry
			}
		}
		if failed == nil {
			time.Sleep(200 * time.Millisecond)
		}
	}
	if failed == nil {
		t.Fatal("the failed sign-in never reached the activity trail")
	}
	details, _ := failed["details"].(map[string]any)
	if failed["actor_email"] != me.String("email") || failed["actor_ip"] == "" || details["reason"] != "wrong_password" {
		t.Errorf("failed sign-in = %v, want the account, its address and the reason", failed)
	}
	verify := pingletest.Call(t, http.MethodGet, "/auditlog/verify", ownerToken, nil)
	if verify.Status != http.StatusOK || verify.Body["intact"] != true {
		t.Errorf("the chain does not verify with the failure in it: %d %s", verify.Status, verify.Raw)
	}
}
