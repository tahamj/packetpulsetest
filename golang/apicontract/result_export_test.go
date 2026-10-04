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

	pingletest "github.com/tahamj/pingletest"
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

func readCsv(t *testing.T, response pingletest.Response) [][]string {
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
	pingletest.RequireServer(t)
	ownerToken, _ := pingletest.SignUpOrganisation(t, "csvcontract")
	pingletest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"site_name": "loopback", "ip_address": "127.0.0.1",
	})
	ttNumber := fmt.Sprintf("TT-CSV-%d", time.Now().UnixNano())
	submit := pingletest.Call(t, http.MethodPost, "/diagnostic/submit", ownerToken, map[string]any{
		"customer_id": "CUST-CSV", "tt_number": ttNumber, "packet_count": 1, "timeout_ms": 900,
	})
	if submit.Status != http.StatusCreated {
		t.Skipf("could not run a diagnostic: %d %s", submit.Status, submit.Raw)
	}
	request, _ := submit.Body["request"].(map[string]any)
	requestId, _ := request["request_id"].(string)
	columns := exportColumns(t)

	ticket := pingletest.Call(t, http.MethodGet, "/diagnostic/"+requestId+"/report.csv", ownerToken, nil)
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

	issued := pingletest.Call(t, http.MethodPost, "/apikey/add", ownerToken, map[string]any{"label": "csv contract"})
	if issued.Status != http.StatusCreated {
		t.Skipf("could not issue an API key: %d %s", issued.Status, issued.Raw)
	}
	from := url.QueryEscape(time.Now().UTC().Add(-time.Hour).Format(time.RFC3339))
	pulled := readCsv(t, pingletest.Call(t, http.MethodGet, "/result/export.csv?from="+from, issued.String("api_key"), nil))
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
	refused := pingletest.Call(t, http.MethodGet, "/result/export.csv?from="+tooLong, issued.String("api_key"), nil)
	if refused.Status != http.StatusUnprocessableEntity || strings.HasPrefix(refused.Header.Get("Content-Type"), "text/csv") {
		t.Errorf("a 40-day pull = %d %s", refused.Status, refused.Header.Get("Content-Type"))
	}
}

func TestTheBulkPullRefusesTheWrongCredential(t *testing.T) {
	pingletest.RequireServer(t)
	sessionToken, _ := pingletest.SignUpOrganisation(t, "csvauth")
	for name, token := range map[string]string{
		"no credential":         "",
		"a human session token": sessionToken,
		"a malformed API key":   "pingle_development_000000000000_nope",
	} {
		if response := pingletest.Call(t, http.MethodGet, "/result/export.csv", token, nil); response.Status != http.StatusUnauthorized {
			t.Errorf("%s: expected 401, got %d", name, response.Status)
		}
	}
}

// The connection test sends a request to a server somebody typed, which is
// the same door the probe engine guards. It must not open onto the cloud
// metadata service, or onto Pingle's own host.
func TestAConnectionTestCannotReachPinglesOwnHost(t *testing.T) {
	pingletest.RequireServer(t)
	ownerToken, _ := pingletest.SignUpOrganisation(t, "exportssrf")
	for _, host := range []string{"169.254.169.254", "127.0.0.1", "localhost", "::1"} {
		for _, protocol := range []string{"sftp", "ftp"} {
			response := pingletest.Call(t, http.MethodPost, "/export/target/test", ownerToken, map[string]any{
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
