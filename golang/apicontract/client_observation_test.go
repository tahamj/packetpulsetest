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
	"testing"
	"time"

	pingletest "github.com/tahamj/pingletest"
)

// oneRun is a real browser measurement: the numbers below were produced by
// pingle-probe.js against Cloudflare, plus a target that never answered.
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
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "clientobs")
	ttNumber := fmt.Sprintf("TT-CLIENT-%d", time.Now().UnixNano())

	attached := pingletest.Call(t, http.MethodPost, "/diagnostic/clientobservation",
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

	results := pingletest.ListOf(t, attached, "results")
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
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "clientobs2")
	ttNumber := fmt.Sprintf("TT-CLIENT-%d", time.Now().UnixNano())

	attached := pingletest.Call(t, http.MethodPost, "/diagnostic/clientobservation",
		ownerToken, oneRun(ttNumber))
	if attached.Status != http.StatusCreated {
		t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
	}
	request, _ := attached.Body["request"].(map[string]any)
	requestId, _ := request["request_id"].(string)

	byTt := pingletest.Call(t, http.MethodGet, "/diagnostic/tt/"+ttNumber, ownerToken, nil)
	if byTt.Status != http.StatusOK {
		t.Fatalf("lookup by ticket returned %d: %s", byTt.Status, byTt.Raw)
	}

	report := pingletest.Call(t, http.MethodGet,
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
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "clientobs3")

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

			response := pingletest.Call(t, http.MethodPost,
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
	pingletest.RequireServer(t)

	unassignedToken, _, _, _ := pingletest.SignUpUnassigned(t, "clientobs4")

	response := pingletest.Call(t, http.MethodPost, "/diagnostic/clientobservation",
		unassignedToken, oneRun("TT-UNASSIGNED"))
	if response.Status == http.StatusCreated {
		t.Fatal("an account in the holding organisation attached a diagnostic")
	}
	if response.Status < 400 {
		t.Errorf("status = %d, want a refusal: %s", response.Status, response.Raw)
	}
}

func TestClientObservationRefusesAnAnonymousCaller(t *testing.T) {
	pingletest.RequireServer(t)

	response := pingletest.Call(t, http.MethodPost, "/diagnostic/clientobservation",
		"", oneRun("TT-ANON"))
	if response.Status != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401: %s", response.Status, response.Raw)
	}
}
