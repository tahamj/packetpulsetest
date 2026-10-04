// Package loadtest measures the platform under the load one licensed
// organisation puts on it: twenty people signing in at once, reading and
// running diagnostics side by side, a sweep of two hundred sites, and the
// machines that pull results out of it.
//
// It is opt-in - `./pingletest.sh load` - and not part of the default run.
// It takes minutes, and its budgets describe the machine it runs on as much
// as the code, so a slow laptop failing it says something different from a
// regression failing it. The budgets are set from measured runs with room to
// spare (docs/PingleTestStrategy.md §4.7 records them); what they catch is
// the pathological - a lock held across a request, a query that scans a
// partition per row, a sweep that drops what it measured.
//
// Three properties are absolute, whatever the machine: no request answers
// 5xx, every result a sweep measures is stored, and everything stored can be
// pulled back out.
package loadtest

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	pingletest "github.com/tahamj/pingletest"
)

// The licence size the plan is sold at, and the inventory it allows.
const (
	people = 20
	sites  = 200
)

// Budgets, as the 95th percentile of each kind of request. Measured on a
// developer laptop against a local server and database; see the strategy
// document for the figures they were set from.
const (
	budgetPasswordStep = 3 * time.Second // bcrypt, twenty at once
	budgetSecondStep   = 1 * time.Second
	budgetReads        = 750 * time.Millisecond
	budgetSubmit       = 20 * time.Second // five sites, ten people at once
	budgetSweep        = 120 * time.Second
	budgetResultsApi   = 1 * time.Second // a ticket of two hundred results
	budgetCsv          = 10 * time.Second
)

var client = &http.Client{
	Timeout:   3 * time.Minute,
	Transport: &http.Transport{MaxIdleConnsPerHost: 64, MaxConnsPerHost: 64},
}

// sample is one request as it was measured.
type sample struct {
	took   time.Duration
	status int
	err    error
}

// meter collects one scenario's samples. Requests run on many goroutines, so
// it records; the test goroutine judges.
type meter struct {
	name    string
	budget  time.Duration
	mu      sync.Mutex
	samples []sample
}

type reply struct {
	status int
	raw    []byte
	body   map[string]any
}

func (r reply) string(key string) string {
	value, _ := r.body[key].(string)
	return value
}

// do makes one request and times it. It never fails the test itself: it may
// run on any goroutine, and t.Fatal belongs to the test's own.
func (m *meter) do(method, path, token string, payload any) reply {
	var body io.Reader
	if payload != nil {
		encoded, _ := json.Marshal(payload)
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, pingletest.BaseURL()+"/api/v1"+path, body)
	if err != nil {
		m.record(sample{err: err})
		return reply{}
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		m.record(sample{took: time.Since(started), err: err})
		return reply{}
	}
	raw, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	took := time.Since(started)
	m.record(sample{took: took, status: response.StatusCode, err: err})
	decoded := map[string]any{}
	_ = json.Unmarshal(raw, &decoded)
	return reply{status: response.StatusCode, raw: raw, body: decoded}
}

func (m *meter) record(s sample) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.samples = append(m.samples, s)
}

// percentile is the nearest-rank percentile of the durations.
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	return sorted[max(rank, 0)]
}

// judge fails the test for any transport error, any 5xx, any answer outside
// 2xx, or a 95th percentile over budget - and logs the figures either way,
// because a passing run's numbers are what the next budget is set from.
func (m *meter) judge(t *testing.T) {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	durations := make([]time.Duration, 0, len(m.samples))
	serverErrors, refused, failed := 0, 0, 0
	for _, s := range m.samples {
		switch {
		case s.err != nil:
			failed++
		case s.status >= 500:
			serverErrors++
		case s.status < 200 || s.status >= 300:
			refused++
		}
		durations = append(durations, s.took)
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p50, p95 := percentile(durations, 50), percentile(durations, 95)
	t.Logf("%-28s n=%-4d p50=%-9v p95=%-9v max=%-9v budget=%v", m.name, len(durations),
		p50.Round(time.Millisecond), p95.Round(time.Millisecond),
		percentile(durations, 100).Round(time.Millisecond), m.budget)
	if len(durations) == 0 {
		t.Errorf("%s: nothing was measured", m.name)
	}
	if failed > 0 {
		t.Errorf("%s: %d requests failed in transport", m.name, failed)
	}
	if serverErrors > 0 {
		t.Errorf("%s: %d requests answered 5xx", m.name, serverErrors)
	}
	if refused > 0 {
		t.Errorf("%s: %d requests were refused", m.name, refused)
	}
	if p95 > m.budget {
		t.Errorf("%s: p95 %v is over its budget of %v", m.name, p95, m.budget)
	}
}

// together runs work for each index at once, at most limit at a time.
func together(count, limit int, work func(index int)) {
	semaphore := make(chan struct{}, limit)
	var group sync.WaitGroup
	for index := range count {
		group.Add(1)
		go func() {
			defer group.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			work(index)
		}()
	}
	group.Wait()
}

type person struct {
	email, password, token string
}

// organisation provisions what one licensed customer has: its owner and
// nineteen engineers, each with an authenticator set up, and two hundred
// sites. Provisioning is not measured - it is how the stage is set.
func organisation(t *testing.T) ([]person, []string) {
	t.Helper()
	ownerToken, _ := pingletest.SignUpOrganisationWithSeats(t, "load", people+5)

	roles := pingletest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	engineerRole := ""
	for _, role := range pingletest.ListOf(t, roles, "roles") {
		if role["role_name"] == "NOC Engineer" {
			engineerRole, _ = role["role_id"].(string)
		}
	}
	if engineerRole == "" {
		t.Fatal("there is no NOC Engineer role to give the engineers")
	}

	owner := pingletest.Call(t, http.MethodGet, "/user/me", ownerToken, nil)
	// The harness signs every owner up with this password.
	team := []person{{email: owner.String("email"), password: "PingleTest2026x", token: ownerToken}}
	stamp := time.Now().UnixNano()
	for index := 1; index < people; index++ {
		engineer := person{
			email:    fmt.Sprintf("load-%d-%02d@pingletest.local", stamp, index),
			password: "LoadEngineer2026",
		}
		added := pingletest.Call(t, http.MethodPost, "/user/add", ownerToken, map[string]any{
			"email": engineer.email, "password": engineer.password,
			"full_name": fmt.Sprintf("Engineer %02d", index), "role_id": engineerRole,
		})
		if added.Status != http.StatusCreated {
			t.Fatalf("adding engineer %d: %d %s", index, added.Status, added.Raw)
		}
		engineer.token = pingletest.SignIn(t, engineer.email, engineer.password)
		team = append(team, engineer)
	}

	// Two hundred black-holed addresses (TEST-NET-3): every probe times out
	// and every failure is traced, which is the load that once lost a
	// sweep's results when tracing ran past the sweep's deadline.
	var paste strings.Builder
	for index := 1; index <= sites; index++ {
		fmt.Fprintf(&paste, "Load site %03d=203.0.113.%d\n", index, index)
	}
	imported := pingletest.Call(t, http.MethodPost, "/dnssite/bulkimport", ownerToken, map[string]any{
		"entries": paste.String(), "region": "Load",
	})
	if imported.Status != http.StatusOK {
		t.Fatalf("importing the sites: %d %s", imported.Status, imported.Raw)
	}
	listed := pingletest.Call(t, http.MethodGet, "/dnssite/list", ownerToken, nil)
	siteIds := make([]string, 0, sites)
	for _, site := range pingletest.ListOf(t, listed, "dns_sites") {
		if id, ok := site["dns_site_id"].(string); ok {
			siteIds = append(siteIds, id)
		}
	}
	if len(siteIds) != sites {
		t.Fatalf("the organisation has %d sites, want %d", len(siteIds), sites)
	}
	return team, siteIds
}

// results reads a diagnostic back and answers how many results are stored.
func storedResults(t *testing.T, token, requestId string) int {
	t.Helper()
	detail := pingletest.Call(t, http.MethodGet, "/diagnostic/"+requestId, token, nil)
	if detail.Status != http.StatusOK {
		t.Fatalf("reading diagnostic %s back: %d %s", requestId, detail.Status, detail.Raw)
	}
	stored, _ := detail.Body["results"].([]any)
	return len(stored)
}

func requestIdOf(r reply) string {
	request, _ := r.body["request"].(map[string]any)
	id, _ := request["request_id"].(string)
	return id
}

func TestTheLicensedLoad(t *testing.T) {
	pingletest.RequireServer(t)
	started := time.Now().UTC().Add(-time.Minute)
	team, siteIds := organisation(t)
	owner := team[0]

	t.Run("twenty people sign in at once", func(t *testing.T) {
		passwordStep := &meter{name: "sign-in: password", budget: budgetPasswordStep}
		secondStep := &meter{name: "sign-in: authenticator", budget: budgetSecondStep}
		everyone := team
		challenges := make([]string, len(everyone))
		together(len(everyone), len(everyone), func(index int) {
			challenges[index] = passwordStep.do(http.MethodPost, "/user/signin", "", map[string]any{
				"email": everyone[index].email, "password": everyone[index].password,
			}).string("challenge_id")
		})
		// The code each person reads off their phone: a fresh one, read
		// before the clock starts, as it is in life.
		codes := make([]string, len(everyone))
		for index := range everyone {
			codes[index] = pingletest.NextCode(t, everyone[index].email)
		}
		together(len(everyone), len(everyone), func(index int) {
			verified := secondStep.do(http.MethodPost, "/user/signin/verify", "", map[string]any{
				"challenge_id": challenges[index], "code": codes[index],
			})
			if token := verified.string("token"); token != "" {
				everyone[index].token = token
			}
		})
		passwordStep.judge(t)
		secondStep.judge(t)
	})

	t.Run("everyone reads at once", func(t *testing.T) {
		reads := &meter{name: "reads", budget: budgetReads}
		paths := []string{"/diagnostic/list", "/dnssite/list", "/ping/dashboard", "/user/me"}
		together(len(team)*10, len(team), func(index int) {
			who := team[index%len(team)]
			reads.do(http.MethodGet, paths[index%len(paths)], who.token, nil)
		})
		reads.judge(t)
	})

	var sweepTicket, sweepId string
	t.Run("a sweep of two hundred sites keeps every result", func(t *testing.T) {
		sweep := &meter{name: "sweep: 200 sites, traced", budget: budgetSweep}
		sweepTicket = fmt.Sprintf("TT-LOAD-%d", time.Now().UnixNano())
		submitted := sweep.do(http.MethodPost, "/diagnostic/submit", owner.token, map[string]any{
			"customer_id": "CUST-LOAD", "tt_number": sweepTicket,
			"packet_count": 1, "timeout_ms": 1000, "trace_failures": true,
		})
		sweep.judge(t)
		if submitted.status != http.StatusCreated {
			t.Fatalf("the sweep: %d %s", submitted.status, submitted.raw)
		}
		sweepId = requestIdOf(submitted)
		answered, _ := submitted.body["results"].([]any)
		if len(answered) != sites {
			t.Errorf("the sweep answered %d results, want %d", len(answered), sites)
		}
		// Stored, not just answered: tracing two hundred dead addresses runs
		// past its budget, and a sweep must not lose what it measured to that.
		if stored := storedResults(t, owner.token, sweepId); stored != sites {
			t.Errorf("%d results are stored, want all %d", stored, sites)
		}
	})

	var submitIds []string
	t.Run("ten people run diagnostics at once", func(t *testing.T) {
		submits := &meter{name: "submit: 5 sites", budget: budgetSubmit}
		const submitters, each = 10, 5
		ids := make([]string, submitters)
		together(submitters, submitters, func(index int) {
			who := team[1+index]
			chosen := siteIds[index*each : (index+1)*each]
			ids[index] = requestIdOf(submits.do(http.MethodPost, "/diagnostic/submit", who.token, map[string]any{
				"customer_id": "CUST-LOAD", "tt_number": fmt.Sprintf("TT-LOAD-%d-%d", index, time.Now().UnixNano()),
				"dns_site_ids": chosen, "packet_count": 1, "timeout_ms": 1000,
			}))
		})
		submits.judge(t)
		for index, id := range ids {
			if id == "" {
				continue // already counted as refused
			}
			if stored := storedResults(t, owner.token, id); stored != each {
				t.Errorf("diagnostic %d stored %d results, want %d", index, stored, each)
			}
		}
		submitIds = ids
	})

	t.Run("a ticketing system pulls results steadily", func(t *testing.T) {
		if sweepTicket == "" {
			t.Skip("the sweep did not run")
		}
		issued := pingletest.Call(t, http.MethodPost, "/apikey/add", owner.token, map[string]any{"label": "load"})
		if issued.Status != http.StatusCreated {
			t.Fatalf("issuing a key: %d %s", issued.Status, issued.Raw)
		}
		key := issued.String("api_key")
		pulls := &meter{name: "results API: 200 results", budget: budgetResultsApi}
		together(200, 10, func(int) {
			pulls.do(http.MethodGet, "/result/bytt/"+sweepTicket, key, nil)
		})
		pulls.judge(t)

		// Everything stored comes back out as CSV: the sweep's results and
		// each of the ten diagnostics', one row each.
		export := &meter{name: "CSV export of the run", budget: budgetCsv}
		pulled := export.do(http.MethodGet,
			"/result/export.csv?from="+url.QueryEscape(started.Format(time.RFC3339)), key, nil)
		export.judge(t)
		records, err := csv.NewReader(bytes.NewReader(pulled.raw)).ReadAll()
		if err != nil {
			t.Fatalf("the export is not CSV: %v", err)
		}
		want := sites
		for _, id := range submitIds {
			if id != "" {
				want += 5
			}
		}
		if rows := len(records) - 1; rows != want {
			t.Errorf("the export holds %d rows, want %d - every stored result once", rows, want)
		}
	})
}
