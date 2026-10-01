// Package translationcoverage reports what is not yet translated.
//
// It does NOT fail on an incomplete language: shipping English where a
// translation is missing is correct behaviour, and a test that failed on it
// would only train people to skip the suite. It fails on the things that are
// actually wrong - a missing language, or a catalogue the two sides disagree
// about.
package translationcoverage

import (
	"net/http"
	"sort"
	"testing"

	pingletest "github.com/tahamj/pingletest"
)

const expectedLanguageCount = 23

func TestAllLanguagesArePresent(t *testing.T) {
	pingletest.RequireServer(t)

	response := pingletest.Call(t, http.MethodGet, "/init/language/list", "", nil)
	if response.Status != http.StatusOK {
		t.Fatalf("listing languages: %d", response.Status)
	}

	languages := pingletest.ListOf(t, response, "languages")
	if len(languages) != expectedLanguageCount {
		t.Errorf("expected %d languages, got %d", expectedLanguageCount, len(languages))
	}

	rightToLeft := 0
	for _, language := range languages {
		if isRtl, _ := language["is_right_to_left"].(bool); isRtl {
			rightToLeft++
		}
		for _, field := range []string{"language_id", "code", "english_name", "native_name"} {
			if value, _ := language[field].(string); value == "" {
				t.Errorf("language %v has no %s", language["language_id"], field)
			}
		}
	}

	// Arabic, Persian, Sindhi and Urdu. Sindhi is the one that gets missed,
	// because it is grouped with the Indic languages by name but written in
	// Perso-Arabic script.
	if rightToLeft != 4 {
		t.Errorf("expected 4 right-to-left languages, got %d", rightToLeft)
	}
}

func TestEveryLanguageServesACompleteStringArray(t *testing.T) {
	pingletest.RequireServer(t)

	english := pingletest.Call(t, http.MethodGet, "/init?language_id=0", "", nil)
	expected := len(pingletest.ListOf(t, english, "strings"))
	if expected == 0 {
		t.Fatal("the English catalogue is empty")
	}

	for languageId := 0; languageId < expectedLanguageCount; languageId++ {
		response := pingletest.Call(t, http.MethodGet,
			"/init?language_id="+itoa(languageId), "", nil)
		if response.Status != http.StatusOK {
			t.Errorf("language %d: %d", languageId, response.Status)
			continue
		}

		strings := pingletest.ListOf(t, response, "strings")
		// The array must be DENSE and complete for every language. The client
		// looks strings up by position, so a short array is an
		// index-out-of-range crash on screens that reference a later string.
		if len(strings) != expected {
			t.Errorf("language %d serves %d strings, English has %d: "+
				"positional lookup would crash the client",
				languageId, len(strings), expected)
		}
		for index, entry := range strings {
			if value, _ := entry["string_value"].(string); value == "" {
				t.Errorf("language %d has an empty string at index %d: "+
					"the English fallback did not fill it", languageId, index)
			}
		}
	}
}

func TestCoverageIsReported(t *testing.T) {
	pingletest.RequireServer(t)

	response := pingletest.Call(t, http.MethodGet, "/init/coverage", "", nil)
	if response.Status != http.StatusOK {
		t.Fatalf("reading coverage: %d", response.Status)
	}

	rows := pingletest.ListOf(t, response, "languages")
	sort.Slice(rows, func(i, j int) bool {
		return toInt(rows[i]["coverage_percent"]) > toInt(rows[j]["coverage_percent"])
	})

	t.Logf("%d strings in the catalogue", int(response.Float("string_count")))
	t.Logf("%-12s %-12s %s", "LANGUAGE", "NATIVE", "COVERAGE")

	incomplete := 0
	for _, row := range rows {
		coverage := toInt(row["coverage_percent"])
		if coverage < 100 {
			incomplete++
		}
		marker := ""
		if isRtl, _ := row["is_right_to_left"].(bool); isRtl {
			marker = "  RTL"
		}
		t.Logf("%-12v %-12v %3d%%%s",
			row["english_name"], row["native_name"], coverage, marker)
	}

	if incomplete > 0 {
		t.Logf("\n%d language(s) are not fully translated. Untranslated entries "+
			"serve English, which is why this is a report and not a failure.", incomplete)
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return digits
}

func toInt(value any) int {
	if number, ok := value.(float64); ok {
		return int(number)
	}
	return 0
}
