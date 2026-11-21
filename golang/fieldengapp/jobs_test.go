package fieldengapp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image/png"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The app's day as an engineer works it (18 November 2026): jobs in the
// order they matter, how long each has left, what was done to each, and a
// result read as a verdict - and the files a phone installs it from.

// at sets the app's clock: started at nine in the morning, IST, and now
// minutes later.
func (r *rig) at(minutes int) {
	started := time.Date(2026, 11, 18, 9, 0, 0, 0, ist)
	r.fieldApp.started = started
	r.fieldApp.now = func() time.Time { return started.Add(time.Duration(minutes) * time.Minute) }
}

// order is the order the TT numbers appear in a page.
func order(body string, tts ...string) bool {
	last := -1
	for _, tt := range tts {
		at := strings.Index(body, `href="/ticket?tt=`+tt+`"`)
		if at <= last {
			return false
		}
		last = at
	}
	return true
}

func TestTheDayIsOrderedByWhatMattersFirst(t *testing.T) {
	r := newRig(t, nil)
	r.at(0)
	r.signIn(t)

	_, body := r.get(t, "/")
	// P1 first, then by what runs out first; closed jobs last.
	if !order(body, "TT-2026-00752", "TT-2026-01458", "TT-2026-00744", "TT-2026-00768") {
		t.Errorf("the jobs are not P1 first, then soonest due:\n%s", body)
	}
	for _, want := range []string{
		"Good morning, Asha", "Wednesday 18 November",
		"<b>4</b><span>open jobs</span>", `class="hot"><b>1</b><span>due within the hour</span>`, "<b>0</b><span>closed today</span>",
		"Due in 50m", "Due in 2h 30m", `data-search="tt-2026-00752 cust-77230 p. nair powai`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the jobs list lacks %q", want)
		}
	}

	// An hour on, the P1 is overdue, and says so.
	r.at(60)
	if _, body := r.get(t, "/"); !strings.Contains(body, "Overdue by 10m") || !strings.Contains(body, "Good morning, Asha") {
		t.Errorf("an hour on: %s", body)
	}
}

// Opening the test marks the job as opened; closing it records how, takes
// away the test button and moves it to the end of the day.
func TestAJobIsOpenedInPacketPulseThenClosed(t *testing.T) {
	r := newRig(t, nil)
	r.at(30)
	r.signIn(t)

	if _, body := r.get(t, "/ticket?tt=TT-2026-00752"); !strings.Contains(body, "Assigned") || !strings.Contains(body, "Test the line in PacketPulse") {
		t.Fatalf("a job not yet tested: %s", body)
	}
	if response, _ := r.post(t, "/launch", url.Values{"tt": {"TT-2026-00752"}}); response.StatusCode != http.StatusSeeOther {
		t.Fatalf("launch: %d", response.StatusCode)
	}
	if _, body := r.get(t, "/ticket?tt=TT-2026-00752"); !strings.Contains(body, "Test opened") ||
		!strings.Contains(body, "Line test opened in PacketPulse</b><small>Today, 09:30") {
		t.Errorf("after the test was opened: %s", body)
	}

	for _, refused := range []url.Values{
		{"tt": {"TT-2026-00752"}, "resolution": {"Because I said so"}},
		{"tt": {"TT-2026-00752"}},
	} {
		if response, _ := r.post(t, "/close", refused); response.StatusCode != http.StatusBadRequest {
			t.Errorf("closing with %v: %d, want refused", refused, response.StatusCode)
		}
	}
	if response, _ := r.post(t, "/close", url.Values{"tt": {"TT-NONE"}, "resolution": {Resolutions[0]}}); response.StatusCode != http.StatusNotFound {
		t.Errorf("closing a job the app does not have: %d", response.StatusCode)
	}

	response, _ := r.post(t, "/close", url.Values{"tt": {"TT-2026-00752"}, "resolution": {Resolutions[2]}})
	if response.StatusCode != http.StatusSeeOther || response.Header.Get("Location") != "/ticket?tt=TT-2026-00752" {
		t.Fatalf("closing: %d %s", response.StatusCode, response.Header.Get("Location"))
	}
	_, body := r.get(t, "/ticket?tt=TT-2026-00752")
	if !strings.Contains(body, "Closed: Escalated to the network team</b><small>Today, 09:30") ||
		strings.Contains(body, `action="/launch"`) || strings.Contains(body, "Close this job") || strings.Contains(body, "Overdue") {
		t.Errorf("a closed job still offers its test, or no record of how it closed: %s", body)
	}
	_, body = r.get(t, "/")
	if !order(body, "TT-2026-01458", "TT-2026-00744", "TT-2026-00768", "TT-2026-00752") ||
		!strings.Contains(body, "<b>3</b><span>open jobs</span>") || !strings.Contains(body, "<b>1</b><span>closed today</span>") ||
		!strings.Contains(body, `data-status="closed"`) {
		t.Errorf("the closed job is not last, or not counted: %s", body)
	}
}

func TestAResultReadsAsAVerdict(t *testing.T) {
	site := func(grade string, reportOnly bool) map[string]any {
		return map[string]any{"site_name": "Mumbai edge", "reachable": grade != "breach", "packet_loss": 0.25,
			"avg_rtt_ms": 12.6, "jitter_ms": 1.04, "sla_status": grade, "report_only": reportOnly}
	}
	answer := func(status string, sites ...map[string]any) map[string]any {
		rows := make([]any, 0, len(sites))
		for _, s := range sites {
			rows = append(rows, s)
		}
		return map[string]any{"status": status, "results": rows, "tested_by": "Asha Rao",
			"started_on": "2026-11-18T10:12:00Z", "download_mbps": 94.27, "upload_mbps": nil}
	}
	for name, test := range map[string]struct {
		body          map[string]any
		verdict, tone string
	}{
		"a breach anywhere":            {answer("completed", site("ok", false), site("breach", false)), "Fault on the line", "bad"},
		"a degraded endpoint":          {answer("completed", site("ok", false), site("degraded", false)), "Degraded", "warn"},
		"every endpoint in its target": {answer("completed", site("ok", false)), "Healthy", "good"},
		"a breach kept for the report": {answer("completed", site("ok", false), site("breach", true)), "Healthy", "good"},
		"nothing graded":               {answer("failed", site("", false)), "Not graded", "neutral"},
		"a test still running":         {answer("running"), "Still testing", "info"},
	} {
		view := viewResult(test.body)
		if view.Verdict != test.verdict || view.Tone != test.tone {
			t.Errorf("%s: %s (%s), want %s (%s)", name, view.Verdict, view.Tone, test.verdict, test.tone)
		}
	}

	view := viewResult(answer("completed", site("degraded", false)))
	if view.When != "18 Nov, 15:42" || view.Download != "94.3" || view.Upload != "" || view.TestedBy != "Asha Rao" {
		t.Errorf("the figures: %+v", view)
	}
	if got := view.Sites[0]; got.Loss != "0.2" && got.Loss != "0.3" || got.Latency != "13" || got.Jitter != "1.0" || got.Tone != "warn" || !got.Reached {
		t.Errorf("an endpoint: %+v", got)
	}
}

func TestTheResultPageLeadsWithTheVerdict(t *testing.T) {
	r := newRig(t, nil)
	r.signIn(t)
	r.upstream.status = http.StatusOK
	r.upstream.answer = `{"tt_number":"TT-2026-01458","status":"completed","tested_by":"Asha Rao","started_on":"2026-11-18T10:12:00Z",` +
		`"download_mbps":94.2,"upload_mbps":48.5,"results":[{"site_name":"Mumbai edge","reachable":true,"packet_loss":0,"avg_rtt_ms":11.8,"jitter_ms":0.9,"sla_status":"ok"}]}`
	_, body := r.get(t, "/result?tt=TT-2026-01458")
	for _, want := range []string{
		`class="verdict tone-good"`, "<h1>Healthy</h1>", "<dt>Tested by</dt><dd>Asha Rao</dd>", "<b>94.2</b>", "<b>48.5</b>",
		"<strong>Mumbai edge</strong>", "Reached", "<small>Latency</small>12 ms", `href="/result?tt=TT-2026-01458"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the result page lacks %q", want)
		}
	}
}

func TestTheHelpersSayItAsAnEngineerReadsIt(t *testing.T) {
	for at, want := range map[int]string{8: "Good morning", 13: "Good afternoon", 19: "Good evening"} {
		if got := greeting(time.Date(2026, 11, 18, at, 0, 0, 0, ist)); got != want {
			t.Errorf("at %d: %q, want %q", at, got, want)
		}
	}
	for name, want := range map[string]string{"Asha Rao": "AR", "meera": "M", "Ana Maria Lopez": "AM", "": ""} {
		if got := initials(name); got != want {
			t.Errorf("initials(%q) = %q, want %q", name, got, want)
		}
	}
	if firstName("") != "" || firstName("Asha Rao") != "Asha" {
		t.Error("first names")
	}
	if span(45*time.Minute) != "45m" || span(125*time.Minute) != "2h 05m" {
		t.Error("spans")
	}
	for priority, want := range map[string]string{"P1": "bad", "P2": "warn", "P3": "info", "": "info"} {
		if got := (jobView{Ticket: Ticket{Priority: priority}}).PriorityTone(); got != want {
			t.Errorf("tone of %q = %q, want %q", priority, got, want)
		}
	}
	// A job with no priority and no SLA sorts after the rest, and promises nothing.
	jobs := []jobView{viewJob(Ticket{TtNumber: "TT-PLAIN"}, jobState{}, time.Now(), time.Now()), viewJob(Ticket{TtNumber: "TT-P3", Priority: "P3"}, jobState{}, time.Now(), time.Now())}
	sortJobs(jobs)
	if jobs[0].TtNumber != "TT-P3" || jobs[1].Due != "" {
		t.Errorf("jobs = %+v", jobs)
	}
}

// A phone installs the app from its manifest and icons, served by the app
// itself; every page lets only the app's own files run.
func TestThePhoneCanInstallTheApp(t *testing.T) {
	r := newRig(t, nil)

	response, body := r.get(t, "/manifest.webmanifest")
	var installed struct {
		Name    string              `json:"name"`
		Display string              `json:"display"`
		Start   string              `json:"start_url"`
		Icons   []map[string]string `json:"icons"`
	}
	if err := json.Unmarshal([]byte(body), &installed); err != nil || response.Header.Get("Content-Type") != "application/manifest+json" ||
		installed.Name != "Field Eng App" || installed.Display != "standalone" || installed.Start != "/" || len(installed.Icons) != 3 {
		t.Errorf("manifest: %s %v %+v", response.Header.Get("Content-Type"), err, installed)
	}

	for path, size := range map[string]int{"/icon-192.png": 192, "/icon-512.png": 512} {
		response, body := r.get(t, path)
		icon, err := png.Decode(bytes.NewReader([]byte(body)))
		if err != nil || response.Header.Get("Content-Type") != "image/png" || response.Header.Get("Cache-Control") != "public, max-age=86400" {
			t.Fatalf("%s: %v %s", path, err, response.Header)
		}
		if bounds := icon.Bounds(); bounds.Dx() != size || bounds.Dy() != size {
			t.Errorf("%s is %v", path, bounds)
		}
		// The corner is cut away, the line is white, and the ground is the brand.
		if _, _, _, a := icon.At(0, 0).RGBA(); a != 0 {
			t.Errorf("%s: the corner is not transparent", path)
		}
		unit := float64(size) / 64
		if red, green, blue, _ := icon.At(int(16*unit), int(36*unit)).RGBA(); red < 0xf000 || green < 0xf000 || blue < 0xf000 {
			t.Errorf("%s: the line is not white", path)
		}
		if red, _, blue, _ := icon.At(int(32*unit), int(10*unit)).RGBA(); blue < red {
			t.Errorf("%s: the ground is not the brand's blue", path)
		}
	}

	if response, body := r.get(t, "/icon.svg"); response.Header.Get("Content-Type") != "image/svg+xml" || !strings.HasPrefix(body, "<svg") {
		t.Errorf("icon.svg: %s", response.Header.Get("Content-Type"))
	}
	if response, body := r.get(t, "/app.js"); !strings.HasPrefix(response.Header.Get("Content-Type"), "text/javascript") ||
		!strings.Contains(body, "form[data-launch]") || !strings.Contains(body, "data-filter") ||
		// Installed on a phone, PacketPulse opens over the app: coming back
		// to the app takes the engineer to the job, waiting for its result,
		// never to an overlay left standing.
		!strings.Contains(body, "'visibilitychange'") || !strings.Contains(body, "'&back=1'") || !strings.Contains(body, "'opening-dismiss'") {
		t.Errorf("app.js: %s", response.Header.Get("Content-Type"))
	}
	response, body = r.get(t, "/")
	if policy := response.Header.Get("Content-Security-Policy"); !strings.Contains(policy, "script-src 'self'") || !strings.Contains(policy, "frame-ancestors 'none'") ||
		!strings.Contains(body, `<link rel="manifest" href="/manifest.webmanifest">`) || !regexp.MustCompile(`<script src="/app.js\?v=[0-9a-f]{12}" defer></script>`).MatchString(body) {
		t.Errorf("a page's policy %q, or no manifest and script", policy)
	}
}

// Each engineer's day is their own, and a sign-in begins it afresh: what one
// engineer closed is still open for another, and signing in again - as a
// demonstration does - finds the jobs fresh, their SLAs counting from then.
func TestEachEngineersDayIsTheirOwnAndBeginsAtSignIn(t *testing.T) {
	r := newRig(t, nil)
	r.at(0)
	r.signIn(t)
	r.post(t, "/close", url.Values{"tt": {"TT-2026-00752"}, "resolution": {Resolutions[0]}})
	if _, body := r.get(t, "/ticket?tt=TT-2026-00752"); !strings.Contains(body, "Closed:") {
		t.Fatalf("not closed for Asha: %s", body)
	}

	r.at(180)
	if response, _ := r.post(t, "/signin", url.Values{"username": {"FE-1002"}, "password": {"field1002"}}); response.StatusCode != http.StatusSeeOther {
		t.Fatal("Vikram's sign-in")
	}
	if _, body := r.get(t, "/ticket?tt=TT-2026-00752"); strings.Contains(body, "Closed:") || !strings.Contains(body, "Due in 50m") {
		t.Errorf("Asha's closing reached Vikram's day, or his SLAs do not count from his sign-in: %s", body)
	}

	r.at(240)
	r.signIn(t)
	if _, body := r.get(t, "/"); strings.Contains(body, `data-status="closed"`) || !strings.Contains(body, "Due in 50m") {
		t.Errorf("signing in again did not begin Asha's day afresh: %s", body)
	}
}

// A sign-in from before the app last started counts its SLAs from when
// the app did.
func TestASignInFromBeforeARestartCountsFromTheRestart(t *testing.T) {
	book := newJobBook()
	restarted := time.Date(2026, 11, 18, 9, 0, 0, 0, ist)
	if got := book.day("FE-1001", restarted); !got.Equal(restarted) {
		t.Errorf("day = %v, want the restart", got)
	}
	book.begin("FE-1001", restarted.Add(time.Hour))
	if got := book.day("FE-1001", restarted); !got.Equal(restarted.Add(time.Hour)) {
		t.Errorf("day = %v, want the sign-in", got)
	}
}

// The app names its own page for the ticket as the way back from
// PacketPulse - https only, as PacketPulse takes nothing else: its public
// address when set, else the one an https request came in by, else none.
func TestTheAppNamesItsOwnPageAsTheWayBack(t *testing.T) {
	sent := func(change func(*Config), header map[string]string) (string, bool) {
		t.Helper()
		r := newRig(t, change)
		r.signIn(t)
		request, _ := http.NewRequest(http.MethodPost, r.app.URL+"/launch", strings.NewReader(url.Values{"tt": {"TT-2026-01458"}}.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		for key, value := range header {
			request.Header.Set(key, value)
		}
		response, err := r.browser.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		if response.StatusCode != http.StatusSeeOther {
			t.Fatalf("launch: %d", response.StatusCode)
		}
		value, present := r.upstream.body["return_url"]
		return value, present
	}

	if got, _ := sent(func(c *Config) { c.PublicURL = "https://fieldapp.example" }, nil); got != "https://fieldapp.example/ticket?tt=TT-2026-01458&back=1" {
		t.Errorf("with a public address: %q", got)
	}
	if got, _ := sent(nil, map[string]string{"X-Forwarded-Proto": "https"}); !strings.HasPrefix(got, "https://127.0.0.1:") ||
		!strings.HasSuffix(got, "/ticket?tt=TT-2026-01458&back=1") {
		t.Errorf("from an https request: %q", got)
	}
	if got, present := sent(nil, nil); present {
		t.Errorf("over plain http: %q, want none", got)
	}
	if got, present := sent(func(c *Config) { c.PublicURL = "http://fieldapp.example" }, nil); present {
		t.Errorf("a plain http public address: %q, want none", got)
	}
}

// Back on the job, the page greets the engineer and waits for the result;
// the result reads as PacketPulse's verdict once filed, which the job then
// keeps in the list - and each other state says what it is.
func TestTheResultLandsOnTheJobItself(t *testing.T) {
	r := newRig(t, nil)
	r.signIn(t)
	if _, body := r.get(t, "/ticket?tt=TT-2026-01458"); strings.Contains(body, `id="live-result"`) {
		t.Errorf("a job not yet tested waits for a result: %s", body)
	}
	_, body := r.get(t, "/ticket?tt=TT-2026-01458&back=1")
	if !strings.Contains(body, "Back from PacketPulse") || !strings.Contains(body, `id="live-result" data-tt="TT-2026-01458"`) {
		t.Errorf("back from PacketPulse: %s", body)
	}

	live := func() map[string]string {
		t.Helper()
		response, body := r.get(t, "/result.json?tt=TT-2026-01458")
		var result map[string]string
		if err := json.Unmarshal([]byte(body), &result); err != nil || response.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("result.json: %v %s", err, body)
		}
		return result
	}
	r.upstream.status, r.upstream.answer = http.StatusNotFound, `{"error":{"code":"not_found","message":"No diagnostic has been run for that TT number."}}`
	if got := live(); got["state"] != "none" {
		t.Errorf("nothing filed yet: %v", got)
	}
	r.upstream.status, r.upstream.answer = http.StatusOK, `{"status":"running","results":[]}`
	if got := live(); got["state"] != "running" {
		t.Errorf("still testing: %v", got)
	}
	if _, body := r.get(t, "/"); strings.Contains(body, "Healthy") {
		t.Error("a verdict shown before PacketPulse graded the test")
	}
	r.upstream.answer = `{"status":"completed","tested_by":"Asha Rao","started_on":"2026-11-18T10:12:00Z","download_mbps":94.2,` +
		`"results":[{"site_name":"Mumbai edge","reachable":true,"packet_loss":0,"avg_rtt_ms":11.8,"sla_status":"ok"}]}`
	if got := live(); got["state"] != "done" || got["verdict"] != "Healthy" || got["tone"] != "good" || got["tested_by"] != "Asha Rao" ||
		got["when"] != "18 Nov, 15:42" || got["download"] != "94.2" {
		t.Errorf("graded: %v", got)
	}
	if _, body := r.get(t, "/"); !strings.Contains(body, `<span class="chip good">Healthy</span>`) {
		t.Errorf("the jobs list does not show the verdict: %s", body)
	}
	r.upstream.status, r.upstream.answer = http.StatusForbidden, `{"error":{"code":"forbidden","message":"That key may not read results."}}`
	if got := live(); got["state"] != "unavailable" || got["message"] != "That key may not read results." {
		t.Errorf("refused: %v", got)
	}
	if response, _ := r.get(t, "/result.json?tt=TT-NONE"); response.StatusCode != http.StatusNotFound {
		t.Errorf("a job the app does not have: %d", response.StatusCode)
	}

	keyless := newRig(t, func(c *Config) { c.ApiKey = "" })
	keyless.signIn(t)
	_, body = keyless.get(t, "/result.json?tt=TT-2026-01458")
	if !strings.Contains(body, `"state":"unavailable"`) || !strings.Contains(body, "no PacketPulse Results API key") {
		t.Errorf("without a key: %s", body)
	}
}

// Installed on a phone, the app opens PacketPulse in a browser layer over
// itself, and its own page stays under the "Opening PacketPulse" overlay -
// which was left standing when the engineer closed PacketPulse, with no way
// back (18 November 2026). The page script, run in Node against a stand-in
// page: back in the app after a tap, the engineer lands on the job waiting
// for its result; after twenty seconds the overlay says what to do; the
// browser's Back lands on the job too; and coming back without a tap
// changes nothing.
func TestTheAppNeverLeavesTheEngineerOnItsOverlay(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal("node is needed to run the app's page script")
	}
	page := filepath.Join(t.TempDir(), "app.js")
	if err := os.WriteFile(page, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(node, "testdata/app_script_check.js", page).CombinedOutput()
	if err != nil || strings.Count(string(output), "ok   ") != 6 {
		t.Errorf("the page script: %v\n%s", err, output)
	}
}

// PacketPulse files no figure for a way of the line it could not measure
// fully, so a result can carry a download and no upload. The app showed the
// download alone, which read as the upload gone missing (customer review,
// 19 November 2026): both renderers - the result page and the job page's
// live card - show both ways once either has a figure, and the way with
// none says Not recorded. With neither, no speed is shown at all.
func TestAWayOfTheLineWithNoFigureReadsNotRecorded(t *testing.T) {
	r := newRig(t, nil)
	r.signIn(t)
	r.upstream.status = http.StatusOK
	r.upstream.answer = `{"tt_number":"TT-2026-01458","status":"completed","tested_by":"Asha Rao","started_on":"2026-11-18T10:12:00Z",` +
		`"download_mbps":81.9,"upload_mbps":null,"results":[{"site_name":"Mumbai edge","reachable":true,"sla_status":"ok"}]}`
	_, body := r.get(t, "/result?tt=TT-2026-01458")
	for _, want := range []string{
		`<div><small>Download</small><b>81.9</b><small>Mbit/s</small></div>`,
		`<div><small>Upload</small><b class="unrecorded">Not recorded</b></div>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the result page lacks %q", want)
		}
	}

	for answer, want := range map[string]string{
		`{"state":"done","verdict":"Healthy","tone":"good","download":"81.9"}`:                `["Download 81.9 Mbit/s","Upload Not recorded"]`,
		`{"state":"done","verdict":"Healthy","tone":"good","upload":"3.6"}`:                   `["Download Not recorded","Upload 3.6 Mbit/s"]`,
		`{"state":"done","verdict":"Healthy","tone":"good","download":"81.9","upload":"3.6"}`: `["Download 81.9 Mbit/s","Upload 3.6 Mbit/s"]`,
		`{"state":"done","verdict":"Healthy","tone":"good"}`:                                  `[]`,
	} {
		drawn := liveCard(t, answer)
		if tiles, _ := json.Marshal(drawn.Tiles); string(tiles) != want {
			t.Errorf("the live card's tiles for %s: %s, want %s", answer, tiles, want)
		}
	}
}

// liveDial is one dial as the page script drew it on the live card.
type liveDial struct {
	Class, Way, Figure, Arc, Track, Hidden string
	Needle                                 []string
	Ticks                                  [][]string
}

// liveCard runs the page script against a stand-in for the job page's live
// card, given the /result.json answer, and says what it drew.
func liveCard(t *testing.T, answer string) (drawn struct {
	Tiles []string
	Dials []liveDial
}) {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal("node is needed to run the app's page script")
	}
	page := filepath.Join(t.TempDir(), "app.js")
	if err := os.WriteFile(page, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(node, "testdata/live_result_check.js", page, answer).CombinedOutput()
	if err != nil {
		t.Fatalf("the live card for %s: %v\n%s", answer, err, output)
	}
	if err := json.Unmarshal(output, &drawn); err != nil {
		t.Fatalf("the live card for %s: %v\n%s", answer, err, output)
	}
	return drawn
}

// The line's speed as dials (the user, 19 November 2026: "keep the speed dial
// UX in field eng app as well", "subtle and clear in reports"): the result
// page and the live card each draw both ways resting at their figures, above
// the tiles - a way with no figure at nought, a dash - on PacketPulse's own
// scale. The page is drawn by the server and the card by the page script,
// so the two are held to the same needle, sweep and ticks; and a screen
// reader passes the drawing by, as the tiles say the same in words.
func TestTheDialsAreDrawnAlikeInPageAndScript(t *testing.T) {
	// The scale: each step between two marks the same share of the sweep.
	for i, mark := range speedMarks {
		if got := dialPosition(mark); math.Abs(got-float64(i)/8) > 1e-12 {
			t.Errorf("mark %v sits at %v of the sweep, want %v", mark, got, float64(i)/8)
		}
	}
	for mbps, want := range map[float64]float64{-1: 0, 2: 1.25 / 8, 300: 7.2 / 8, 900: 1} {
		if got := dialPosition(mbps); math.Abs(got-want) > 1e-12 {
			t.Errorf("%v Mbit/s sits at %v of the sweep, want %v", mbps, got, want)
		}
	}

	r := newRig(t, nil)
	r.signIn(t)
	r.upstream.status = http.StatusOK
	r.upstream.answer = `{"tt_number":"TT-2026-01458","status":"completed","started_on":"2026-11-18T10:12:00Z",` +
		`"download_mbps":207.7,"upload_mbps":null,"results":[{"site_name":"Mumbai edge","reachable":true,"sla_status":"ok"}]}`
	_, page := r.get(t, "/result?tt=TT-2026-01458")
	down, up := newSpeedDial("Download", "207.7", false), newSpeedDial("Upload", "", true)
	dials := strings.Index(page, `<div class="dials">`)
	if dials < 0 || dials > strings.Index(page, `<div class="tiles">`) {
		t.Fatalf("the result page has no dials above its tiles:\n%s", page)
	}
	for _, want := range []string{
		`<div class="dial"><svg viewBox="0 0 200 150" aria-hidden="true" focusable="false">`,
		`<path class="dial-arc" d="` + down.Arc + `"/>`,
		`x2="` + down.NeedleX + `" y2="` + down.NeedleY + `"/>`,
		`<text class="dial-figure" x="100" y="124">207.7</text>`,
		`<div class="dial up none">`,
		`<text class="dial-figure" x="100" y="124">—</text>`,
		`x2="` + up.NeedleX + `" y2="` + up.NeedleY + `"/>`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the result page lacks %q", want)
		}
	}
	if up.Arc != "" || strings.Count(page, `class="dial-tick"`) != 2*len(speedMarks) {
		t.Errorf("a way with no figure swept %q; ticks drawn %d", up.Arc, strings.Count(page, `class="dial-tick"`))
	}

	// The live card, from the page script: the same dials.
	drawn := liveCard(t, `{"state":"done","verdict":"Healthy","tone":"good","download":"207.7"}`)
	if len(drawn.Dials) != 2 {
		t.Fatalf("the live card drew %d dials, want 2", len(drawn.Dials))
	}
	near := func(got, want string) bool {
		g, errG := strconv.ParseFloat(got, 64)
		w, errW := strconv.ParseFloat(want, 64)
		return errG == nil && errW == nil && math.Abs(g-w) <= 0.11
	}
	for i, want := range []speedDial{down, up} {
		got := drawn.Dials[i]
		wantClass := map[bool]string{false: "dial", true: "dial up none"}[want.Upload]
		if got.Class != wantClass || got.Hidden != "true" || !near(got.Needle[0], want.NeedleX) || !near(got.Needle[1], want.NeedleY) ||
			len(got.Ticks) != len(want.Ticks) || got.Track != want.Track || (got.Arc == "") != (want.Arc == "") {
			t.Errorf("the live card's %s dial = %+v; the page's = %+v", want.Way, got, want)
		}
		for k, tick := range want.Ticks {
			if !near(got.Ticks[k][0], tick.X1) || !near(got.Ticks[k][3], tick.Y2) {
				t.Errorf("the live card's %s tick %d = %v, the page's %+v", want.Way, k, got.Ticks[k], tick)
			}
		}
	}
	if drawn.Dials[0].Figure != "207.7" || drawn.Dials[1].Figure != "—" ||
		drawn.Dials[0].Way != "↓ Download" || drawn.Dials[1].Way != "↑ Upload" {
		t.Errorf("the live card's dials say %+v", drawn.Dials)
	}
}

// The script's address is named by its content: a changed script is a new
// address, so no phone keeps an old one for its day of caching - an old
// script once kept the "Opening PacketPulse" overlay standing after a fix.
func TestTheScriptsAddressChangesWithTheScript(t *testing.T) {
	sum := sha256.Sum256([]byte(script))
	if want := "/app.js?v=" + hex.EncodeToString(sum[:6]); scriptPath != want {
		t.Errorf("script path %q, want %q", scriptPath, want)
	}
	r := newRig(t, nil)
	if _, body := r.get(t, "/"); !strings.Contains(body, `src="`+scriptPath+`"`) {
		t.Errorf("the page does not load the script by its versioned address")
	}
	if response, body := r.get(t, scriptPath); response.StatusCode != http.StatusOK || body != script {
		t.Errorf("the versioned address: %d", response.StatusCode)
	}
}
