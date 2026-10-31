package apicontract

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// repoRoot is the checkout holding every PacketPulse repository.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func readSource(t *testing.T, parts ...string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(append([]string{repoRoot(t)}, parts...)...))
	if err != nil {
		t.Fatalf("read %s: %v", filepath.Join(parts...), err)
	}
	return string(raw)
}

var (
	drawnCityPattern = regexp.MustCompile(`(?s)PacketPulseCity\(\s*key: '([a-z_]+)',\s*names: \[(.*?)\],`)
	quotedPattern    = regexp.MustCompile(`'([^']+)'`)
	aliasPattern     = regexp.MustCompile(`"([^"]+)": "([^"]+)"`)
)

// The app draws a city by matching its name
// (packetpulseflutter/lib/common/cities/PacketPulseCities.dart), and the
// server files a report's reflector sites and sweeps the nearest city by its
// own list (packetpulsego/pkg/common/cityname). For every city the app draws,
// the two lists are the same names: a name only the server knows is a site
// the report files as Mumbai and the app draws as nowhere, and a name only
// the app knows is two reflector sites drawn as one city.
func TestTheAppAndTheServerKnowEveryDrawnCityByTheSameNames(t *testing.T) {
	client := readSource(t, "packetpulseflutter", "lib", "common", "cities", "PacketPulseCities.dart")
	server := readSource(t, "packetpulsego", "pkg", "common", "cityname", "CityName.go")

	byCity := map[string][]string{} // filed-under name -> its aliases, lower case
	for _, alias := range aliasPattern.FindAllStringSubmatch(server, -1) {
		byCity[alias[2]] = append(byCity[alias[2]], alias[1])
	}
	drawn := 0
	for _, city := range drawnCityPattern.FindAllStringSubmatch(client, -1) {
		quoted := quotedPattern.FindAllStringSubmatch(city[2], -1)
		if len(quoted) == 0 {
			continue // Pan India: a drawing, not a city a site is filed under
		}
		drawn++
		var names []string
		for _, name := range quoted {
			names = append(names, strings.ToLower(strings.Join(strings.Fields(name[1]), " ")))
		}
		filedAs := quoted[0][1]
		served := slices.Clone(byCity[filedAs])
		slices.Sort(names)
		slices.Sort(served)
		if !slices.Equal(names, served) {
			t.Errorf("%s: the app knows it as %q, the server files it under %q", city[1], names, served)
		}
	}
	if drawn < 9 {
		t.Errorf("found %d drawn cities in the app, want at least the nine it shipped with", drawn)
	}
}
