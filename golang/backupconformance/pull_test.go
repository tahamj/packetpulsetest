package backupconformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// pullRig runs PacketPulsePullBackups.sh against the fixture's backup root as the
// "server". ssh is replaced by a stand-in that runs the remote command on
// this machine, so rsync, the listing, the verification and the receipts all
// take their real code paths; only the network hop is missing.
type pullRig struct {
	f     fixture
	local string
	env   map[string]string
}

func newPullRig(t *testing.T) pullRig {
	t.Helper()
	f := newFixture(t)
	tools := t.TempDir()
	fakeSSH := filepath.Join(tools, "ssh")
	writeExecutable(t, fakeSSH, `#!/bin/bash
# Drop ssh's options, then the host, and run what is left as the remote shell would.
no_stdin=0
while [ $# -gt 0 ]; do
    case "$1" in
        -i|-o|-l|-p) shift 2 ;;
        -n)          no_stdin=1; shift ;;
        -*)          shift ;;
        *)           break ;;
    esac
done
shift
# Real ssh forwards stdin to the remote side and so consumes it, unless -n.
# rsync's own connection is the exception: its protocol runs over stdin.
case "$*" in
    rsync*) ;;
    *) [ "$no_stdin" -eq 1 ] || cat > /dev/null ;;
esac
exec /bin/bash -c "$*"
`)
	key := filepath.Join(tools, "id_test")
	writeEmpty(t, key)
	local := filepath.Join(t.TempDir(), "offsite")
	return pullRig{f: f, local: local, env: map[string]string{
		"PACKETPULSE_BACKUP_SSH":           fakeSSH,
		"PACKETPULSE_BACKUP_SSH_KEY":       key,
		"PACKETPULSE_BACKUP_HOST":          "backup-host.invalid",
		"PACKETPULSE_BACKUP_REMOTE_DIR":    f.root,
		"PACKETPULSE_BACKUP_LOCAL_DIR":     local,
		"PACKETPULSE_BACKUP_RETRY_SECONDS": "0",
	}}
}

func (p pullRig) pull(t *testing.T) result {
	t.Helper()
	return run(t, "PacketPulsePullBackups.sh", p.env)
}

func (p pullRig) heldLocally(name string) bool {
	return exists(filepath.Join(p.local, name, ".verified"))
}

// Every finished backup is copied, verified here, and only then receipted on
// the server; a second run fetches nothing again.
func TestABackupIsVerifiedHereBeforeTheServerMayPruneIt(t *testing.T) {
	p := newPullRig(t)
	first := p.f.takeBackup(t, nil)
	second := p.f.takeBackup(t, nil)

	out := p.pull(t)
	if out.code != 0 {
		t.Fatalf("the pull failed:\n%s", out.output)
	}
	for _, dir := range []string{first, second} {
		name := filepath.Base(dir)
		if !p.heldLocally(name) {
			t.Errorf("%s was not pulled and verified", name)
		}
		if !exists(filepath.Join(dir, ".pulled")) {
			t.Errorf("%s carries no receipt on the server", name)
		}
		if readFile(t, filepath.Join(dir, "db.sql.gz")) != readFile(t, filepath.Join(p.local, name, "db.sql.gz")) {
			t.Errorf("%s: the offsite copy differs from the server's", name)
		}
	}

	again := p.pull(t)
	if again.code != 0 || !strings.Contains(again.output, "pulled 0, skipped 2") {
		t.Errorf("a second pull did not recognise what it already holds:\n%s", again.output)
	}
}

// A copy that does not match its manifest is not a backup. No receipt, so the
// server keeps its own copy.
func TestACorruptCopyIsNeverReceipted(t *testing.T) {
	p := newPullRig(t)
	dir := p.f.takeBackup(t, nil)
	corrupt(t, filepath.Join(dir, "db.sql.gz"))

	out := p.pull(t)
	if out.code == 0 {
		t.Fatalf("a corrupt backup was pulled successfully:\n%s", out.output)
	}
	if exists(filepath.Join(dir, ".pulled")) {
		t.Error("the server was told a corrupt copy is safe offsite")
	}
	if p.heldLocally(filepath.Base(dir)) {
		t.Error("the corrupt copy counts as held")
	}
}

// A backup that failed, or one still being written (no manifest yet), is not
// fetched or receipted; the finished one beside it is.
func TestUnfinishedBackupsAreLeftOnTheServer(t *testing.T) {
	p := newPullRig(t)
	good := p.f.takeBackup(t, nil)
	failed := p.f.fakeBackup(t, -time.Hour, false, true)
	writing := filepath.Join(p.f.root, "29990102-000000")
	if err := os.MkdirAll(writing, 0o700); err != nil {
		t.Fatal(err)
	}
	writeEmpty(t, filepath.Join(writing, "db.sql.gz"))

	out := p.pull(t)
	if out.code != 0 {
		t.Fatalf("the pull failed:\n%s", out.output)
	}
	if !p.heldLocally(filepath.Base(good)) {
		t.Error("the finished backup was not pulled")
	}
	for _, dir := range []string{failed, writing} {
		if exists(filepath.Join(p.local, filepath.Base(dir))) {
			t.Errorf("%s was pulled although it never finished", filepath.Base(dir))
		}
		if exists(filepath.Join(dir, ".pulled")) {
			t.Errorf("%s was receipted although it never finished", filepath.Base(dir))
		}
	}
}

// Pulling everything there is says nothing about whether the server is still
// making backups. A newest backup older than the threshold is reported.
func TestAServerThatStoppedBackingUpIsReported(t *testing.T) {
	p := newPullRig(t)
	dir := p.f.takeBackup(t, nil)
	threeDaysOld := filepath.Join(p.f.root, time.Now().UTC().Add(-72*time.Hour).Format("20060102-150405"))
	if err := os.Rename(dir, threeDaysOld); err != nil {
		t.Fatal(err)
	}

	out := p.pull(t)
	if out.code == 0 || !strings.Contains(out.output, "the nightly backup is not running") {
		t.Fatalf("a three-day-old newest backup was not reported (exit %d):\n%s", out.code, out.output)
	}
}

// Backups that expired on the server before they were ever collected are gone
// for good, and that has to be said plainly.
func TestBackupsLostBeforeCollectionAreReported(t *testing.T) {
	p := newPullRig(t)
	stale := filepath.Join(p.local, "20200101-000000")
	if err := os.MkdirAll(stale, 0o700); err != nil {
		t.Fatal(err)
	}
	writeEmpty(t, filepath.Join(stale, ".verified"))
	dir := p.f.takeBackup(t, nil)
	corrupt(t, filepath.Join(dir, "db.sql.gz"))

	out := p.pull(t)
	if out.code == 0 || !strings.Contains(out.output, "PERMANENTLY LOST") {
		t.Fatalf("the gap between the copies was not reported (exit %d):\n%s", out.code, out.output)
	}
}
