package fieldengapp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
)

// App is the mock's web app and server, as one http.Handler.
type App struct {
	config  Config
	partner *Partner
	secret  []byte
	pages   *template.Template
	jobs    *jobBook
	// started is when the app began: the jobs' SLAs count from an
	// engineer's sign-in, or from this for a sign-in before it. now is the
	// clock, which a test may set.
	started time.Time
	now     func() time.Time
}

const sessionCookie = "field_eng_session"

// NewApp builds the app. The session secret is random when none is set,
// which signs everyone out when the app restarts - acceptable for a mock.
func NewApp(config Config, partner *Partner) *App {
	secret := config.SessionSecret
	if len(secret) == 0 {
		secret = make([]byte, 32)
		_, _ = rand.Read(secret)
	}
	if config.AppName == "" {
		config.AppName = DefaultAppName
	}
	app := &App{config: config, partner: partner, secret: secret, jobs: newJobBook(), now: time.Now}
	app.started = app.now()
	app.pages = template.Must(template.New("pages").Funcs(template.FuncMap{"initials": initials, "script": func() string { return scriptPath }, "speed": newSpeedTile, "dial": newSpeedDial}).Parse(pages))
	return app
}

// icons are the PNG icons, drawn once, when first asked for.
var icons = sync.OnceValue(func() map[string][]byte {
	return map[string][]byte{"/icon-192.png": iconPNG(192), "/icon-512.png": iconPNG(512)}
})

// contentPolicy lets a page load only the app's own files. Styles are in
// the page; the one script is /app.js.
const contentPolicy = "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; " +
	"script-src 'self'; base-uri 'none'; frame-ancestors 'none'"

// ServeHTTP routes the app's few pages.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", contentPolicy)
	switch {
	case r.URL.Path == "/healthz":
		_, _ = w.Write([]byte("ok"))
	case r.URL.Path == "/manifest.webmanifest":
		asset(w, "application/manifest+json", manifest(a.config.AppName))
	case r.URL.Path == "/icon.svg":
		asset(w, "image/svg+xml", []byte(iconSVG))
	case r.URL.Path == "/icon-192.png" || r.URL.Path == "/icon-512.png":
		asset(w, "image/png", icons()[r.URL.Path])
	case r.URL.Path == "/app.js":
		asset(w, "text/javascript; charset=utf-8", []byte(script))
	case r.URL.Path == "/" && r.Method == http.MethodGet:
		a.home(w, r)
	case r.URL.Path == "/signin" && r.Method == http.MethodPost:
		a.signIn(w, r)
	case r.URL.Path == "/signout" && r.Method == http.MethodPost:
		http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	case r.URL.Path == "/ticket" && r.Method == http.MethodGet:
		a.withEngineer(w, r, func(engineer Engineer) { a.ticket(w, r, engineer, nil) })
	case r.URL.Path == "/launch" && r.Method == http.MethodPost:
		a.withEngineer(w, r, func(engineer Engineer) { a.launch(w, r, engineer) })
	case r.URL.Path == "/result" && r.Method == http.MethodGet:
		a.withEngineer(w, r, func(engineer Engineer) { a.result(w, r, engineer) })
	case r.URL.Path == "/result.json" && r.Method == http.MethodGet:
		a.withEngineer(w, r, func(engineer Engineer) { a.resultJSON(w, r, engineer) })
	case r.URL.Path == "/close" && r.Method == http.MethodPost:
		a.withEngineer(w, r, func(engineer Engineer) { a.closeJob(w, r, engineer) })
	default:
		http.NotFound(w, r)
	}
}

// --- the engineer's own sign-in to the app ---------------------------------

func (a *App) sign(id string) string {
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(id))
	return id + "." + hex.EncodeToString(mac.Sum(nil))
}

// signedIn is the engineer the app's cookie names, when it is genuine.
func (a *App) signedIn(r *http.Request) (Engineer, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return Engineer{}, false
	}
	id, _, found := strings.Cut(cookie.Value, ".")
	if !found || !hmac.Equal([]byte(cookie.Value), []byte(a.sign(id))) {
		return Engineer{}, false
	}
	return a.engineer(id)
}

func (a *App) engineer(id string) (Engineer, bool) {
	for _, engineer := range a.config.Engineers {
		if engineer.Id == id {
			return engineer, true
		}
	}
	return Engineer{}, false
}

// byUsername is the engineer with this employee ID or email, in any case.
func (a *App) byUsername(username string) (Engineer, bool) {
	for _, engineer := range a.config.Engineers {
		if strings.EqualFold(engineer.Id, username) || strings.EqualFold(engineer.Email, username) {
			return engineer, true
		}
	}
	return Engineer{}, false
}

func (a *App) withEngineer(w http.ResponseWriter, r *http.Request, then func(Engineer)) {
	engineer, ok := a.signedIn(r)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	then(engineer)
}

// demoAccount is the demo engineer the sign-in form is filled in with: the
// one a row of the demo accounts names (?as=), or else the first. The app is
// only ever a demonstration, so a presenter signs in with one tap.
func (a *App) demoAccount(id string) Engineer {
	if engineer, ok := a.engineer(id); ok {
		return engineer
	}
	if len(a.config.Engineers) == 0 {
		return Engineer{}
	}
	return a.config.Engineers[0]
}

// signIn is the app's own sign-in, as an engineer does it every morning:
// their employee ID or email, and their password.
func (a *App) signIn(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.PostFormValue("username"))
	engineer, known := a.byUsername(username)
	if !known || subtle.ConstantTimeCompare([]byte(engineer.Password), []byte(r.PostFormValue("password"))) != 1 {
		a.render(w, http.StatusUnauthorized, "signin", page{App: a.config.AppName, Engineers: a.config.Engineers,
			Username: username, Problem: "That employee ID or email and password do not match."})
		return
	}
	a.jobs.begin(engineer.Id, a.now())
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: a.sign(engineer.Id), Path: "/",
		HttpOnly: true, Secure: r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https", SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// --- pages -------------------------------------------------------------------

// page is what every template is given.
type page struct {
	App string
	// Title names the screen; Back, when set, is where its back arrow goes.
	Title     string
	Back      string
	Engineer  *Engineer
	Engineers []Engineer
	Problem   string
	Username  string
	// Password fills in the sign-in form: the app is a demonstration, and
	// its accounts are listed on the same page.
	Password string
	Refusal  *Refusal

	// The jobs list.
	Greeting string
	Today    string
	Jobs     []jobView
	Open     int
	Urgent   int
	Closed   int

	// One job, and its result. Returned is a job page reached back from
	// PacketPulse, which greets the engineer.
	Returned    bool
	Job         *jobView
	Resolutions []string
	Result      *resultView
}

func (a *App) render(w http.ResponseWriter, status int, name string, data page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := a.pages.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("render", "page", name, "error", err)
	}
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	engineer, ok := a.signedIn(r)
	if !ok {
		demo := a.demoAccount(r.URL.Query().Get("as"))
		a.render(w, http.StatusOK, "signin", page{App: a.config.AppName, Title: "Sign in", Engineers: a.config.Engineers,
			Username: demo.Id, Password: demo.Password})
		return
	}
	now := a.now()
	data := page{App: a.config.AppName, Title: "Today's jobs", Engineer: &engineer,
		Greeting: greeting(now) + ", " + firstName(engineer.Name), Today: now.In(ist).Format("Monday 2 January")}
	began := a.jobs.day(engineer.Id, a.started)
	for _, ticket := range a.config.Tickets {
		job := viewJob(ticket, a.jobs.get(engineer.Id, ticket.TtNumber), began, now)
		data.Jobs = append(data.Jobs, job)
		switch {
		case job.Status == "closed":
			data.Closed++
		case job.Urgent:
			data.Open++
			data.Urgent++
		default:
			data.Open++
		}
	}
	sortJobs(data.Jobs)
	a.render(w, http.StatusOK, "tickets", data)
}

func (a *App) findTicket(ttNumber string) (Ticket, bool) {
	for _, ticket := range a.config.Tickets {
		if ticket.TtNumber == ttNumber {
			return ticket, true
		}
	}
	return Ticket{}, false
}

// job is the job a request names, as the engineer's day shows it now.
func (a *App) job(engineer Engineer, ttNumber string) (*jobView, bool) {
	ticket, ok := a.findTicket(ttNumber)
	if !ok {
		return nil, false
	}
	job := viewJob(ticket, a.jobs.get(engineer.Id, ticket.TtNumber), a.jobs.day(engineer.Id, a.started), a.now())
	return &job, true
}

func (a *App) ticket(w http.ResponseWriter, r *http.Request, engineer Engineer, refusal *Refusal) {
	job, ok := a.job(engineer, r.FormValue("tt"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	status := http.StatusOK
	if refusal != nil {
		status = http.StatusBadGateway
	}
	a.render(w, status, "ticket", page{App: a.config.AppName, Title: job.TtNumber, Back: "/", Engineer: &engineer,
		Job: job, Refusal: refusal, Resolutions: Resolutions, Returned: r.FormValue("back") == "1"})
}

// launch is the button on the ticket: the app's server asks PacketPulse for
// a link, over mutual TLS, and the phone is sent to it at once. A link that
// waits runs out, so it is never shown or stored.
func (a *App) launch(w http.ResponseWriter, r *http.Request, engineer Engineer) {
	ticket, ok := a.findTicket(r.PostFormValue("tt"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	link, err := a.partner.Launch(r.Context(), engineer, ticket, a.returnURL(r, ticket))
	if err != nil {
		var refusal *Refusal
		if !errors.As(err, &refusal) {
			refusal = &Refusal{Status: http.StatusBadGateway, Code: "unreachable", Message: err.Error()}
		}
		slog.Warn("PacketPulse refused a launch", "engineer", engineer.Id, "tt", ticket.TtNumber, "error", err)
		a.ticket(w, r, engineer, refusal)
		return
	}
	a.jobs.opened(engineer.Id, ticket.TtNumber, a.now())
	slog.Info("launched", "engineer", engineer.Id, "tt", ticket.TtNumber, "made_now", link.Engineer.Created)
	http.Redirect(w, r, link.Url, http.StatusSeeOther)
}

// returnURL is the job's page as the phone reaches it, for PacketPulse to
// send the engineer back to - marked, so the page greets them and waits for
// the result. https only: over plain http there is none.
func (a *App) returnURL(r *http.Request, ticket Ticket) string {
	base := a.config.PublicURL
	if base == "" && (r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https") {
		base = "https://" + r.Host
	}
	if !strings.HasPrefix(base, "https://") {
		return ""
	}
	return base + "/ticket?tt=" + url.QueryEscape(ticket.TtNumber) + "&back=1"
}

// liveResult is what the job page shows of the ticket's result as it
// waits for it: whether there is one yet, and its verdict when there is.
type liveResult struct {
	// State is done, running, none (nothing filed yet) or unavailable.
	State   string `json:"state"`
	Message string `json:"message,omitempty"`
	Verdict string `json:"verdict,omitempty"`
	Tone    string `json:"tone,omitempty"`
	Note    string `json:"note,omitempty"`
	By      string `json:"tested_by,omitempty"`
	When    string `json:"when,omitempty"`
	Down    string `json:"download,omitempty"`
	Up      string `json:"upload,omitempty"`
}

// resultJSON is the job page's view of the ticket's latest result, read as
// the page waits: the verdict once PacketPulse has one, which the job then
// keeps for the jobs list.
func (a *App) resultJSON(w http.ResponseWriter, r *http.Request, engineer Engineer) {
	ticket, ok := a.findTicket(r.FormValue("tt"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	live := liveResult{State: "unavailable"}
	body, err := a.partner.Result(r.Context(), ticket.TtNumber)
	var refusal *Refusal
	switch {
	case err == nil:
		view := viewResult(body)
		live = liveResult{State: "done", Verdict: view.Verdict, Tone: view.Tone, Note: view.VerdictNote,
			By: view.TestedBy, When: view.When, Down: view.Download, Up: view.Upload}
		if view.Running {
			live.State = "running"
		} else {
			a.jobs.graded(engineer.Id, ticket.TtNumber, view.Verdict, view.Tone)
		}
	case errors.As(err, &refusal) && refusal.Status == http.StatusNotFound:
		live.State = "none"
	case errors.As(err, &refusal):
		live.Message = refusal.Message
	default:
		live.Message = "This app has no PacketPulse Results API key yet, so the result shows in PacketPulse only."
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(live)
}

// closeJob closes a job with one of the resolutions.
func (a *App) closeJob(w http.ResponseWriter, r *http.Request, engineer Engineer) {
	ticket, ok := a.findTicket(r.PostFormValue("tt"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	resolution := r.PostFormValue("resolution")
	if !slices.Contains(Resolutions, resolution) {
		http.Error(w, "Choose how the job was resolved.", http.StatusBadRequest)
		return
	}
	a.jobs.close(engineer.Id, ticket.TtNumber, resolution, a.now())
	slog.Info("closed", "engineer", engineer.Id, "tt", ticket.TtNumber, "resolution", resolution)
	http.Redirect(w, r, "/ticket?tt="+ticket.TtNumber, http.StatusSeeOther)
}

func (a *App) result(w http.ResponseWriter, r *http.Request, engineer Engineer) {
	job, ok := a.job(engineer, r.FormValue("tt"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	data := page{App: a.config.AppName, Title: "Line test", Back: "/ticket?tt=" + job.TtNumber, Engineer: &engineer, Job: job}
	body, err := a.partner.Result(r.Context(), job.TtNumber)
	if err != nil {
		var refusal *Refusal
		if !errors.As(err, &refusal) {
			refusal = &Refusal{Status: http.StatusBadGateway, Code: "unavailable", Message: err.Error()}
		}
		data.Refusal = refusal
	} else {
		data.Result = viewResult(body)
	}
	a.render(w, http.StatusOK, "result", data)
}
