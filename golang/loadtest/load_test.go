// Package loadtest measures the platform under the load one licensed
// organisation puts on it: twenty people signing in at once, reading and
// filing the tests their devices ran side by side, and the machines that pull
// results out of it.
//
// It is opt-in - `./packetpulsetest.sh load` - and not part of the default run.
// It takes minutes, and its budgets describe the machine it runs on as much
// as the code, so a slow laptop failing it says something different from a
// regression failing it. The budgets are set from measured runs with room to
// spare (docs/PacketPulseTestStrategy.md §4.7 records them); what they catch is
// the pathological - a lock held across a request, a query that scans a
// partition per row, a sweep that drops what it measured.
//
// Three properties are absolute, whatever the machine: no request answers
// 5xx, every result a test files is stored, and everything stored can be
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

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// The licence size the plan is sold at, and the inventory it allows.
const (
	people = 20
	sites  = 200
)

// largestFiling is the most results one filed test may hold: the server's
// diagnosticconstants.MaxClientObservations.
const largestFiling = 16

// deviceFiling is a test a device ran against one endpoint, with as many
// results as asked, as Run diagnostic files it.
func deviceFiling(customerId, ttNumber, dnsSiteId string, results int) map[string]any {
	observations := make([]map[string]any, 0, results)
	for index := range results {
		observations = append(observations, map[string]any{
			"target_label": fmt.Sprintf("Load target %d", index+1), "target_address": "203.0.113.1", "method": "icmp",
			"packets_sent": 25, "packets_recv": 25, "packet_loss": 0,
			"min_rtt_ms": 12.0, "avg_rtt_ms": 20.0, "max_rtt_ms": 45.0, "jitter_ms": 2.5,
		})
	}
	return map[string]any{
		"customer_id": customerId, "tt_number": ttNumber, "dns_site_id": dnsSiteId,
		"sample_count": 25, "timeout_ms": 2000, "observations": observations,
	}
}

// Budgets, as the 95th percentile of each kind of request. Measured on a
// developer laptop against a local server and database; see the strategy
// document for the figures they were set from.
const (
	budgetPasswordStep = 3 * time.Second // bcrypt, twenty at once
	budgetSecondStep   = 1 * time.Second
	budgetReads        = 750 * time.Millisecond
	budgetSubmit       = 20 * time.Second // a filing, ten people at once
	budgetResultsApi   = 1 * time.Second  // a ticket of sixteen results
	budgetCsv          = 10 * time.Second
	// A network report of a busy day and of its year, every test listed
	// (customer request, 28 October 2026: as tests grow, a report reads its
	// period, not everything an organisation has run).
	budgetReport = 5 * time.Second
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
	request, err := http.NewRequest(method, packetpulsetest.BaseURL()+"/api/v1"+path, body)
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
// nineteen engineers, each signed in once already, and two hundred
// sites. Provisioning is not measured - it is how the stage is set.
func organisation(t *testing.T) ([]person, []string) {
	t.Helper()
	ownerToken, _ := packetpulsetest.SignUpOrganisationWithSeats(t, "load", people+5)

	roles := packetpulsetest.Call(t, http.MethodGet, "/staff/role/list", ownerToken, nil)
	engineerRole := ""
	for _, role := range packetpulsetest.ListOf(t, roles, "roles") {
		if role["role_name"] == "NOC Engineer" {
			engineerRole, _ = role["role_id"].(string)
		}
	}
	if engineerRole == "" {
		t.Fatal("there is no NOC Engineer role to give the engineers")
	}

	owner := packetpulsetest.Call(t, http.MethodGet, "/user/me", ownerToken, nil)
	// The harness signs every owner up with this password.
	team := []person{{email: owner.String("email"), password: "PacketPulseTest2026x", token: ownerToken}}
	stamp := time.Now().UnixNano()
	for index := 1; index < people; index++ {
		engineer := person{
			email:    fmt.Sprintf("load-%d-%02d@packetpulsetest.local", stamp, index),
			password: "LoadEngineer2026",
		}
		added := packetpulsetest.Call(t, http.MethodPost, "/user/add", ownerToken, map[string]any{
			"email": engineer.email, "password": engineer.password,
			"full_name": fmt.Sprintf("Engineer %02d", index), "role_id": engineerRole,
		})
		if added.Status != http.StatusCreated {
			t.Fatalf("adding engineer %d: %d %s", index, added.Status, added.Raw)
		}
		engineer.token = packetpulsetest.SignIn(t, engineer.email, engineer.password)
		team = append(team, engineer)
	}

	// Two hundred black-holed addresses (TEST-NET-3): every probe times out,
	// the slowest sweep there is - and the load that once lost a sweep's
	// results when the work after probing ran past the sweep's deadline.
	var paste strings.Builder
	for index := 1; index <= sites; index++ {
		fmt.Fprintf(&paste, "Load site %03d=203.0.113.%d\n", index, index)
	}
	imported := packetpulsetest.Call(t, http.MethodPost, "/dnssite/bulkimport", ownerToken, map[string]any{
		"entries": paste.String(), "region": "Load",
	})
	if imported.Status != http.StatusOK {
		t.Fatalf("importing the sites: %d %s", imported.Status, imported.Raw)
	}
	listed := packetpulsetest.Call(t, http.MethodGet, "/dnssite/list", ownerToken, nil)
	siteIds := make([]string, 0, sites)
	for _, site := range packetpulsetest.ListOf(t, listed, "dns_sites") {
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
	detail := packetpulsetest.Call(t, http.MethodGet, "/diagnostic/"+requestId, token, nil)
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
	packetpulsetest.RequireServer(t)
	started := time.Now().UTC().Add(-time.Minute)
	team, siteIds := organisation(t)
	owner := team[0]

	t.Run("twenty people sign in at once", func(t *testing.T) {
		passwordStep := &meter{name: "sign-in: password", budget: budgetPasswordStep}
		secondStep := &meter{name: "sign-in: code", budget: budgetSecondStep}
		everyone := team
		challenges := make([]string, len(everyone))
		together(len(everyone), len(everyone), func(index int) {
			challenges[index] = passwordStep.do(http.MethodPost, "/user/signin", "", map[string]any{
				"email": everyone[index].email, "password": everyone[index].password,
			}).string("challenge_id")
		})
		// The code each person reads from their inbox, read before the clock
		// starts, as it is in life.
		codes := make([]string, len(everyone))
		for index := range everyone {
			codes[index] = packetpulsetest.CodeFor(t, challenges[index])
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

	// A test is filed from a person's device; the server's own sweep is
	// switched off (decision of 30 October 2026). The largest filing holds as
	// many results as one test can (the server's MaxClientObservations).
	var sweepTicket, sweepId string
	t.Run("the largest filing keeps every result", func(t *testing.T) {
		largest := &meter{name: "filing: 16 results", budget: budgetSubmit}
		sweepTicket = fmt.Sprintf("TT-LOAD-%d", time.Now().UnixNano())
		submitted := largest.do(http.MethodPost, "/diagnostic/clientobservation", owner.token,
			deviceFiling("CUST-LOAD", sweepTicket, siteIds[0], largestFiling))
		largest.judge(t)
		if submitted.status != http.StatusCreated {
			t.Fatalf("the filing: %d %s", submitted.status, submitted.raw)
		}
		sweepId = requestIdOf(submitted)
		if stored := storedResults(t, owner.token, sweepId); stored != largestFiling {
			t.Errorf("%d results are stored, want all %d", stored, largestFiling)
		}
	})

	var submitIds []string
	t.Run("ten people file tests at once", func(t *testing.T) {
		submits := &meter{name: "filing: 5 results", budget: budgetSubmit}
		const submitters, each = 10, 5
		ids := make([]string, submitters)
		together(submitters, submitters, func(index int) {
			who := team[index%len(team)]
			ids[index] = requestIdOf(submits.do(http.MethodPost, "/diagnostic/clientobservation", who.token,
				deviceFiling("CUST-LOAD", fmt.Sprintf("TT-LOAD-%d-%d", index, time.Now().UnixNano()), siteIds[index], each)))
		})
		submits.judge(t)
		for index, id := range ids {
			if id == "" {
				continue // already counted as refused
			}
			if stored := storedResults(t, owner.token, id); stored != each {
				t.Errorf("filing %d stored %d results, want %d", index, stored, each)
			}
		}
		submitIds = ids
	})

	t.Run("a ticketing system pulls results steadily", func(t *testing.T) {
		if sweepTicket == "" {
			t.Skip("the filing did not run")
		}
		issued := packetpulsetest.Call(t, http.MethodPost, "/apikey/add", owner.token, map[string]any{"label": "load"})
		if issued.Status != http.StatusCreated {
			t.Fatalf("issuing a key: %d %s", issued.Status, issued.Raw)
		}
		key := issued.String("api_key")
		pulls := &meter{name: "results API: 16 results", budget: budgetResultsApi}
		together(200, 10, func(int) {
			pulls.do(http.MethodGet, "/result/bytt/"+sweepTicket, key, nil)
		})
		pulls.judge(t)

		// Everything stored comes back out as CSV: the largest filing's
		// results and each of the ten filings', one row each.
		export := &meter{name: "CSV export of the run", budget: budgetCsv}
		pulled := export.do(http.MethodGet,
			"/result/export.csv?from="+url.QueryEscape(started.Format(time.RFC3339)), key, nil)
		export.judge(t)
		records, err := csv.NewReader(bytes.NewReader(pulled.raw)).ReadAll()
		if err != nil {
			t.Fatalf("the export is not CSV: %v", err)
		}
		want := largestFiling
		for _, id := range submitIds {
			if id != "" {
				want += 5
			}
		}
		if rows := len(records) - 1; rows != want {
			t.Errorf("the export holds %d rows, want %d - every stored result once", rows, want)
		}
	})

	// The team files a busy day of device tests - a thousand tickets of four
	// results each - and the network report of the day and of the year reads
	// them within its budget, every test accounted for: listed under its
	// city, counted once, and every one in the CSV.
	t.Run("a busy day's network report lists every test", func(t *testing.T) {
		const tickets, each = 1000, 4
		files := &meter{name: "device tests filed", budget: budgetSubmit}
		stamp := time.Now().UnixNano()
		together(tickets, len(team), func(index int) {
			observations := make([]map[string]any, 0, each)
			for range each {
				observations = append(observations, map[string]any{
					"target_label": "Load site", "target_address": "203.0.113.1", "method": "icmp",
					"packets_sent": 10, "packets_recv": 10 - index%2, "packet_loss": float64(index%2) * 10,
					"min_rtt_ms": 12.0, "avg_rtt_ms": 20.0 + float64(index%7), "max_rtt_ms": 45.0, "jitter_ms": 2.5,
				})
			}
			files.do(http.MethodPost, "/diagnostic/clientobservation", team[index%len(team)].token, map[string]any{
				"customer_id": "CUST-LOAD-DAY", "tt_number": fmt.Sprintf("TT-DAY-%d-%04d", stamp, index),
				"sample_count": 25, "timeout_ms": 2000, "dns_site_id": siteIds[index%len(siteIds)],
				"observations": observations,
			})
		})
		files.judge(t)

		reports := &meter{name: "network report: day and year", budget: budgetReport}
		for _, period := range []string{"day", "year"} {
			read := reports.do(http.MethodGet, "/report/network?period="+period, owner.token, nil)
			sites, _ := read.body["sites"].([]any)
			listed := 0
			for _, site := range sites {
				count, _ := site.(map[string]any)["entry_count"].(float64)
				listed += int(count)
			}
			if listed < tickets*each {
				t.Errorf("the %s's report lists %d tests, want at least the %d filed", period, listed, tickets*each)
			}
			csvFile := reports.do(http.MethodGet, "/report/network.csv?period="+period, owner.token, nil)
			records, err := csv.NewReader(bytes.NewReader(csvFile.raw)).ReadAll()
			if err != nil {
				t.Fatalf("the %s's CSV is not CSV: %v", period, err)
			}
			tests := 0
			for _, record := range records[1:] {
				if len(record) > 3 && record[3] == "test" {
					tests++
				}
			}
			if tests != listed {
				t.Errorf("the %s's CSV holds %d tests, the report %d", period, tests, listed)
			}
		}
		reports.judge(t)
	})
}
