// A browser UDP test filed on a ticket from Run diagnostic (October 2026),
// over the real stack: the split - where each missing packet went - is kept
// and read back, a test to PacketPulse's own reflector is filed under
// PacketPulse's name rather than the endpoint's, and an endpoint's reflector
// address survives the edits made to it.
package apicontract

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// udpRun is a 30 s UDP test of [endpoint]: 1,500 sent, 1,490 back on time,
// 4 lost on the way up, 2 on the way down and 4 late.
func udpRun(ttNumber, endpoint, reflector string) map[string]any {
	return map[string]any{
		"customer_id": "CUST-UDP", "tt_number": ttNumber,
		"sample_count": 1500, "timeout_ms": 1000,
		"dns_site_id": endpoint,
		"observations": []map[string]any{{
			"target_label": "what the device called it", "target_address": "198.51.100.77",
			"method":       "udp-webrtc",
			"packets_sent": 1500, "packets_recv": 1490, "packet_loss": 0.67,
			"min_rtt_ms": 41.2, "avg_rtt_ms": 54.7, "max_rtt_ms": 190.3,
			"jitter_ms": 8.6, "stddev_rtt_ms": 11.2, "mos_score": 4.37,
			"started_on":      time.Now().UTC().Format(time.RFC3339),
			"packets_lost_up": 4, "packets_lost_down": 2, "packets_late": 4,
			"packets_not_sent": 1, "jitter_up_ms": 9.4, "jitter_down_ms": 4.6,
			"longest_burst": 3, "reflector": reflector,
		}},
	}
}

// A test to PacketPulse's reflector measured the customer's access line. The
// ticket files it under PacketPulse's name and an address the server chose -
// never the endpoint's, never what the device claimed - with its split.
func TestABrowserUdpTestIsFiledWithWhereItsPacketsWent(t *testing.T) {
	packetpulsetest.RequireServer(t)

	token, _ := packetpulsetest.SignUpOrganisation(t, "udpfiled")
	endpoint := addEndpoint(t, token, "Chennai Core", "192.0.2.71")
	ttNumber := fmt.Sprintf("TT-UDP-%d", time.Now().UnixNano())

	attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", token,
		udpRun(ttNumber, endpoint, "packetpulse"))
	if attached.Status != http.StatusCreated {
		t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
	}
	requestId, _ := attached.Body["request"].(map[string]any)["request_id"].(string)

	detail := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId, token, nil)
	results := packetpulsetest.ListOf(t, detail, "results")
	if len(results) != 1 {
		t.Fatalf("read back %d results: %s", len(results), detail.Raw)
	}
	row := results[0]
	if row["site_name"] != "PacketPulse network" || row["dns_site_id"] != nil ||
		row["ip_address"] == "198.51.100.77" || row["ip_address"] == "192.0.2.71" {
		t.Errorf("filed as %v at %v (endpoint %v), want the PacketPulse network at the server's own address",
			row["site_name"], row["ip_address"], row["dns_site_id"])
	}
	for field, want := range map[string]float64{
		"packets_lost_up": 4, "packets_lost_down": 2, "packets_late": 4, "packets_not_sent": 1,
		"jitter_up_ms": 9.4, "jitter_down_ms": 4.6, "longest_burst": 3,
	} {
		if got, _ := row[field].(float64); got != want {
			t.Errorf("%s = %v, want %v", field, row[field], want)
		}
	}
	if row["method"] != "udp-webrtc" {
		t.Errorf("method = %v", row["method"])
	}
}

// The split is a claim only a UDP test can make, and one to an endpoint's
// reflector needs an endpoint that has one. Refused on its own field, and
// nothing is filed.
func TestAUdpSplitIsRefusedWhereItCannotBeTrue(t *testing.T) {
	packetpulsetest.RequireServer(t)

	token, _ := packetpulsetest.SignUpOrganisation(t, "udprefused")
	endpoint := addEndpoint(t, token, "No Reflector", "192.0.2.72")

	https := func(run map[string]any) map[string]any {
		run["observations"].([]map[string]any)[0]["method"] = "https-reach"
		return run
	}
	for name, build := range map[string]func(tt string) map[string]any{
		"an HTTPS test with a split":       func(tt string) map[string]any { return https(udpRun(tt, endpoint, "")) },
		"a reflector the endpoint has not": func(tt string) map[string]any { return udpRun(tt, endpoint, "endpoint") },
		"a reflector that is neither":      func(tt string) map[string]any { return udpRun(tt, endpoint, "nearest") },
		"an endpoint's reflector, no endpoint": func(tt string) map[string]any {
			run := udpRun(tt, endpoint, "endpoint")
			delete(run, "dns_site_id")
			return run
		},
	} {
		t.Run(name, func(t *testing.T) {
			ttNumber := fmt.Sprintf("TT-UDPNO-%d", time.Now().UnixNano())
			refused := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", token, build(ttNumber))
			if refused.Status != http.StatusUnprocessableEntity || refused.ErrorCode() != "validation_failed" {
				t.Fatalf("status = %d, want 422: %s", refused.Status, refused.Raw)
			}
			if byTt := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/tt/"+ttNumber, token, nil); byTt.Status != http.StatusNotFound {
				t.Errorf("the refused test was filed: %d", byTt.Status)
			}
		})
	}
}

// An endpoint's reflector address is kept across edits - including an edit
// from an app that predates it, which sends every other field and not this
// one - and a UDP test to it is then filed under the endpoint.
func TestAnEndpointsReflectorAddressSurvivesItsEdits(t *testing.T) {
	packetpulsetest.RequireServer(t)

	token, _ := packetpulsetest.SignUpOrganisation(t, "udpreflector")
	added := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", token, map[string]any{
		"site_name": "Pune Edge", "ip_address": "192.0.2.73", "is_enabled": true,
		"latitude": 18.5204, "longitude": 73.8567,
		"reflector_address": "Reflector.Pune.Example:50001",
	})
	if added.Status != http.StatusCreated || added.Body["reflector_address"] != "reflector.pune.example:50001" {
		t.Fatalf("add = %d %s", added.Status, added.Raw)
	}
	endpoint, _ := added.Body["dns_site_id"].(string)

	renamed := packetpulsetest.Call(t, http.MethodPut, "/dnssite/"+endpoint, token, map[string]any{
		"site_name": "Pune Edge 2", "ip_address": "192.0.2.73", "is_enabled": true,
		"latitude": 18.5204, "longitude": 73.8567,
	})
	if renamed.Status != http.StatusOK || renamed.Body["reflector_address"] != "reflector.pune.example:50001" {
		t.Errorf("an edit without the field = %d %s, want the reflector kept", renamed.Status, renamed.Raw)
	}

	ttNumber := fmt.Sprintf("TT-UDPEP-%d", time.Now().UnixNano())
	attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", token,
		udpRun(ttNumber, endpoint, "endpoint"))
	if attached.Status != http.StatusCreated {
		t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
	}
	requestId, _ := attached.Body["request"].(map[string]any)["request_id"].(string)
	row := packetpulsetest.ListOf(t, packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId, token, nil), "results")[0]
	if row["site_name"] != "Pune Edge 2" || row["dns_site_id"] != endpoint {
		t.Errorf("filed as %v (endpoint %v), want Pune Edge 2", row["site_name"], row["dns_site_id"])
	}

	cleared := packetpulsetest.Call(t, http.MethodPut, "/dnssite/"+endpoint, token, map[string]any{
		"site_name": "Pune Edge 2", "ip_address": "192.0.2.73", "is_enabled": true, "reflector_address": "",
		"latitude": 18.5204, "longitude": 73.8567,
	})
	if cleared.Status != http.StatusOK || cleared.Body["reflector_address"] != "" {
		t.Errorf("clearing it = %d %s", cleared.Status, cleared.Raw)
	}
}

// The relay and the key are signed-in routes: anonymous, each is refused
// before anything is relayed or issued.
func TestTheUdpRoutesRefuseAnAnonymousCaller(t *testing.T) {
	packetpulsetest.RequireServer(t)

	for _, path := range []string{
		"/diagnostic/endpoint/0b5e9a4c-7d2f-4e1a-8c3b-6f5a4d3c2b1a/webrtc-offer",
		"/diagnostic/endpoint/0b5e9a4c-7d2f-4e1a-8c3b-6f5a4d3c2b1a/reflector-key",
	} {
		refused := packetpulsetest.Call(t, http.MethodPost, path, "", map[string]any{"sdp": "v=0", "type": "offer"})
		if refused.Status != http.StatusUnauthorized {
			t.Errorf("%s anonymous = %d, want 401: %s", path, refused.Status, refused.Raw)
		}
	}
}

// A UDP test reached its reflector over IPv6, by the address the browser
// connected to: it is filed under IPv6 with its reflector and the endpoint's
// resolved IP, and a report counts it as an IPv6 row (customer report, 22
// October 2026: device tests had no family, so IPv6 was missing, and the
// resolved IP was blank). An endpoint set to show its reflector only keeps
// its address from an engineer, and not from whoever manages endpoints.
func TestAUdpTestOverIpv6IsFiledAndReportedAsIpv6(t *testing.T) {
	packetpulsetest.RequireServer(t)

	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "udpipv6")
	added := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", ownerToken, map[string]any{
		"site_name": "Andheri Core", "ip_address": "192.0.2.74", "is_enabled": true, "region": "Mumbai",
		"latitude": 19.1136, "longitude": 72.8697,
		"reflector_address": "reflector.mumbai.example:50001", "report_visibility": "reflector_only",
	})
	if added.Status != http.StatusCreated || added.Body["report_visibility"] != "reflector_only" {
		t.Fatalf("add = %d %s", added.Status, added.Raw)
	}
	endpoint, _ := added.Body["dns_site_id"].(string)
	engineerToken := engineerIn(t, ownerToken)

	run := udpRun(fmt.Sprintf("TT-UDP6-%d", time.Now().UnixNano()), endpoint, "endpoint")
	run["observations"].([]map[string]any)[0]["reflector_ip"] = "2001:db8::50"
	attached := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", engineerToken, run)
	if attached.Status != http.StatusCreated {
		t.Fatalf("attach returned %d: %s", attached.Status, attached.Raw)
	}
	requestId, _ := attached.Body["request"].(map[string]any)["request_id"].(string)

	asOwner := packetpulsetest.ListOf(t, packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId, ownerToken, nil), "results")[0]
	if asOwner["ip_address"] != "192.0.2.74" || asOwner["resolved_ip"] != "192.0.2.74" ||
		asOwner["reflector_address"] != "reflector.mumbai.example:50001" || asOwner["reflector_ip"] != "2001:db8::50" ||
		asOwner["ip_version"] != 6.0 {
		t.Errorf("read by the owner = %v", asOwner)
	}
	asEngineer := packetpulsetest.ListOf(t, packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId, engineerToken, nil), "results")[0]
	if asEngineer["ip_address"] != "" || asEngineer["resolved_ip"] != "" ||
		asEngineer["reflector_address"] != "reflector.mumbai.example:50001" || asEngineer["reflector_ip"] != "2001:db8::50" {
		t.Errorf("read by an engineer = %v, want the reflector alone", asEngineer)
	}

	// Today, in the organisation's zone: India's, as no other was chosen.
	kolkata, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	report := packetpulsetest.Call(t, http.MethodGet,
		"/report/network?period=day&date="+time.Now().In(kolkata).Format("2006-01-02"), ownerToken, nil)
	sites := packetpulsetest.ListOf(t, report, "sites")
	if len(sites) != 1 || sites[0]["city"] != "Mumbai" {
		t.Fatalf("sites = %v", sites)
	}
	endpoints, _ := sites[0]["endpoints"].([]any)
	if len(endpoints) != 1 {
		t.Fatalf("endpoint rows = %v", endpoints)
	}
	row, _ := endpoints[0].(map[string]any)
	if row["endpoint_name"] != "Andheri Core" || row["ip_version"] != 6.0 || row["method_label"] != "UDP (reflector) IPv6" {
		t.Errorf("endpoint row = %v", row)
	}
	if coverage, _ := report.Body["coverage"].(map[string]any); coverage["over_ipv6"] != 1.0 {
		t.Errorf("coverage = %v", coverage)
	}
}

// attachThroughProxy attaches a test as it arrives through PacketPulse's own
// nginx: the suite connects from loopback, a trusted proxy, carrying the
// X-Forwarded-For nginx writes - whatever the device sent, then the address
// nginx saw.
func attachThroughProxy(t *testing.T, token, forwardedFor string, run map[string]any) (int, map[string]any) {
	t.Helper()
	encoded, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, packetpulsetest.BaseURL()+"/api/v1/diagnostic/clientobservation", bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Forwarded-For", forwardedFor)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	raw, _ := io.ReadAll(response.Body)
	body := map[string]any{}
	_ = json.Unmarshal(raw, &body)
	return response.StatusCode, body
}

// The customer's report of 26 October 2026: "Resolved IP is taking the
// reflector IP instead of User IP". An endpoint whose address is its own
// reflector had nothing else on its ticket. A device test is now filed with
// the user IP it came from - the address nginx saw, not one the device put
// first - and the ticket and its CSV carry it. A device that names its own
// user IP is refused.
func TestADeviceTestIsFiledWithTheUserIpItCameFrom(t *testing.T) {
	packetpulsetest.RequireServer(t)

	token, _ := packetpulsetest.SignUpOrganisation(t, "userip")
	added := packetpulsetest.Call(t, http.MethodPost, "/dnssite/add", token, map[string]any{
		"site_name": "TATA-MUMBAI", "ip_address": "192.0.2.80", "is_enabled": true, "region": "Mumbai",
		"latitude": 19.076, "longitude": 72.8777, "reflector_address": "192.0.2.80",
	})
	if added.Status != http.StatusCreated {
		t.Fatalf("add = %d %s", added.Status, added.Raw)
	}
	endpoint, _ := added.Body["dns_site_id"].(string)
	run := udpRun(fmt.Sprintf("TT-USERIP-%d", time.Now().UnixNano()), endpoint, "endpoint")
	run["observations"].([]map[string]any)[0]["reflector_ip"] = "192.0.2.80"
	// The device's IPv6, as its own lookup found it: the server is reached
	// over IPv4 and never sees it.
	run["egress_ipv6"] = "2401:4900:1c5a:8f2e::1"

	status, attached := attachThroughProxy(t, token, "198.51.100.66, 203.0.113.9", run)
	if status != http.StatusCreated {
		t.Fatalf("attach returned %d: %v", status, attached)
	}
	requestId, _ := attached["request"].(map[string]any)["request_id"].(string)

	detail := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId, token, nil)
	request, _ := detail.Body["request"].(map[string]any)
	if request["user_ip"] != "203.0.113.9" || request["user_ipv6"] != "2401:4900:1c5a:8f2e::1" {
		t.Errorf("user_ip = %v, user_ipv6 = %v; want the address nginx saw and the device's IPv6",
			request["user_ip"], request["user_ipv6"])
	}
	row := packetpulsetest.ListOf(t, detail, "results")[0]
	if row["resolved_ip"] != "192.0.2.80" || row["reflector_ip"] != "192.0.2.80" {
		t.Errorf("resolved %v, reflector %v: the endpoint's addresses should be as before", row["resolved_ip"], row["reflector_ip"])
	}

	ticketCsv := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId+"/report.csv", token, nil)
	records, err := csv.NewReader(strings.NewReader(string(ticketCsv.Raw))).ReadAll()
	if err != nil || len(records) != 2 {
		t.Fatalf("ticket csv = %d %s", ticketCsv.Status, ticketCsv.Raw)
	}
	column := map[string]int{}
	for index, name := range records[0] {
		column[name] = index
	}
	if at, ok := column["user_ip"]; !ok || records[1][at] != "203.0.113.9" {
		t.Errorf("ticket csv = %v, want user_ip 203.0.113.9", records)
	}
	if at, ok := column["user_ipv6"]; !ok || records[1][at] != "2401:4900:1c5a:8f2e::1" {
		t.Errorf("ticket csv = %v, want user_ipv6 2401:4900:1c5a:8f2e::1", records)
	}

	// An IPv6 that is not a public one is refused on its field.
	private := udpRun(fmt.Sprintf("TT-USERIP-ULA-%d", time.Now().UnixNano()), endpoint, "endpoint")
	private["observations"].([]map[string]any)[0]["reflector_ip"] = "192.0.2.80"
	private["egress_ipv6"] = "fd12:3456::1"
	if refused := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", token, private); refused.Status != http.StatusUnprocessableEntity {
		t.Errorf("a private IPv6 answered %d, want 422: %s", refused.Status, refused.Raw)
	}

	run["tt_number"], run["user_ip"] = fmt.Sprintf("TT-USERIP-SAID-%d", time.Now().UnixNano()), "198.51.100.66"
	if said := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/clientobservation", token, run); said.Status != http.StatusBadRequest {
		t.Errorf("a body naming its own user IP answered %d, want 400", said.Status)
	}
}
