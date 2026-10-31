package apicontract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// cityBannerContract is what the app's city banners read from a network
// report, from the contract file the client's own test parses too.
type cityBannerContract struct {
	Report              []string `json:"report"`
	Site                []string `json:"site"`
	PanIndiaCountryCode string   `json:"pan_india_country_code"`
}

func readCityBannerContract(t *testing.T) cityBannerContract {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "contracts", "city_banner_fields.json"))
	if err != nil {
		t.Fatalf("read the contract: %v", err)
	}
	var contract cityBannerContract
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatal(err)
	}
	if len(contract.Report) == 0 || len(contract.Site) == 0 || contract.PanIndiaCountryCode == "" {
		t.Fatalf("the contract names nothing: %s", raw)
	}
	return contract
}

// The banners show each site's city as the report files it and match the
// drawing in the app: an endpoint typed with an older name or in Devanagari
// is filed under the city's own name (pkg/common/cityname), the name the app
// draws - "Bombay" is Mumbai's skyline, never a site of its own. A report of
// every site combined says which country it is for: the Pan India drawing
// and tagline are India's alone, and it counts the organisation's cities.
func TestTheNetworkReportCarriesWhatTheCityBannersRead(t *testing.T) {
	packetpulsetest.RequireServer(t)
	contract := readCityBannerContract(t)
	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "citybanner")
	stamp := time.Now().UnixNano()
	written := []string{"Bombay", "पुणे"}
	filed := []string{"Mumbai", "Pune"}
	for i, city := range written {
		site := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
			"latitude": 19.0 + float64(i), "longitude": 73.0, "site_name": fmt.Sprintf("Banner POP %d", i),
			"ip_address": fmt.Sprintf("192.0.2.%d", 90+i), "is_enabled": true, "region": city,
		})
		endpoint, _ := site.Body["dns_site_id"].(string)
		if site.Status != http.StatusCreated || endpoint == "" {
			t.Fatalf("add the %s endpoint: %d %s", city, site.Status, site.Raw)
		}
		run := oneRun(fmt.Sprintf("TT-BANNER-%d-%d", stamp, i))
		run["dns_site_id"] = endpoint
		run["observations"] = run["observations"].([]map[string]any)[:1]
		if attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", ownerToken, run); attached.Status != http.StatusCreated {
			t.Fatalf("a test of the %s endpoint: %d %s", city, attached.Status, attached.Raw)
		}
	}
	today := kolkataToday(t)

	report := packetpulsetest.Call(t, http.MethodGet, "/report/network?period=day&date="+today, ownerToken, nil)
	if report.Status != http.StatusOK || report.String("kind") != "city" {
		t.Fatalf("report by site: %d %s", report.Status, report.Raw)
	}
	for _, field := range contract.Report {
		// Only a report of every site combined is for a country.
		if _, present := report.Body[field]; !present && field != "country_code" {
			t.Errorf("the report by site has no %q: %s", field, report.Raw)
		}
	}
	var sites []string
	for _, site := range packetpulsetest.ListOf(t, report, "sites") {
		for _, field := range contract.Site {
			if _, present := site[field]; !present {
				t.Errorf("a site has no %q: %v", field, site)
			}
		}
		city, _ := site["city"].(string)
		sites = append(sites, city)
	}
	slices.Sort(sites)
	if !slices.Equal(sites, filed) {
		t.Errorf("sites are %q, want %q: each city under the name the app draws", sites, filed)
	}
	assertCities := func(name string, body map[string]any) {
		listed, _ := body["cities"].([]any)
		for _, city := range filed {
			if !slices.Contains(listed, any(city)) {
				t.Errorf("the %s report's cities %v leave out %q", name, listed, city)
			}
		}
	}
	assertCities("by-site", report.Body)

	country := packetpulsetest.Call(t, http.MethodGet, "/report/network?view=country&period=day&date="+today, ownerToken, nil)
	if country.Status != http.StatusOK || country.String("kind") != "country" {
		t.Fatalf("report of every site: %d %s", country.Status, country.Raw)
	}
	for _, field := range contract.Report {
		if _, present := country.Body[field]; !present {
			t.Errorf("the report of every site has no %q: %s", field, country.Raw)
		}
	}
	if code := country.String("country_code"); code != contract.PanIndiaCountryCode {
		t.Errorf("an organisation in India has a whole-country report for %q, want %q", code, contract.PanIndiaCountryCode)
	}
	assertCities("whole-country", country.Body)
}
