package docsparity

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ── who starts a test ───────────────────────────────────────────────────

// startedBy ties something outside PacketPulse - a scheduler, a ticketing
// system, an integration - to starting a test, within a few words of it.
var (
	outsider   = `\b(?:schedulers?|ticketing systems?|own systems?|your systems?|IT systems?|integrations?)\b`
	startsTest = `\b(?:sweeps?|start\w*|trigger\w*|asks? for one|request\w*|call\w*)\b`
	startedBy  = regexp.MustCompile(`(?i)` + outsider + `.{0,50}?` + startsTest + `|` + startsTest + `.{0,50}?` + outsider)
	// readsOnly is wording that already says the outsider only reads.
	readsOnly = regexp.MustCompile(`(?i)\b(?:cannot|can't|never|no API key|only read|pull\w*|fetch\w*|reads?)\b`)
	markup    = regexp.MustCompile(`<[^>]+>|&[a-z]+;`)
)

// readingPassage is a match that is about reading results, not starting a test.
type readingPassage struct{ file, text, why string }

var readingPassages = []readingPassage{
	{"docs/PacketPulse_Future_Features_Estimates.html", "integration team can call",
		"FF-16: an integration team calling the Results API, which only reads"},
	{"docs/PacketPulse_Product_Briefing_and_Feature_Guide.html", "The ticketing system calls",
		"UC-6: the ticketing system calls /result/bytt, which only reads"},
	{"docs/PacketPulse_Product_Briefing_and_Feature_Guide.html", "request). Integration",
		"a retention set on request, then the next row's integration secrets"},
	{"docs/PacketPulse_Product_Briefing_and_Feature_Guide.html", "started by a signed-in person; your ticketing system",
		"a person starts tests; the ticketing system pulls results"},
	{"docs/PacketPulse_Product_Briefing_and_Feature_Guide.html", "your systems by API or SFTP; tests are started",
		"results flow out to the customer's systems; a signed-in person starts tests"},
}

// whoStartsSources is every text a customer or a colleague reads about the
// product: the documents, the testing guide's parts, the in-app help and the
// string catalogue.
func whoStartsSources(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	for _, pattern := range []string{"docs/*.html", "docs/*.md", "packetpulsetest/manual-testing/parts/*.md"} {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil || len(matches) == 0 {
			t.Fatalf("no files match %s: %v", pattern, err)
		}
		files = append(files, matches...)
	}
	return append(files,
		filepath.Join(root, "packetpulsego/pkg/initmicroservice/initconstants/PacketPulseHelp.go"),
		filepath.Join(root, "scripts/intellicodegen/packetpulsestrings.py"))
}

// startClaimsIn is every passage of a text that says something outside
// PacketPulse starts a test.
func startClaimsIn(text string) []string {
	plain := strings.Join(strings.Fields(markup.ReplaceAllString(text, " ")), " ")
	var claims []string
	for _, passage := range startedBy.FindAllString(plain, -1) {
		if !readsOnly.MatchString(passage) {
			claims = append(claims, passage)
		}
	}
	return claims
}

// Only a signed-in person starts a test: API keys read results and nothing
// else (confirmed 29 October 2026, and held by apicontract's
// TestAnApiKeyCannotStartATest). Until then the in-app help, six documents and
// the features register told customers that their scheduler or ticketing
// system could start a server sweep through the API. A passage tying an
// outside system to starting a test fails here, unless it is listed above as
// being about reading; a listed passage that is gone fails too.
func TestNoTextSaysAnOutsideSystemStartsATest(t *testing.T) {
	root := repoRoot(t)
	used := make([]bool, len(readingPassages))
	for _, path := range whoStartsSources(t, root) {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		for _, claim := range startClaimsIn(string(source)) {
			listed := false
			for i, reading := range readingPassages {
				if reading.file == relative && strings.Contains(claim, reading.text) {
					used[i], listed = true, true
				}
			}
			if !listed {
				t.Errorf("%s: %q reads as an outside system starting a test; only a signed-in person starts one, "+
					"and API keys only read results", relative, claim)
			}
		}
	}
	for i, reading := range readingPassages {
		if !used[i] {
			t.Errorf("readingPassages lists %s %q, which no longer matches: remove it", reading.file, reading.text)
		}
	}
}

// The guard is shown the sentences it was written to catch, and the right
// ones, so it cannot pass by matching nothing.
func TestTheStartWordingGuardCatchesTheOldClaims(t *testing.T) {
	for _, claim := range []string{
		"<p>PacketPulse's server sweeps your endpoints only when your own systems ask it to through the API.</p>",
		"diagnostics run against tickets, and your own scheduler can call the sweep API.",
		"A customer's ticketing system or scheduler can ask for one through the API.",
		"Only scripts and integrations can start a server sweep today.",
		"<h4>Server sweeps when your systems ask</h4>",
	} {
		if len(startClaimsIn(claim)) == 0 {
			t.Errorf("not caught: %q", claim)
		}
	}
	for _, right := range []string{
		"Outside systems connect with read-only API keys: they can fetch results but cannot start a test.",
		"A key lets your ticketing system pull results into the ticket.",
		"A signed-in member can still start one through the API; no API key can.",
	} {
		if claims := startClaimsIn(right); len(claims) != 0 {
			t.Errorf("flagged a right sentence %q: %v", right, claims)
		}
	}
}
