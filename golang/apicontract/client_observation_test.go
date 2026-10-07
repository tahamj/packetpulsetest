// The opt-in attach path: a measurement the customer's own device took,
// filed against a trouble ticket.
//
// This is tested end to end rather than in a unit because the feature's whole
// claim is that the run never reaches the server unless somebody chooses to
// send it - and that the thing which arrives is labelled honestly once it
// does. Both are properties of the stack, not of a function.
package apicontract

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// oneRun is a real browser measurement: the numbers below were produced by
// packetpulse-probe.js against Cloudflare, plus a target that never answered.
func oneRun(ttNumber string) map[string]any {
	return map[string]any{
		"customer_id":  "CUST-CLIENT",
		"tt_number":    ttNumber,
		"notes":        "Measured in the customer's browser.",
		"sample_count": 12,
		"timeout_ms":   2000,
		"observations": []map[string]any{
			{
				"target_label": "Cloudflare DNS", "target_address": "1.1.1.1",
				"method":       "https-dns",
				"packets_sent": 12, "packets_recv": 12, "packet_loss": 0,
				"min_rtt_ms": 34.8, "avg_rtt_ms": 49.8, "max_rtt_ms": 60.4,
				"jitter_ms": 7.0, "stddev_rtt_ms": 9.6, "mos_score": 4.39,
				"started_on": time.Now().UTC().Format(time.RFC3339),
			},
			{
				"target_label": "Branch endpoint", "target_address": "10.255.255.1",
				"method":       "https-reach",
				"packets_sent": 12, "packets_recv": 0, "packet_loss": 100,
				"error_message": "timed out after 2000 ms",
				"started_on":    time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
}

func TestClientObservationIsRecordedAsClientSourced(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "clientobs")
	ttNumber := fmt.Sprintf("TT-CLIENT-%d", time.Now().UnixNano())

	attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation",
		ownerToken, oneRun(ttNumber))
	if attached.Status != http.StatusCreated {
		t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
	}

	request, ok := attached.Body["request"].(map[string]any)
	if !ok {
		t.Fatalf("no request in the response: %s", attached.Raw)
	}

	// The source is the field that lets an evidence file hold both vantage
	// points. Recorded as 'manual' it would be indistinguishable from a server
	// sweep, and the two are not comparable.
	if request["source"] != "client" {
		t.Errorf("source = %v, want \"client\"", request["source"])
	}
	if request["status"] != "completed" {
		t.Errorf("status = %v, want completed", request["status"])
	}
	// No SLA grading, so no breaches invented from an HTTPS round-trip.
	if breaches, _ := request["breach_count"].(float64); breaches != 0 {
		t.Errorf("breach_count = %v, want 0", breaches)
	}
	if success, _ := request["success_count"].(float64); success != 1 {
		t.Errorf("success_count = %v, want 1 - the dead target must not count", success)
	}
	// No speed test with this run: none recorded, not a zero.
	if _, has := request["download_mbps"]; has {
		t.Errorf("download_mbps = %v on a run with no speed test", request["download_mbps"])
	}

	results := packetpulsetest.ListOf(t, attached, "results")
	if len(results) != 2 {
		t.Fatalf("stored %d results, want 2", len(results))
	}
	for _, result := range results {
		method, _ := result["method"].(string)
		if method != "https-dns" && method != "https-reach" {
			t.Errorf("method = %q, want an https method", method)
		}
		// 'unknown' is the honest value: this server applied no SLA to a
		// measurement it did not take.
		if result["sla_status"] != "unknown" {
			t.Errorf("sla_status = %v, want unknown", result["sla_status"])
		}
	}
}

// Once attached it is ordinary evidence: it appears in the history, opens as a
// detail, and exports as a PDF, because it shares one results model with a
// server sweep rather than living in a parallel one.
func TestAttachedClientRunBehavesLikeAnyOtherDiagnostic(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "clientobs2")
	ttNumber := fmt.Sprintf("TT-CLIENT-%d", time.Now().UnixNano())

	attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation",
		ownerToken, oneRun(ttNumber))
	if attached.Status != http.StatusCreated {
		t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
	}
	request, _ := attached.Body["request"].(map[string]any)
	requestId, _ := request["request_id"].(string)

	byTt := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/tt/"+ttNumber, ownerToken, nil)
	if byTt.Status != http.StatusOK {
		t.Fatalf("lookup by ticket returned %d: %s", byTt.Status, byTt.Raw)
	}

	report := packetpulsetest.Call(t, http.MethodGet,
		"/diagnostic/"+requestId+"/report.pdf", ownerToken, nil)
	if report.Status != http.StatusOK {
		t.Fatalf("PDF export returned %d: %s", report.Status, report.Raw)
	}
	// The report must describe what was actually counted. A browser sent no
	// packet, and this document is the one in which that gets challenged.
	if bytes := report.Raw; len(bytes) < 1000 {
		t.Errorf("the PDF is %d bytes, which is too small to be a report", len(bytes))
	}
}

func TestClientObservationRefusesWhatItCannotRecordHonestly(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "clientobs3")

	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"no ticket", func(body map[string]any) { body["tt_number"] = "" }},
		{"no customer", func(body map[string]any) { body["customer_id"] = "" }},
		{"nothing measured", func(body map[string]any) {
			body["observations"] = []map[string]any{}
		}},
		// 'icmp' from a browser is a lie the server cannot detect, but an
		// unrecognised method it can and must: the method decides whether a
		// loss figure counts packets or unanswered queries.
		{"unrecognised method", func(body map[string]any) {
			body["observations"] = []map[string]any{{
				"target_label": "x", "target_address": "1.1.1.1", "method": "ping",
				"packets_sent": 4, "packets_recv": 4,
			}}
		}},
		{"unnamed target", func(body map[string]any) {
			body["observations"] = []map[string]any{{
				"target_label": "  ", "target_address": "1.1.1.1", "method": "https-dns",
				"packets_sent": 4, "packets_recv": 4,
			}}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := oneRun(fmt.Sprintf("TT-BAD-%d", time.Now().UnixNano()))
			tc.mutate(body)

			response := packetpulsetest.Call(t, http.MethodPost,
				"/diagnostic/clientobservation", ownerToken, body)
			// 422, as every other validated form on this API answers: the
			// request parsed, it just described something that cannot be
			// recorded honestly.
			if response.Status != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422: %s", response.Status, response.Raw)
			}
			if response.ErrorCode() != "validation_failed" {
				t.Errorf("code = %q, want validation_failed", response.ErrorCode())
			}
			// The message has to name the field, or the form cannot show it
			// where the person is looking.
			details, _ := response.Body["error"].(map[string]any)
			if fields, ok := details["details"].(map[string]any); !ok || len(fields) == 0 {
				t.Errorf("no per-field detail; got %s", response.Raw)
			}
		})
	}
}

// Attaching writes a diagnostic against a ticket, which is the licensed act,
// so it is gated exactly like a submission. Running the test is not gated at
// all and does not come through here.
func TestClientObservationRequiresAnOrganisation(t *testing.T) {
	packetpulsetest.RequireServer(t)

	// The console operator is signed in but belongs to no organisation.
	noOrganisation := packetpulsetest.SuperUserToken(t)

	response := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation",
		noOrganisation, oneRun("TT-NO-ORGANISATION"))
	if response.Status == http.StatusCreated {
		t.Fatal("an account with no organisation attached a diagnostic")
	}
	if response.Status < 400 {
		t.Errorf("status = %d, want a refusal: %s", response.Status, response.Raw)
	}
}

func TestClientObservationRefusesAnAnonymousCaller(t *testing.T) {
	packetpulsetest.RequireServer(t)

	response := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation",
		"", oneRun("TT-ANON"))
	if response.Status != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401: %s", response.Status, response.Raw)
	}
}

// A speed measured with the run is filed with it and read back with the ticket.
func TestALineSpeedIsFiledWithTheDevicesRun(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "clientspeed")
	ttNumber := fmt.Sprintf("TT-SPEED-%d", time.Now().UnixNano())
	run := oneRun(ttNumber)
	run["download_mbps"], run["upload_mbps"] = 87.4, 12.25

	attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", ownerToken, run)
	if attached.Status != http.StatusCreated {
		t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
	}
	requestId, _ := attached.Body["request"].(map[string]any)["request_id"].(string)
	detail := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId, ownerToken, nil)
	request, _ := detail.Body["request"].(map[string]any)
	if request["download_mbps"] != 87.4 || request["upload_mbps"] != 12.25 {
		t.Errorf("read back %v / %v, want 87.4 / 12.25", request["download_mbps"], request["upload_mbps"])
	}

	run["tt_number"], run["upload_mbps"] = ttNumber+"-X", -3
	if refused := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", ownerToken, run); refused.Status != http.StatusUnprocessableEntity {
		t.Errorf("a negative speed answered %d, want 422", refused.Status)
	}
}

// addEndpoint adds an enabled endpoint to the caller's organisation and
// returns its id.
func addEndpoint(t *testing.T, token, name, address string) string {
	t.Helper()
	site := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", token, map[string]any{
		"site_name": name, "ip_address": address, "is_enabled": true,
	})
	if site.Status != http.StatusCreated {
		t.Fatalf("add endpoint: %d %s", site.Status, site.Raw)
	}
	id, _ := site.Body["dns_site_id"].(string)
	if id == "" {
		t.Fatalf("no dns_site_id in %s", site.Raw)
	}
	return id
}

// Run diagnostic tests one of the organisation's endpoints from the device
// (October 2026). The run is filed under that endpoint and read back by its
// name - on the ticket, in History and in the CSV - not as the internal
// device target.
func TestADeviceRunOfAnEndpointIsFiledUnderItsName(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "clientendpoint")
	endpoint := addEndpoint(t, ownerToken, "Chennai Core", "192.0.2.61")
	ttNumber := fmt.Sprintf("TT-ENDPOINT-%d", time.Now().UnixNano())
	run := oneRun(ttNumber)
	run["dns_site_id"] = endpoint
	run["observations"] = run["observations"].([]map[string]any)[:1]

	attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", ownerToken, run)
	if attached.Status != http.StatusCreated {
		t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
	}
	requestId, _ := attached.Body["request"].(map[string]any)["request_id"].(string)

	detail := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId, ownerToken, nil)
	results := packetpulsetest.ListOf(t, detail, "results")
	if len(results) != 1 {
		t.Fatalf("read back %d results, want 1: %s", len(results), detail.Raw)
	}
	if results[0]["site_name"] != "Chennai Core" || results[0]["ip_address"] != "192.0.2.61" ||
		results[0]["dns_site_id"] != endpoint {
		t.Errorf("result = %v, want Chennai Core at 192.0.2.61, endpoint %s", results[0], endpoint)
	}

	csv := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId+"/report.csv", ownerToken, nil)
	if csv.Status != http.StatusOK || !strings.Contains(string(csv.Raw), "Chennai Core") {
		t.Errorf("the ticket's CSV (%d) does not name the endpoint:\n%s", csv.Status, csv.Raw)
	}
}

// An endpoint a device run is filed against must be the caller's own: another
// organisation's is refused exactly like one that does not exist, and
// nothing is filed.
func TestADeviceRunCannotBeFiledAgainstAnotherOrganisationsEndpoint(t *testing.T) {
	packetpulsetest.RequireServer(t)

	theirToken, _ := packetpulsetest.SignUpOrganisation(t, "clientendpointtheirs")
	theirs := addEndpoint(t, theirToken, "Their Core", "192.0.2.62")
	ourToken, _ := packetpulsetest.SignUpOrganisation(t, "clientendpointours")

	for name, dnsSiteId := range map[string]string{
		"another organisation's": theirs,
		"nobody's":               "0b5e9a4c-7d2f-4e1a-8c3b-6f5a4d3c2b1a",
		"not an id":              "chennai-core",
	} {
		t.Run(name, func(t *testing.T) {
			ttNumber := fmt.Sprintf("TT-THEIRS-%d", time.Now().UnixNano())
			run := oneRun(ttNumber)
			run["dns_site_id"] = dnsSiteId

			refused := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", ourToken, run)
			if refused.Status != http.StatusUnprocessableEntity || refused.ErrorCode() != "validation_failed" {
				t.Fatalf("status = %d, want 422 validation_failed: %s", refused.Status, refused.Raw)
			}
			details, _ := refused.Body["error"].(map[string]any)["details"].(map[string]any)
			if _, named := details["dns_site_id"]; !named {
				t.Errorf("no message on dns_site_id: %s", refused.Raw)
			}
			if strings.Contains(string(refused.Raw), "Their Core") {
				t.Errorf("the refusal names the other organisation's endpoint: %s", refused.Raw)
			}
			byTt := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/tt/"+ttNumber, ourToken, nil)
			if byTt.Status != http.StatusNotFound {
				t.Errorf("the refused run was filed: %d %s", byTt.Status, byTt.Raw)
			}
		})
	}
}
