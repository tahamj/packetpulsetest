// Package docsparity holds the manual testing guide to the code it describes.
//
// The guide (packetpulsetest/manual-testing/) states numbers, cites paths, lists
// every route, capability and migration, and promises things about itself:
// four lenses per chapter, a unique ID per test case. A document like that
// drifts the week after it is written - MShop's did, through three versions,
// until a suite like this one was added there. Everything here is a fact that
// can be re-derived from source, so it is checked rather than trusted.
//
// No server and no database: it reads files, and runs the guide's generator in
// --check mode.
package docsparity

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func guide(t *testing.T, root string) string {
	t.Helper()
	return read(t, filepath.Join(root, "packetpulsetest/manual-testing/MANUAL_TESTING_GUIDE.md"))
}

// section is the guide's text from one heading to the next of the same level.
func section(t *testing.T, text, heading string) string {
	t.Helper()
	start := strings.Index(text, heading)
	if start < 0 {
		t.Fatalf("the guide has no %q", heading)
	}
	level := strings.SplitN(heading, " ", 2)[0] + " "
	rest := text[start+len(heading):]
	if end := strings.Index(rest, "\n"+level); end >= 0 {
		rest = rest[:end]
	}
	return rest
}

// ── 1. the generated editions are current ───────────────────────────────

func TestTheGeneratedEditionsAreCurrent(t *testing.T) {
	root := repoRoot(t)
	python := filepath.Join(root, "packetpulsetest/.venv/bin/python3")
	if _, err := os.Stat(python); err != nil {
		t.Fatalf("the guide's generator needs packetpulsetest/.venv (markdown-it-py); ./packetpulsetest.sh docs makes it: %v", err)
	}
	command := exec.Command(python, "build/build_guide.py", "--check")
	command.Dir = filepath.Join(root, "packetpulsetest/manual-testing")
	if output, err := command.CombinedOutput(); err != nil {
		t.Errorf("the generated guide is stale or will not build:\n%s", output)
	}
}

// ── 2. every repository path the guide cites exists ─────────────────────

var citedPath = regexp.MustCompile("`((?:packetpulsego|packetpulseflutter|packetpulsetest|packetpulseweb|scripts|docs|\\.github)/[^`\\s]+)`")

func TestEveryCitedRepoPathExists(t *testing.T) {
	root := repoRoot(t)
	seen := map[string]bool{}
	for _, match := range citedPath.FindAllStringSubmatch(guide(t, root), -1) {
		path := strings.TrimRight(match[1], ".,;:")
		if seen[path] {
			continue
		}
		seen[path] = true
		check := path
		// "dir/**" cites a tree: the directory must exist.
		check = strings.TrimSuffix(check, "/**")
		if strings.ContainsAny(check, "*?[") {
			if found, _ := filepath.Glob(filepath.Join(root, check)); len(found) == 0 {
				t.Errorf("the guide cites %s, which matches nothing", path)
			}
			continue
		}
		if _, err := os.Stat(filepath.Join(root, check)); err != nil {
			t.Errorf("the guide cites %s, which does not exist", path)
		}
	}
	if len(seen) < 40 {
		t.Errorf("only %d cited paths found; the pattern has stopped matching the guide", len(seen))
	}
}

// ── 3. the headline numbers ─────────────────────────────────────────────

// The metric strip is the most-read part of the guide and the first to go
// stale, so every number in it is re-derived here.
func TestTheHeadlineNumbersMatchSource(t *testing.T) {
	root := repoRoot(t)
	text := guide(t, root)
	header := regexp.MustCompile(`\| 🧭 Screens \|[^\n]*\n\|[^\n]*\n(\|[^\n]*)`).FindStringSubmatch(text)
	if header == nil {
		t.Fatal("the guide has no metric strip")
	}
	var stated []int
	for _, match := range regexp.MustCompile(`\*\*(\d+)\*\*`).FindAllStringSubmatch(header[1], -1) {
		number, _ := strconv.Atoi(match[1])
		stated = append(stated, number)
	}

	shell := read(t, filepath.Join(root, "packetpulseflutter/lib/common/presentation/PacketPulseShell.dart"))
	screens := strings.Count(shell, "    _Destination(\n")
	capabilityCodes, _ := capabilities(t, root)
	countMatch := regexp.MustCompile(`static const int stringCount = (\d+);`).FindStringSubmatch(
		read(t, filepath.Join(root, "packetpulseflutter/lib/common/localization/PacketPulseStringsIndex.dart")))
	if countMatch == nil {
		t.Fatal("PacketPulseStringsIndex.dart declares no stringCount")
	}
	stringCount, _ := strconv.Atoi(countMatch[1])
	helpMatch := regexp.MustCompile(`HELP_SCREEN_COUNT = (\d+)`).FindStringSubmatch(
		read(t, filepath.Join(root, "packetpulsego/pkg/initmicroservice/initconstants/PacketPulseHelp.go")))
	help, _ := strconv.Atoi(helpMatch[1])
	all, _ := suites(t, root)

	derived := []struct {
		name  string
		value int
	}{
		{"screens", screens},
		{"API routes", len(sourceRoutes(t, root))},
		{"capabilities", len(capabilityCodes)},
		{"migrations", len(migrations(t, root))},
		{"catalogue strings", stringCount},
		{"help topics", help},
		{"guard suites", len(all)},
	}
	if len(stated) != len(derived) {
		t.Fatalf("the strip states %d numbers, want %d", len(stated), len(derived))
	}
	for index, fact := range derived {
		if stated[index] != fact.value {
			t.Errorf("the strip says %d %s; source has %d", stated[index], fact.name, fact.value)
		}
	}
}

// ── 4. the appendices are the source's lists ────────────────────────────

// A capability cell is one capability, or several joined by "or" (either)
// and "and" (both); "or" binds first.
var routeRow = regexp.MustCompile("(?m)^\\| `(GET|POST|PUT|PATCH|DELETE)` \\| `([^`]+)` \\| ([a-z ]+) \\| (`[a-z_]+`(?: (?:or|and) `[a-z_]+`)*|—) \\| (required|—) \\| (✅|—) \\|$")

var capabilityCell = strings.NewReplacer("`", "", " or ", "|", " and ", "+")

// capabilityCodes turns a requirement in capability names into one in codes.
func capabilityCodes(requirement string, codes map[string]string) string {
	return regexp.MustCompile(`[A-Za-z]+`).ReplaceAllStringFunc(requirement, func(name string) string {
		return codes[name]
	})
}

func TestAppendixAIsEveryRouteAsRegistered(t *testing.T) {
	root := repoRoot(t)
	written := map[string]route{}
	for _, match := range routeRow.FindAllStringSubmatch(section(t, guide(t, root), "## Appendix A"), -1) {
		entry := route{Method: match[1], Path: match[2], Access: match[3],
			Capability: capabilityCell.Replace(match[4]), Licensed: match[5] == "required", Audited: match[6] == "✅"}
		if entry.Capability == "—" {
			entry.Capability = ""
		}
		if _, twice := written[entry.key()]; twice {
			t.Errorf("Appendix A lists %s twice", entry.key())
		}
		written[entry.key()] = entry
	}
	codes, _ := capabilities(t, root)
	source := sourceRoutes(t, root)
	for _, actual := range source {
		if actual.Capability != "" {
			actual.Capability = capabilityCodes(actual.Capability, codes)
		}
		stated, ok := written[actual.key()]
		if !ok {
			t.Errorf("Appendix A is missing %s", actual.key())
			continue
		}
		if stated != actual {
			t.Errorf("Appendix A says %+v; source has %+v", stated, actual)
		}
		delete(written, actual.key())
	}
	for key := range written {
		t.Errorf("Appendix A lists %s, which is not a route", key)
	}
}

var capabilityRow = regexp.MustCompile("(?m)^\\| `([a-z_]+)` \\| ([^|]+) \\| (✅|—) \\| (✅|—) \\| (✅|—) \\|$")

func TestAppendixBIsEveryCapabilityAndBuiltInGrant(t *testing.T) {
	root := repoRoot(t)
	_, descriptions := capabilities(t, root)
	grants := builtInGrants(t, root)
	roles := []string{"Administrator", "NOC Engineer", "Viewer"}
	rows := capabilityRow.FindAllStringSubmatch(section(t, guide(t, root), "## Appendix B"), -1)
	if len(rows) != len(descriptions) {
		t.Errorf("Appendix B lists %d capabilities; source has %d", len(rows), len(descriptions))
	}
	for _, row := range rows {
		description, ok := descriptions[row[1]]
		if !ok {
			t.Errorf("Appendix B lists %s, which is not a capability", row[1])
			continue
		}
		if strings.TrimSpace(row[2]) != description {
			t.Errorf("Appendix B describes %s as %q; source says %q", row[1], strings.TrimSpace(row[2]), description)
		}
		for index, role := range roles {
			if (row[3+index] == "✅") != grants[role][row[1]] {
				t.Errorf("Appendix B says %s %v %s; the migrations disagree", role, row[3+index] == "✅", row[1])
			}
		}
	}
	// The counts quoted in the roles chapter: what each role holds. A grant a
	// later migration took away is in the map as false, and is not held.
	text := guide(t, root)
	for _, role := range roles {
		held := 0
		for _, granted := range grants[role] {
			if granted {
				held++
			}
		}
		want := fmt.Sprintf("%s %d", role, held)
		if !strings.Contains(text, want) {
			t.Errorf("the guide does not say %q", want)
		}
	}
}

func TestAppendixCIsEveryMigration(t *testing.T) {
	root := repoRoot(t)
	var written []string
	for _, match := range regexp.MustCompile("(?m)^\\| `(\\d{4}_[^`]+\\.sql)` \\|").FindAllStringSubmatch(section(t, guide(t, root), "## Appendix C"), -1) {
		written = append(written, match[1])
	}
	actual := migrations(t, root)
	if strings.Join(written, ",") != strings.Join(actual, ",") {
		t.Errorf("Appendix C lists\n  %v\nsource has\n  %v", written, actual)
	}
}

// ── 5. every named suite exists ─────────────────────────────────────────

func TestEveryNamedSuiteExists(t *testing.T) {
	root := repoRoot(t)
	all, optIn := suites(t, root)
	known := map[string]bool{}
	for _, name := range append(all, optIn...) {
		known[name] = true
	}
	for _, match := range regexp.MustCompile("\\./packetpulsetest\\.sh((?: [a-z]+)+)").FindAllStringSubmatch(guide(t, root), -1) {
		for _, name := range strings.Fields(match[1]) {
			if !known[name] {
				t.Errorf("the guide names suite %q, which packetpulsetest.sh does not have", name)
			}
		}
	}
	if !known["load"] || containsString(all, "load") {
		t.Error("load must be a suite, and opt-in: the guide says so")
	}
}

func containsString(list []string, value string) bool {
	for _, each := range list {
		if each == value {
			return true
		}
	}
	return false
}

// ── 6. the promises the guide makes about itself ─────────────────────────

var lenses = []string{
	"🌟 **Commercial Presentation & Sales Pitch**",
	"📖 **User Guide & Operational Flow**",
	"🧪 **Manual Testing Playbook**",
	"⚙️ **Developer Guide & Release Confidence**",
}

func TestEveryChapterCarriesAllFourLenses(t *testing.T) {
	root := repoRoot(t)
	part := section(t, guide(t, root), "# Part II")
	chapters := regexp.MustCompile(`(?m)^### (\d+\.\d+) `).FindAllStringSubmatchIndex(part, -1)
	if len(chapters) < 15 {
		t.Fatalf("only %d chapters found in Part II", len(chapters))
	}
	for index, chapter := range chapters {
		end := len(part)
		if index+1 < len(chapters) {
			end = chapters[index+1][0]
		}
		body := part[chapter[0]:end]
		name := part[chapter[2]:chapter[3]]
		for _, lens := range lenses {
			if !strings.Contains(body, lens) {
				t.Errorf("chapter %s is missing %s", name, lens)
			}
		}
	}
}

var (
	definedId    = regexp.MustCompile("(?m)^\\s*\\| `([A-Z0-9]+-\\d{3})` \\|")
	journeyId    = regexp.MustCompile(`(?m)^### (JRN-\d{3}) `)
	referencedId = regexp.MustCompile("`([A-Z0-9]+-\\d{3})`")
)

func TestTestIdsAreUniqueAndEveryReferenceIsDefined(t *testing.T) {
	text := guide(t, repoRoot(t))
	defined := map[string]int{}
	for _, match := range definedId.FindAllStringSubmatch(text, -1) {
		defined[match[1]]++
	}
	for _, match := range journeyId.FindAllStringSubmatch(text, -1) {
		defined[match[1]]++
	}
	var duplicates []string
	for id, count := range defined {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}
	sort.Strings(duplicates)
	if len(duplicates) > 0 {
		t.Errorf("test IDs used for more than one case: %v", duplicates)
	}
	for _, match := range referencedId.FindAllStringSubmatch(text, -1) {
		if defined[match[1]] == 0 {
			t.Errorf("the guide refers to %s, which no case defines", match[1])
		}
	}
	if len(defined) < 150 {
		t.Errorf("only %d test cases found; the pattern has stopped matching the guide", len(defined))
	}
}
