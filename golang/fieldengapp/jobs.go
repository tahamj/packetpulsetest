package fieldengapp

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// What the engineer has done with a job in this app, and how a job and its
// result read on screen. The state is kept in memory, and each engineer's
// day is their own and begins at their sign-in: a demonstration signs in,
// and finds the jobs fresh, their SLAs counting down from then.

// Resolutions are the ways an engineer may close a job.
var Resolutions = []string{
	"Fixed at the customer’s premises",
	"Customer equipment replaced",
	"Escalated to the network team",
	"No fault found",
}

// ist is the zone the demo's engineers work in. A fixed offset, so the
// host needs no time-zone database.
var ist = time.FixedZone("IST", 5*3600+30*60)

// jobState is one job's progress in the app.
type jobState struct {
	TestedAt   time.Time
	ClosedAt   time.Time
	Resolution string
	// Verdict and Tone are PacketPulse's grading of the job's latest test,
	// once the app has read it.
	Verdict string
	Tone    string
}

// jobBook holds each engineer's day: when it began, and their jobs' state.
type jobBook struct {
	mu     sync.Mutex
	began  map[string]time.Time
	states map[[2]string]jobState
}

func newJobBook() *jobBook {
	return &jobBook{began: map[string]time.Time{}, states: map[[2]string]jobState{}}
}

// begin starts an engineer's day afresh: no job opened or closed yet.
func (b *jobBook) begin(engineerId string, at time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.began[engineerId] = at
	for key := range b.states {
		if key[0] == engineerId {
			delete(b.states, key)
		}
	}
}

// day is when an engineer's day began, or otherwise when the app did - for
// a sign-in from before the app last started.
func (b *jobBook) day(engineerId string, otherwise time.Time) time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	if began, ok := b.began[engineerId]; ok {
		return began
	}
	return otherwise
}

func (b *jobBook) get(engineerId, ttNumber string) jobState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.states[[2]string{engineerId, ttNumber}]
}

// opened records that the engineer opened the job's test in PacketPulse.
func (b *jobBook) opened(engineerId, ttNumber string, at time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := [2]string{engineerId, ttNumber}
	state := b.states[key]
	state.TestedAt = at
	b.states[key] = state
}

// graded keeps PacketPulse's verdict on the job's latest test.
func (b *jobBook) graded(engineerId, ttNumber, verdict, tone string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := [2]string{engineerId, ttNumber}
	state := b.states[key]
	state.Verdict, state.Tone = verdict, tone
	b.states[key] = state
}

// close closes the job with a resolution.
func (b *jobBook) close(engineerId, ttNumber, resolution string, at time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := [2]string{engineerId, ttNumber}
	state := b.states[key]
	state.ClosedAt = at
	state.Resolution = resolution
	b.states[key] = state
}

// jobView is a ticket as the app shows it: with what was done to it, and
// how long is left to do it.
type jobView struct {
	Ticket
	// Status is assigned, testing or closed; StatusLabel says it in words.
	Status      string
	StatusLabel string
	// Due is how long is left on the job's SLA, in words; Urgent when it is
	// under an hour or past.
	Due    string
	Urgent bool
	// TestedAt and ClosedAt are when, in words; Resolution how it closed.
	TestedAt   string
	ClosedAt   string
	Resolution string
	// Verdict and VerdictTone are PacketPulse's grading of the latest test.
	Verdict     string
	VerdictTone string
	// Search is everything a search may match, in lower case.
	Search string
}

// PriorityTone is the colour a priority wears.
func (j jobView) PriorityTone() string {
	switch j.Priority {
	case "P1":
		return "bad"
	case "P2":
		return "warn"
	default:
		return "info"
	}
}

func priorityRank(priority string) int {
	switch priority {
	case "P1":
		return 1
	case "P2":
		return 2
	case "P3":
		return 3
	default:
		return 4
	}
}

// viewJob is a ticket's view at now, for an app that started at started.
func viewJob(ticket Ticket, state jobState, started, now time.Time) jobView {
	view := jobView{Ticket: ticket, Status: "assigned", StatusLabel: "Assigned",
		Verdict: state.Verdict, VerdictTone: state.Tone,
		Search: strings.ToLower(strings.Join([]string{ticket.TtNumber, ticket.CustomerId, ticket.Customer,
			ticket.Area, ticket.Address, ticket.Fault, ticket.Plan, ticket.Priority}, " "))}
	if !state.TestedAt.IsZero() {
		view.Status, view.StatusLabel = "testing", "Test opened"
		view.TestedAt = clock(state.TestedAt)
	}
	if !state.ClosedAt.IsZero() {
		view.Status, view.StatusLabel = "closed", "Closed"
		view.ClosedAt = clock(state.ClosedAt)
		view.Resolution = state.Resolution
		return view
	}
	if ticket.DueInMinutes > 0 {
		left := started.Add(time.Duration(ticket.DueInMinutes) * time.Minute).Sub(now)
		view.Urgent = left < time.Hour
		if left < 0 {
			view.Due = "Overdue by " + span(-left)
		} else {
			view.Due = "Due in " + span(left)
		}
	}
	return view
}

// sortJobs puts open jobs before closed ones, then by priority, then by
// what runs out first.
func sortJobs(jobs []jobView) {
	sort.SliceStable(jobs, func(i, k int) bool {
		closedI, closedK := jobs[i].Status == "closed", jobs[k].Status == "closed"
		if closedI != closedK {
			return !closedI
		}
		if rankI, rankK := priorityRank(jobs[i].Priority), priorityRank(jobs[k].Priority); rankI != rankK {
			return rankI < rankK
		}
		return jobs[i].DueInMinutes < jobs[k].DueInMinutes
	})
}

// span is a duration as an engineer reads it: 2h 05m, or 45m.
func span(d time.Duration) string {
	minutes := int(math.Round(d.Minutes()))
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dh %02dm", minutes/60, minutes%60)
}

// clock is a time of day in the engineers' zone.
func clock(at time.Time) string { return at.In(ist).Format("15:04") }

// greeting is how the jobs list greets the engineer at this time of day.
func greeting(now time.Time) string {
	switch hour := now.In(ist).Hour(); {
	case hour < 12:
		return "Good morning"
	case hour < 17:
		return "Good afternoon"
	default:
		return "Good evening"
	}
}

// firstName is the name a greeting uses.
func firstName(name string) string {
	if fields := strings.Fields(name); len(fields) > 0 {
		return fields[0]
	}
	return name
}

// initials are the two letters on an engineer's avatar.
func initials(name string) string {
	var out []rune
	for _, field := range strings.Fields(name) {
		out = append(out, []rune(field)[0])
		if len(out) == 2 {
			break
		}
	}
	return strings.ToUpper(string(out))
}

// --- a result -----------------------------------------------------------------

// siteView is one endpoint's result.
type siteView struct {
	Name    string
	Reached bool
	Loss    string
	Latency string
	Jitter  string
	// Tone is the endpoint's grade's colour: good, warn, bad or neutral.
	Tone string
}

// resultView is a ticket's latest test as the app shows it: a verdict an
// engineer acts on, the figures behind it, and the whole answer below.
type resultView struct {
	Verdict     string
	VerdictNote string
	Tone        string
	Running     bool
	TestedBy    string
	TestedEmail string
	When        string
	Download    string
	Upload      string
	Sites       []siteView
	Raw         string
}

// viewResult reads a Results API answer. The verdict is PacketPulse's own
// grading, as it leads a ticket: a breach anywhere is a fault, else a
// degraded endpoint degrades the line, else it is healthy; with nothing
// graded it says so. A result kept only for the report counts for nothing.
func viewResult(body map[string]any) *resultView {
	view := &resultView{Raw: indent(body), TestedBy: text(body["tested_by"]), TestedEmail: text(body["tested_by_email"]),
		Download: number(body["download_mbps"], 1), Upload: number(body["upload_mbps"], 1)}
	if started, err := time.Parse(time.RFC3339, text(body["started_on"])); err == nil {
		view.When = started.In(ist).Format("2 Jan, 15:04")
	}

	tally := map[string]int{}
	rows, _ := body["results"].([]any)
	for _, row := range rows {
		site, _ := row.(map[string]any)
		reached, _ := site["reachable"].(bool)
		grade := text(site["sla_status"])
		if reportOnly, _ := site["report_only"].(bool); !reportOnly {
			tally[grade]++
		}
		view.Sites = append(view.Sites, siteView{
			Name: text(site["site_name"]), Reached: reached,
			Loss: number(site["packet_loss"], 1), Latency: number(site["avg_rtt_ms"], 0),
			Jitter: number(site["jitter_ms"], 1), Tone: gradeTone(grade),
		})
	}

	switch status := text(body["status"]); {
	case status != "completed" && status != "failed":
		view.Running = true
		view.Verdict, view.Tone, view.VerdictNote = "Still testing", "info", "PacketPulse has not finished this test. Refresh in a moment."
	case tally["breach"] > 0:
		view.Verdict, view.Tone, view.VerdictNote = "Fault on the line", "bad", "At least one endpoint is outside its targets."
	case tally["degraded"] > 0:
		view.Verdict, view.Tone, view.VerdictNote = "Degraded", "warn", "Every endpoint answered, but not every one within its targets."
	case tally["ok"] > 0:
		view.Verdict, view.Tone, view.VerdictNote = "Healthy", "good", "Every endpoint within its targets."
	default:
		view.Verdict, view.Tone, view.VerdictNote = "Not graded", "neutral", "PacketPulse has no targets to grade this test against."
	}
	return view
}

func gradeTone(grade string) string {
	switch grade {
	case "ok":
		return "good"
	case "degraded":
		return "warn"
	case "breach":
		return "bad"
	default:
		return "neutral"
	}
}

// text is a JSON value as text, empty for null.
func text(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

// speedTile is one way of the line's speed as a tile shows it: the way, and
// its figure, empty where PacketPulse filed none.
type speedTile struct{ Way, Figure string }

// newSpeedTile is the template's "speed": both ways are shown once either has
// a figure, so a way PacketPulse did not record reads as such, not as missing.
func newSpeedTile(way, figure string) speedTile { return speedTile{Way: way, Figure: figure} }

// The line's speed as dials (19 November 2026), as PacketPulse draws them:
// subtle in a report, each resting at the figure filed. speedMarks are the
// marks, in Mbit/s - nought to the 500 a reflector measures at most - each
// step between two the same share of the sweep, so a 2 Mbit/s line moves
// the needle as plainly as a 300 Mbit/s one. The page script draws the live
// card's dials by the same numbers (TestTheDialsAreDrawnAlikeInPageAndScript).
var speedMarks = []float64{0, 1, 5, 10, 25, 50, 100, 250, 500}

// A dial's geometry in its 200 x 150 view: a sweep of 240 degrees from lower
// left, over the top, to lower right, about its hub.
const (
	dialCenterX, dialCenterY = 100.0, 86.0
	dialRadius               = 70.0
	dialStart, dialSweep     = 150.0, 240.0
)

// dialPosition is how far round the sweep a speed sits, from 0 to 1.
func dialPosition(mbps float64) float64 {
	if mbps <= speedMarks[0] {
		return 0
	}
	for i := 1; i < len(speedMarks); i++ {
		if mbps <= speedMarks[i] {
			within := (mbps - speedMarks[i-1]) / (speedMarks[i] - speedMarks[i-1])
			return (float64(i-1) + within) / float64(len(speedMarks)-1)
		}
	}
	return 1
}

// dialPoint is the point a share of the way round the sweep, at a radius.
func dialPoint(share, radius float64) (float64, float64) {
	angle := (dialStart + dialSweep*share) * math.Pi / 180
	return dialCenterX + radius*math.Cos(angle), dialCenterY + radius*math.Sin(angle)
}

// dialArc is the SVG path of the sweep from its start to a share of it.
func dialArc(share float64) string {
	x0, y0 := dialPoint(0, dialRadius)
	x1, y1 := dialPoint(share, dialRadius)
	large := 0
	if dialSweep*share > 180 {
		large = 1
	}
	return fmt.Sprintf("M%.1f,%.1f A%.0f,%.0f 0 %d 1 %.1f,%.1f", x0, y0, dialRadius, dialRadius, large, x1, y1)
}

// speedDial is one way of the line's speed as a dial draws it: its sweep,
// the part swept to the figure, its ticks and its needle - at nought, with
// a dash, for a way PacketPulse filed no figure for.
type speedDial struct {
	Way, Figure      string
	Upload           bool
	Track, Arc       string
	NeedleX, NeedleY string
	Ticks            []dialTick
}

// dialTick is one mark's tick, inside the sweep.
type dialTick struct{ X1, Y1, X2, Y2 string }

// newSpeedDial is the template's "dial".
func newSpeedDial(way, figure string, upload bool) speedDial {
	share := 0.0
	if mbps, err := strconv.ParseFloat(figure, 64); err == nil {
		share = dialPosition(mbps)
	}
	dial := speedDial{Way: way, Figure: figure, Upload: upload, Track: dialArc(1)}
	if share > 0 {
		dial.Arc = dialArc(share)
	}
	x, y := dialPoint(share, dialRadius-26)
	dial.NeedleX, dial.NeedleY = fmt.Sprintf("%.1f", x), fmt.Sprintf("%.1f", y)
	for i := range speedMarks {
		at := float64(i) / float64(len(speedMarks)-1)
		x1, y1 := dialPoint(at, dialRadius-9)
		x2, y2 := dialPoint(at, dialRadius-15)
		dial.Ticks = append(dial.Ticks, dialTick{
			fmt.Sprintf("%.1f", x1), fmt.Sprintf("%.1f", y1), fmt.Sprintf("%.1f", x2), fmt.Sprintf("%.1f", y2),
		})
	}
	return dial
}

// number is a JSON number to so many places, empty for null or not a number.
func number(value any, places int) string {
	figure, ok := value.(float64)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%.*f", places, figure)
}

func indent(body map[string]any) string {
	keys := make([]string, 0, len(body))
	for key := range body {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&out, "%s: %v\n", key, body[key])
	}
	return out.String()
}
