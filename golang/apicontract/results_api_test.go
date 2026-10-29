// Package apicontract pins the shape of the Results API.
//
// That endpoint is consumed by the CUSTOMER'S ticketing system, which nobody
// here can redeploy. Its field names are therefore a contract: this test fails
// if one is renamed or removed, which is the point - an internal rename should
// not be able to silently break someone else's integration.
package apicontract

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// contractFields are the top-level fields the Results API promises.
var contractFields = []string{
	"tt_number", "customer_id", "status", "site_count",
	"success_count", "failure_count", "started_on", "results",
}

// resultFields are the per-site fields it promises.
var resultFields = []string{
	"site_name", "ip_address", "ip_version", "report_only", "reachable",
	"packets_sent", "packets_recv", "packets_lost", "packet_loss",
}

func TestResultsApiShapeIsStable(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "contract")

	packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"latitude": 12.9716, "longitude": 77.5946, "site_name": "loopback", "ip_address": "127.0.0.1",
	})

	ttNumber := fmt.Sprintf("TT-CONTRACT-%d", time.Now().UnixNano())
	submit := packetpulsetest.FileDeviceTest(t, ownerToken, "CUST-CONTRACT", ttNumber, nil)
	if submit.Status != http.StatusCreated {
		t.Fatalf("could not file a test: %d %s", submit.Status, submit.Raw)
	}

	issued := packetpulsetest.Call(t, http.MethodPost, "/apikey/add", ownerToken, map[string]any{
		"label": "contract test",
	})
	if issued.Status != http.StatusCreated {
		t.Fatalf("could not issue an API key: %d %s", issued.Status, issued.Raw)
	}
	apiKey := issued.String("api_key")
	if apiKey == "" {
		t.Fatal("the API key must be returned in full exactly once, at creation")
	}

	response := packetpulsetest.Call(t, http.MethodGet, "/result/bytt/"+ttNumber, apiKey, nil)
	if response.Status != http.StatusOK {
		t.Fatalf("Results API: %d %s", response.Status, response.Raw)
	}

	for _, field := range contractFields {
		if _, present := response.Body[field]; !present {
			t.Errorf("the Results API no longer returns %q, which an integration may rely on", field)
		}
	}

	rows := packetpulsetest.ListOf(t, response, "results")
	if len(rows) == 0 {
		t.Fatal("the Results API returned no per-site results")
	}
	for _, field := range resultFields {
		if _, present := rows[0][field]; !present {
			t.Errorf("each result no longer carries %q", field)
		}
	}
	// Taken out on 28 October 2026 at the customer's request: it read as the
	// user's IP, and was the address the endpoint resolved to.
	if resolved, present := rows[0]["resolved_ip"]; present {
		t.Errorf("each result still carries resolved_ip (%v)", resolved)
	}
}

// The Results API reads a customer's tickets by Customer ID for a period
// (customer request, 28 October 2026): newest first, each in the shape
// /result/bytt gives one ticket, the last 30 days unless asked otherwise. Another
// customer's tickets are never among them, nor another organisation's.
func TestTheResultsApiReadsACustomersTicketsNewestFirst(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "bycustomer")
	packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"latitude": 12.9716, "longitude": 77.5946, "site_name": "loopback", "ip_address": "127.0.0.1",
	})
	stamp := time.Now().UnixNano()
	customerId := fmt.Sprintf("CUST-BYCUSTOMER-%d", stamp)
	submit := func(customer, ttNumber string) {
		t.Helper()
		submitted := packetpulsetest.FileDeviceTest(t, ownerToken, customer, ttNumber, nil)
		if submitted.Status != http.StatusCreated {
			t.Fatalf("could not file a test: %d %s", submitted.Status, submitted.Raw)
		}
	}
	older, newer := fmt.Sprintf("TT-BYCUSTOMER-%d-1", stamp), fmt.Sprintf("TT-BYCUSTOMER-%d-2", stamp)
	submit(customerId, older)
	submit(fmt.Sprintf("CUST-ELSE-%d", stamp), fmt.Sprintf("TT-ELSE-%d", stamp))
	submit(customerId, newer)
	apiKey := issueKey(t, ownerToken)

	read := func(key string, period url.Values) packetpulsetest.Response {
		t.Helper()
		path := "/result/bycustomer/" + url.PathEscape(customerId)
		if len(period) > 0 {
			path += "?" + period.Encode()
		}
		return packetpulsetest.Call(t, http.MethodGet, path, key, nil)
	}
	response := read(apiKey, nil)
	if response.Status != http.StatusOK {
		t.Fatalf("by customer: %d %s", response.Status, response.Raw)
	}
	if response.Body["customer_id"] != customerId || response.Body["count"] != 2.0 || response.Body["more"] != false {
		t.Errorf("by customer = %v", response.Body)
	}
	from, _ := time.Parse(time.RFC3339, fmt.Sprint(response.Body["from"]))
	to, _ := time.Parse(time.RFC3339, fmt.Sprint(response.Body["to"]))
	if days := to.Sub(from).Hours() / 24; days != 30 {
		t.Errorf("the period read unasked = %v to %v, want the last 30 days", response.Body["from"], response.Body["to"])
	}
	tickets := packetpulsetest.ListOf(t, response, "tickets")
	if len(tickets) != 2 || tickets[0]["tt_number"] != newer || tickets[1]["tt_number"] != older {
		t.Fatalf("tickets = %v, want %s then %s", tickets, newer, older)
	}
	for _, ticket := range tickets {
		for _, field := range contractFields {
			if _, present := ticket[field]; !present {
				t.Errorf("a ticket read by customer has no %q", field)
			}
		}
		results, _ := ticket["results"].([]any)
		if len(results) == 0 {
			t.Fatalf("ticket %v has no results", ticket["tt_number"])
		}
		row, _ := results[0].(map[string]any)
		for _, field := range resultFields {
			if _, present := row[field]; !present {
				t.Errorf("a result read by customer has no %q", field)
			}
		}
		if _, present := row["resolved_ip"]; present {
			t.Errorf("a result read by customer carries resolved_ip")
		}
	}

	// A period before both tickets were run reads none of them.
	earlier := read(apiKey, url.Values{
		"from": {time.Now().Add(-72 * time.Hour).UTC().Format(time.RFC3339)},
		"to":   {time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)},
	})
	if earlier.Status != http.StatusOK || earlier.Body["count"] != 0.0 || len(packetpulsetest.ListOf(t, earlier, "tickets")) != 0 {
		t.Errorf("an earlier period = %d %s, want none", earlier.Status, earlier.Raw)
	}
	// Another organisation's key reads none of them.
	strangerToken, _ := packetpulsetest.SignUpOrganisation(t, "bycustomerother")
	if stranger := read(issueKey(t, strangerToken), nil); stranger.Status != http.StatusOK || stranger.Body["count"] != 0.0 {
		t.Errorf("another organisation's read = %d %s, want none", stranger.Status, stranger.Raw)
	}
	// The count bounds a call, not the period (customer decision, 29 October
	// 2026): years are read, newest first, the newest 200 at most.
	years := read(apiKey, url.Values{"from": {"2020-01-01T00:00:00Z"}})
	if years.Status != http.StatusOK || years.Body["count"] != 2.0 {
		t.Errorf("a period of years = %d %s, want both tickets", years.Status, years.Raw)
	}
	// A time that cannot be read is 400, as the CSV export answers it; a
	// period that ends before it starts is 422.
	for name, refusal := range map[string]struct {
		period url.Values
		status int
	}{
		"a period that is not a time":   {url.Values{"from": {"yesterday"}}, http.StatusBadRequest},
		"a period ending as it started": {url.Values{"from": {"2026-10-01T00:00:00Z"}, "to": {"2026-10-01T00:00:00Z"}}, http.StatusUnprocessableEntity},
	} {
		if refused := read(apiKey, refusal.period); refused.Status != refusal.status {
			t.Errorf("%s: %d %s, want %d", name, refused.Status, refused.Raw, refusal.status)
		}
	}
}

// issueKey issues an API key for an organisation's owner.
func issueKey(t *testing.T, ownerToken string) string {
	t.Helper()
	issued := packetpulsetest.Call(t, http.MethodPost, "/apikey/add", ownerToken, map[string]any{"label": "contract test"})
	apiKey := issued.String("api_key")
	if issued.Status != http.StatusCreated || apiKey == "" {
		t.Fatalf("could not issue an API key: %d %s", issued.Status, issued.Raw)
	}
	return apiKey
}

func TestResultsApiRefusesTheWrongCredential(t *testing.T) {
	packetpulsetest.RequireServer(t)

	sessionToken, _ := packetpulsetest.SignUpOrganisation(t, "contractauth")

	for name, token := range map[string]string{
		"no credential":         "",
		"a human session token": sessionToken,
		"a malformed API key":   "packetpulse_development_000000000000_nope",
	} {
		for _, path := range []string{"/result/bytt/anything", "/result/bycustomer/anything"} {
			response := packetpulsetest.Call(t, http.MethodGet, path, token, nil)
			if response.Status != http.StatusUnauthorized {
				t.Errorf("%s on %s: expected 401, got %d", name, path, response.Status)
			}
		}
	}
}

// An API key reads results; it never starts a test. The SOW and the in-app
// help promise the customer that only a signed-in person starts one
// (29 October 2026), so a key presented to either route that starts a sweep
// is refused, and nothing is run or opened against the ticket.
func TestAnApiKeyCannotStartATest(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "contractstart")
	packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"latitude": 12.9716, "longitude": 77.5946, "site_name": "loopback", "ip_address": "127.0.0.1",
	})
	issued := packetpulsetest.Call(t, http.MethodPost, "/apikey/add", ownerToken, map[string]any{
		"label": "start test",
	})
	if issued.Status != http.StatusCreated {
		t.Fatalf("could not issue an API key: %d %s", issued.Status, issued.Raw)
	}
	apiKey := issued.String("api_key")

	ttNumber := fmt.Sprintf("TT-KEYSTART-%d", time.Now().UnixNano())
	for _, path := range []string{"/diagnostic/submit", "/ping/run"} {
		response := packetpulsetest.Call(t, http.MethodPost, path, apiKey, map[string]any{
			"customer_id": "CUST-KEYSTART", "tt_number": ttNumber,
		})
		// 401 where the server's sweep is switched on; absent (404 or 405)
		// where it is off, as it is by default. Never a test.
		if refused := response.Status == http.StatusUnauthorized || response.Status == http.StatusNotFound ||
			response.Status == http.StatusMethodNotAllowed; !refused {
			t.Errorf("an API key on %s: %d %s, want it refused", path, response.Status, response.Raw)
		}
	}

	if opened := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/tt/"+ttNumber, ownerToken, nil); opened.Status != http.StatusNotFound {
		t.Errorf("a refused start still opened a diagnostic against the ticket: %d %s", opened.Status, opened.Raw)
	}
	runs := packetpulsetest.Call(t, http.MethodGet, "/ping/runlist", ownerToken, nil)
	if _, listed := runs.Body["runs"].([]any); runs.Status != http.StatusOK || !listed {
		t.Fatalf("run list: %d %s, want the organisation's runs", runs.Status, runs.Raw)
	}
	if listed := packetpulsetest.ListOf(t, runs, "runs"); len(listed) != 0 {
		t.Errorf("a refused start still ran %d sweep(s)", len(listed))
	}
}
