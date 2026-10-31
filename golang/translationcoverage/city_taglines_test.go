package translationcoverage

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

var (
	drawnCityTaglinePattern = regexp.MustCompile(`(?s)PacketPulseCity\(\s*key: '([a-z_]+)',.*?tagline: PacketPulseStringsIndex\.(\w+),`)
	stringEntryPattern      = regexp.MustCompile(`\{StringId: (\d+), StringValue: ("(?:[^"\\]|\\.)*")\}`)
)

// drawnCityTagline is a city the app draws and the catalogue entry its
// banner reads its tagline from.
type drawnCityTagline struct {
	key, index string
	position   int
	text       string
}

func drawnCityTaglines(t *testing.T) []drawnCityTagline {
	t.Helper()
	root := repoRoot(t)
	cities, err := os.ReadFile(filepath.Join(root, "packetpulseflutter", "lib", "common", "cities", "PacketPulseCities.dart"))
	if err != nil {
		t.Fatal(err)
	}
	constants := filepath.Join(root, "packetpulsego", "pkg", "initmicroservice", "initconstants")
	goIndex := indexFrom(t, filepath.Join(constants, "PacketPulseStringsIndex.go"), goIndexPattern)
	table, err := os.ReadFile(filepath.Join(constants, "PacketPulseStrings.go"))
	if err != nil {
		t.Fatal(err)
	}
	texts := map[int]string{}
	for _, entry := range stringEntryPattern.FindAllStringSubmatch(string(table), -1) {
		position, _ := strconv.Atoi(entry[1])
		text, err := strconv.Unquote(entry[2])
		if err != nil {
			t.Fatalf("string %d: %v", position, err)
		}
		texts[position] = text
	}

	var found []drawnCityTagline
	for _, city := range drawnCityTaglinePattern.FindAllStringSubmatch(string(cities), -1) {
		position, ok := goIndex[strings.ToUpper(city[2])]
		if !ok {
			t.Errorf("%s reads %s, which the server's catalogue does not have", city[1], city[2])
			continue
		}
		found = append(found, drawnCityTagline{key: city[1], index: city[2], position: position, text: texts[position]})
	}
	if len(found) < 10 {
		t.Fatalf("found %d drawn cities' taglines, want the nine cities and Pan India at least", len(found))
	}
	return found
}

// Every city the app draws reads its tagline by catalogue position
// (packetpulseflutter/lib/common/cities/PacketPulseCities.dart). Each reads
// the entry named for it - a city copied from another and left with its
// tagline would read "Queen of the Deccan" over Nagpur - and the entry has
// words in the table the server serves.
func TestEveryDrawnCityHasItsOwnTaglineInTheCatalogue(t *testing.T) {
	for _, city := range drawnCityTaglines(t) {
		if city.index != "city_tagline_"+city.key {
			t.Errorf("%s reads %s, want city_tagline_%s", city.key, city.index, city.key)
		}
		if strings.TrimSpace(city.text) == "" {
			t.Errorf("%s's tagline (string %d) is empty", city.key, city.position)
		}
	}
}

// And the server serves each tagline at the position the app reads it from:
// one missing is a blank line under the city, one a place out is another
// city's nickname.
func TestEveryDrawnCityTaglineIsServedWhereTheAppReadsIt(t *testing.T) {
	packetpulsetest.RequireServer(t)
	response := packetpulsetest.Call(t, http.MethodGet, "/init", "", nil)
	if response.Status != http.StatusOK {
		t.Fatalf("/init: %d", response.Status)
	}
	served := packetpulsetest.ListOf(t, response, "strings")
	for _, city := range drawnCityTaglines(t) {
		if city.position >= len(served) {
			t.Errorf("%s's tagline is string %d; the server serves %d", city.key, city.position, len(served))
			continue
		}
		if got, _ := served[city.position]["string_value"].(string); got != city.text {
			t.Errorf("%s's tagline is served as %q, want %q", city.key, got, city.text)
		}
	}
}
