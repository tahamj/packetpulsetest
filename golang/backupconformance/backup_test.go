package backupconformance

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The whole point: a nightly backup, restored into a scratch database, gives
// back every table with exactly the rows that were dumped, and the scratch
// database is gone afterwards.
func TestANightlyBackupRestoresExactlyIntoAScratchDatabase(t *testing.T) {
	f := newFixture(t)
	dir := f.takeBackup(t, nil)

	for _, file := range []string{"db.sql.gz", "state.json", "manifest.txt"} {
		if !exists(filepath.Join(dir, file)) {
			t.Fatalf("the backup has no %s", file)
		}
	}
	state := readFile(t, filepath.Join(dir, "state.json"))
	live := sql(t, liveDatabaseName(t), "SELECT count(*) FROM schema_migrations")
	if !strings.Contains(state, `"public.schema_migrations": `+live) {
		t.Errorf("state.json does not record the %s migrations the database holds:\n%s", live, state)
	}

	scratch := uniqueName("pingle_bk_drill")
	dropAfter(t, scratch)
	drill := run(t, "PingleRestoreDrill.sh", f.env(nil), "--into", scratch)
	if drill.code != 0 {
		t.Fatalf("the drill failed on a good backup:\n%s", drill.output)
	}
	verdict := readFile(t, filepath.Join(f.root, "last_drill_result"))
	if !strings.HasPrefix(verdict, "PASS ") || !strings.HasSuffix(strings.TrimSpace(verdict), dir) {
		t.Errorf("last_drill_result = %q, want PASS for %s", verdict, dir)
	}
	if databaseExists(t, scratch) {
		t.Error("the drill left its scratch database behind")
	}
	if !strings.Contains(drill.output, "every table restored exactly") {
		t.Errorf("the drill did not report verifying the tables:\n%s", drill.output)
	}
}

// pg_dump exiting 0 is not success. A dump that stops early (a dropped
// connection, a full pipe) must be marked incomplete, never counted.
func TestADumpThatExitsZeroButStopsEarlyIsAFailure(t *testing.T) {
	f := newFixture(t)
	fakes := t.TempDir()
	// Plenty of bytes, a valid gzip, a zero exit, but no completion trailer.
	writeExecutable(t, filepath.Join(fakes, "pg_dump"), `#!/bin/bash
echo "-- PostgreSQL database dump"
echo "COPY public.schema_migrations (version) FROM stdin;"
head -c 20000 /dev/urandom | base64
exit 0
`)
	out := run(t, "PingleBackup.sh", f.env(map[string]string{"PATH": fakes + ":" + os.Getenv("PATH")}))
	assertFailedBackup(t, f, out, "cut short")
}

// A dump that FAILS must never be recorded as a success: without pipefail the
// status tested would be gzip's, which is always 0.
func TestAFailedDumpIsNeverRecordedAsASuccess(t *testing.T) {
	f := newFixture(t)
	fakes := t.TempDir()
	writeExecutable(t, filepath.Join(fakes, "pg_dump"), "#!/bin/bash\necho 'pg_dump: error: connection refused' >&2\nexit 1\n")
	out := run(t, "PingleBackup.sh", f.env(map[string]string{"PATH": fakes + ":" + os.Getenv("PATH")}))
	assertFailedBackup(t, f, out, "pg_dump FAILED")
}

// A complete dump of some other database is not a Pingle backup.
func TestADumpWithoutSchemaMigrationsIsNotAPingleBackup(t *testing.T) {
	f := newFixture(t)
	other := uniqueName("pingle_bk_other")
	dropAfter(t, other)
	sql(t, "postgres", `CREATE DATABASE "`+other+`"`)
	sql(t, other, "CREATE TABLE filler AS SELECT g, md5(g::text) AS h FROM generate_series(1, 2000) g")
	envFile := filepath.Join(t.TempDir(), "other.env")
	if err := os.WriteFile(envFile, []byte("DATABASE_URL="+withDatabase(t, adminURL(), other)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out := run(t, "PingleBackup.sh", f.env(map[string]string{"PINGLE_ENV_FILE": envFile}))
	assertFailedBackup(t, f, out, "not a Pingle database")
}

func assertFailedBackup(t *testing.T, f fixture, out result, why string) {
	t.Helper()
	if out.code == 0 {
		t.Fatalf("the backup reported success:\n%s", out.output)
	}
	if !strings.Contains(out.output, why) {
		t.Errorf("the failure does not say %q:\n%s", why, out.output)
	}
	backups := f.backups(t)
	if len(backups) != 1 {
		t.Fatalf("want exactly one (failed) backup directory, found %v", backups)
	}
	if !exists(filepath.Join(backups[0], ".incomplete")) {
		t.Error("the failed backup is not marked .incomplete")
	}
	if exists(filepath.Join(backups[0], "manifest.txt")) {
		t.Error("the failed backup has a manifest, which is what marks a backup complete")
	}
	if exists(filepath.Join(f.root, "last_success")) {
		t.Error("a failed backup updated last_success")
	}
}

// argv is world-readable through ps; the database password must reach
// pg_dump through its environment only.
func TestThePasswordNeverReachesACommandLine(t *testing.T) {
	f := newFixture(t)
	parsed := regexp.MustCompile(`^postgres(ql)?://[^:@/]*:([^@]+)@`).FindStringSubmatch(databaseURL(t))
	if parsed == nil {
		t.Skip("DATABASE_URL carries no password, so there is none to leak")
	}
	password := parsed[2]
	realDump, err := exec.LookPath("pg_dump")
	if err != nil {
		t.Fatal(err)
	}
	fakes := t.TempDir()
	argvLog := filepath.Join(fakes, "argv")
	envLog := filepath.Join(fakes, "env")
	writeExecutable(t, filepath.Join(fakes, "pg_dump"), `#!/bin/bash
printf '%s\n' "$0" "$@" > '`+argvLog+`'
[ -n "${PGPASSWORD:-}" ] && echo PGPASSWORD > '`+envLog+`'
exec '`+realDump+`' "$@"
`)
	f.takeBackup(t, map[string]string{"PATH": fakes + ":" + os.Getenv("PATH")})

	argv := readFile(t, argvLog)
	if strings.Contains(argv, password) || strings.Contains(argv, "://") {
		t.Errorf("pg_dump's command line carries the connection secret:\n%s", argv)
	}
	if !exists(envLog) {
		t.Error("pg_dump was not given the password through PGPASSWORD")
	}
}

// PgBouncer's transaction pooling cannot hold the dump's one snapshot; the
// result would be inconsistent while looking fine.
func TestBackupsRefuseToGoThroughPgBouncer(t *testing.T) {
	f := newFixture(t)
	envFile := filepath.Join(t.TempDir(), "pooled.env")
	pooled := regexp.MustCompile(`:\d+/`).ReplaceAllString(databaseURL(t), ":6432/")
	if !strings.Contains(pooled, ":6432/") {
		pooled = strings.Replace(databaseURL(t), "@localhost/", "@localhost:6432/", 1)
	}
	if err := os.WriteFile(envFile, []byte("DATABASE_URL="+pooled+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out := run(t, "PingleBackup.sh", f.env(map[string]string{"PINGLE_ENV_FILE": envFile}))
	if out.code == 0 || !strings.Contains(out.output, "PgBouncer") {
		t.Fatalf("a backup through port 6432 was not refused (exit %d):\n%s", out.code, out.output)
	}
	if len(f.backups(t)) != 0 {
		t.Error("a refused backup still created a backup directory")
	}
}

// MShop's backup looked for its .env beside itself, never found the real one,
// and failed every night for the life of the deployment. On the server these
// scripts live in /opt/pingle/backup/ and the .env one level up.
func TestTheEnvFileIsFoundOneLevelAboveTheScripts(t *testing.T) {
	f := newFixture(t)
	install := filepath.Join(t.TempDir(), "opt", "pingle")
	if err := os.MkdirAll(filepath.Join(install, "backup"), 0o755); err != nil {
		t.Fatal(err)
	}
	script := readFile(t, scriptPath(t, "PingleBackup.sh"))
	writeExecutable(t, filepath.Join(install, "backup", "PingleBackup.sh"), script)
	if err := os.WriteFile(filepath.Join(install, ".env"), []byte(readFile(t, f.envFile)), 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command("/bin/bash", filepath.Join(install, "backup", "PingleBackup.sh"))
	command.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"),
		"PINGLE_BACKUP_DIR=" + f.root, "PINGLE_BACKUP_KEY_FILE=/nonexistent"}
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("with no PINGLE_ENV_FILE the backup did not find ../.env: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), filepath.Join(install, ".env")) {
		t.Errorf("the log does not say which .env it used:\n%s", out)
	}
}

// The .env goes into the backup encrypted to the backup key, and nowhere in
// the clear: whoever holds the server can write backups but not read one.
func TestTheEnvTravelsEncryptedToTheBackupKeyOnly(t *testing.T) {
	f := newFixture(t)
	keys := newTestKey(t)
	dir := f.takeBackup(t, map[string]string{"PINGLE_BACKUP_KEY_FILE": keys.publicKey})

	sealed := readFile(t, filepath.Join(dir, "secrets.gpg"))
	if strings.Contains(sealed, f.jwtSecret) {
		t.Fatal("secrets.gpg contains JWT_SECRET in the clear")
	}
	if got := keys.decrypt(t, filepath.Join(dir, "secrets.gpg")); got != readFile(t, f.envFile) {
		t.Errorf("secrets.gpg does not decrypt to the .env:\n%s", got)
	}
	// The property that matters, stated directly: without the private key it
	// cannot be read. Grepping for the secret is not enough, because gpg
	// compresses, and a merely compressed .env would hide it from a grep.
	if readable, err := exec.Command("gpg", "--homedir", emptyGPGHome(t), "--batch", "--quiet",
		"--decrypt", filepath.Join(dir, "secrets.gpg")).Output(); err == nil {
		t.Errorf("secrets.gpg is readable without the backup key:\n%s", readable)
	}
	if !strings.Contains(readFile(t, filepath.Join(dir, "state.json")), `"has_secrets": true`) {
		t.Error("state.json does not record that the key material is present")
	}
	if !strings.Contains(readFile(t, filepath.Join(dir, "manifest.txt")), "  secrets.gpg") {
		t.Error("the manifest does not cover secrets.gpg")
	}
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if entry.Name() != "secrets.gpg" && strings.Contains(readFile(t, filepath.Join(dir, entry.Name())), f.jwtSecret) {
			t.Errorf("%s holds JWT_SECRET in the clear", entry.Name())
		}
	}
}

// Without a key the data is still backed up, but the gap is said out loud
// and recorded, not discovered on the day of a restore.
func TestAMissingBackupKeyIsRecordedNotSilent(t *testing.T) {
	f := newFixture(t)
	dir := f.takeBackup(t, nil)
	if exists(filepath.Join(dir, "secrets.gpg")) {
		t.Fatal("secrets.gpg exists although no key was configured")
	}
	if !strings.Contains(readFile(t, filepath.Join(dir, "state.json")), `"has_secrets": false`) {
		t.Error("state.json does not record that the .env is missing from this backup")
	}
}

// The dump holds every password hash and sealed directory password: nobody
// but the backup's owner may read it.
func TestBackupsAreReadableOnlyByTheirOwner(t *testing.T) {
	f := newFixture(t)
	dir := f.takeBackup(t, nil)
	assertMode(t, dir, 0o700)
	for _, file := range []string{"db.sql.gz", "state.json", "manifest.txt"} {
		assertMode(t, filepath.Join(dir, file), 0o600)
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("%s is mode %o, want %o", filepath.Base(path), got, want)
	}
}

// Two backups at once would write over each other; a marker left by a run
// that was killed must not block every backup after it.
func TestASecondBackupWaitsButADeadOneDoesNotBlock(t *testing.T) {
	f := newFixture(t)
	if err := os.MkdirAll(f.root, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(f.root, ".in_progress")
	if err := os.WriteFile(marker, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		t.Fatal(err)
	}
	out := run(t, "PingleBackup.sh", f.env(nil))
	if out.code == 0 || !strings.Contains(out.output, "in progress") {
		t.Fatalf("a second backup started while one was running:\n%s", out.output)
	}
	if len(f.backups(t)) != 0 {
		t.Error("the refused run created a backup directory")
	}

	ageBy(t, marker, 7*3600)
	f.takeBackup(t, nil)
	if exists(marker) {
		t.Error("the progress marker outlived the run")
	}
}

// state.json's row counts must come from the dump itself, so the drill can
// demand exact equality without racing the writes that follow the dump.
func TestStateRowCountsAreReadFromTheDumpItself(t *testing.T) {
	f := newFixture(t)
	dir := f.takeBackup(t, nil)
	dump, err := exec.Command("bash", "-c", "gzip -dc '"+filepath.Join(dir, "db.sql.gz")+"'").Output()
	if err != nil {
		t.Fatal(err)
	}
	copies := regexp.MustCompile(`(?m)^COPY ([^ ]+) `).FindAllStringSubmatch(string(dump), -1)
	state := readFile(t, filepath.Join(dir, "state.json"))
	if len(copies) == 0 {
		t.Fatal("the dump has no COPY blocks")
	}
	for _, copy := range copies {
		if !strings.Contains(state, `"`+copy[1]+`": `) {
			t.Errorf("state.json has no row count for %s", copy[1])
		}
	}
	// row_counts is the last object in state.json, so everything after its key
	// is row counts.
	rowCounts := state[strings.Index(state, `"row_counts": {`)+len(`"row_counts": {`):]
	if got := strings.Count(rowCounts, `": `); got != len(copies) {
		t.Errorf("state.json records %d tables, the dump holds %d", got, len(copies))
	}
}
