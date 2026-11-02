package docsparity

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// ── who chooses the endpoint ────────────────────────────────────────────

// handPicked is a passage saying that an engineer, or everyone, picks the
// endpoint of a test themselves.
var handPicked = regexp.MustCompile(`(?i)` + strings.Join([]string{
	`\beveryone\b.{0,40}?\b(?:picks?|chooses?|choosing|picking)\b`,
	`\bengineers?\b.{0,40}?\b(?:picks?|chooses?)\b.{0,40}?\b(?:themselves|by hand|yourself)\b`,
	`\byou (?:choose|pick) (?:it|the endpoint) yourself\b`,
	`\bengineers? (?:may|can) (?:pick|choose) any\b`,
}, "|"))

// choicePassage is a match that is right where it stands.
type choicePassage struct{ file, text, why string }

var choicePassages = []choicePassage{
	{"docs/PacketPulseDeploymentRecord.md", "everyone picks one by hand",
		"the record of the 17 October 2026 release, as it shipped"},
	{"docs/PacketPulseDeploymentRecord.md", "everyone picks and is",
		"the record of the 17 October 2026 release, as it shipped"},
	{"docs/PacketPulseDeploymentRecord.md", "everyone chooses, and the page says why",
		"the record of a release before 2 November 2026, as it shipped"},
}

// claimsOfHandPicking is every passage of a text that says an engineer, or
// everyone, picks the endpoint: the words that say it, with up to forty
// bytes either side, so a listed passage can be named exactly.
func claimsOfHandPicking(text string) []string {
	plain := strings.Join(strings.Fields(markup.ReplaceAllString(text, " ")), " ")
	var claims []string
	for _, at := range handPicked.FindAllStringIndex(plain, -1) {
		start, end := max(at[0]-40, 0), min(at[1]+40, len(plain))
		// Whole characters: a Devanagari or a typographic one is several bytes.
		for start > 0 && !utf8.RuneStart(plain[start]) {
			start--
		}
		for end < len(plain) && !utf8.RuneStart(plain[end]) {
			end++
		}
		claims = append(claims, plain[start:end])
	}
	return claims
}

// The endpoint of an engineer's test is chosen by where the device is, and
// an engineer cannot change it (the SOW, and gap G6 closed on 2 November
// 2026). Until then, a device that would not say where it was let anyone
// pick, and the in-app help, the testing guide and three customer documents
// said so: "everyone picks by hand". Now whoever manages the endpoints
// picks, and an engineer gets no test until there is a position. A passage
// saying otherwise fails here, unless it is listed above with why it is
// right; a listed passage that is gone fails too.
func TestNoTextSaysAnEngineerPicksTheEndpoint(t *testing.T) {
	root := repoRoot(t)
	used := make([]bool, len(choicePassages))
	for _, path := range whoStartsSources(t, root) {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		for _, claim := range claimsOfHandPicking(string(source)) {
			listed := false
			for i, passage := range choicePassages {
				if passage.file == relative && strings.Contains(claim, passage.text) {
					used[i], listed = true, true
				}
			}
			if !listed {
				t.Errorf("%s: %q says an engineer picks the endpoint; it is chosen by where the device is, "+
					"and without a position an engineer gets no test", relative, claim)
			}
		}
	}
	for i, passage := range choicePassages {
		if !used[i] {
			t.Errorf("choicePassages lists %s %q, which no longer matches: remove it", passage.file, passage.text)
		}
	}
}

// The guard is shown the sentences it was written to catch, and the right
// ones, so it cannot pass by matching nothing.
func TestTheEndpointChoiceGuardCatchesTheOldClaims(t *testing.T) {
	for _, claim := range []string{
		"With no position to go by, everyone picks by hand, and the page says why.",
		"nothing can be chosen for you, so everyone picks the endpoint by hand",
		"until every enabled endpoint has a position, engineers choose the endpoint themselves.",
		"or when this device cannot say where it is, you choose it yourself.",
		"until then everyone chooses, and the page says why.",
		"The engineer may pick any one, and that is what is probed.",
	} {
		if len(claimsOfHandPicking(claim)) == 0 {
			t.Errorf("not caught: %q", claim)
		}
	}
	for _, right := range []string{
		"With no position to go by, whoever manages the endpoints picks by hand, and the page says why.",
		"An engineer does not: no test starts without a position.",
		"An engineer cannot change it; whoever manages the endpoints may choose another.",
		"Whoever manages the endpoints still picks by hand.",
		"Sets what everyone in the organisation tests.",
	} {
		if claims := claimsOfHandPicking(right); len(claims) != 0 {
			t.Errorf("flagged a right sentence %q: %v", right, claims)
		}
	}
}
