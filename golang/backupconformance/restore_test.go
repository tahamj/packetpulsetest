package backupconformance

import (
	"path/filepath"
	"strings"
	"testing"
)

// restoreRig is a backup, a throwaway "live" database holding one marker row,
// and a service controller that records what it was asked to do instead of
// touching systemd.
type restoreRig struct {
	f          fixture
	live       string
	staging    string
	serviceLog string
	migrations string
}

func newRestoreRig(t *testing.T) restoreRig {
	t.Helper()
	f := newFixture(t)
	f.takeBackup(t, nil)
	prefix := uniqueName("pingle_bk_rst")
	dropAfter(t, prefix)
	rig := restoreRig{
		f:          f,
		live:       prefix + "_live",
		staging:    prefix + "_staging",
		serviceLog: filepath.Join(t.TempDir(), "service.log"),
		migrations: sql(t, liveDatabaseName(t), "SELECT count(*) FROM schema_migrations"),
	}
	sql(t, "postgres", `CREATE DATABASE "`+rig.live+`"`)
	sql(t, rig.live, "CREATE TABLE original_marker (id int); INSERT INTO original_marker VALUES (1)")
	return rig
}

func (r restoreRig) restore(t *testing.T, healthCmd string, args ...string) result {
	t.Helper()
	controller := filepath.Join(t.TempDir(), "servicectl")
	writeExecutable(t, controller, "#!/bin/bash\necho \"$*\" >> '"+r.serviceLog+"'\n")
	return run(t, "PingleRestore.sh", r.f.env(map[string]string{
		"PINGLE_RESTORE_STAGING_DB": r.staging,
		"PINGLE_LIVE_DB":            r.live,
		"PINGLE_SERVICE_CTL":        controller,
		"PINGLE_HEALTH_CMD":         healthCmd,
		"PINGLE_HEALTH_WAIT":        "1",
	}), args...)
}

func (r restoreRig) serviceCalls(t *testing.T) string {
	if !exists(r.serviceLog) {
		return ""
	}
	return strings.TrimSpace(readFile(t, r.serviceLog))
}

// holdsOriginal reports whether a database is the throwaway "live" original.
func holdsOriginal(t *testing.T, database string) bool {
	return sql(t, database, "SELECT count(*) FROM pg_tables WHERE tablename = 'original_marker'") == "1"
}

func holdsBackup(t *testing.T, database, migrations string) bool {
	return sql(t, database, "SELECT count(*) FROM pg_tables WHERE tablename = 'schema_migrations'") == "1" &&
		sql(t, database, "SELECT count(*) FROM schema_migrations") == migrations
}

// Step one restores and verifies into staging and nothing else: the service
// keeps running on the live database while someone inspects the copy.
func TestARestoreWithoutSwapLeavesTheLiveDatabaseAlone(t *testing.T) {
	r := newRestoreRig(t)
	out := r.restore(t, "true")
	if out.code != 0 {
		t.Fatalf("restore failed:\n%s", out.output)
	}
	if !holdsBackup(t, r.staging, r.migrations) {
		t.Error("staging does not hold the verified backup")
	}
	if !holdsOriginal(t, r.live) {
		t.Error("the live database was changed by a restore without --swap")
	}
	if calls := r.serviceCalls(t); calls != "" {
		t.Errorf("the service was touched without --swap: %q", calls)
	}
}

// The swap puts the verified copy into service under the live name and keeps
// the database it replaced, renamed, rather than dropping anything.
func TestASwapPutsTheBackupIntoServiceAndKeepsTheOriginal(t *testing.T) {
	r := newRestoreRig(t)
	out := r.restore(t, "true", "--swap")
	if out.code != 0 {
		t.Fatalf("swap failed:\n%s", out.output)
	}
	if !holdsBackup(t, r.live, r.migrations) {
		t.Error("the live name does not hold the backup after the swap")
	}
	aside := sql(t, "postgres", "SELECT datname FROM pg_database WHERE datname LIKE '"+r.live+"_before_%'")
	if aside == "" || !holdsOriginal(t, aside) {
		t.Errorf("the replaced database was not kept aside (found %q)", aside)
	}
	if databaseExists(t, r.staging) {
		t.Error("staging still exists after being swapped into service")
	}
	if calls := r.serviceCalls(t); calls != "stop pingle\nstart pingle" {
		t.Errorf("service calls = %q, want a stop then a start", calls)
	}
}

// If Pingle does not come up on the restored data, the swap is undone: the
// service goes back onto the database it had, and the restored copy is kept
// under the staging name for inspection.
func TestASwapThatDoesNotComeUpIsRolledBack(t *testing.T) {
	r := newRestoreRig(t)
	out := r.restore(t, "false", "--swap")
	if out.code == 0 {
		t.Fatalf("an unhealthy swap reported success:\n%s", out.output)
	}
	if !holdsOriginal(t, r.live) {
		t.Error("the original database is not back under the live name")
	}
	if !holdsBackup(t, r.staging, r.migrations) {
		t.Error("the restored copy was not kept in staging")
	}
	if aside := sql(t, "postgres", "SELECT datname FROM pg_database WHERE datname LIKE '"+r.live+"_before_%'"); aside != "" {
		t.Errorf("a renamed-aside database was left behind: %s", aside)
	}
	if calls := r.serviceCalls(t); calls != "stop pingle\nstart pingle\nstop pingle\nstart pingle" {
		t.Errorf("service calls = %q, want the swap's stop/start and the rollback's stop/start", calls)
	}
}

// Nothing that fails verification may reach service: the swap never starts.
func TestAnUnverifiableBackupIsNeverSwappedIn(t *testing.T) {
	r := newRestoreRig(t)
	backups := r.f.backups(t)
	corrupt(t, filepath.Join(backups[len(backups)-1], "db.sql.gz"))

	out := r.restore(t, "true", "--swap")
	if out.code == 0 {
		t.Fatalf("a corrupt backup was restored:\n%s", out.output)
	}
	if !holdsOriginal(t, r.live) {
		t.Error("the live database was changed")
	}
	if databaseExists(t, r.staging) {
		t.Error("a staging copy of a corrupt backup was left where it could be swapped in")
	}
	if calls := r.serviceCalls(t); calls != "" {
		t.Errorf("the service was stopped for a backup that never verified: %q", calls)
	}
}

// A copy that restored but did not verify is dropped, never left in staging
// where a later --swap could put it into service.
func TestACopyThatFailsVerificationIsNotLeftInStaging(t *testing.T) {
	r := newRestoreRig(t)
	backups := r.f.backups(t)
	falsifyRowCount(t, backups[len(backups)-1])

	out := r.restore(t, "true")
	if out.code == 0 {
		t.Fatalf("a backup that does not verify was restored:\n%s", out.output)
	}
	if !strings.Contains(out.output, "did not restore exactly") {
		t.Errorf("the restore did not fail on the row counts:\n%s", out.output)
	}
	if databaseExists(t, r.staging) {
		t.Error("the unverified copy was left in staging")
	}
}
