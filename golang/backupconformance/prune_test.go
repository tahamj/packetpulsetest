package backupconformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeBackup creates a backup directory under the fixture's root, named for
// the moment it was taken, age ago, as PingleBackup.sh names them: complete
// (it has a manifest) unless incomplete is set, and receipted if pulled. The
// scripts read a backup's age from that name, never from the directory's
// mtime, which the puller's receipt resets.
func (f fixture) fakeBackup(t *testing.T, age time.Duration, pulled, incomplete bool) string {
	t.Helper()
	dir := filepath.Join(f.root, time.Now().UTC().Add(-age).Format("20060102-150405"))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if incomplete {
		writeEmpty(t, filepath.Join(dir, ".incomplete"))
	} else {
		writeEmpty(t, filepath.Join(dir, "manifest.txt"))
	}
	if pulled {
		writeEmpty(t, filepath.Join(dir, ".pulled"))
	}
	return dir
}

func writeEmpty(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertKept(t *testing.T, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		if !exists(dir) {
			t.Errorf("%s was pruned and should have been kept", filepath.Base(dir))
		}
	}
}

func assertPruned(t *testing.T, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		if exists(dir) {
			t.Errorf("%s was kept and should have been pruned", filepath.Base(dir))
		}
	}
}

// Deleting a backup the offsite copy never received turns two copies into
// one without anybody noticing. Age alone is never grounds for deletion.
func TestOnlyBackupsConfirmedOffsiteArePruned(t *testing.T) {
	f := newFixture(t)
	pulledOld := f.fakeBackup(t, 62*24*time.Hour, true, false)
	unpulledOld := f.fakeBackup(t, 61*24*time.Hour, false, false)
	protected := f.fakeBackup(t, 60*24*time.Hour, true, false)

	out := run(t, "PingleBackup.sh", f.env(map[string]string{"PINGLE_BACKUP_MIN_KEEP": "2"}))
	if out.code != 0 {
		t.Fatalf("backup failed:\n%s", out.output)
	}
	assertPruned(t, pulledOld)
	assertKept(t, unpulledOld, protected)
	if !strings.Contains(out.output, "never pulled offsite") {
		t.Errorf("keeping an unpulled backup past retention was not reported:\n%s", out.output)
	}
}

// However old and however safely copied, the newest few stay on the server:
// a restore should never have to wait for a download.
func TestTheNewestBackupsAreKeptHoweverOld(t *testing.T) {
	f := newFixture(t)
	oldest := f.fakeBackup(t, 90*24*time.Hour, true, false)
	second := f.fakeBackup(t, 89*24*time.Hour, true, false)
	third := f.fakeBackup(t, 88*24*time.Hour, true, false)

	f.takeBackup(t, map[string]string{"PINGLE_BACKUP_MIN_KEEP": "3"})
	assertPruned(t, oldest)
	assertKept(t, second, third)
}

// A full disk stops postgres, so below the free-space floor retention yields
// even without an offsite copy, but never below the minimum kept.
func TestDiskPressureOverridesRetentionButKeepsTheMinimum(t *testing.T) {
	f := newFixture(t)
	first := f.fakeBackup(t, 2*24*time.Hour, false, false)
	second := f.fakeBackup(t, 24*time.Hour, false, false)

	out := run(t, "PingleBackup.sh", f.env(map[string]string{
		"PINGLE_BACKUP_MIN_KEEP":    "2",
		"PINGLE_BACKUP_MIN_FREE_MB": "999999999",
	}))
	if out.code != 0 {
		t.Fatalf("backup failed:\n%s", out.output)
	}
	assertPruned(t, first)
	assertKept(t, second)
	if got := len(f.backups(t)); got != 2 {
		t.Errorf("%d backups left under disk pressure, want the minimum of 2", got)
	}
	if !strings.Contains(out.output, "DISK PRESSURE") {
		t.Errorf("removing a backup without an offsite copy was not shouted:\n%s", out.output)
	}
}

// A failed night must not take a slot among the newest kept backups: if it
// did, a run of failures would push every good backup out of protection.
func TestAFailedBackupNeverCountsTowardTheMinimumKept(t *testing.T) {
	f := newFixture(t)
	older := f.fakeBackup(t, 62*24*time.Hour, true, false)
	good := f.fakeBackup(t, 61*24*time.Hour, true, false)
	recentFailure := f.fakeBackup(t, time.Hour, false, true)
	oldFailure := f.fakeBackup(t, 63*24*time.Hour, false, true)

	f.takeBackup(t, map[string]string{"PINGLE_BACKUP_MIN_KEEP": "2"})
	assertPruned(t, older, oldFailure)
	assertKept(t, good, recentFailure)
}
