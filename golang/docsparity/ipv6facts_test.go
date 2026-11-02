package docsparity

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ── whether a test runs over IPv6 ───────────────────────────────────────

// ipv6TestsDefault reads the default of IPV6_TESTS_ENABLED from the server's
// configuration, so the guard follows the code: when the default is switched
// back on, the sentences it refuses become true again.
func ipv6TestsDefault(t *testing.T, root string) bool {
	t.Helper()
	config := read(t, filepath.Join(root, "packetpulsego/pkg/common/config/Config.go"))
	m := regexp.MustCompile(`getBool\("IPV6_TESTS_ENABLED",\s*(true|false)\)`).FindStringSubmatch(config)
	if m == nil {
		t.Fatal(`no getBool("IPV6_TESTS_ENABLED", …) in config/Config.go: the guard cannot tell whether tests run over IPv6`)
	}
	return m[1] == "true"
}

var (
	// A sentence that says a test runs over IPv6 as well as IPv4: "the test
	// runs over IPv4 and, where both ends have it, IPv6", "again over IPv6",
	// "runs IPv4 then IPv6", "the test runs over both".
	// A lookup over each family ("asks Cloudflare over IPv4 and over IPv6")
	// is not a test, so a run verb has to lead.
	ipv6Run = regexp.MustCompile(`(?i)\bagain over IPv6\b|\bruns? IPv4,? then IPv6\b|\b(?:runs?|tests?|tested)\b(?: \S+){0,4}? over IPv4(?: first)?,? (?:and|then)\b[^.]{0,120}?\bIPv6\b|\bthe test runs over both\b`)
	// A passage that says the IPv6 run happens only once it is switched on.
	ipv6Qualified = regexp.MustCompile(`(?i)\bswitch(?:es|ed)? (?:(?:IPv6 )?tests |it |them )?(?:back )?on\b|\bturned (?:back )?on\b|IPV6_TESTS_ENABLED|\bby default\b|\bset aside\b|\btests IPv6 as well\b`)
)

// sentenceAround is the sentence holding [start, end) and the one before it.
func sentenceAround(text string, start, end int) string {
	from := strings.LastIndex(text[:start], ". ")
	if from > 0 {
		if before := strings.LastIndex(text[:from], ". "); before >= 0 {
			from = before
		} else {
			from = 0
		}
	}
	if from < 0 {
		from = 0
	}
	to := strings.Index(text[end:], ". ")
	if to < 0 {
		to = len(text)
	} else {
		to += end
	}
	return text[from:to]
}

// staleIpv6Claims is every passage of a plain text that says a test runs over
// IPv6 without saying that happens only once IPv6 tests are switched on, and
// does not date itself as history.
func staleIpv6Claims(text string) []string {
	var claims []string
	for _, m := range ipv6Run.FindAllStringIndex(text, -1) {
		if ipv6Qualified.MatchString(sentenceAround(text, m[0], m[1])) || aboutThePast(text, m[0], m[1]) {
			continue
		}
		claims = append(claims, text[max(0, m[0]-60):min(len(text), m[1]+40)])
	}
	return claims
}

// ipv6Passages are matches that are true today.
var ipv6Passages = []readingPassage{
	{"packetpulsetest/manual-testing/parts/20-part-L-learn.md", "browser UDP test runs over IPv4 and then over IPv6 when the reflector offers it",
		`the paragraph opens "Tests run over IPv4 alone unless IPV6_TESTS_ENABLED is set ... With it set", four sentences earlier`},
}

// Tests have run over IPv4 alone since 28 October 2026, at the customer's
// request (IPV6_TESTS_ENABLED, off by default). Five days later the README
// and both guides in the reflector zip about to be shipped to Tata, and three
// documents, still said the UDP test runs again over IPv6 - telling the
// customer to set up and accept something their own minutes switched off. A
// passage that still says so fails here, unless it says the run happens once
// IPv6 tests are switched on, dates itself as history, or is listed above.
func TestNoTextSaysTestsRunOverIpv6WhileItIsSwitchedOff(t *testing.T) {
	root := repoRoot(t)
	if ipv6TestsDefault(t, root) {
		return // IPv6 tests are on by default: every such sentence is true
	}
	used := make([]bool, len(ipv6Passages))
	for _, path := range udpSources(t, root) {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		for _, claim := range staleIpv6Claims(plainText(string(source))) {
			listed := false
			for i, passage := range ipv6Passages {
				if passage.file == relative && strings.Contains(claim, passage.text) {
					used[i], listed = true, true
				}
			}
			if !listed {
				t.Errorf("%s: %q says tests run over IPv6, but IPV6_TESTS_ENABLED is off by default: tests run over IPv4 alone", relative, claim)
			}
		}
	}
	for i, passage := range ipv6Passages {
		if !used[i] {
			t.Errorf("ipv6Passages lists %s %q, which no longer matches: remove it", passage.file, passage.text)
		}
	}
}

// The guard is shown the sentences it was written to catch, and the right
// ones, so it cannot pass by matching nothing.
func TestTheIpv6GuardCatchesTheOldClaims(t *testing.T) {
	for _, old := range []string{
		"IPv6 (OPTIONAL) - the test runs over IPv4 and, where both ends have it, IPv6.",
		"PacketPulse tests each connection over IPv4 and, where the reflector and the client both have IPv6, again over IPv6, and reports each as its own path.",
		"If IPv6 is set up (2.4): the test runs over both.",
		"The app runs each UDP test over IPv4 and then, where the reflector offers an IPv6 address and the client has IPv6, over IPv6, and files each as its own result.",
		"IPv4 and IPv6: the UDP test runs over IPv4, then again over IPv6 when the reflector offers it; each run is filed.",
		"A browser UDP test runs IPv4 then IPv6; a server sweep runs both at once.",
		"The browser's UDP test runs over IPv4 first and then, when the reflector offers an IPv6 address and the run was not stopped, runs the whole test again over IPv6.",
	} {
		if len(staleIpv6Claims(plainText(old))) == 0 {
			t.Errorf("not caught: %q", old)
		}
	}
	for _, right := range []string{
		"Tests run over IPv4 alone by default (28 October 2026). With IPv6 tests switched on, a browser UDP test runs over IPv4, then over IPv6 when the reflector offers it, and both count.",
		"Tests run over IPv4 alone for now. Once PacketPulse switches IPv6 tests on, the test runs over IPv4 and, where both ends have it, again over IPv6.",
		"Until 28 October 2026 the test ran over IPv4 and then again over IPv6.",
		"The device asks Cloudflare over IPv4 and over IPv6 which address it saw.",
		"The app's ICMP test runs over IPv4 only.",
	} {
		if claims := staleIpv6Claims(plainText(right)); len(claims) != 0 {
			t.Errorf("flagged a right sentence %q: %v", right, claims)
		}
	}
}
