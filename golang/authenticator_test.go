package pingletest

import (
	"encoding/base32"
	"fmt"
	"strings"
	"testing"
	"time"
)

// The suite's authenticator must compute what every real app computes, or
// every sign-in in every suite fails for a reason that is the harness's own.
func TestTheSuitesAuthenticatorMatchesTheRfcVectors(t *testing.T) {
	key := []byte("12345678901234567890")
	for counter, want := range []string{"755224", "287082", "359152", "969429", "338314"} {
		if got := hotp(key, uint64(counter)); got != want {
			t.Errorf("HOTP(%d) = %s, want %s", counter, got, want)
		}
	}
	// RFC 6238 at T=59: step 1, whose six-digit code is 287082.
	if got := hotp(key, uint64(59/stepSeconds)); got != "287082" {
		t.Errorf("TOTP at 59 = %s, want 287082", got)
	}
}

// Each code is spent once: a burst gets three without waiting (the server
// accepts one step either side of its own), and never the same step twice.
func TestCodesAreNeverReusedAndABurstNeedsNoWait(t *testing.T) {
	// A fresh account per run: spent steps are shared between runs.
	account := fmt.Sprintf("burst-%d@pingletest.local", time.Now().UnixNano())
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("an-authenticator-key"))
	if err := rememberAuthenticator(strings.ToUpper(account), secret); err != nil {
		t.Fatal(err)
	}
	if err := rememberAuthenticator(account, "!!"); err == nil {
		t.Error("a secret that is not base32 was accepted")
	}

	started := time.Now()
	seen := map[string]bool{}
	for range 3 {
		code := NextCode(t, account)
		if seen[code] {
			t.Errorf("code %s was handed out twice", code)
		}
		seen[code] = true
	}
	if time.Since(started) > 2*time.Second {
		t.Errorf("three codes took %v, want no wait", time.Since(started))
	}
	if !knowsAuthenticator(strings.ToUpper(account)) || knowsAuthenticator("nobody@pingletest.local") {
		t.Error("authenticators are not matched by address, case-insensitively")
	}
}
