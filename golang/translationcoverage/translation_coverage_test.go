// Package translationcoverage verifies the string catalogue the client draws
// every label from.
//
// PacketPulse is English only (customer requirement, October 2026). The suite kept
// its name so CI and packetpulsetest.sh need no change, but what it guards is now:
//
//   - The Go and Dart index constants agree, name for name and position for
//     position. The client looks strings up BY POSITION, so a constant that
//     resolves to a different index on one side puts a label on the wrong
//     control, and nothing else notices.
//   - No other language's tables have been put back. A table nobody keeps
//     translated degrades with every string added.
//
// When the PacketPulse server is running it also verifies the wire: one language
// offered, and a complete English catalogue served for every language id -
// including the ids accounts chose before PacketPulse went English-only.
package translationcoverage

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("working directory: %v", err)
	}
	for range 10 {
		if isDir(filepath.Join(dir, "packetpulsego")) && isDir(filepath.Join(dir, "packetpulseflutter")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("repo root not found")
	return ""
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// indexFrom reads `name = position` pairs out of a generated index file.
func indexFrom(t *testing.T, path string, pattern *regexp.Regexp) map[string]int {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	index := map[string]int{}
	for _, match := range pattern.FindAllStringSubmatch(string(source), -1) {
		position, _ := strconv.Atoi(match[2])
		if _, seen := index[match[1]]; seen {
			t.Errorf("%s: %s is declared twice", filepath.Base(path), match[1])
		}
		index[match[1]] = position
	}
	return index
}

var (
	goIndexPattern   = regexp.MustCompile(`(?m)^\s*STR_([A-Z0-9_]+)_INDEX\s*=\s*(\d+)\s*$`)
	dartIndexPattern = regexp.MustCompile(`(?m)^\s*static const int ([a-z0-9_]+) = (\d+);\s*$`)
	goCountPattern   = regexp.MustCompile(`const StringCount = (\d+)`)
	stringIdPattern  = regexp.MustCompile(`\{StringId: (\d+), StringValue:`)
)

func TestGoAndDartCataloguesAgreePositionally(t *testing.T) {
	root := repoRoot(t)
	constants := filepath.Join(root, "packetpulsego", "pkg", "initmicroservice", "initconstants")

	goIndex := indexFrom(t, filepath.Join(constants, "PacketPulseStringsIndex.go"), goIndexPattern)
	dartIndex := indexFrom(t, filepath.Join(root, "packetpulseflutter", "lib", "common", "localization", "PacketPulseStringsIndex.dart"), dartIndexPattern)
	delete(dartIndex, "stringcount") // the count, not a string

	countSource, err := os.ReadFile(filepath.Join(constants, "PacketPulseStringsIndex.go"))
	if err != nil {
		t.Fatal(err)
	}
	count := goCountPattern.FindStringSubmatch(string(countSource))
	if count == nil {
		t.Fatal("PacketPulseStringsIndex.go declares no StringCount")
	}
	stringCount, _ := strconv.Atoi(count[1])

	if len(goIndex) != stringCount || len(dartIndex) != stringCount {
		t.Fatalf("Go has %d constants, Dart %d, StringCount says %d", len(goIndex), len(dartIndex), stringCount)
	}

	// Every position is used exactly once, so the constants describe one table.
	used := make([]string, stringCount)
	for name, position := range goIndex {
		if position < 0 || position >= stringCount || used[position] != "" {
			t.Fatalf("STR_%s_INDEX = %d collides or is out of range", name, position)
		}
		used[position] = name

		lower := toDartName(name)
		if dart, ok := dartIndex[lower]; !ok {
			t.Errorf("STR_%s_INDEX has no Dart constant %s", name, lower)
		} else if dart != position {
			t.Errorf("%s is %d in Go and %d in Dart: the client would draw the wrong label", name, position, dart)
		}
	}

	// And the table itself is in that order, entry for entry.
	tableSource, err := os.ReadFile(filepath.Join(constants, "PacketPulseStrings.go"))
	if err != nil {
		t.Fatal(err)
	}
	entries := stringIdPattern.FindAllStringSubmatch(string(tableSource), -1)
	if len(entries) != stringCount {
		t.Fatalf("PacketPulseStrings.go holds %d entries, want %d", len(entries), stringCount)
	}
	for position, entry := range entries {
		if entry[1] != strconv.Itoa(position) {
			t.Fatalf("entry %d of PacketPulseStrings.go claims id %s", position, entry[1])
		}
	}
}

func toDartName(goName string) string {
	lower := []byte(goName)
	for i, c := range lower {
		if c >= 'A' && c <= 'Z' {
			lower[i] = c + ('a' - 'A')
		}
	}
	return string(lower)
}

func TestNoOtherLanguageTablesShip(t *testing.T) {
	constants := filepath.Join(repoRoot(t), "packetpulsego", "pkg", "initmicroservice", "initconstants")
	for _, retired := range []string{"langstrings", "langhelp"} {
		if isDir(filepath.Join(constants, retired)) {
			t.Errorf("initconstants/%s is back: PacketPulse is English only, and a table nobody translates degrades with every new string", retired)
		}
	}
}

// ── Against a running server ─────────────────────────────────────────────────

func TestOnlyEnglishIsOffered(t *testing.T) {
	packetpulsetest.RequireServer(t)

	response := packetpulsetest.Call(t, http.MethodGet, "/init/language/list", "", nil)
	if response.Status != http.StatusOK {
		t.Fatalf("listing languages: %d", response.Status)
	}
	languages := packetpulsetest.ListOf(t, response, "languages")
	if len(languages) != 1 {
		t.Fatalf("%d languages offered, want English alone: %v", len(languages), languages)
	}
	english := languages[0]
	if english["language_id"] != "0" || english["code"] != "en" || english["english_name"] != "English" || english["is_right_to_left"] != false {
		t.Errorf("the one language = %v, want English, left to right", english)
	}
}

// "3" was Arabic and "22" Swahili. Accounts that chose them still send them,
// and must be drawn in complete English, left to right.
func TestEveryLanguageIdIsServedTheCompleteEnglishCatalogue(t *testing.T) {
	packetpulsetest.RequireServer(t)

	var english []map[string]any
	for _, languageId := range []string{"0", "3", "22"} {
		response := packetpulsetest.Call(t, http.MethodGet, "/init?language_id="+languageId, "", nil)
		if response.Status != http.StatusOK {
			t.Fatalf("language %s: %d", languageId, response.Status)
		}
		if rtl, _ := response.Body["is_right_to_left"].(bool); rtl {
			t.Errorf("language %s is laid out right to left", languageId)
		}
		served := packetpulsetest.ListOf(t, response, "strings")
		if len(served) == 0 {
			t.Fatalf("language %s: the catalogue is empty", languageId)
		}
		for position, entry := range served {
			if id, _ := entry["string_id"].(float64); int(id) != position {
				t.Fatalf("language %s: position %d holds string %v", languageId, position, entry["string_id"])
			}
			if value, _ := entry["string_value"].(string); value == "" {
				t.Errorf("language %s: string %d is empty on screen", languageId, position)
			}
		}
		if english == nil {
			english = served
			continue
		}
		if fmt.Sprint(served) != fmt.Sprint(english) {
			t.Errorf("language %s is not served the English catalogue", languageId)
		}
	}
}
