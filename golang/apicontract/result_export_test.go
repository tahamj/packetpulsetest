package apicontract

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// exportColumns is the CSV header every export promises, from the contract
// file the server's own writer is tested against too.
func exportColumns(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "contracts", "result_export_columns.json"))
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

func readCsv(t *testing.T, response packetpulsetest.Response) [][]string {
	t.Helper()
	if response.Status != http.StatusOK {
		t.Fatalf("export: %d %s", response.Status, response.Raw)
	}
	if got := response.Header.Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}
	records, err := csv.NewReader(strings.NewReader(string(response.Raw))).ReadAll()
	if err != nil {
		t.Fatalf("not CSV: %v\n%s", err, response.Raw)
	}
	return records
}

// A ticket's CSV and the Results API's bulk pull are the same file, under
// the header an IT system's importer is written against.
func TestTheExportsCarryTheContractedHeader(t *testing.T) {
	packetpulsetest.RequireServer(t)
	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "csvcontract")
	packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"latitude": 12.9716, "longitude": 77.5946, "site_name": "loopback", "ip_address": "127.0.0.1",
	})
	ttNumber := fmt.Sprintf("TT-CSV-%d", time.Now().UnixNano())
	submit := packetpulsetest.FileDeviceTest(t, ownerToken, "CUST-CSV", ttNumber, nil)
	if submit.Status != http.StatusCreated {
		t.Fatalf("could not file a test: %d %s", submit.Status, submit.Raw)
	}
	request, _ := submit.Body["request"].(map[string]any)
	requestId, _ := request["request_id"].(string)
	columns := exportColumns(t)

	ticket := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId+"/report.csv", ownerToken, nil)
	records := readCsv(t, ticket)
	if !reflect.DeepEqual(records[0], columns) {
		t.Errorf("ticket CSV header = %v\nwant %v", records[0], columns)
	}
	if len(records) < 2 || records[1][3] != ttNumber {
		t.Errorf("ticket CSV = %v, want a row for %s", records, ttNumber)
	}
	if !strings.HasSuffix(ticket.Header.Get("Content-Disposition"), `.csv"`) {
		t.Errorf("Content-Disposition = %q", ticket.Header.Get("Content-Disposition"))
	}

	issued := packetpulsetest.Call(t, http.MethodPost, "/apikey/add", ownerToken, map[string]any{"label": "csv contract"})
	if issued.Status != http.StatusCreated {
		t.Fatalf("could not issue an API key: %d %s", issued.Status, issued.Raw)
	}
	from := url.QueryEscape(time.Now().UTC().Add(-time.Hour).Format(time.RFC3339))
	pulled := readCsv(t, packetpulsetest.Call(t, http.MethodGet, "/result/export.csv?from="+from, issued.String("api_key"), nil))
	if !reflect.DeepEqual(pulled[0], columns) {
		t.Errorf("pulled CSV header = %v", pulled[0])
	}
	found := false
	for _, record := range pulled[1:] {
		found = found || record[3] == ttNumber
	}
	if !found {
		t.Errorf("the pull did not hold %s: %v", ttNumber, pulled)
	}

	// A period longer than a month is refused, as an error, not a file.
	tooLong := url.QueryEscape(time.Now().UTC().AddDate(0, 0, -40).Format(time.RFC3339))
	refused := packetpulsetest.Call(t, http.MethodGet, "/result/export.csv?from="+tooLong, issued.String("api_key"), nil)
	if refused.Status != http.StatusUnprocessableEntity || strings.HasPrefix(refused.Header.Get("Content-Type"), "text/csv") {
		t.Errorf("a 40-day pull = %d %s", refused.Status, refused.Header.Get("Content-Type"))
	}
}

// The line's speed a device filed is in every place an IT system reads a
// result from, as the same two figures (customer request, 17 November 2026):
// the ticket's CSV, the Results API's bulk pull, and the Results API's ticket
// as JSON - megabits a second, two places in a file and as filed in JSON, on
// every row of the test; and a test that measured none says nothing in a file
// and null in JSON, never 0.
func TestTheLinesSpeedIsInTheCsvAndTheResultsApi(t *testing.T) {
	packetpulsetest.RequireServer(t)
	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "csvspeed")
	packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"latitude": 12.9716, "longitude": 77.5946, "site_name": "loopback", "ip_address": "127.0.0.1",
	})
	stamp := time.Now().UnixNano()
	measured, plain := fmt.Sprintf("TT-CSVSPEED-%d", stamp), fmt.Sprintf("TT-CSVNOSPEED-%d", stamp)
	filed := packetpulsetest.FileDeviceTest(t, ownerToken, "CUST-CSVSPEED", measured,
		map[string]any{"download_mbps": 87.436, "upload_mbps": 12.254})
	if filed.Status != http.StatusCreated {
		t.Fatalf("could not file a test: %d %s", filed.Status, filed.Raw)
	}
	if none := packetpulsetest.FileDeviceTest(t, ownerToken, "CUST-CSVSPEED", plain, nil); none.Status != http.StatusCreated {
		t.Fatalf("could not file a test: %d %s", none.Status, none.Raw)
	}
	requestId, _ := filed.Body["request"].(map[string]any)["request_id"].(string)
	columns := exportColumns(t)
	if len(columns) < 2 || columns[len(columns)-2] != "download_mbps" || columns[len(columns)-1] != "upload_mbps" {
		t.Fatalf("the last columns of the contract are %v, want download_mbps and upload_mbps", columns[max(0, len(columns)-2):])
	}
	speeds := func(records [][]string, ticket string) (download, upload string, found bool) {
		for _, record := range records[1:] {
			if record[3] == ticket {
				return record[len(columns)-2], record[len(columns)-1], true
			}
		}
		return "", "", false
	}

	fromTicket := readCsv(t, packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId+"/report.csv", ownerToken, nil))
	if down, up, ok := speeds(fromTicket, measured); !ok || down != "87.44" || up != "12.25" {
		t.Errorf("ticket CSV speed = %q / %q (found %v), want 87.44 / 12.25", down, up, ok)
	}

	apiKey := issueKey(t, ownerToken)
	from := url.QueryEscape(time.Now().UTC().Add(-time.Hour).Format(time.RFC3339))
	pulled := readCsv(t, packetpulsetest.Call(t, http.MethodGet, "/result/export.csv?from="+from, apiKey, nil))
	if down, up, ok := speeds(pulled, measured); !ok || down != "87.44" || up != "12.25" {
		t.Errorf("pulled CSV speed = %q / %q (found %v), want 87.44 / 12.25", down, up, ok)
	}
	if down, up, ok := speeds(pulled, plain); !ok || down != "" || up != "" {
		t.Errorf("a test that measured no speed has %q / %q in the pull (found %v), want empty cells", down, up, ok)
	}

	ticketJson := packetpulsetest.Call(t, http.MethodGet, "/result/bytt/"+measured, apiKey, nil)
	if ticketJson.Body["download_mbps"] != 87.436 || ticketJson.Body["upload_mbps"] != 12.254 {
		t.Errorf("Results API speed = %v / %v, want 87.436 / 12.254", ticketJson.Body["download_mbps"], ticketJson.Body["upload_mbps"])
	}
	noneJson := packetpulsetest.Call(t, http.MethodGet, "/result/bytt/"+plain, apiKey, nil)
	for _, field := range []string{"download_mbps", "upload_mbps"} {
		if value, present := noneJson.Body[field]; !present || value != nil {
			t.Errorf("%s = %v (present %v) for a test that measured none, want null", field, value, present)
		}
	}
	// And by Customer ID: each ticket in the shape of one by TT number.
	byCustomer := packetpulsetest.Call(t, http.MethodGet, "/result/bycustomer/CUST-CSVSPEED?from="+from, apiKey, nil)
	sawSpeed := false
	for _, ticket := range packetpulsetest.ListOf(t, byCustomer, "tickets") {
		if ticket["tt_number"] == measured && ticket["download_mbps"] == 87.436 && ticket["upload_mbps"] == 12.254 {
			sawSpeed = true
		}
	}
	if !sawSpeed {
		t.Errorf("a customer's tickets do not carry the line's speed: %s", byCustomer.Raw)
	}
}

func TestTheBulkPullRefusesTheWrongCredential(t *testing.T) {
	packetpulsetest.RequireServer(t)
	sessionToken, _ := packetpulsetest.SignUpOrganisation(t, "csvauth")
	for name, token := range map[string]string{
		"no credential":         "",
		"a human session token": sessionToken,
		"a malformed API key":   "packetpulse_development_000000000000_nope",
	} {
		if response := packetpulsetest.Call(t, http.MethodGet, "/result/export.csv", token, nil); response.Status != http.StatusUnauthorized {
			t.Errorf("%s: expected 401, got %d", name, response.Status)
		}
	}
}

// The connection test sends a request to a server somebody typed, which is
// the same door the probe engine guards. It must not open onto the cloud
// metadata service, or onto PacketPulse's own host.
func TestAConnectionTestCannotReachPacketPulsesOwnHost(t *testing.T) {
	packetpulsetest.RequireServer(t)
	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "exportssrf")
	for _, host := range []string{"169.254.169.254", "127.0.0.1", "localhost", "::1"} {
		for _, protocol := range []string{"sftp", "ftp"} {
			response := packetpulsetest.Call(t, http.MethodPost, "/export/target/test", ownerToken, map[string]any{
				"protocol": protocol, "host": host, "username": "probe", "password": "probe",
			})
			envelope, _ := response.Body["error"].(map[string]any)
			details, _ := envelope["details"].(map[string]any)
			if response.Status != http.StatusUnprocessableEntity || details["host"] == nil {
				t.Errorf("%s over %s: %d %s, want refused against the host field", host, protocol, response.Status, response.Raw)
			}
		}
	}
}
