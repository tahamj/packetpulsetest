package backupconformance

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// testKey is a throwaway gpg key pair standing in for the real backup key,
// whose private half is deliberately nowhere this suite can reach.
type testKey struct {
	home      string
	publicKey string
}

func newTestKey(t *testing.T) testKey {
	t.Helper()
	// A short path on purpose: gpg-agent's socket lives in the home, and a
	// macOS temp path is long enough to exceed the 104-byte socket limit.
	home, err := os.MkdirTemp("/tmp", "pgk")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(home, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = exec.Command("gpgconf", "--homedir", home, "--kill", "gpg-agent").Run()
		_ = os.RemoveAll(home)
	})
	gpg := func(args ...string) []byte {
		out, err := exec.Command("gpg", append([]string{"--homedir", home, "--batch",
			"--pinentry-mode", "loopback", "--passphrase", ""}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("gpg %v: %v\n%s", args, err, out)
		}
		return out
	}
	gpg("--quick-gen-key", "Pingle Backup Test <backup-test@example.invalid>", "default", "default", "never")
	publicKey := filepath.Join(home, "public.asc")
	if err := os.WriteFile(publicKey, gpg("--armor", "--export"), 0o644); err != nil {
		t.Fatal(err)
	}
	return testKey{home: home, publicKey: publicKey}
}

func (k testKey) decrypt(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("gpg", "--homedir", k.home, "--batch", "--quiet",
		"--pinentry-mode", "loopback", "--passphrase", "", "--decrypt", path).Output()
	if err != nil {
		t.Fatalf("cannot decrypt %s with the backup key: %v", filepath.Base(path), err)
	}
	return string(out)
}

// emptyGPGHome is a keyring holding no keys at all: what anyone without the
// backup key has.
func emptyGPGHome(t *testing.T) string {
	t.Helper()
	home, err := os.MkdirTemp("/tmp", "pge")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(home, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = exec.Command("gpgconf", "--homedir", home, "--kill", "gpg-agent").Run()
		_ = os.RemoveAll(home)
	})
	return home
}

// ageBy moves a file's or directory's modification time into the past.
func ageBy(t *testing.T, path string, seconds int) {
	t.Helper()
	then := time.Now().Add(-time.Duration(seconds) * time.Second)
	if err := os.Chtimes(path, then, then); err != nil {
		t.Fatal(err)
	}
}
