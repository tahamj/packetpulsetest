// Package gradingconformance asserts, through the API, how a ticket's results
// are graded against the organisation's targets.
//
// The service tests prove the rules against the database; this suite proves
// them against the running stack, through the routes the app uses: a
// device's UDP test is graded like a sweep and its breach counted on the
// ticket, an HTTPS round-trip is never graded, a ticket counts its results by
// grade wherever it is read, and automatic monitoring - removed in October
// 2026 - answers nothing.
//
// It replaces the monitor suite, whose alerting, damping and planned-work
// tests went with the feature they tested.
package gradingconformance

import (
	"net/http"
	"testing"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// udpObservation is a 30-second UDP test to the PacketPulse network: the
// run the 13 October 2026 review saw filed "UNKNOWN", 5.4% lost at 190 ms.
func udpObservation(ttNumber string) map[string]any {
	return map[string]any{
		"customer_id": "CUST-GRADE", "tt_number": ttNumber, "sample_count": 1500, "timeout_ms": 1000,
		"observations": []map[string]any{{
			"method": "udp-webrtc", "reflector": "packetpulse",
			"target_label": "PacketPulse network", "target_address": "203.0.113.20",
			"packets_sent": 1500, "packets_recv": 1419, "packet_loss": 5.4,
			"avg_rtt_ms": 189.8, "jitter_ms": 9.3, "mos_score": 3.1,
			"packets_lost_up": 4, "packets_lost_down": 73, "packets_late": 4,
		}},
	}
}

// grades is a ticket's results counted by grade: OK, Degraded, Breached.
func grades(request map[string]any) [3]float64 {
	count := func(name string) float64 { value, _ := request[name].(float64); return value }
	return [3]float64{count("ok_count"), count("degraded_count"), count("breach_count")}
}

func file(t *testing.T, token string, body map[string]any) (map[string]any, []any) {
	t.Helper()
	response := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", token, body)
	if response.Status != http.StatusOK && response.Status != http.StatusCreated {
		t.Fatalf("attach observation: %d %s", response.Status, response.Raw)
	}
	request, _ := response.Body["request"].(map[string]any)
	results, _ := response.Body["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results = %v, want one", results)
	}
	return request, results
}

// A device's UDP test sends real packets, the way a call does, so it is
// graded against the organisation's targets and its breach counted on the
// ticket. It was filed "unknown", with no breach, until October 2026.
func TestADevicesUdpTestIsGradedAndItsBreachCounted(t *testing.T) {
	packetpulsetest.RequireServer(t)
	token, _ := packetpulsetest.SignUpOrganisation(t, "grdudp")

	request, results := file(t, token, udpObservation("TT-GRADE-1"))
	if status := results[0].(map[string]any)["sla_status"]; status != "breached" {
		t.Errorf("sla_status = %v, want breached", status)
	}
	if got := grades(request); got != [3]float64{0, 0, 1} {
		t.Errorf("OK, degraded, breached = %v, want 0, 0, 1", got)
	}
}

// The organisation's own default targets are what a test to the PacketPulse
// network is held to - not the built-in ones.
func TestAUdpTestIsHeldToTheOrganisationsOwnTargets(t *testing.T) {
	packetpulsetest.RequireServer(t)
	token, _ := packetpulsetest.SignUpOrganisation(t, "grdown")

	policy := packetpulsetest.Call(t, http.MethodPost, "/monitor/sla/add", token, map[string]any{
		"policy_name": "Lenient", "max_latency_ms": 400, "max_jitter_ms": 50,
		"max_loss_pct": 10, "min_mos_score": 2, "is_default": true,
	})
	if policy.Status != http.StatusCreated && policy.Status != http.StatusOK {
		t.Fatalf("add policy: %d %s", policy.Status, policy.Raw)
	}

	request, results := file(t, token, udpObservation("TT-GRADE-2"))
	if status := results[0].(map[string]any)["sla_status"]; status != "ok" {
		t.Errorf("sla_status = %v, want ok against the organisation's lenient targets", status)
	}
	if got := grades(request); got != [3]float64{1, 0, 0} {
		t.Errorf("OK, degraded, breached = %v, want 1, 0, 0", got)
	}
}

// A ticket counts its results by grade, and is read with them wherever it
// is listed, so History and the Dashboard lead it with its worst: a test
// graded Degraded read OK on its row while the network report called it
// Degraded (review of 30 October 2026). The dashboard counts the days' tests
// by grade for whoever reads everyone's, as the owner does.
func TestATicketIsCountedByGradeWhereverItIsRead(t *testing.T) {
	packetpulsetest.RequireServer(t)
	token, _ := packetpulsetest.SignUpOrganisation(t, "grddeg")

	// 189.8 ms is past 80% of 220 ms, short of it; every other figure well
	// inside its target: Degraded.
	policy := packetpulsetest.Call(t, http.MethodPost, "/monitor/sla/add", token, map[string]any{
		"policy_name": "Close", "max_latency_ms": 220, "max_jitter_ms": 50,
		"max_loss_pct": 10, "min_mos_score": 2, "is_default": true,
	})
	if policy.Status != http.StatusCreated && policy.Status != http.StatusOK {
		t.Fatalf("add policy: %d %s", policy.Status, policy.Raw)
	}

	request, results := file(t, token, udpObservation("TT-GRADE-DEG"))
	if status := results[0].(map[string]any)["sla_status"]; status != "degraded" {
		t.Fatalf("sla_status = %v, want degraded", status)
	}
	if got := grades(request); got != [3]float64{0, 1, 0} {
		t.Errorf("filed: OK, degraded, breached = %v, want 0, 1, 0", got)
	}

	listed := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/list?tt_number=TT-GRADE-DEG", token, nil)
	rows, _ := listed.Body["diagnostics"].([]any)
	if listed.Status != http.StatusOK || len(rows) != 1 {
		t.Fatalf("list: %d %s", listed.Status, listed.Raw)
	}
	if got := grades(rows[0].(map[string]any)); got != [3]float64{0, 1, 0} {
		t.Errorf("listed: OK, degraded, breached = %v, want 0, 1, 0", got)
	}

	dashboard := packetpulsetest.Call(t, http.MethodGet, "/ping/dashboard", token, nil)
	counted, _ := dashboard.Body["recent_grades"].(map[string]any)
	if dashboard.Status != http.StatusOK || counted == nil {
		t.Fatalf("dashboard: %d %s, want the recent tests by grade", dashboard.Status, dashboard.Raw)
	}
	if counted["ok"] != 0.0 || counted["degraded"] != 1.0 || counted["breached"] != 0.0 || counted["ungraded"] != 0.0 {
		t.Errorf("recent_grades = %v, want the one degraded test", counted)
	}
	if _, kept := dashboard.Body["last_run"]; kept {
		t.Error("the dashboard still sends the latest run, which read as a sweep's")
	}
}

// An HTTPS round-trip from a customer's own device is a query, not a packet:
// TCP turns loss into delay. The targets are set for packets, so it is never
// graded, and never a breach, however bad.
func TestAnHttpsObservationIsNeverABreach(t *testing.T) {
	packetpulsetest.RequireServer(t)
	token, _ := packetpulsetest.SignUpOrganisation(t, "grdhttps")

	request, results := file(t, token, map[string]any{
		"customer_id": "CUST-OBS", "tt_number": "TT-OBS-1", "sample_count": 25, "timeout_ms": 2000,
		"observations": []map[string]any{{
			"method": "https-dns", "target_label": "Cloudflare", "target_address": "1.1.1.1",
			"packets_sent": 10, "packets_recv": 2, "packets_lost": 8, "packet_loss": 80, "avg_rtt_ms": 950,
		}},
	})
	// Counted under no grade at all: its ticket reads Not graded, never OK.
	if got := grades(request); got != [3]float64{0, 0, 0} {
		t.Errorf("OK, degraded, breached = %v, want none", got)
	}
	if status := results[0].(map[string]any)["sla_status"]; status != "unknown" {
		t.Errorf("sla_status = %v, want unknown", status)
	}
}

// Automatic monitoring was removed (review of 13 October 2026): none of its
// routes answers, so nothing can be scheduled, alerted or reported on.
func TestAutomaticMonitoringIsGone(t *testing.T) {
	packetpulsetest.RequireServer(t)
	token, _ := packetpulsetest.SignUpOrganisation(t, "grdgone")

	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/monitor/schedule/list"},
		{http.MethodPost, "/monitor/schedule/add"},
		{http.MethodGet, "/monitor/channel/list"},
		{http.MethodGet, "/monitor/alert/list"},
		{http.MethodGet, "/monitor/maintenance/list"},
		{http.MethodGet, "/monitor/slareport"},
		{http.MethodGet, "/monitor/trend"},
	} {
		response := packetpulsetest.Call(t, route.method, route.path, token, nil)
		if response.Status != http.StatusNotFound && response.Status != http.StatusMethodNotAllowed {
			t.Errorf("%s %s answered %d, want it gone", route.method, route.path, response.Status)
		}
	}
}
