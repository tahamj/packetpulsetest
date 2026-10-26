package translationcoverage

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// dartLiteral is a Dart string literal in single or double quotes.
var dartLiteral = regexp.MustCompile(`'((?:[^'\\]|\\.)*)'|"((?:[^"\\]|\\.)*)"`)

// prose is two words of a sentence: a word, then a lower-case word.
var prose = regexp.MustCompile(`(^|[^A-Za-z_$])[A-Za-z][a-z]+ [a-z]+`)

// notUserText is a line whose literals are never shown to a person.
var notUserText = regexp.MustCompile(`^\s*(//|\*)|^\s*(import|export|part) |Key\(|RegExp\(|Uri\.parse\(|debugPrint\(|assert\(`)

// No sentence is written in a widget: every sentence a person reads comes
// from the catalogue (AGENTS.md section 4), so it can be reworded without a
// client release and is held to the catalogue's checks. Two chart captions
// were written in the widget as "endpoint(s)" until 26 October 2026, and
// nothing noticed.
//
// It reads every screen and widget for a literal holding two words of prose.
// The engines' probe-log reasons and the API client's fallbacks are outside
// it: they have no BuildContext to read the catalogue from.
func TestNoSentenceIsWrittenInAWidget(t *testing.T) {
	lib := filepath.Join(repoRoot(t), "packetpulseflutter", "lib")
	checked := 0
	err := filepath.WalkDir(lib, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".dart") {
			return err
		}
		relative, _ := filepath.Rel(lib, path)
		if !strings.Contains(relative, "/presentation/") && !strings.HasPrefix(relative, "common/widgets/") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		checked++
		for number, line := range strings.Split(string(source), "\n") {
			if notUserText.MatchString(line) {
				continue
			}
			for _, match := range dartLiteral.FindAllStringSubmatch(line, -1) {
				text := match[1] + match[2]
				if prose.MatchString(text) && !strings.HasPrefix(text, "http") {
					t.Errorf("%s:%d writes a sentence in a widget - add it to the catalogue: %q", relative, number+1, text)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 20 {
		t.Fatalf("read %d screen and widget files; the walk is not finding them", checked)
	}
}
