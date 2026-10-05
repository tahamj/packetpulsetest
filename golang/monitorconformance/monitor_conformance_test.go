// Package monitorconformance asserts the alerting rules through the API.
//
// The service tests prove these against the database; this suite proves them
// against the running stack, through the same routes the client uses, because
// what a NOC relies on is the deployed behaviour: a breach is announced once,
// a flapping circuit is damped, planned work is silent, and a recovery is
// announced once.
//
// Schedules are created DISABLED and driven with "run now", so the server's
// own minute ticker never adds a run in the middle of a test.
package monitorconformance

import (
	"net/http"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

const (
	// Unroutable (TEST-NET-1): every probe times out.
	failing = "192.0.2.1"
	// Public and answering everywhere this suite runs.
	healthy = "1.1.1.1"
)

type monitoredOrg struct {
	token    string
	site     string
	schedule string
}

// setUp gives a fresh organisation one failing site, a default target set
// with the given damping, a channel, and a disabled alerting schedule.
func setUp(t *testing.T, prefix string, thresholdRuns int) monitoredOrg {
	t.Helper()
	packetpulsetest.RequireServer(t)
	token, _ := packetpulsetest.SignUpOrganisation(t, prefix)

	site := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", token, map[string]any{
		"site_name": "Branch", "ip_address": failing, "is_enabled": true,
	})
	if site.Status != http.StatusCreated && site.Status != http.StatusOK {
		t.Fatalf("add site: %d %s", site.Status, site.Raw)
	}

	// Generous thresholds: only reachability decides, never internet latency.
	policy := packetpulsetest.Call(t, http.MethodPost, "/monitor/sla/add", token, map[string]any{
		"policy_name": "Reachability", "max_latency_ms": 2000, "max_jitter_ms": 1000,
		"max_loss_pct": 50, "min_mos_score": 0, "is_default": true,
		"breach_threshold_runs": thresholdRuns, "alert_cooldown_minutes": 60,
	})
	if policy.Status != http.StatusCreated && policy.Status != http.StatusOK {
		t.Fatalf("add policy: %d %s", policy.Status, policy.Raw)
	}

	channel := packetpulsetest.Call(t, http.MethodPost, "/monitor/channel/add", token, map[string]any{
		"channel_name": "NOC mailbox", "channel_type": "email", "target": "noc@operator.test",
	})
	if channel.Status != http.StatusCreated && channel.Status != http.StatusOK {
		t.Fatalf("add channel: %d %s", channel.Status, channel.Raw)
	}

	schedule := packetpulsetest.Call(t, http.MethodPost, "/monitor/schedule/add", token, map[string]any{
		"schedule_name": "Branch watch", "interval_minutes": 60,
		"dns_site_ids": []string{site.String("dns_site_id")}, "customer_id": "CUST-MON",
		"tt_number_prefix": "MON", "packet_count": 2, "alert_on_failure": true, "is_enabled": false,
	})
	if schedule.Status != http.StatusCreated && schedule.Status != http.StatusOK {
		t.Fatalf("add schedule: %d %s", schedule.Status, schedule.Raw)
	}
	return monitoredOrg{token: token, site: site.String("dns_site_id"), schedule: schedule.String("schedule_id")}
}

// run points the site at an address and runs the schedule once.
func (m monitoredOrg) run(t *testing.T, address string) packetpulsetest.Response {
	t.Helper()
	update := packetpulsetest.Call(t, http.MethodPut, "/dnssite/"+m.site, m.token, map[string]any{
		"site_name": "Branch", "ip_address": address, "is_enabled": true,
	})
	if update.Status != http.StatusOK {
		t.Fatalf("point site at %s: %d %s", address, update.Status, update.Raw)
	}
	run := packetpulsetest.Call(t, http.MethodPost, "/monitor/schedule/"+m.schedule+"/run", m.token, nil)
	if run.Status != http.StatusOK {
		t.Fatalf("run schedule: %d %s", run.Status, run.Raw)
	}
	return run
}

func (m monitoredOrg) alerts(t *testing.T) []map[string]any {
	t.Helper()
	response := packetpulsetest.Call(t, http.MethodGet, "/monitor/alert/list", m.token, nil)
	if response.Status != http.StatusOK {
		t.Fatalf("list alerts: %d %s", response.Status, response.Raw)
	}
	return packetpulsetest.ListOf(t, response, "alerts")
}

func severities(alerts []map[string]any) []string {
	out := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		severity, _ := alert["severity"].(string)
		out = append(out, severity)
	}
	return out
}

// A breach is announced once, however many failing runs follow it inside the
// cooldown, and its recovery is announced once.
func TestABreachIsAnnouncedOnceAndItsRecoveryOnce(t *testing.T) {
	m := setUp(t, "monbrk", 1)

	m.run(t, failing)
	m.run(t, failing)
	if got := severities(m.alerts(t)); len(got) != 1 || got[0] != "critical" {
		t.Fatalf("after two failing runs: alerts %v, want exactly one critical", got)
	}

	m.run(t, healthy)
	m.run(t, healthy)
	got := severities(m.alerts(t))
	if len(got) != 2 {
		t.Fatalf("after recovery: alerts %v, want the outage and one recovery", got)
	}
	recoveries := 0
	for _, severity := range got {
		if severity == "info" {
			recoveries++
		}
	}
	if recoveries != 1 {
		t.Errorf("alerts %v, want exactly one recovery", got)
	}
}

// With a threshold of two, one failing run is a blip and announces nothing;
// the second consecutive one is an outage.
func TestDampingWaitsForConsecutiveBreaches(t *testing.T) {
	m := setUp(t, "mondmp", 2)

	m.run(t, failing)
	if got := m.alerts(t); len(got) != 0 {
		t.Fatalf("one failing run alerted %v, want nothing below the threshold", severities(got))
	}
	m.run(t, failing)
	if got := severities(m.alerts(t)); len(got) != 1 || got[0] != "critical" {
		t.Fatalf("two failing runs: alerts %v, want exactly one critical", got)
	}
}

// A run inside declared planned work grades the site "excluded" and alerts
// nobody - the outage is the work, not a fault.
func TestPlannedWorkIsExcludedAndSilent(t *testing.T) {
	m := setUp(t, "monwin", 1)
	now := time.Now().UTC()
	window := packetpulsetest.Call(t, http.MethodPost, "/monitor/maintenance/add", m.token, map[string]any{
		"window_name": "Line card swap", "scope": "all",
		"starts_on": now.Add(-time.Hour).Format(time.RFC3339), "ends_on": now.Add(time.Hour).Format(time.RFC3339),
		"recurrence": "none", "timezone": "UTC", "is_enabled": true,
	})
	if window.Status != http.StatusCreated && window.Status != http.StatusOK {
		t.Fatalf("add window: %d %s", window.Status, window.Raw)
	}

	run := m.run(t, failing)
	results, _ := run.Body["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results = %v, want one", results)
	}
	if status := results[0].(map[string]any)["sla_status"]; status != "excluded" {
		t.Errorf("sla_status = %v, want excluded", status)
	}
	if got := m.alerts(t); len(got) != 0 {
		t.Errorf("planned work alerted %v, want nothing", severities(got))
	}
}

// A measurement from a customer's own device cannot be graded against the
// server's targets: it grades "unknown" and counts no breach, however bad.
func TestAClientObservationIsNeverABreach(t *testing.T) {
	packetpulsetest.RequireServer(t)
	token, _ := packetpulsetest.SignUpOrganisation(t, "monobs")

	response := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", token, map[string]any{
		"customer_id": "CUST-OBS", "tt_number": "TT-OBS-1", "sample_count": 10, "timeout_ms": 2000,
		"observations": []map[string]any{{
			"method": "https-dns", "target_label": "Cloudflare", "target_address": "1.1.1.1",
			"packets_sent": 10, "packets_recv": 2, "packets_lost": 8, "packet_loss": 80, "avg_rtt_ms": 950,
		}},
	})
	if response.Status != http.StatusOK && response.Status != http.StatusCreated {
		t.Fatalf("attach observation: %d %s", response.Status, response.Raw)
	}
	request, _ := response.Body["request"].(map[string]any)
	if breaches, _ := request["breach_count"].(float64); breaches != 0 {
		t.Errorf("breach_count = %v, want 0", breaches)
	}
	results, _ := response.Body["results"].([]any)
	if len(results) != 1 || results[0].(map[string]any)["sla_status"] != "unknown" {
		t.Errorf("results = %v, want one graded unknown", results)
	}
}
