package apicontract

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// networkReportColumns is the header every network report CSV promises, from
// the contract file the server's own writer is tested against too.
func networkReportColumns(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "contracts", "network_report_columns.json"))
	if err != nil {
		t.Fatalf("read the contract: %v", err)
	}
	var contract struct {
		Columns []string `json:"columns"`
	}
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatal(err)
	}
	return contract.Columns
}

// kolkataToday is today in Asia/Kolkata: the zone of an organisation in
// India that chose none, which its reports' days begin in.
func kolkataToday(t *testing.T) string {
	t.Helper()
	kolkata, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	return time.Now().In(kolkata).Format("2006-01-02")
}

// The report by reflector site and the Pan India report of every site
// combined: as data for the screen, and as the CSV and PDF a team keeps - the
// CSV under its contracted header, and the files named for what they hold.
func TestTheNetworkReportsKeepTheirContract(t *testing.T) {
	packetpulsetest.RequireServer(t)
	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "reportcontract")
	city := fmt.Sprintf("Contractpur%d", time.Now().UnixNano())
	site := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"latitude": 12.9716, "longitude": 77.5946, "site_name": "Contract POP", "ip_address": "192.0.2.81",
		"is_enabled": true, "region": city,
	})
	endpoint, _ := site.Body["dns_site_id"].(string)
	if site.Status != http.StatusCreated || endpoint == "" {
		t.Fatalf("add endpoint: %d %s", site.Status, site.Raw)
	}
	// A device's test of the endpoint, over HTTPS.
	run := oneRun(fmt.Sprintf("TT-REPORT-%d", time.Now().UnixNano()))
	run["dns_site_id"] = endpoint
	run["observations"] = run["observations"].([]map[string]any)[:1]
	attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", ownerToken, run)
	if attached.Status != http.StatusCreated {
		t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
	}
	today := kolkataToday(t)

	report := packetpulsetest.Call(t, http.MethodGet, "/report/network?period=day&date="+today, ownerToken, nil)
	if report.Status != http.StatusOK {
		t.Fatalf("report: %d %s", report.Status, report.Raw)
	}
	period, _ := report.Body["period"].(map[string]any)
	if report.String("kind") != "city" || period["time_zone"] != "Asia/Kolkata" || period["slug"] != today {
		t.Errorf("report = %s", report.Raw)
	}
	// The city where a test was run is not reported (customer, 22 October
	// 2026): only the reflector site - the endpoint's city.
	if _, present := report.Body["tested_in"]; present {
		t.Errorf("the report still files tests by where they were run: %s", report.Raw)
	}
	sites := packetpulsetest.ListOf(t, report, "sites")
	if len(sites) != 1 || sites[0]["city"] != city {
		t.Fatalf("sites = %v", sites)
	}
	totals, _ := sites[0]["totals"].([]any)
	endpoints, _ := sites[0]["endpoints"].([]any)
	if len(totals) != 1 || len(endpoints) != 1 {
		t.Fatalf("site rows = %v / %v", totals, endpoints)
	}
	total, _ := totals[0].(map[string]any)
	row, _ := endpoints[0].(map[string]any)
	// An HTTPS test counts queries, not packets, and states no packet loss;
	// its latency is its average, minimum and maximum.
	if total["level"] != "city" || total["method"] != "https-dns" || total["tests"] != 1.0 ||
		total["loss_avg_pct"] != nil || math.Abs(total["rtt_avg_ms"].(float64)-49.8) > 1e-9 ||
		total["rtt_min_ms"] != 34.8 || total["rtt_max_ms"] != 60.4 {
		t.Errorf("site row = %v", total)
	}
	if row["level"] != "endpoint" || row["endpoint_name"] != "Contract POP" || row["tests"] != 1.0 {
		t.Errorf("endpoint row = %v", row)
	}
	coverage, _ := report.Body["coverage"].(map[string]any)
	if coverage["from_device"] != 1.0 || coverage["to_endpoints"] != 1.0 {
		t.Errorf("coverage = %v", coverage)
	}

	csv := readCsv(t, packetpulsetest.Call(t, http.MethodGet, "/report/network.csv?period=day&date="+today, ownerToken, nil))
	if !reflect.DeepEqual(csv[0], networkReportColumns(t)) {
		t.Errorf("CSV header = %v\nwant %v", csv[0], networkReportColumns(t))
	}
	if !strings.Contains(strings.Join(csv[1], ","), city) {
		t.Errorf("CSV first row = %v, want the endpoint's city", csv[1])
	}
	// Every row says whose tests it adds up, as the report itself does.
	organisation := report.String("organisation_name")
	if organisation == "" {
		t.Errorf("the report names no organisation: %s", report.Raw)
	}
	for _, row := range csv[1:] {
		if row[0] != organisation {
			t.Errorf("a %s row names %q, want %q", row[2], row[0], organisation)
		}
	}

	country := packetpulsetest.Call(t, http.MethodGet, "/report/network?view=country&period=week&date="+today, ownerToken, nil)
	if country.Status != http.StatusOK || country.String("kind") != "country" || country.String("country_code") != "IN" ||
		country.String("title") != "Pan India network performance" || country.Body["previous_period"] == nil ||
		len(packetpulsetest.ListOf(t, country, "all")) != 1 {
		t.Errorf("country report = %d %s", country.Status, country.Raw)
	}

	pdf := packetpulsetest.Call(t, http.MethodGet, "/report/network.pdf?view=country&period=month&date="+today, ownerToken, nil)
	if pdf.Status != http.StatusOK || pdf.Header.Get("Content-Type") != "application/pdf" || !strings.HasPrefix(string(pdf.Raw), "%PDF-") ||
		pdf.Header.Get("Content-Disposition") != `attachment; filename="packetpulse-pan-india-monthly-`+today[:7]+`-so-far.pdf"` {
		t.Errorf("PDF = %d %q %q", pdf.Status, pdf.Header.Get("Content-Type"), pdf.Header.Get("Content-Disposition"))
	}
	// Asked for no day, the report is today so far - the test just filed is
	// in it - and says so; its files are named so too.
	open := packetpulsetest.Call(t, http.MethodGet, "/report/network", ownerToken, nil)
	if openPeriod, _ := open.Body["period"].(map[string]any); open.Status != http.StatusOK || openPeriod["slug"] != today ||
		openPeriod["in_progress"] != true || openPeriod["as_of"] == nil || len(packetpulsetest.ListOf(t, open, "sites")) != 1 {
		t.Errorf("an open report = %d %s", open.Status, open.Raw)
	}
	if soFar := packetpulsetest.Call(t, http.MethodGet, "/report/network.csv", ownerToken, nil); soFar.Header.Get("Content-Disposition") !=
		`attachment; filename="packetpulse-city-daily-`+today+`-so-far.csv"` {
		t.Errorf("today's CSV = %q", soFar.Header.Get("Content-Disposition"))
	}
	yearly := packetpulsetest.Call(t, http.MethodGet, "/report/network.csv?period=year&date="+today, ownerToken, nil)
	if yearly.Status != http.StatusOK ||
		yearly.Header.Get("Content-Disposition") != `attachment; filename="packetpulse-city-yearly-`+today[:4]+`-so-far.csv"` {
		t.Errorf("yearly CSV = %d %q", yearly.Status, yearly.Header.Get("Content-Disposition"))
	}
}

// What a report is asked for is checked, and a refusal names its field and
// is an error, never a file.
func TestANetworkReportQueryIsChecked(t *testing.T) {
	packetpulsetest.RequireServer(t)
	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "reportquery")
	for query, field := range map[string]string{
		"view=town":        "view",
		"period=fortnight": "period",
		"date=21-10-2026":  "date",
		"date=" + time.Now().UTC().AddDate(0, 0, 3).Format("2006-01-02"): "date",
		// A city is one of the organisation's own, and only narrows a report
		// by site (customer request, 28 October 2026).
		"city=Atlantis":            "city",
		"view=country&city=Mumbai": "city",
	} {
		for _, path := range []string{"/report/network", "/report/network.csv"} {
			response := packetpulsetest.Call(t, http.MethodGet, path+"?"+query, ownerToken, nil)
			envelope, _ := response.Body["error"].(map[string]any)
			details, _ := envelope["details"].(map[string]any)
			if response.Status != http.StatusUnprocessableEntity || details[field] == nil ||
				strings.HasPrefix(response.Header.Get("Content-Type"), "text/csv") {
				t.Errorf("%s?%s = %d %s, want refused against %s", path, query, response.Status, response.Raw, field)
			}
		}
	}
}

// The report settings round-trip; a zone PacketPulse does not know is
// refused; and scheduled reports cannot be switched on with no file server
// to send them to.
func TestTheReportSettingsAreCheckedAndKept(t *testing.T) {
	packetpulsetest.RequireServer(t)
	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "reportsettings")
	fresh := packetpulsetest.Call(t, http.MethodGet, "/report/settings", ownerToken, nil)
	// An organisation in India that chose no zone is in India's.
	if deliveries := packetpulsetest.ListOf(t, fresh, "deliveries"); fresh.String("time_zone") != "Asia/Kolkata" || len(deliveries) != 4 ||
		deliveries[0]["cadence"] != "daily" || deliveries[0]["is_enabled"] != false || deliveries[3]["cadence"] != "yearly" {
		t.Errorf("fresh settings = %s", fresh.Raw)
	}

	off := map[string]any{"daily": false, "weekly": false, "monthly": false, "yearly": false}
	saved := packetpulsetest.Call(t, http.MethodPut, "/report/settings", ownerToken, map[string]any{"time_zone": "Asia/Dubai", "deliveries": off})
	if saved.Status != http.StatusOK || saved.String("time_zone") != "Asia/Dubai" {
		t.Errorf("save = %d %s", saved.Status, saved.Raw)
	}
	if read := packetpulsetest.Call(t, http.MethodGet, "/report/settings", ownerToken, nil); read.String("time_zone") != "Asia/Dubai" {
		t.Errorf("read back = %s", read.Raw)
	}
	report := packetpulsetest.Call(t, http.MethodGet, "/report/network", ownerToken, nil)
	if period, _ := report.Body["period"].(map[string]any); period["time_zone"] != "Asia/Dubai" {
		t.Errorf("a report after the save is cut in %v", period["time_zone"])
	}

	for name, body := range map[string]map[string]any{
		"an unknown zone":        {"time_zone": "Mars/Olympus_Mons", "deliveries": off},
		"no zone":                {"deliveries": off},
		"a schedule left out":    {"time_zone": "UTC", "deliveries": map[string]any{"daily": true, "weekly": false, "monthly": false}},
		"on with no file server": {"time_zone": "UTC", "deliveries": map[string]any{"daily": false, "weekly": false, "monthly": false, "yearly": true}},
	} {
		response := packetpulsetest.Call(t, http.MethodPut, "/report/settings", ownerToken, body)
		if response.Status != http.StatusUnprocessableEntity {
			t.Errorf("%s = %d %s", name, response.Status, response.Raw)
		}
	}
}
