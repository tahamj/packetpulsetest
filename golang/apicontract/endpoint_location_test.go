package apicontract

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/tahamj/packetpulsetest"
)

// onePixelPng is the smallest image there is.
const onePixelPng = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="

// engineerIn adds a NOC Engineer to the owner's organisation and signs them in.
func engineerIn(t *testing.T, ownerToken string) string {
	t.Helper()
	roles := packetpulsetest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	roleId := ""
	for _, role := range packetpulsetest.ListOf(t, roles, "roles") {
		if role["role_name"] == "NOC Engineer" {
			roleId, _ = role["role_id"].(string)
		}
	}
	if roleId == "" {
		t.Fatalf("no NOC Engineer role: %s", roles.Raw)
	}
	email := fmt.Sprintf("engineer-%d@packetpulsetest.local", time.Now().UnixNano())
	created := packetpulsetest.Call(t, http.MethodPost, "/user/add", ownerToken, map[string]any{
		"email": email, "password": "EngineerPass2026", "full_name": "Field engineer", "role_id": roleId,
	})
	if created.Status != http.StatusCreated {
		t.Fatalf("adding an engineer: %d %s", created.Status, created.Raw)
	}
	return packetpulsetest.SignIn(t, email, "EngineerPass2026")
}

func sweptSites(t *testing.T, response packetpulsetest.Response) []string {
	t.Helper()
	names := []string{}
	for _, result := range packetpulsetest.ListOf(t, response, "results") {
		name, _ := result["site_name"].(string)
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// The whole chain over HTTP, through the server as it is wired at boot: an
// administrator imports a spreadsheet of placed endpoints, an engineer's test
// sweeps the city nearest their device - whatever endpoints their request
// names - and the ticket, its CSV and its PDF say so.
func TestAnEngineersTestSweepsTheCityNearestTheirDevice(t *testing.T) {
	packetpulsetest.RequireServer(t)
	ownerToken, _ := packetpulsetest.SignUpOrganisation(t, "nearest")

	// Addresses the probe policy refuses at once: the test is about which
	// endpoints a sweep takes, and refused ones are still reported.
	imported := packetpulsetest.Call(t, http.MethodPost, "/dnssite/importfile", ownerToken, map[string]any{
		"text": "Name,Address,City,Latitude,Longitude\n" +
			"Bengaluru DNS 1,127.0.0.1,Bengaluru,12.9716,77.5946\n" +
			"Bengaluru DNS 2,127.0.0.2,bengaluru,,\n" +
			"Mumbai DNS 1,169.254.169.254,Mumbai,19.0760,72.8777\n",
	})
	if imported.Status != http.StatusOK || imported.Float("added") != 3 {
		t.Fatalf("import: %d %s", imported.Status, imported.Raw)
	}
	mumbaiId := ""
	for _, site := range packetpulsetest.ListOf(t,
		packetpulsetest.Call(t, http.MethodGet, "/dnssite/list", ownerToken, nil), "dns_sites") {
		if site["site_name"] == "Mumbai DNS 1" {
			mumbaiId, _ = site["dns_site_id"].(string)
		}
	}

	engineer := engineerIn(t, ownerToken)
	ttNumber := fmt.Sprintf("TT-NEAREST-%d", time.Now().UnixNano())
	submitted := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/submit", engineer, map[string]any{
		"customer_id": "CUST-NEAR", "tt_number": ttNumber, "packet_count": 1, "timeout_ms": 900,
		// An engineer's own choice is never read.
		"dns_site_ids": []string{mumbaiId},
		"location":     map[string]any{"status": "captured", "latitude": 12.9698, "longitude": 77.75, "accuracy_m": 20},
	})
	if submitted.Status != http.StatusCreated {
		t.Fatalf("submit: %d %s", submitted.Status, submitted.Raw)
	}
	if got := sweptSites(t, submitted); strings.Join(got, ",") != "Bengaluru DNS 1,Bengaluru DNS 2" {
		t.Errorf("swept %v, want Bengaluru's two and not the Mumbai endpoint the request named", got)
	}
	request, _ := submitted.Body["request"].(map[string]any)
	if request["endpoint_selection"] != "nearest" || request["selection_city"] != "Bengaluru" {
		t.Errorf("recorded %v / %v", request["endpoint_selection"], request["selection_city"])
	}
	location, _ := request["test_location"].(map[string]any)
	if location["status"] != "captured" || location["latitude"] != 12.9698 {
		t.Errorf("test location = %v", location)
	}
	requestId, _ := request["request_id"].(string)

	csv := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId+"/report.csv", engineer, nil)
	lines := strings.Split(strings.TrimSpace(string(csv.Raw)), "\n")
	if csv.Status != http.StatusOK || len(lines) < 2 || !strings.HasSuffix(strings.TrimSpace(lines[1]), ",12.969800,77.750000") {
		t.Errorf("the ticket's CSV: %d %q", csv.Status, lines)
	}

	// The owner's own logo is on the engineer's report.
	logo := packetpulsetest.Call(t, http.MethodPut, "/organisation/brand/logo", ownerToken, map[string]any{"image": onePixelPng})
	if logo.Status != http.StatusOK || logo.String("logo_png") == "" {
		t.Fatalf("saving the logo: %d %s", logo.Status, logo.Raw)
	}
	pdf := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId+"/report.pdf", engineer, nil)
	if pdf.Status != http.StatusOK || !bytes.Contains(pdf.Raw, []byte("/Subtype /Image")) {
		t.Errorf("the report: %d, logo embedded: %v", pdf.Status, bytes.Contains(pdf.Raw, []byte("/Subtype /Image")))
	}

	// With no position, the administrator's sweep takes every endpoint.
	everything := packetpulsetest.Call(t, http.MethodPost, "/diagnostic/submit", ownerToken, map[string]any{
		"customer_id": "CUST-NEAR", "tt_number": ttNumber + "-ALL", "packet_count": 1, "timeout_ms": 900,
	})
	if got := sweptSites(t, everything); len(got) != 3 {
		t.Errorf("with no position swept %v, want all three", got)
	}
	if all, _ := everything.Body["request"].(map[string]any); all["endpoint_selection"] != "all" {
		t.Errorf("recorded %v, want all", all["endpoint_selection"])
	}
}

// A brand is its organisation's own: another organisation reads none of it,
// and an image that is not one is refused.
func TestABrandIsItsOrganisationsOwn(t *testing.T) {
	packetpulsetest.RequireServer(t)
	ours, _ := packetpulsetest.SignUpOrganisation(t, "brand-a")
	theirs, _ := packetpulsetest.SignUpOrganisation(t, "brand-b")

	saved := packetpulsetest.Call(t, http.MethodPut, "/organisation/brand/icon", ours, map[string]any{"image": onePixelPng})
	if saved.Status != http.StatusOK || saved.String("icon_png") == "" {
		t.Fatalf("saving the icon: %d %s", saved.Status, saved.Raw)
	}
	if theirBrand := packetpulsetest.Call(t, http.MethodGet, "/organisation/brand", theirs, nil); theirBrand.Status != http.StatusOK ||
		theirBrand.Body["icon_png"] != nil || theirBrand.Body["logo_png"] != nil {
		t.Errorf("another organisation's brand: %d %s", theirBrand.Status, theirBrand.Raw)
	}

	svg := base64.StdEncoding.EncodeToString([]byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`))
	refused := packetpulsetest.Call(t, http.MethodPut, "/organisation/brand/logo", ours, map[string]any{"image": svg})
	if refused.Status != http.StatusUnprocessableEntity {
		t.Errorf("an SVG: %d %s", refused.Status, refused.Raw)
	}
	removed := packetpulsetest.Call(t, http.MethodDelete, "/organisation/brand/icon", ours, nil)
	if removed.Status != http.StatusOK || removed.Body["icon_png"] != nil {
		t.Errorf("removing the icon: %d %s", removed.Status, removed.Raw)
	}
}
