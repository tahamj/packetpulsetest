package apicontract

import (
	"net/http"
	"testing"
	"time"

	packetpulsetest "github.com/tahamj/packetpulsetest"
)

// A session left unused ends (request of 2 November 2026: thirty minutes of
// inactivity for everyone), and the app asks the server how long its session
// has left before it returns to sign-in, so a tab left alone never signs out
// someone working in another tab of the same session.
//
// The server under test runs with the idle timeout on, as it is by default.
// Over HTTP, against the server as it boots: the session tells the app the
// timeout, asking how long is left is not a use of the session - the time
// runs down - and any other request is, so the time starts again.
func TestASessionsIdleTimeRunsDownUntilItIsUsed(t *testing.T) {
	packetpulsetest.RequireServer(t)
	token, _ := packetpulsetest.SignUpOrganisation(t, "idletime")

	me := packetpulsetest.Call(t, http.MethodGet, "/user/me", token, nil)
	idle, _ := me.Body["idle_timeout_seconds"].(float64)
	if me.Status != http.StatusOK || idle < 60 {
		t.Fatalf("/user/me = %d, idle_timeout_seconds %v; want the server's idle timeout, on as by default: %s",
			me.Status, me.Body["idle_timeout_seconds"], me.Raw)
	}

	ask := func() float64 {
		t.Helper()
		answer := packetpulsetest.Call(t, http.MethodGet, "/staff/session/current", token, nil)
		left, _ := answer.Body["seconds_left"].(float64)
		if answer.Status != http.StatusOK || answer.Body["idle_timeout_seconds"] != idle || left <= 0 || left > idle {
			t.Fatalf("/staff/session/current = %d %s, want up to %v seconds left of the same timeout",
				answer.Status, answer.Raw, idle)
		}
		return left
	}

	first := ask()
	if first < idle-60 {
		t.Errorf("just used, the session has %v of %v seconds left", first, idle)
	}
	time.Sleep(2500 * time.Millisecond)
	second := ask()
	if second > first-2 {
		t.Errorf("%v seconds left, then %v after 2.5 s of only asking: asking kept the session in use", first, second)
	}

	if used := packetpulsetest.Call(t, http.MethodGet, "/user/me", token, nil); used.Status != http.StatusOK {
		t.Fatalf("/user/me = %d %s", used.Status, used.Raw)
	}
	if third := ask(); third < second+2 {
		t.Errorf("%v seconds left before a request and %v after it: the request was not a use of the session", second, third)
	}

	// Without a session there is nothing to ask about.
	if refused := packetpulsetest.Call(t, http.MethodGet, "/staff/session/current", "", nil); refused.Status != http.StatusUnauthorized {
		t.Errorf("asking with no session = %d %s, want 401", refused.Status, refused.Raw)
	}
}
