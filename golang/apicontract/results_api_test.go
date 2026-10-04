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
	"testing"
	"time"

	pingletest "github.com/tahamj/pingletest"
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
	pingletest.RequireServer(t)

	ownerToken, _ := pingletest.SignUpOrganisation(t, "contract")

	pingletest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"site_name": "loopback", "ip_address": "127.0.0.1",
	})

	ttNumber := fmt.Sprintf("TT-CONTRACT-%d", time.Now().UnixNano())
	submit := pingletest.Call(t, http.MethodPost, "/diagnostic/submit", ownerToken, map[string]any{
		"customer_id": "CUST-CONTRACT", "tt_number": ttNumber,
		"packet_count": 1, "timeout_ms": 900,
	})
	if submit.Status != http.StatusCreated {
		t.Skipf("could not run a diagnostic: %d %s", submit.Status, submit.Raw)
	}

	issued := pingletest.Call(t, http.MethodPost, "/apikey/add", ownerToken, map[string]any{
		"label": "contract test",
	})
	if issued.Status != http.StatusCreated {
		t.Skipf("could not issue an API key: %d %s", issued.Status, issued.Raw)
	}
	apiKey := issued.String("api_key")
	if apiKey == "" {
		t.Fatal("the API key must be returned in full exactly once, at creation")
	}

	response := pingletest.Call(t, http.MethodGet, "/result/bytt/"+ttNumber, apiKey, nil)
	if response.Status != http.StatusOK {
		t.Fatalf("Results API: %d %s", response.Status, response.Raw)
	}

	for _, field := range contractFields {
		if _, present := response.Body[field]; !present {
			t.Errorf("the Results API no longer returns %q, which an integration may rely on", field)
		}
	}

	rows := pingletest.ListOf(t, response, "results")
	if len(rows) == 0 {
		t.Fatal("the Results API returned no per-site results")
	}
	for _, field := range resultFields {
		if _, present := rows[0][field]; !present {
			t.Errorf("each result no longer carries %q", field)
		}
	}
}

func TestResultsApiRefusesTheWrongCredential(t *testing.T) {
	pingletest.RequireServer(t)

	sessionToken, _ := pingletest.SignUpOrganisation(t, "contractauth")

	for name, token := range map[string]string{
		"no credential":         "",
		"a human session token": sessionToken,
		"a malformed API key":   "pingle_development_000000000000_nope",
	} {
		response := pingletest.Call(t, http.MethodGet, "/result/bytt/anything", token, nil)
		if response.Status != http.StatusUnauthorized {
			t.Errorf("%s: expected 401, got %d", name, response.Status)
		}
	}
}
