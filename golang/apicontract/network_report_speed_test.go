package apicontract

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// The line's speed a browser's UDP test filed reaches the network reports
// (customer request, 17 November 2026): the row of what it measured - here
// the PacketPulse network - states each way's tests, average, minimum and
// maximum, its test lists its own speed, and the CSV carries it in its last
// eight columns. A test filed with no speed adds none: absent, never 0.
func TestALinesSpeedReachesTheNetworkReports(t *testing.T) {
	packetpulsetest.RequireServer(t)
	token, _ := packetpulsetest.SignUpOrganisation(t, "reportspeed")
	endpoint := addEndpoint(t, token, "Speed Core", "192.0.2.72")
	stamp := time.Now().UnixNano()
	measured := udpRun(fmt.Sprintf("TT-RPTSPEED-%d", stamp), endpoint, "packetpulse")
	measured["download_mbps"], measured["upload_mbps"] = 87.44, 12.25
	for _, filing := range []map[string]any{measured, udpRun(fmt.Sprintf("TT-RPTNOSPEED-%d", stamp), endpoint, "packetpulse")} {
		if attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", token, filing); attached.Status != http.StatusCreated {
			t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
		}
	}
	today := kolkataToday(t)

	report := packetpulsetest.Call(t, http.MethodGet, "/report/network?period=day&date="+today, token, nil)
	if report.Status != http.StatusOK {
		t.Fatalf("report: %d %s", report.Status, report.Raw)
	}
	network := packetpulsetest.ListOf(t, report, "network")
	if len(network) != 1 {
		t.Fatalf("network rows = %v", network)
	}
	row := network[0]
	if row["tests"] != 2.0 || row["download_tests"] != 1.0 || row["download_avg_mbps"] != 87.44 ||
		row["download_min_mbps"] != 87.44 || row["download_max_mbps"] != 87.44 || row["upload_tests"] != 1.0 ||
		row["upload_avg_mbps"] != 12.25 {
		t.Errorf("the network row's speed = %v", row)
	}
	withSpeed := 0
	for _, entry := range packetpulsetest.ListOf(t, report, "network_entries") {
		switch entry["download_mbps"] {
		case 87.44:
			withSpeed++
		case nil:
		default:
			t.Errorf("a test lists %v down", entry["download_mbps"])
		}
	}
	if withSpeed != 1 {
		t.Errorf("%d tests list the speed, want the one that filed it", withSpeed)
	}

	csv := readCsv(t, packetpulsetest.Call(t, http.MethodGet, "/report/network.csv?period=day&date="+today, token, nil))
	columns := csv[0]
	at := map[string]int{}
	for index, name := range columns {
		at[name] = index
	}
	found := false
	for _, record := range csv[1:] {
		if record[at["section"]] == "network" && record[at["level"]] == "network" {
			found = true
			if record[at["download_tests"]] != "1" || record[at["download_avg_mbps"]] != "87.44" ||
				record[at["upload_avg_mbps"]] != "12.25" || record[at["upload_max_mbps"]] != "12.25" {
				t.Errorf("the CSV's network row = %v", record)
			}
		}
	}
	if !found {
		t.Errorf("the CSV has no network row: %v", csv)
	}
}
