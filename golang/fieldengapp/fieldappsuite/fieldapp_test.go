// Package fieldappsuite drives the field app's sign-in end to end against a
// running PacketPulse (migration 0043): "Field Eng App", the mock of the
// operator's own app, asks the partner address for a link over mutual TLS -
// through fieldengapp.FrontHandler standing in for nginx's partner site, as
// on a developer's machine - and the link opens a session on Run tests with
// the ticket filled in, for the engineer the app named.
//
// The server must take launches: PARTNER_LAUNCH_APP_URL set, as .env.example
// and CI set it. Live, the same mock runs against the real nginx
// (scripts/deploy/PacketPulseFieldEngAppDeploy.sh).
package fieldappsuite

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tahamj/packetpulsetest"
	"github.com/tahamj/packetpulsetest/fieldengapp"
)

// world is one organisation's PacketPulse, its registered app and the app.
type world struct {
	organisation packetpulsetest.Organisation
	certificate  fieldengapp.KeyPair
	certId       string
	partner      *fieldengapp.Partner
	app          *httptest.Server
	front        *httptest.Server
	engineers    []fieldengapp.Engineer
	browser      *http.Client
}

// unique makes the app's engineers this run's own: their addresses are
// unique across PacketPulse.
func unique(t *testing.T) []fieldengapp.Engineer {
	t.Helper()
	stamp := time.Now().UnixNano()
	return []fieldengapp.Engineer{
		{Id: "FE-1001", Email: fmt.Sprintf("asha.rao-%d@field-eng-app.example", stamp), Name: "Asha Rao", Password: "field1001"},
		{Id: "FE-1002", Email: fmt.Sprintf("vikram.shah-%d@field-eng-app.example", stamp), Name: "Vikram Shah", Password: "field1002"},
	}
}

func roleId(t *testing.T, ownerToken, name string) string {
	t.Helper()
	roles := packetpulsetest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	for _, role := range packetpulsetest.ListOf(t, roles, "roles") {
		if role["role_name"] == name {
			return role["role_id"].(string)
		}
	}
	t.Fatalf("no %s role: %s", name, roles.Raw)
	return ""
}

// newWorld makes a licensed organisation whose owner registers a fresh
// certificate for the app, and runs the app and the partner site's
// stand-in in front of the server under test.
func newWorld(t *testing.T, seats int) *world {
	t.Helper()
	packetpulsetest.RequireServer(t)
	w := &world{organisation: packetpulsetest.NewOrganisation(t, "fieldapp", seats), engineers: unique(t)}

	var err error
	if w.certificate, err = fieldengapp.NewClientCertificate("Field Eng App suite", 30*24*time.Hour); err != nil {
		t.Fatal(err)
	}
	registered := packetpulsetest.Call(t, http.MethodPost, "/partner/certificate/add", w.organisation.OwnerToken, map[string]any{
		"label": "Field Eng App server", "certificate": string(w.certificate.CertificatePEM),
		"engineer_role_id": roleId(t, w.organisation.OwnerToken, "NOC Engineer"),
	})
	if registered.Status != http.StatusCreated {
		t.Fatalf("the owner registering the app's certificate: %d %s", registered.Status, registered.Raw)
	}
	w.certId = registered.String("certificate_id")
	if fingerprint, _ := fieldengapp.Fingerprint(w.certificate.CertificatePEM); registered.String("fingerprint_sha256") != fingerprint {
		t.Errorf("PacketPulse shows %q, the app's team computes %q: they must match by eye", registered.String("fingerprint_sha256"), fingerprint)
	}

	w.front = startFront(t)
	w.partner, w.app = w.startApp(t, w.certificate, "pp_none")
	jar, _ := cookiejar.New(nil)
	w.browser = &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return w
}

// startFront is nginx's partner site as fieldengapp stands in for it,
// passing launches on to the server under test.
func startFront(t *testing.T) *httptest.Server {
	t.Helper()
	upstream, err := url.Parse(packetpulsetest.BaseURL())
	if err != nil {
		t.Fatal(err)
	}
	pair, err := fieldengapp.NewServerCertificate()
	if err != nil {
		t.Fatal(err)
	}
	certificate, err := tls.X509KeyPair(pair.CertificatePEM, pair.KeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	front := httptest.NewUnstartedServer(fieldengapp.FrontHandler(upstream))
	front.TLS = fieldengapp.FrontTLS(certificate)
	front.StartTLS()
	t.Cleanup(front.Close)
	return front
}

func (w *world) startApp(t *testing.T, certificate fieldengapp.KeyPair, apiKey string) (*fieldengapp.Partner, *httptest.Server) {
	t.Helper()
	dir := t.TempDir()
	write := func(name string, content []byte) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	trust := w.front.Certificate()
	config := fieldengapp.Config{
		AppName: fieldengapp.DefaultAppName, PartnerURL: w.front.URL, PacketPulseURL: packetpulsetest.BaseURL(), ApiKey: apiKey,
		// Where phones reach the app, for the way back from Run tests.
		PublicURL:       "https://fieldapp.test",
		CertificateFile: write("app.crt", certificate.CertificatePEM), KeyFile: write("app.key", certificate.KeyPEM),
		TrustFile: write("front.crt", pemOf(trust)), Engineers: w.engineers, Tickets: fieldengapp.DefaultTickets(),
		// One secret for every instance, as a deployment sets it, so an
		// engineer signed in to the app stays signed in across instances.
		SessionSecret: []byte("fieldappsuite session secret"),
	}
	partner, err := fieldengapp.NewPartner(config)
	if err != nil {
		t.Fatal(err)
	}
	app := httptest.NewServer(fieldengapp.NewApp(config, partner))
	t.Cleanup(app.Close)
	return partner, app
}

func pemOf(certificate *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
}

// tap is the engineer, signed in to the app, tapping the button on a
// ticket: the answer is where the phone is sent, or the app's page.
func (w *world) tap(t *testing.T, app *httptest.Server, engineer fieldengapp.Engineer, ttNumber string) *http.Response {
	t.Helper()
	signedIn, err := w.browser.PostForm(app.URL+"/signin", url.Values{"username": {engineer.Id}, "password": {engineer.Password}})
	if err != nil || signedIn.StatusCode != http.StatusSeeOther {
		t.Fatalf("signing in to the app: %v %v", err, signedIn)
	}
	_ = signedIn.Body.Close()
	response, err := w.browser.PostForm(app.URL+"/launch", url.Values{"tt": {ttNumber}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = response.Body.Close() })
	return response
}

// codeIn is the one-time code in the address the phone was sent to.
func codeIn(t *testing.T, location string) string {
	t.Helper()
	address, err := url.Parse(location)
	if err != nil || !strings.HasSuffix(address.Path, "/app/") || !strings.HasPrefix(address.Fragment, "launch=") {
		t.Fatalf("the phone was sent to %q, want the web app with a launch code in its fragment", location)
	}
	return strings.TrimPrefix(address.Fragment, "launch=")
}

func redeem(t *testing.T, code string) packetpulsetest.Response {
	t.Helper()
	return packetpulsetest.Call(t, http.MethodPost, "/user/signin/launch", "", map[string]any{"code": code})
}

func body(t *testing.T, response *http.Response) string {
	t.Helper()
	raw, _ := io.ReadAll(response.Body)
	return string(raw)
}

// The whole of it, as the customer will see it: the engineer taps the
// button in their app; PacketPulse makes their account the first time,
// named as the app names them; the link opens Run tests signed in as them
// with the ticket filled in, once; the test they run is theirs; and the app
// reads the result back naming them.
func TestTheFieldAppOpensRunTestsSignedInWithTheTicket(t *testing.T) {
	w := newWorld(t, 10)
	asha := w.engineers[0]

	response := w.tap(t, w.app, asha, "TT-2026-01458")
	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("tapping the button: %d %s", response.StatusCode, body(t, response))
	}
	code := codeIn(t, response.Header.Get("Location"))

	session := redeem(t, code)
	if session.Status != http.StatusOK {
		t.Fatalf("opening the link: %d %s", session.Status, session.Raw)
	}
	launch, _ := session.Body["launch"].(map[string]any)
	if session.String("organisation_id") != w.organisation.Id || session.String("staff_name") != "Asha Rao" ||
		session.String("email") != asha.Email || session.String("role_name") != "NOC Engineer" ||
		session.Body["sign_in_from_app"] != true || launch["tt_number"] != "TT-2026-01458" || launch["customer_id"] != "CUST-40217" ||
		launch["return_url"] != "https://fieldapp.test/ticket?tt=TT-2026-01458&back=1" {
		t.Errorf("the session = %s, want Asha's, from the app, with the ticket", session.Raw)
	}
	token := session.String("token")
	if me := packetpulsetest.Call(t, http.MethodGet, "/user/me", token, nil); me.Status != http.StatusOK || me.Body["sign_in_from_app"] != true {
		t.Errorf("/user/me = %d %s", me.Status, me.Raw)
	}

	if again := redeem(t, code); again.Status != http.StatusUnauthorized || again.ErrorCode() != "launch_link_invalid" {
		t.Errorf("the same link twice: %d %s, want launch_link_invalid", again.Status, again.Raw)
	}

	// The engineer presses Start: the test is filed as theirs.
	filed := packetpulsetest.FileDeviceTest(t, token, "CUST-40217", "TT-2026-01458", nil)
	if filed.Status != http.StatusCreated {
		t.Fatalf("filing the engineer's test: %d %s", filed.Status, filed.Raw)
	}

	// And the app reads it back, naming the engineer it sent.
	issued := packetpulsetest.Call(t, http.MethodPost, "/apikey/add", w.organisation.OwnerToken, map[string]any{"label": "Field Eng App results"})
	if issued.Status != http.StatusCreated {
		t.Fatalf("issuing the app's read-only key: %d %s", issued.Status, issued.Raw)
	}
	_, app := w.startApp(t, w.certificate, issued.String("api_key"))
	result, err := w.browser.Get(app.URL + "/result?tt=TT-2026-01458")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = result.Body.Close() }()
	if page := body(t, result); !strings.Contains(page, "<dt>Tested by</dt><dd>Asha Rao</dd>") {
		t.Errorf("the app's result page does not name the engineer:\n%s", page)
	}

	// A second tap is the same person: no second account, no second seat.
	again := w.tap(t, w.app, asha, "TT-2026-00744")
	second := redeem(t, codeIn(t, again.Header.Get("Location")))
	if second.String("user_id") != session.String("user_id") {
		t.Errorf("a second tap opened %s, want the same person %s", second.String("user_id"), session.String("user_id"))
	}
}

// Someone the app signs in has no password: the website's sign-in refuses
// them as it refuses a wrong password.
func TestAnEngineerTheAppSignsInCannotSignInOnTheWebsite(t *testing.T) {
	w := newWorld(t, 10)
	response := w.tap(t, w.app, w.engineers[0], "TT-2026-01458")
	codeIn(t, response.Header.Get("Location"))

	for _, password := range []string{"!", "field1001", "PacketPulseTest2026x"} {
		signIn := packetpulsetest.Call(t, http.MethodPost, "/user/signin", "", map[string]any{
			"email": w.engineers[0].Email, "password": password})
		if signIn.Status != http.StatusUnauthorized {
			t.Errorf("website sign-in with %q: %d %s, want refused", password, signIn.Status, signIn.Raw)
		}
	}
}

// Without a registered certificate in use, the app gets nothing: none shown,
// one never registered, and one the owner has revoked.
func TestOnlyTheRegisteredCertificateGetsALink(t *testing.T) {
	w := newWorld(t, 10)
	asha := w.engineers[0]

	// No certificate: the partner site passes on an empty header.
	plain := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool(w.front.Certificate())}}}
	request, _ := http.NewRequest(http.MethodPost, w.front.URL+fieldengapp.LaunchPath, strings.NewReader(`{"engineer_id":"FE-1001","email":"x@y.example","name":"X"}`))
	request.Header.Set("Content-Type", "application/json")
	if response, err := plain.Do(request); err != nil || response.StatusCode != http.StatusUnauthorized {
		t.Errorf("no certificate: %v %v, want 401", err, response)
	}

	stranger, _ := fieldengapp.NewClientCertificate("Not registered", time.Hour)
	_, strangersApp := w.startApp(t, stranger, "")
	if response := w.tap(t, strangersApp, asha, "TT-2026-01458"); response.StatusCode != http.StatusBadGateway ||
		!strings.Contains(body(t, response), "did not present a certificate registered in PacketPulse") {
		t.Errorf("a certificate never registered: %d", response.StatusCode)
	}

	revoked := packetpulsetest.Call(t, http.MethodPost, "/partner/certificate/"+w.certId+"/revoke", w.organisation.OwnerToken, nil)
	if revoked.Status != http.StatusOK || revoked.Body["revoked_on"] == nil {
		t.Fatalf("revoking: %d %s", revoked.Status, revoked.Raw)
	}
	if response := w.tap(t, w.app, asha, "TT-2026-01458"); response.StatusCode != http.StatusBadGateway {
		t.Errorf("after the owner revoked it: %d, want refused", response.StatusCode)
	}
}

// A link the app asked for stops opening anything once its certificate is
// revoked, even inside its minute.
func TestRevokingTheCertificateSpoilsItsUnusedLinks(t *testing.T) {
	w := newWorld(t, 10)
	response := w.tap(t, w.app, w.engineers[0], "TT-2026-01458")
	code := codeIn(t, response.Header.Get("Location"))
	if revoked := packetpulsetest.Call(t, http.MethodPost, "/partner/certificate/"+w.certId+"/revoke", w.organisation.OwnerToken, nil); revoked.Status != http.StatusOK {
		t.Fatalf("revoking: %d %s", revoked.Status, revoked.Raw)
	}
	if opened := redeem(t, code); opened.Status != http.StatusUnauthorized {
		t.Errorf("a link of a revoked certificate: %d %s, want refused", opened.Status, opened.Raw)
	}
}

// One organisation's app never reaches another's people, and one owner never
// reads or revokes another's certificates.
func TestOneOrganisationsAppNeverReachesAnother(t *testing.T) {
	ours, theirs := newWorld(t, 10), newWorld(t, 10)

	// Their engineer exists; our app names them by their address.
	response := theirs.tap(t, theirs.app, theirs.engineers[0], "TT-2026-01458")
	codeIn(t, response.Header.Get("Location"))
	borrowed := ours.engineers[0]
	borrowed.Email = theirs.engineers[0].Email
	ours.engineers[0] = borrowed
	_, ourApp := ours.startApp(t, ours.certificate, "")
	refused := ours.tap(t, ourApp, borrowed, "TT-2026-01458")
	if page := body(t, refused); refused.StatusCode != http.StatusBadGateway || !strings.Contains(page, "409 conflict") {
		t.Errorf("our app naming their engineer's address: %d, want 409", refused.StatusCode)
	}

	list := packetpulsetest.Call(t, http.MethodGet, "/partner/certificate/list", ours.organisation.OwnerToken, nil)
	for _, certificate := range packetpulsetest.ListOf(t, list, "certificates") {
		if certificate["certificate_id"] == theirs.certId {
			t.Error("our owner's list holds their certificate")
		}
	}
	if revoked := packetpulsetest.Call(t, http.MethodPost, "/partner/certificate/"+theirs.certId+"/revoke", ours.organisation.OwnerToken, nil); revoked.Status != http.StatusNotFound {
		t.Errorf("our owner revoking theirs: %d %s, want 404", revoked.Status, revoked.Raw)
	}
}

// The register is kept with the Results API keys, by whoever may issue them
// (apikey_manage; until 18 November 2026 the owner alone): an administrator
// lists the certificates and revokes one, while a NOC Engineer, who may not
// issue keys, is refused both.
func TestTheRegisterIsKeptByWhoeverMayIssueResultsAPIKeys(t *testing.T) {
	w := newWorld(t, 10)
	administrator := w.colleague(t, "Administrator")
	engineer := w.colleague(t, "NOC Engineer")

	for _, call := range [][2]string{
		{http.MethodGet, "/partner/certificate/list"},
		{http.MethodPost, "/partner/certificate/" + w.certId + "/revoke"},
	} {
		if refused := packetpulsetest.Call(t, call[0], call[1], engineer, nil); refused.Status != http.StatusForbidden {
			t.Errorf("a NOC Engineer %s %s: %d, want 403", call[0], call[1], refused.Status)
		}
	}

	listed := packetpulsetest.Call(t, http.MethodGet, "/partner/certificate/list", administrator, nil)
	if listed.Status != http.StatusOK || !strings.Contains(string(listed.Raw), w.certId) {
		t.Fatalf("an administrator listing: %d %s, want the organisation's certificate", listed.Status, listed.Raw)
	}
	if revoked := packetpulsetest.Call(t, http.MethodPost, "/partner/certificate/"+w.certId+"/revoke", administrator, nil); revoked.Status != http.StatusOK {
		t.Errorf("an administrator revoking: %d %s, want 200", revoked.Status, revoked.Raw)
	}
	if refused := w.tap(t, w.app, w.engineers[0], "TT-2026-01458"); refused.StatusCode != http.StatusBadGateway {
		t.Errorf("a launch after the administrator revoked: %d, want refused", refused.StatusCode)
	}
}

// colleague signs in a new person of the organisation holding roleName.
func (w *world) colleague(t *testing.T, roleName string) string {
	t.Helper()
	email := fmt.Sprintf("fieldapp-%d@packetpulsetest.local", time.Now().UnixNano())
	added := packetpulsetest.Call(t, http.MethodPost, "/user/add", w.organisation.OwnerToken, map[string]any{
		"email": email, "password": "ColleaguePass2026", "full_name": "A " + roleName,
		"role_id": roleId(t, w.organisation.OwnerToken, roleName),
	})
	if added.Status != http.StatusCreated {
		t.Fatalf("adding a %s: %d %s", roleName, added.Status, added.Raw)
	}
	return packetpulsetest.SignIn(t, email, "ColleaguePass2026")
}

// A new engineer takes a place on the licence; with none left, the app is
// told so, and nobody is made.
func TestANewEngineerNeedsAPlaceOnTheLicence(t *testing.T) {
	w := newWorld(t, 2)
	if first := w.tap(t, w.app, w.engineers[0], "TT-2026-01458"); first.StatusCode != http.StatusSeeOther {
		t.Fatalf("the first engineer, into the last place: %d %s", first.StatusCode, body(t, first))
	}
	full := w.tap(t, w.app, w.engineers[1], "TT-2026-01458")
	if page := body(t, full); full.StatusCode != http.StatusBadGateway || !strings.Contains(page, "seat_limit_reached") {
		t.Errorf("past the licence: %d %s", full.StatusCode, page)
	}
}

func pool(certificate *x509.Certificate) *x509.CertPool {
	roots := x509.NewCertPool()
	roots.AddCert(certificate)
	return roots
}

// Whoever may issue API keys but not manage people - as the live demo's
// Demo Administrator, with every right but people, roles, ACL and the
// directory - registers a certificate only with a role whose rights they
// hold: offered no Administrator, and refused one, so they cannot make
// administrators through the app (18 November 2026).
func TestACertificatesEngineersHoldNoMoreThanItsRegistrant(t *testing.T) {
	w := newWorld(t, 10)
	// Every right PacketPulse has but those over people, as the live role.
	grants := map[string]bool{}
	catalogue := packetpulsetest.Call(t, http.MethodGet, "/user/capability/list", w.organisation.OwnerToken, nil)
	for _, capability := range packetpulsetest.ListOf(t, catalogue, "capabilities") {
		switch name := capability["name"].(string); name {
		case "user_manage", "staff_manage", "acl_manage", "ldap_manage":
		default:
			grants[name] = true
		}
	}
	if !grants["apikey_manage"] || !grants["diagnostic_run"] {
		t.Fatalf("the capability catalogue was not read: %v", grants)
	}
	role := packetpulsetest.Call(t, http.MethodPost, "/staff/role/add", w.organisation.OwnerToken, map[string]any{
		"role_name": fmt.Sprintf("Demo Administrator %d", time.Now().UnixNano()), "role_description": "No people rights", "grants": grants,
	})
	if role.Status != http.StatusCreated {
		t.Fatalf("making the limited role: %d %s", role.Status, role.Raw)
	}
	limited := w.colleague(t, role.String("role_name"))

	offered := map[string]bool{}
	listed := packetpulsetest.Call(t, http.MethodGet, "/partner/role/list", limited, nil)
	for _, choice := range packetpulsetest.ListOf(t, listed, "roles") {
		offered[choice["role_name"].(string)] = true
	}
	if listed.Status != http.StatusOK || offered["Administrator"] || !offered["NOC Engineer"] {
		t.Errorf("offered %v (%d), want NOC Engineer and never Administrator", offered, listed.Status)
	}

	certificate, err := fieldengapp.NewClientCertificate("Escalation", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	add := func(roleName string) packetpulsetest.Response {
		return packetpulsetest.Call(t, http.MethodPost, "/partner/certificate/add", limited, map[string]any{
			"label": "By the limited", "certificate": string(certificate.CertificatePEM),
			"engineer_role_id": roleId(t, w.organisation.OwnerToken, roleName),
		})
	}
	if refused := add("Administrator"); refused.Status != http.StatusUnprocessableEntity || !strings.Contains(string(refused.Raw), "engineer_role_id") {
		t.Errorf("the limited giving Administrator: %d %s, want refused on the role", refused.Status, refused.Raw)
	}
	if registered := add("NOC Engineer"); registered.Status != http.StatusCreated {
		t.Errorf("the limited giving NOC Engineer: %d %s", registered.Status, registered.Raw)
	}
}
