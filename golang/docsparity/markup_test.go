package docsparity

import (
	"regexp"
	"strings"
	"testing"
)

// htmlTag is one HTML tag or comment, and nothing else: a tag opens with a
// letter, "/" or "!", and holds no "<". The guards read each text with its
// tags taken out.
//
// It was `<[^>]+>`, which took any "<" for a tag: a loss written "<0.1%" in
// Markdown or in the string catalogue ran on to the next ">" and was taken
// out with everything between (found 2 November 2026). That hid 2,437 lines
// of the string catalogue, 649 of the deployment record and 27 of the
// testing guide's diagnostics part from every guard that reads them.
var htmlTag = regexp.MustCompile(`<!--(?s:.*?)-->|<[!/]?[A-Za-z][^<>]*>`)

func TestATagIsOnlyATag(t *testing.T) {
	for _, text := range []string{
		"Average loss *<0.1%* and *25%*, and everyone picks by hand. Then a > b.",
		"seats_in_use <= limit, and everyone picks by hand -> done",
		"| Reflector turnaround | < 0.1 ms | everyone picks by hand |",
	} {
		if plain := htmlTag.ReplaceAllString(text, " "); !strings.Contains(plain, "everyone picks by hand") {
			t.Errorf("text that is not a tag was taken out: %q became %q", text, plain)
		}
	}
	for _, tagged := range []string{
		`<p class="small">`, "</td>", "<br/>", "<!DOCTYPE html>", "<!-- a note\nover two lines -->",
		`<img src="data:image/png;base64,iVBORw0KGgo=" alt="">`, "<svg viewBox=\"0 0 700 214\"\n  role=\"img\">",
	} {
		if plain := strings.TrimSpace(htmlTag.ReplaceAllString("a"+tagged+"b", "")); plain != "ab" {
			t.Errorf("a tag was left in: %q became %q", tagged, plain)
		}
	}
}
