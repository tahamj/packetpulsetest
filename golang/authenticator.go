package pingletest

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// The authenticator app every account in the suite "carries".
//
// Written here, from the standard library, rather than imported from
// pinglego: a bug in the server's implementation must not be mirrored by the
// suite that is meant to catch it. It is checked against the RFC 6238 vectors
// in authenticator_test.go.

// authenticator is what a person's phone holds: the secret their app was set
// up with. The steps already spent are kept in a shared file - see NextCode.
type authenticator struct {
	key []byte
}

var (
	authenticatorsMu sync.Mutex
	authenticators   = map[string]*authenticator{}
)

const stepSeconds = 30

func accountKey(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

// rememberAuthenticator records the secret an account's app was set up with.
func rememberAuthenticator(email, secret string) error {
	cleaned := strings.ToUpper(strings.NewReplacer(" ", "", "=", "").Replace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(cleaned)
	if err != nil || len(key) == 0 {
		return fmt.Errorf("not a base32 authenticator secret")
	}
	authenticatorsMu.Lock()
	defer authenticatorsMu.Unlock()
	authenticators[accountKey(email)] = &authenticator{key: key}
	return nil
}

func knowsAuthenticator(email string) bool {
	authenticatorsMu.Lock()
	defer authenticatorsMu.Unlock()
	_, ok := authenticators[accountKey(email)]
	return ok
}

// hotp is RFC 4226, six digits.
func hotp(key []byte, counter uint64) string {
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(message[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", value%1_000_000)
}

// NextCode is the next unspent code from an account's app.
//
// The server accepts the step before and after its own, so a burst of three
// sign-ins needs no wait; a fourth within the same half minute waits for the
// next step, exactly as a person would.
//
// Spent steps are recorded in a file shared by every test process, because
// each suite is its own process and the server remembers what was spent
// across all of them: a fresh process starting again at "now - 1" replays a
// code the previous suite already used, and is rightly refused.
func NextCode(t *testing.T, email string) string {
	t.Helper()

	authenticatorsMu.Lock()
	app, ok := authenticators[accountKey(email)]
	authenticatorsMu.Unlock()
	if !ok {
		t.Fatalf("no authenticator is known for %s", email)
	}

	step, err := claimStep(email, time.Now().Unix()/stepSeconds-1)
	if err != nil {
		t.Fatalf("recording the spent authenticator step for %s: %v", email, err)
	}
	// A step more than one ahead of the server's clock is not accepted yet.
	if wait := time.Until(time.Unix((step-1)*stepSeconds, 0)); wait > 0 {
		time.Sleep(wait + 200*time.Millisecond)
	}
	return hotp(app.key, uint64(step))
}

// claimStep spends the first step at or after earliest that no test process
// has spent for this account, under an exclusive lock on the account's file.
func claimStep(email string, earliest int64) (int64, error) {
	digest := sha256.Sum256([]byte(accountKey(email)))
	directory := filepath.Join(os.TempDir(), "pingletest-authenticators")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return 0, err
	}
	file, err := os.OpenFile(filepath.Join(directory, hex.EncodeToString(digest[:8])), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return 0, err
	}
	defer func() { _ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN) }()

	recorded, _ := io.ReadAll(file)
	spent, _ := strconv.ParseInt(strings.TrimSpace(string(recorded)), 10, 64)
	step := max(earliest, spent+1)
	if err := file.Truncate(0); err != nil {
		return 0, err
	}
	if _, err := file.WriteAt([]byte(strconv.FormatInt(step, 10)), 0); err != nil {
		return 0, err
	}
	return step, nil
}
