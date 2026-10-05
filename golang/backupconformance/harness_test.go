// Package backupconformance runs PacketPulse's backup scripts for real: against a
// real PostgreSQL, with real gpg, rsync and gzip, exactly as cron runs them on
// the server. The scripts are shell, so the only honest test is to run them
// and look at what they left behind: the dump, the receipts, the databases.
//
// The suite reads the live development database (pg_dump is read-only) and
// writes only to databases and directories it creates itself, with unique
// names, dropped afterwards. It never restores into, renames or drops the
// database DATABASE_URL names.
//
// Environment:
//
//	DATABASE_URL                     the database to back up (else the repo's .env)
//	PACKETPULSE_TEST_ADMIN_DATABASE_URL   a superuser connection for the drill
//	                                 (default: the local socket, as the OS user)
//	PACKETPULSE_TEST_REQUIRE_DATABASE=1   fail rather than skip without a database
package backupconformance

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "packetpulsetest.sh")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find the repository root (no packetpulsetest.sh above the test)")
		}
		dir = parent
	}
}

func scriptPath(t *testing.T, name string) string {
	return filepath.Join(repoRoot(t), "scripts", "deploy", "backup", name)
}

// databaseURL is the database the suite backs up: DATABASE_URL, else the
// repository's .env, as every DB-backed suite in this repo reads it.
func databaseURL(t *testing.T) string {
	t.Helper()
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value
	}
	file, err := os.Open(filepath.Join(repoRoot(t), ".env"))
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "DATABASE_URL=") {
				return strings.Trim(strings.TrimPrefix(line, "DATABASE_URL="), `"'`)
			}
		}
	}
	return ""
}

func adminURL() string {
	if value := os.Getenv("PACKETPULSE_TEST_ADMIN_DATABASE_URL"); value != "" {
		return value
	}
	return "postgresql:///postgres?host=/tmp"
}

// withDatabase returns rawURL pointed at another database on the same server.
func withDatabase(t *testing.T, rawURL, database string) string {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("unparseable database URL: %v", err)
	}
	parsed.Path = "/" + database
	return parsed.String()
}

// liveDatabaseName is the database DATABASE_URL names: the one no test may
// ever restore into.
func liveDatabaseName(t *testing.T) string {
	parsed, err := url.Parse(databaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimPrefix(parsed.Path, "/")
}

// requireDatabase skips without a usable database and superuser, unless
// PACKETPULSE_TEST_REQUIRE_DATABASE=1, where a skip would hide a regression.
func requireDatabase(t *testing.T) {
	t.Helper()
	missing := func(why string) {
		if os.Getenv("PACKETPULSE_TEST_REQUIRE_DATABASE") == "1" {
			t.Fatalf("backup suite needs a database: %s", why)
		}
		t.Skipf("no database for the backup suite: %s", why)
	}
	if databaseURL(t) == "" {
		missing("DATABASE_URL is not set and the repo has no .env")
	}
	for _, tool := range []string{"psql", "pg_dump", "gzip", "gpg", "rsync"} {
		if _, err := exec.LookPath(tool); err != nil {
			missing(tool + " is not installed")
		}
	}
	if out, err := exec.Command("psql", databaseURL(t), "-XtAc", "SELECT count(*) FROM schema_migrations").CombinedOutput(); err != nil {
		missing("cannot read schema_migrations (run the unit suite first, which migrates the database): " + strings.TrimSpace(string(out)))
	}
	if out, err := exec.Command("psql", adminURL(), "-XtAc", "SELECT rolsuper FROM pg_roles WHERE rolname = current_user").CombinedOutput(); err != nil || strings.TrimSpace(string(out)) != "t" {
		missing("PACKETPULSE_TEST_ADMIN_DATABASE_URL is not a superuser connection: " + strings.TrimSpace(string(out)))
	}
}

// sql runs one statement against a database as the superuser and returns
// the trimmed output.
func sql(t *testing.T, database, statement string) string {
	t.Helper()
	out, err := exec.Command("psql", withDatabase(t, adminURL(), database), "-XtAq",
		"-v", "ON_ERROR_STOP=1", "-c", statement).CombinedOutput()
	if err != nil {
		t.Fatalf("psql %q on %s: %v\n%s", statement, database, err, out)
	}
	return strings.TrimSpace(string(out))
}

func databaseExists(t *testing.T, name string) bool {
	return sql(t, "postgres", "SELECT count(*) FROM pg_database WHERE datname = '"+name+"'") == "1"
}

// uniqueName is a database name no other test, and no person, is using.
func uniqueName(prefix string) string {
	buffer := make([]byte, 4)
	_, _ = rand.Read(buffer)
	return prefix + "_" + hex.EncodeToString(buffer)
}

// dropAfter drops every database whose name starts with prefix when the test
// ends, so a failing test does not leave databases behind.
func dropAfter(t *testing.T, prefix string) {
	t.Cleanup(func() {
		names := sql(t, "postgres", "SELECT datname FROM pg_database WHERE datname LIKE '"+prefix+"%'")
		for _, name := range strings.Fields(names) {
			sql(t, "postgres", `DROP DATABASE IF EXISTS "`+name+`" WITH (FORCE)`)
		}
	})
}

type result struct {
	output string
	code   int
}

// run executes one of the backup scripts with exactly the environment given,
// plus PATH, HOME and TMPDIR: nothing from the test's own environment can
// leak in and make a broken script look working. /bin/bash is used on
// purpose: on a Mac it is bash 3.2, the oldest the scripts promise to run on.
func run(t *testing.T, script string, env map[string]string, args ...string) result {
	t.Helper()
	command := exec.Command("/bin/bash", append([]string{scriptPath(t, script)}, args...)...)
	command.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
		"TMPDIR=" + os.TempDir(),
	}
	for key, value := range env {
		command.Env = append(command.Env, key+"="+value)
	}
	out, err := command.CombinedOutput()
	code := 0
	if exitError, ok := err.(*exec.ExitError); ok {
		code = exitError.ExitCode()
	} else if err != nil {
		t.Fatalf("could not run %s: %v", script, err)
	}
	return result{output: string(out), code: code}
}

// fixture is a backup root and the .env the scripts read.
type fixture struct {
	root      string
	envFile   string
	jwtSecret string
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	requireDatabase(t)
	dir := t.TempDir()
	secret := uniqueName("jwt-secret-that-must-never-appear-in-clear")
	envFile := filepath.Join(dir, "packetpulse.env")
	content := "# a PacketPulse .env\nDATABASE_URL=" + databaseURL(t) + "\nJWT_SECRET=" + secret + "\n"
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, "backups")
	return fixture{root: root, envFile: envFile, jwtSecret: secret}
}

// env is the environment cron gives the scripts, pointed at this fixture.
func (f fixture) env(extra map[string]string) map[string]string {
	env := map[string]string{
		"PACKETPULSE_ENV_FILE":   f.envFile,
		"PACKETPULSE_BACKUP_DIR": f.root,
		// No key unless a test supplies one, so most tests do not need gpg keys.
		"PACKETPULSE_BACKUP_KEY_FILE": filepath.Join(f.root, "no-key-configured.asc"),
		"PACKETPULSE_DRILL_ADMIN_URL": adminURL(),
	}
	for key, value := range extra {
		env[key] = value
	}
	return env
}

// backups lists the backup directories under the root, oldest first.
func (f fixture) backups(t *testing.T) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(f.root, "2*"))
	if err != nil {
		t.Fatal(err)
	}
	return matches
}

// takeBackup runs PacketPulseBackup.sh and insists it succeeded, returning the
// directory it created.
func (f fixture) takeBackup(t *testing.T, extra map[string]string) string {
	t.Helper()
	before := map[string]bool{}
	for _, dir := range f.backups(t) {
		before[dir] = true
	}
	out := run(t, "PacketPulseBackup.sh", f.env(extra))
	if out.code != 0 {
		t.Fatalf("PacketPulseBackup.sh exited %d:\n%s", out.code, out.output)
	}
	for _, dir := range f.backups(t) {
		if !before[dir] {
			return dir
		}
	}
	t.Fatalf("PacketPulseBackup.sh succeeded but created no backup directory:\n%s", out.output)
	return ""
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// writeExecutable writes a small script, for standing in for a tool.
func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}
