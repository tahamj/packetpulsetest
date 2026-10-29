package apicontract

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// A test is started by a person, from their device (decision of 30 October
// 2026): the server's own sweep is switched off unless SERVER_SWEEPS_ENABLED
// turns it on, and the server under test runs as live does. Neither route
// that started one exists for a member who may run diagnostics, and nothing
// is run or filed. How a sweep runs when switched on - the server's
// PING_COUNT and PING_TIMEOUT, never the caller's - is held by pingservice's
// and diagnosticapp's own tests.
func TestTheServersOwnSweepIsSwitchedOff(t *testing.T) {
	packetpulsetest.RequireServer(t)

	token, _ := packetpulsetest.SignUpOrganisation(t, "sweepsoff")
	addEndpoint(t, token, "loopback", "127.0.0.1")
	ttNumber := fmt.Sprintf("TT-SWEEP-%d", time.Now().UnixNano())

	for route, body := range map[string]map[string]any{
		"/diagnostic/submit": {"customer_id": "CUST-SWEEP", "tt_number": ttNumber},
		"/ping/run":          {},
	} {
		absent := packetpulsetest.Call(t, http.MethodPost, route, token, body)
		if absent.Status != http.StatusNotFound && absent.Status != http.StatusMethodNotAllowed {
			t.Errorf("%s = %d %s, want it absent (404 or 405)", route, absent.Status, absent.Raw)
		}
	}
	if filed := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/tt/"+ttNumber, token, nil); filed.Status != http.StatusNotFound {
		t.Errorf("a refused sweep opened the ticket: %d %s", filed.Status, filed.Raw)
	}
	runs := packetpulsetest.Call(t, http.MethodGet, "/ping/runlist", token, nil)
	if listed, ok := runs.Body["runs"].([]any); runs.Status != http.StatusOK || !ok || len(listed) != 0 {
		t.Errorf("runs after a refused sweep: %d %s, want none", runs.Status, runs.Raw)
	}
}
