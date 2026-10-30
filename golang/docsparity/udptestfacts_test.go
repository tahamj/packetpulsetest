package docsparity

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// ── how the UDP test runs ───────────────────────────────────────────────

// limits is one reflector's caps on a single test, as reflectorcore sets them.
type limits struct {
	sessions, packets, ratePps, payloadBytes int
	durationSeconds, idleSeconds             int
}

// udpFacts is how the UDP test runs, read from the code that runs it.
type udpFacts struct {
	defaultPps, maxPps                     int
	minSeconds, maxSeconds, defaultSeconds int
	hosted, customer                       limits
}

// readUdpFacts reads the test's pace and lengths from reflectorconstants and
// each reflector's caps from reflectorcore, so a document is held to the code
// as it is, not as it was when the document was written.
func readUdpFacts(t *testing.T, root string) udpFacts {
	t.Helper()
	pkg := filepath.Join(root, "packetpulsego/pkg/reflectormicroservice")
	consts := intConsts(t, filepath.Join(pkg, "reflectorconstants/ReflectorAPI.go"))
	session := parseGo(t, filepath.Join(pkg, "reflectorcore/ReflectorSession.go"))
	return udpFacts{
		defaultPps:     need(t, consts, "TestDefaultRatePps"),
		maxPps:         need(t, consts, "TestMaxRatePps"),
		minSeconds:     need(t, consts, "TestMinSeconds"),
		maxSeconds:     need(t, consts, "TestMaxSeconds"),
		defaultSeconds: need(t, consts, "TestDefaultSeconds"),
		hosted:         limitsOf(t, session, "HostedLimits"),
		customer:       limitsOf(t, session, "CustomerLimits"),
	}
}

func parseGo(t *testing.T, path string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return file
}

func need(t *testing.T, values map[string]int, name string) int {
	t.Helper()
	value, ok := values[name]
	if !ok {
		t.Fatalf("no %s in the source", name)
	}
	return value
}

// intConsts is every integer constant a file declares.
func intConsts(t *testing.T, path string) map[string]int {
	t.Helper()
	values := map[string]int{}
	for _, decl := range parseGo(t, path).Decls {
		general, ok := decl.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST {
			continue
		}
		for _, spec := range general.Specs {
			value := spec.(*ast.ValueSpec)
			for i, name := range value.Names {
				if i < len(value.Values) {
					if n, ok := seconds(value.Values[i]); ok {
						values[name.Name] = n
					}
				}
			}
		}
	}
	return values
}

// seconds is an integer literal, or a literal times time.Second or
// time.Minute, in seconds.
func seconds(expression ast.Expr) (int, bool) {
	switch e := expression.(type) {
	case *ast.BasicLit:
		n, err := strconv.Atoi(e.Value)
		return n, err == nil && e.Kind == token.INT
	case *ast.BinaryExpr:
		n, ok := seconds(e.X)
		unit, isUnit := e.Y.(*ast.SelectorExpr)
		if !ok || !isUnit || e.Op != token.MUL {
			return 0, false
		}
		switch unit.Sel.Name {
		case "Second":
			return n, true
		case "Minute":
			return n * 60, true
		}
	}
	return 0, false
}

// limitsOf reads the Limits literal a function such as CustomerLimits returns.
func limitsOf(t *testing.T, file *ast.File, function string) limits {
	t.Helper()
	fields := map[string]int{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != function {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if pair, ok := node.(*ast.KeyValueExpr); ok {
				if n, ok := seconds(pair.Value); ok {
					fields[fmt.Sprint(pair.Key)] = n
				}
			}
			return true
		})
	}
	return limits{
		sessions:        need(t, fields, "MaxSessions"),
		packets:         need(t, fields, "MaxPackets"),
		ratePps:         need(t, fields, "MaxRatePps"),
		payloadBytes:    need(t, fields, "MaxPayloadBytes"),
		durationSeconds: need(t, fields, "MaxDuration"),
		idleSeconds:     need(t, fields, "IdleTimeout"),
	}
}

// udpSources is every text that tells a customer or a colleague how the test
// runs: the documents, the testing guide's parts, the in-app help, the
// reflector release's guides and README, and the features register's source.
// The deployment record is left out: it is a record of what each release was.
func udpSources(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	for _, path := range whoStartsSources(t, root) {
		if strings.HasSuffix(path, "packetpulsestrings.py") || strings.HasSuffix(path, "PacketPulseDeploymentRecord.md") {
			continue // the catalogue keeps retired strings by design; the record is history
		}
		files = append(files, path)
	}
	guides, err := filepath.Glob(filepath.Join(root, "packetpulserelease/reflector/docs/*.html"))
	if err != nil || len(guides) == 0 {
		t.Fatalf("no reflector guide templates: %v", err)
	}
	return append(append(files, guides...),
		filepath.Join(root, "packetpulserelease/reflector/BuildReflectorRelease.sh"),
		filepath.Join(root, "scripts/PacketPulseFutureFeaturesRegister.py"))
}

var (
	tag     = regexp.MustCompile(`<[^>]+>`)
	dataURI = regexp.MustCompile(`data:[a-z]+/[a-z0-9.+-]+;base64,[A-Za-z0-9+/=]+`)
	pace    = regexp.MustCompile(`\b(\d{2,3}) ?(?:UDP )?(?:(?:packets?|probes?) )?(?:a|per) second\b|\b(\d{2,3}) ?(?:/s|pkt/s|packets/s)\b`)
	// The sending windows before 28 October 2026: 75 s on PacketPulse's own
	// reflector, 150 s on a customer's. Today both are the code's MaxDuration.
	oldWindow = regexp.MustCompile(`\b(?:75|150)(?: |-)?(?:s|seconds?)\b`)
	// The lengths a test could run for before 28 October 2026.
	oldLengths = regexp.MustCompile(`\b10(?:, 30 or |–|-| to )60 ?(?:s|seconds)\b`)
	// The method the screen offered as a choice until it became the only one.
	oldControl = regexp.MustCompile(`(?i)\b(?:keep|choose|select)\s+(?:the\s+)?["“]?UDP, like a call`)
	// A passage that says it is about the past: a past-tense word and a date.
	pastWord = regexp.MustCompile(`(?i)\b(?:until|before|was|were|from|raised|when this was written)\b`)
	dated    = regexp.MustCompile(`\b\d{1,2} Oct(?:ober)? 2026\b`)
)

// plainText is a file's words without markup, entities or embedded images.
func plainText(source string) string {
	text := tag.ReplaceAllString(dataURI.ReplaceAllString(source, ""), " ")
	return strings.Join(strings.Fields(html.UnescapeString(text)), " ")
}

// aboutThePast is true when the sentence holding [start, end), or the one
// before it, dates what it says: "until 28 October 2026 it was 150 s".
func aboutThePast(text string, start, end int) bool {
	from := strings.LastIndex(text[:start], ". ")
	if from > 0 {
		if before := strings.LastIndex(text[:from], ". "); before >= 0 {
			from = before
		} else {
			from = 0
		}
	}
	to := strings.Index(text[end:], ". ")
	if from < 0 {
		from = 0
	}
	if to < 0 {
		to = len(text)
	} else {
		to += end
	}
	passage := text[from:to]
	return pastWord.MatchString(passage) && dated.MatchString(passage)
}

// staleUdpClaims is every passage of a plain text that says the test runs in
// a way the code no longer runs it, and does not date itself as history.
func staleUdpClaims(text string, facts udpFacts) []string {
	allowed := map[int]bool{facts.defaultPps: true, facts.maxPps: true,
		facts.hosted.ratePps: true, facts.customer.ratePps: true}
	var claims []string
	claim := func(start, end int) {
		if !aboutThePast(text, start, end) {
			claims = append(claims, text[max(0, start-60):min(len(text), end+40)])
		}
	}
	for _, m := range pace.FindAllStringSubmatchIndex(text, -1) {
		if m[0] > 0 && strings.ContainsAny(text[m[0]-1:m[0]], ",.") {
			continue // "7,000 packets a second" is not a pace of 000
		}
		digits := m[2:4]
		if digits[0] < 0 {
			digits = m[4:6]
		}
		if n, _ := strconv.Atoi(text[digits[0]:digits[1]]); !allowed[n] {
			claim(m[0], m[1])
		}
	}
	for _, rx := range []*regexp.Regexp{oldWindow, oldLengths, oldControl} {
		for _, m := range rx.FindAllStringIndex(text, -1) {
			claim(m[0], m[1])
		}
	}
	return claims
}

// udpPassages are matches that are true today: a fact about something other
// than PacketPulse's test, or a dated record that does not say so in-sentence.
var udpPassages = []readingPassage{
	{"docs/PacketPulseTestStrategy.md", "150-second request timeout",
		"the server's HTTP request timeout, not a test's length"},
	{"docs/PacketPulseUdpFieldGuide.md", "| 1,500 | 150 s |",
		"how long 1,500 packets take at 10 a second, a length a test may run for"},
	{"docs/PacketPulse_Results_API_Samples.html", "150-second test at 10 packets a second",
		"a real test's length, inside 30 s to 5 min"},
	{"docs/PacketPulse_Network_Reports_Explained.html", "A voice call sends about 50 a second",
		"how a voice call is carried, not PacketPulse's pace"},
	{"docs/PacketPulse_Keep_Test_Running_Solution.html", "side by side 50 packets a second",
		"a measurement of 21 October 2026, at the rate it was run at"},
	{"docs/PacketPulse_Keep_Test_Running_Solution.html", "At 100 a second a few slots go unsent",
		"a measurement of 21 October 2026, at the rate it was run at"},
	{"docs/PacketPulse_Keep_Test_Running_Solution.html", "at 100 a second it left 8% of slots",
		"a measurement of 21 October 2026, at the rate it was run at"},
	{"docs/PacketPulse_Keep_Test_Running_Solution.html", "test at 10 and at 50 packets a second",
		"the checks behind a measurement of 21 October 2026"},
	{"docs/PacketPulse_Future_Features_Register.html", "voice (50 a second",
		"the call profiles a future feature would imitate"},
	{"docs/PacketPulse_Future_Features_Register.html", "gaming (60 a second",
		"the call profiles a future feature would imitate"},
	{"docs/PacketPulse_Future_Features_Register.html", "Voice call 50 a second",
		"the call profiles a future feature would imitate"},
	{"docs/PacketPulse_Future_Features_Register.html", "Online game 60 a second",
		"the call profiles a future feature would imitate"},
	{"scripts/PacketPulseFutureFeaturesRegister.py", "voice (50 a second",
		"the call profiles a future feature would imitate"},
	{"scripts/PacketPulseFutureFeaturesRegister.py", "gaming (60 a second",
		"the call profiles a future feature would imitate"},
	{"scripts/PacketPulseFutureFeaturesRegister.py", "# gaming: 60 a second",
		"the call profiles a future feature would imitate"},
	{"scripts/PacketPulseFutureFeaturesRegister.py", "Voice call 50 a second",
		"the call profiles a future feature would imitate"},
	{"scripts/PacketPulseFutureFeaturesRegister.py", "Online game 60 a second",
		"the call profiles a future feature would imitate"},
	{"packetpulsetest/manual-testing/parts/00-front-matter.md", "for 10, 30 or 60 seconds at 10, 25, 50 or 100 packets a second",
		"the v2 change table of 6 October 2026, superseded by the 28 October table above it"},
}

// The test's pace, lengths and limits changed on 28-30 October 2026 (the
// customer's request: 30 s to 5 min at the server's pace of 10 a second, in a
// 6-minute window). Weeks later a dozen documents, and the reflector guide
// shipped to Tata, still said 50 a second, 10, 30 or 60 s, a 150 s window, and
// told engineers to keep a method and a rate the screen no longer offered.
// A passage that still says so fails here, unless it dates itself as history
// or is listed above; a listed passage that is gone fails too.
func TestNoTextDescribesTheUdpTestAsItNoLongerRuns(t *testing.T) {
	root := repoRoot(t)
	facts := readUdpFacts(t, root)
	used := make([]bool, len(udpPassages))
	for _, path := range udpSources(t, root) {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		for _, claim := range staleUdpClaims(plainText(string(source)), facts) {
			listed := false
			for i, passage := range udpPassages {
				if passage.file == relative && strings.Contains(claim, passage.text) {
					used[i], listed = true, true
				}
			}
			if !listed {
				t.Errorf("%s: %q describes the UDP test as it no longer runs: %d to %d s at %d a second, in a %d s window",
					relative, claim, facts.minSeconds, facts.maxSeconds, facts.defaultPps, facts.customer.durationSeconds)
			}
		}
	}
	for i, passage := range udpPassages {
		if !used[i] {
			t.Errorf("udpPassages lists %s %q, which no longer matches: remove it", passage.file, passage.text)
		}
	}
}

// thousands writes 6000 as "6,000", as the documents do.
func thousands(n int) string {
	digits := strconv.Itoa(n)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "," + digits[i:]
	}
	return digits
}

// duration writes 360 as "6 min" and 30 as "30 s".
func duration(seconds int) string {
	if seconds%60 == 0 {
		return fmt.Sprintf("%d min", seconds/60)
	}
	return fmt.Sprintf("%d s", seconds)
}

// The reference guide's summary of a customer reflector's caps is the one
// PacketPulse's engineers quote; it is CustomerLimits, written out.
func TestTheReflectorGuideQuotesTheCustomerLimits(t *testing.T) {
	root := repoRoot(t)
	c := readUdpFacts(t, root).customer
	want := fmt.Sprintf("Per test: %d packets/s, %s packets, %s, %s bytes a packet, %s idle. At most %d tests at once.",
		c.ratePps, thousands(c.packets), duration(c.durationSeconds), thousands(c.payloadBytes), duration(c.idleSeconds), c.sessions)
	guide := plainText(read(t, filepath.Join(root, "packetpulserelease/reflector/docs/PacketPulse_Reflector_Installation_Guide.html")))
	if !strings.Contains(guide, want) {
		t.Errorf("the reflector reference guide does not say %q", want)
	}
}

// The guard is shown the sentences it was written to catch, and the right
// ones, so it cannot pass by matching nothing.
func TestTheUdpFactsGuardCatchesTheOldClaims(t *testing.T) {
	facts := readUdpFacts(t, repoRoot(t))
	for _, old := range []string{
		"a stream of UDP packets, 50 a second by default, sent from a browser.",
		"Keep UDP, like a call, a Test length of 30 s and 50 Packets per second, and press Start test.",
		`c. Keep "UDP, like a call", Test length 30 s, and press Start test.`,
		"One test sends at most 200 packets a second, for at most 150 seconds and 6,000 packets.",
		"In a browser the test is UDP, like a call: 10, 30 or 60 seconds of call-like packets.",
		"Session length 75 s 150 s",
		"TT-2026-00731 · UDP like a call · 30 s at 50/s · UTC",
	} {
		if len(staleUdpClaims(plainText(old), facts)) == 0 {
			t.Errorf("not caught: %q", old)
		}
	}
	for _, right := range []string{
		"A test runs for 30 seconds to 5 minutes, at 10 packets a second, at most 200 packets a second.",
		"Flood protection below about 7,000 packets a second to this host.",
		"The windows were 75 s and 150 s until 28 October 2026, so an old reflector stops a test early.",
		"Choose a Test length of 30 s and press Start test.",
	} {
		if claims := staleUdpClaims(plainText(right), facts); len(claims) != 0 {
			t.Errorf("flagged a right sentence %q: %v", right, claims)
		}
	}
}
