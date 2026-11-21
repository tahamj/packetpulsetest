package fieldengapp

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
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
	"sync"
	"testing"
	"time"
)

// The mock on its own, against a fake PacketPulse: that its stand-in for
// nginx's partner site passes on exactly what nginx would, and that the app
// sends its signed-in engineer to the link PacketPulse gives - or shows
// PacketPulse's refusal as written. The real PacketPulse is ./fieldappsuite.

// upstream is a fake PacketPulse behind the front: it records what reached
// it and answers as the test says.
type upstream struct {
	mu       sync.Mutex
	header   string
	forwards string
	body     map[string]string
	calls    int
	status   int
	answer   string
}

func (u *upstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.calls++
	u.header = r.Header.Get(ClientCertificateHeader)
	u.forwards = r.Header.Get("X-Forwarded-For")
	u.body = map[string]string{}
	_ = json.NewDecoder(r.Body).Decode(&u.body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(u.status)
	_, _ = io.WriteString(w, u.answer)
}

type rig struct {
	upstream *upstream
	front    *httptest.Server
	app      *httptest.Server
	client   KeyPair
	browser  *http.Client
	config   Config
	// fieldApp is the app itself, whose clock a test may set before its
	// first request.
	fieldApp *App
}

// newRig is the app, its front and a fake PacketPulse, wired as on a
// developer's machine.
func newRig(t *testing.T, change func(*Config)) *rig {
	t.Helper()
	fake := &upstream{status: http.StatusCreated, answer: `{"url":"https://packetpulse.test/app/#launch=CODE","expires_in":60,"engineer":{"user_id":"u","name":"Asha Rao","created":true}}`}
	backend := httptest.NewServer(fake)
	t.Cleanup(backend.Close)
	target, _ := url.Parse(backend.URL)

	serverPair, err := NewServerCertificate()
	if err != nil {
		t.Fatal(err)
	}
	serverCert, err := tls.X509KeyPair(serverPair.CertificatePEM, serverPair.KeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	front := httptest.NewUnstartedServer(FrontHandler(target))
	front.TLS = FrontTLS(serverCert)
	front.StartTLS()
	t.Cleanup(front.Close)

	clientPair, err := NewClientCertificate("Field Eng App test", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	write := func(name string, content []byte) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	config := Config{
		PartnerURL: front.URL, PacketPulseURL: backend.URL, ApiKey: "pp_test_key",
		CertificateFile: write("app.crt", clientPair.CertificatePEM), KeyFile: write("app.key", clientPair.KeyPEM),
		TrustFile: write("front.crt", serverPair.CertificatePEM),
		Engineers: DefaultEngineers(), Tickets: DefaultTickets(),
	}
	if change != nil {
		change(&config)
	}
	partner, err := NewPartner(config)
	if err != nil {
		t.Fatal(err)
	}
	fieldApp := NewApp(config, partner)
	app := httptest.NewServer(fieldApp)
	t.Cleanup(app.Close)

	jar, _ := cookiejar.New(nil)
	browser := &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return &rig{upstream: fake, front: front, app: app, client: clientPair, browser: browser, config: config, fieldApp: fieldApp}
}

func (r *rig) get(t *testing.T, path string) (*http.Response, string) {
	t.Helper()
	response, err := r.browser.Get(r.app.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	body, _ := io.ReadAll(response.Body)
	return response, string(body)
}

func (r *rig) post(t *testing.T, path string, form url.Values) (*http.Response, string) {
	t.Helper()
	response, err := r.browser.PostForm(r.app.URL+path, form)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	body, _ := io.ReadAll(response.Body)
	return response, string(body)
}

func (r *rig) signIn(t *testing.T) {
	t.Helper()
	response, _ := r.post(t, "/signin", url.Values{"username": {"FE-1001"}, "password": {"field1001"}})
	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("sign-in answered %d", response.StatusCode)
	}
}

// The front passes on the certificate the caller showed, escaped as nginx
// escapes $ssl_client_escaped_cert, replacing any such header the caller
// sent, and adds the caller to X-Forwarded-For. It serves nothing else.
func TestTheFrontPassesOnTheCertificateAsNginxDoes(t *testing.T) {
	r := newRig(t, nil)
	certificate, _ := tls.X509KeyPair(r.client.CertificatePEM, r.client.KeyPEM)
	roots, _ := os.ReadFile(r.config.TrustFile)
	pool := newPool(t, roots)
	call := func(certificates []tls.Certificate, method, path string) int {
		client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, Certificates: certificates}}}
		request, _ := http.NewRequest(method, r.front.URL+path, strings.NewReader(`{}`))
		request.Header.Set(ClientCertificateHeader, "forged")
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		return response.StatusCode
	}

	if status := call([]tls.Certificate{certificate}, http.MethodPost, LaunchPath); status != http.StatusCreated {
		t.Fatalf("status %d", status)
	}
	unescaped, err := url.PathUnescape(r.upstream.header)
	if err != nil || unescaped != string(r.client.CertificatePEM) {
		t.Errorf("passed on %q, want the caller's certificate", r.upstream.header)
	}
	if strings.ContainsAny(r.upstream.header, " +/=\n") {
		t.Errorf("passed on unescaped characters: %q", r.upstream.header)
	}
	if r.upstream.forwards != "127.0.0.1" {
		t.Errorf("X-Forwarded-For = %q, want the caller", r.upstream.forwards)
	}

	if status := call(nil, http.MethodPost, LaunchPath); status != http.StatusCreated || r.upstream.header != "" {
		t.Errorf("no certificate: status %d, header %q, want it passed on empty", status, r.upstream.header)
	}
	before := r.upstream.calls
	for _, other := range [][2]string{{http.MethodGet, LaunchPath}, {http.MethodPost, "/api/v1/user/signin"}} {
		if status := call([]tls.Certificate{certificate}, other[0], other[1]); status != http.StatusNotFound {
			t.Errorf("%s %s answered %d, want 404", other[0], other[1], status)
		}
	}
	if r.upstream.calls != before {
		t.Error("the front passed on a route other than the launch")
	}
}

// The engineer signs in to the app, picks a ticket and taps the button: the
// app's server asks for a link naming them and the ticket, and the phone is
// sent to it at once.
func TestTheAppSendsItsEngineerToTheLinkPacketPulseGives(t *testing.T) {
	r := newRig(t, nil)

	response, body := r.get(t, "/")
	if response.StatusCode != http.StatusOK || !strings.Contains(body, "<title>Sign in · Field Eng App</title>") ||
		!strings.Contains(body, "Employee ID or email") || !strings.Contains(body, "<small>FE-1001 · field1001</small>") {
		t.Fatalf("the sign-in page, with its demo accounts: %d %s", response.StatusCode, body)
	}
	for _, wrong := range []url.Values{
		{"username": {"FE-1001"}, "password": {"wrong"}},
		{"username": {"FE-9999"}, "password": {"field1001"}},
	} {
		if response, body := r.post(t, "/signin", wrong); response.StatusCode != http.StatusUnauthorized ||
			!strings.Contains(body, "do not match") || !strings.Contains(body, `value="`+wrong.Get("username")+`"`) {
			t.Errorf("%v: %d, want refused, keeping what was typed", wrong, response.StatusCode)
		}
	}
	// The email works in place of the ID, in any case, as it would in the app.
	if response, _ := r.post(t, "/signin", url.Values{"username": {" ASHA.RAO@field-eng-app.example "}, "password": {"field1001"}}); response.StatusCode != http.StatusSeeOther {
		t.Errorf("signing in by email: %d", response.StatusCode)
	}
	r.signIn(t)
	if _, body := r.get(t, "/"); !strings.Contains(body, "TT-2026-01458") || !strings.Contains(body, "Signed in as <strong>Asha Rao</strong>") {
		t.Errorf("the tickets page: %s", body)
	}
	if response, body := r.get(t, "/ticket?tt=TT-2026-01458"); response.StatusCode != http.StatusOK || !strings.Contains(body, "Test this line in PacketPulse") {
		t.Errorf("the ticket page: %d", response.StatusCode)
	}

	response, _ = r.post(t, "/launch", url.Values{"tt": {"TT-2026-01458"}})
	if response.StatusCode != http.StatusSeeOther || response.Header.Get("Location") != "https://packetpulse.test/app/#launch=CODE" {
		t.Errorf("launch: %d to %q, want sent to the link", response.StatusCode, response.Header.Get("Location"))
	}
	want := map[string]string{"engineer_id": "FE-1001", "email": "asha.rao@field-eng-app.example", "name": "Asha Rao",
		"tt_number": "TT-2026-01458", "customer_id": "CUST-40217"}
	for key, value := range want {
		if r.upstream.body[key] != value {
			t.Errorf("sent %s = %q, want %q", key, r.upstream.body[key], value)
		}
	}
	unescaped, _ := url.PathUnescape(r.upstream.header)
	if unescaped != string(r.client.CertificatePEM) {
		t.Error("the app's server did not present its certificate")
	}

	if response, _ := r.post(t, "/launch", url.Values{"tt": {"TT-NOPE"}}); response.StatusCode != http.StatusNotFound {
		t.Errorf("an unknown ticket: %d", response.StatusCode)
	}
	if response, _ := r.get(t, "/ticket?tt=TT-NOPE"); response.StatusCode != http.StatusNotFound {
		t.Errorf("an unknown ticket page: %d", response.StatusCode)
	}
	if response, _ := r.post(t, "/signout", nil); response.StatusCode != http.StatusSeeOther {
		t.Errorf("sign out: %d", response.StatusCode)
	}
	if _, body := r.get(t, "/"); !strings.Contains(body, "Sign in") {
		t.Error("still signed in after signing out")
	}
}

// PacketPulse's refusal is shown as PacketPulse wrote it, with its code, and
// the phone stays in the app.
func TestARefusalIsShownAsPacketPulseWroteIt(t *testing.T) {
	r := newRig(t, nil)
	r.signIn(t)
	r.upstream.status = http.StatusForbidden
	r.upstream.answer = `{"error":{"code":"seat_limit_reached","message":"Your organisation's PacketPulse licence covers no more people.","action":"ask_administrator"}}`

	response, body := r.post(t, "/launch", url.Values{"tt": {"TT-2026-01458"}})
	if response.StatusCode != http.StatusBadGateway || !strings.Contains(body, "licence covers no more people") ||
		!strings.Contains(body, "403 seat_limit_reached") {
		t.Errorf("refusal: %d %s", response.StatusCode, body)
	}

	r.upstream.answer = `not json`
	if _, body := r.post(t, "/launch", url.Values{"tt": {"TT-2026-01458"}}); !strings.Contains(body, "PacketPulse answered 403 without saying why.") ||
		!strings.Contains(body, "Reference: 403<") {
		t.Errorf("an unreadable refusal: %s", body)
	}
	// A 404 that is not PacketPulse's own - a server from before the field
	// app's sign-in, or a wrong address - says what it means, not "404".
	r.upstream.status, r.upstream.answer = http.StatusNotFound, "404 page not found"
	if _, body := r.post(t, "/launch", url.Values{"tt": {"TT-2026-01458"}}); !strings.Contains(body, "PacketPulse does not answer field app sign-ins at this address.") {
		t.Errorf("a bare 404: %s", body)
	}
	r.upstream.status, r.upstream.answer = http.StatusCreated, `not json`
	if _, body := r.post(t, "/launch", url.Values{"tt": {"TT-2026-01458"}}); !strings.Contains(body, "could not be read") {
		t.Errorf("an unreadable link: %s", body)
	}
	r.front.Close()
	if _, body := r.post(t, "/launch", url.Values{"tt": {"TT-2026-01458"}}); !strings.Contains(body, "could not be reached") {
		t.Errorf("an unreachable PacketPulse: %s", body)
	}
}

// Only a signed-in engineer reaches a ticket, a launch or a result; a forged
// cookie signs no one in.
func TestOnlyASignedInEngineerReachesATicket(t *testing.T) {
	r := newRig(t, nil)
	base, _ := url.Parse(r.app.URL)
	r.browser.Jar.SetCookies(base, []*http.Cookie{{Name: sessionCookie, Value: "FE-1001.forged"}})
	for _, path := range []string{"/ticket?tt=TT-2026-01458", "/result?tt=TT-2026-01458"} {
		if response, _ := r.get(t, path); response.StatusCode != http.StatusSeeOther {
			t.Errorf("%s answered %d, want sent to sign in", path, response.StatusCode)
		}
	}
	if response, _ := r.post(t, "/launch", url.Values{"tt": {"TT-2026-01458"}}); response.StatusCode != http.StatusSeeOther || r.upstream.calls != 0 {
		t.Errorf("a forged launch: %d, %d calls", response.StatusCode, r.upstream.calls)
	}
	if response, _ := r.get(t, "/nowhere"); response.StatusCode != http.StatusNotFound {
		t.Errorf("an unknown page: %d", response.StatusCode)
	}
	if response, body := r.get(t, "/healthz"); response.StatusCode != http.StatusOK || body != "ok" {
		t.Errorf("healthz: %d %q", response.StatusCode, body)
	}
}

// The result is read back with the read-only key and shows who ran it.
func TestTheResultShowsWhoRanTheTest(t *testing.T) {
	r := newRig(t, nil)
	r.signIn(t)
	r.upstream.status = http.StatusOK
	r.upstream.answer = `{"tt_number":"TT-2026-01458","status":"completed","tested_by":"Asha Rao","download_mbps":null,
		"results":[{"site_name":"Mumbai reflector","reachable":true,"packet_loss":0,"avg_rtt_ms":12.5}]}`

	response, body := r.get(t, "/result?tt=TT-2026-01458")
	if response.StatusCode != http.StatusOK || !strings.Contains(body, "<dt>Tested by</dt><dd>Asha Rao</dd>") ||
		!strings.Contains(body, "Mumbai reflector") || strings.Contains(body, "Download") {
		t.Errorf("result page: %d %s", response.StatusCode, body)
	}

	r.upstream.status = http.StatusNotFound
	r.upstream.answer = `{"error":{"code":"not_found","message":"No diagnostic has been run against that TT number."}}`
	if _, body := r.get(t, "/result?tt=TT-2026-01458"); !strings.Contains(body, "No diagnostic has been run") {
		t.Errorf("no result yet: %s", body)
	}
	if response, _ := r.get(t, "/result?tt=TT-NOPE"); response.StatusCode != http.StatusNotFound {
		t.Errorf("an unknown ticket: %d", response.StatusCode)
	}

	keyless := newRig(t, func(c *Config) { c.ApiKey = "" })
	keyless.signIn(t)
	if _, body := keyless.get(t, "/result?tt=TT-2026-01458"); !strings.Contains(body, "no Results API key") {
		t.Errorf("no API key: %s", body)
	}
}

func TestTheConfigurationIsReadFromTheEnvironment(t *testing.T) {
	for _, key := range []string{"FIELD_ENG_APP_NAME", "FIELD_ENG_PARTNER_URL", "FIELD_ENG_PACKETPULSE_URL", "FIELD_ENG_CERT_FILE",
		"FIELD_ENG_KEY_FILE", "FIELD_ENG_TRUST_FILE", "FIELD_ENG_API_KEY", "FIELD_ENG_SESSION_SECRET",
		"FIELD_ENG_ENGINEERS_FILE", "FIELD_ENG_TICKETS_FILE"} {
		t.Setenv(key, "")
	}
	if _, err := ConfigFromEnvironment(); err == nil {
		t.Error("no partner address or certificate: want refused")
	}
	t.Setenv("FIELD_ENG_PARTNER_URL", "https://partner.packetpulse.test/")
	t.Setenv("FIELD_ENG_CERT_FILE", "app.crt")
	t.Setenv("FIELD_ENG_KEY_FILE", "app.key")
	config, err := ConfigFromEnvironment()
	if err != nil || config.AppName != "Field Eng App" || config.PartnerURL != "https://partner.packetpulse.test" ||
		len(config.Engineers) != 3 || len(config.Tickets) != 4 {
		t.Errorf("config = %+v (%v)", config, err)
	}

	dir := t.TempDir()
	people := filepath.Join(dir, "engineers.json")
	_ = os.WriteFile(people, []byte(`[{"id":"X-1","email":"x@y.example","name":"X","password":"p"}]`), 0o600)
	t.Setenv("FIELD_ENG_ENGINEERS_FILE", people)
	t.Setenv("FIELD_ENG_APP_NAME", "  Field Eng App (Pune)  ")
	if config, err := ConfigFromEnvironment(); err != nil || len(config.Engineers) != 1 || config.AppName != "Field Eng App (Pune)" {
		t.Errorf("engineers from a file = %+v (%v)", config.Engineers, err)
	}
	broken := filepath.Join(dir, "tickets.json")
	_ = os.WriteFile(broken, []byte(`{`), 0o600)
	t.Setenv("FIELD_ENG_TICKETS_FILE", broken)
	if _, err := ConfigFromEnvironment(); err == nil || !strings.Contains(err.Error(), "not the JSON expected") {
		t.Errorf("a broken tickets file: err = %v", err)
	}
	t.Setenv("FIELD_ENG_TICKETS_FILE", filepath.Join(dir, "missing.json"))
	if _, err := ConfigFromEnvironment(); err == nil {
		t.Error("a missing tickets file: want refused")
	}
}

// The fingerprint is the SHA-256 of the certificate, as OpenSSL prints it.
func TestTheFingerprintIsTheCertificatesSha256(t *testing.T) {
	pair, err := NewClientCertificate("x", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(pair.CertificatePEM)
	sum := sha256.Sum256(block.Bytes)
	fingerprint, err := Fingerprint(pair.CertificatePEM)
	if err != nil || !strings.HasPrefix(fingerprint, fmt.Sprintf("%02X:%02X:", sum[0], sum[1])) || len(fingerprint) != 95 {
		t.Errorf("fingerprint = %q (%v)", fingerprint, err)
	}
	if _, err := Fingerprint([]byte("not a certificate")); err == nil {
		t.Error("a fingerprint of nothing")
	}
}

// The app will not start with a certificate it cannot load or a trust file
// that holds none.
func TestTheAppNeedsItsCertificate(t *testing.T) {
	dir := t.TempDir()
	if _, err := NewPartner(Config{CertificateFile: filepath.Join(dir, "none.crt"), KeyFile: filepath.Join(dir, "none.key")}); err == nil {
		t.Error("no certificate: want refused")
	}
	pair, _ := NewClientCertificate("x", time.Hour)
	certPath, keyPath, junk := filepath.Join(dir, "a.crt"), filepath.Join(dir, "a.key"), filepath.Join(dir, "junk.crt")
	_ = os.WriteFile(certPath, pair.CertificatePEM, 0o600)
	_ = os.WriteFile(keyPath, pair.KeyPEM, 0o600)
	_ = os.WriteFile(junk, []byte("junk"), 0o600)
	if _, err := NewPartner(Config{CertificateFile: certPath, KeyFile: keyPath, TrustFile: junk}); err == nil {
		t.Error("a trust file with no certificate: want refused")
	}
	if _, err := NewPartner(Config{CertificateFile: certPath, KeyFile: keyPath, TrustFile: filepath.Join(dir, "missing")}); err == nil {
		t.Error("a missing trust file: want refused")
	}
}

func newPool(t *testing.T, pemBytes []byte) *x509.CertPool {
	t.Helper()
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		t.Fatal("no certificate to trust")
	}
	return pool
}

// The Field Eng App is shown on PacketPulse's public demo (NWTEL live), where
// Run tests fills in a sample ticket of its own. A ticket of the app's that
// PacketPulse could show by itself would hide a hand-over that failed: the
// demo would look the same either way. So no ticket number or customer ID of
// the app is text in PacketPulse's string catalogue.
func TestNoTicketOfTheAppIsTextPacketPulseShowsByItself(t *testing.T) {
	read, err := os.ReadFile("../../../packetpulsego/pkg/initmicroservice/initconstants/PacketPulseStrings.go")
	if err != nil {
		t.Fatal(err)
	}
	catalogue := string(read)
	shown := func(value string) bool { return strings.Contains(catalogue, `StringValue: "`+value+`"`) }
	if !shown("TT-2026-00731") || !shown("CUST-88412") {
		t.Fatal("the demo's sample ticket is not where this test looks for it; read the catalogue's format again")
	}
	for _, ticket := range DefaultTickets() {
		for _, value := range []string{ticket.TtNumber, ticket.CustomerId} {
			if shown(value) {
				t.Errorf("%s is text PacketPulse shows by itself: on the demo, Run tests would look the same had the app not handed it over", value)
			}
		}
	}
}

// A presenter signs in with one tap: the form is filled in with the first
// demo engineer, and each name under Demo accounts fills it with theirs. A
// refused sign-in keeps the ID that was typed, never a password.
func TestTheSignInIsFilledInWithADemoAccount(t *testing.T) {
	r := newRig(t, nil)
	filled := func(body, id, password string) bool {
		return strings.Contains(body, `name="username" value="`+id+`"`) &&
			strings.Contains(body, `type="password" value="`+password+`"`)
	}
	if _, body := r.get(t, "/"); !filled(body, "FE-1001", "field1001") || !strings.Contains(body, `<a href="/?as=FE-1003">`) || !strings.Contains(body, "<strong>Meera Iyer</strong>") {
		t.Errorf("the sign-in, filled in with the first engineer, naming the others: %s", body)
	}
	if _, body := r.get(t, "/?as=FE-1002"); !filled(body, "FE-1002", "field1002") {
		t.Errorf("chosen from the demo accounts: %s", body)
	}
	if _, body := r.get(t, "/?as=FE-9999"); !filled(body, "FE-1001", "field1001") {
		t.Errorf("an ID the app does not have: %s", body)
	}
	if response, _ := r.post(t, "/signin", url.Values{"username": {"FE-1002"}, "password": {"field1002"}}); response.StatusCode != http.StatusSeeOther {
		t.Errorf("signing in with what the form was filled in with: %d", response.StatusCode)
	}
	if _, body := r.post(t, "/signin", url.Values{"username": {"FE-1001"}, "password": {"wrong"}}); !filled(body, "FE-1001", "") {
		t.Errorf("a refused sign-in: %s", body)
	}

	empty := newRig(t, func(config *Config) { config.Engineers = nil })
	if _, body := empty.get(t, "/"); !filled(body, "", "") {
		t.Errorf("an app with no engineers: %s", body)
	}
}
