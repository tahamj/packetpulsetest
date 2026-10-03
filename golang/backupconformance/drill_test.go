package backupconformance

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// corrupt flips one byte in the middle of a file.
func corrupt(t *testing.T, path string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content[len(content)/2] ^= 0xff
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertDrillFailed(t *testing.T, f fixture, out result, scratch, why string) {
	t.Helper()
	if out.code == 0 {
		t.Fatalf("the drill passed:\n%s", out.output)
	}
	if !strings.Contains(out.output, why) {
		t.Errorf("the drill's failure does not say %q:\n%s", why, out.output)
	}
	if verdict := readFile(t, filepath.Join(f.root, "last_drill_result")); !strings.HasPrefix(verdict, "FAIL ") {
		t.Errorf("last_drill_result = %q, want FAIL", verdict)
	}
	if databaseExists(t, scratch) {
		t.Error("a failed drill left its scratch database behind")
	}
}

// A backup whose bytes changed after it was written must fail before anything
// is restored from it.
func TestACorruptBackupFailsTheDrill(t *testing.T) {
	f := newFixture(t)
	dir := f.takeBackup(t, nil)
	corrupt(t, filepath.Join(dir, "db.sql.gz"))

	scratch := uniqueName("pingle_bk_drill")
	dropAfter(t, scratch)
	out := run(t, "PingleRestoreDrill.sh", f.env(nil), "--into", scratch)
	assertDrillFailed(t, f, out, scratch, "does not match its manifest checksum")
}

// falsifyRowCount raises the row count state.json records for
// schema_migrations by one and re-signs the manifest, so the backup passes its
// checksums and only the row-count comparison can catch it. It returns the
// count the dump really holds.
func falsifyRowCount(t *testing.T, dir string) string {
	t.Helper()
	statePath := filepath.Join(dir, "state.json")
	state := readFile(t, statePath)
	pattern := regexp.MustCompile(`"public\.schema_migrations": (\d+)`)
	match := pattern.FindStringSubmatch(state)
	if match == nil {
		t.Fatalf("state.json has no schema_migrations count:\n%s", state)
	}
	recorded, _ := strconv.Atoi(match[1])
	state = pattern.ReplaceAllString(state, `"public.schema_migrations": `+strconv.Itoa(recorded+1))
	if err := os.WriteFile(statePath, []byte(state), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(state))
	manifestPath := filepath.Join(dir, "manifest.txt")
	manifest := regexp.MustCompile(`(?m)^[0-9a-f]{64}  state\.json$`).
		ReplaceAllString(readFile(t, manifestPath), hex.EncodeToString(sum[:])+"  state.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	return match[1]
}

// The row-count comparison is what turns "psql exited 0" into "every row came
// back". A count that disagrees must fail the drill and name the table.
func TestARestoreThatDisagreesWithTheDumpFailsTheDrill(t *testing.T) {
	f := newFixture(t)
	dir := f.takeBackup(t, nil)
	actual := falsifyRowCount(t, dir)

	scratch := uniqueName("pingle_bk_drill")
	dropAfter(t, scratch)
	out := run(t, "PingleRestoreDrill.sh", f.env(nil), "--into", scratch)
	assertDrillFailed(t, f, out, scratch, "public.schema_migrations: restored "+actual)
}

// The drill drops and recreates its target, so the live database must be
// refused by name. The "live" database here is a throwaway that the env file
// names, so that if the guard ever breaks, this test destroys only its own.
func TestTheDrillNeverRestoresIntoTheLiveDatabase(t *testing.T) {
	f := newFixture(t)
	f.takeBackup(t, nil)

	live := uniqueName("pingle_bk_live")
	dropAfter(t, live)
	sql(t, "postgres", `CREATE DATABASE "`+live+`"`)
	sql(t, live, "CREATE TABLE still_here (id int); INSERT INTO still_here VALUES (1)")
	liveEnv := filepath.Join(t.TempDir(), "live.env")
	if err := os.WriteFile(liveEnv, []byte("DATABASE_URL="+withDatabase(t, databaseURL(t), live)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out := run(t, "PingleRestoreDrill.sh", f.env(map[string]string{"PINGLE_ENV_FILE": liveEnv}), "--into", live)
	if out.code == 0 || !strings.Contains(out.output, "refusing to restore into") {
		t.Fatalf("the drill did not refuse the live database (exit %d):\n%s", out.code, out.output)
	}
	if got := sql(t, live, "SELECT count(*) FROM still_here"); got != "1" {
		t.Errorf("the live database was touched: still_here holds %s rows", got)
	}
}

// A failed night must not make the weekly drill test the failure: it drills
// the newest backup that finished.
func TestTheDrillTestsTheNewestFinishedBackup(t *testing.T) {
	f := newFixture(t)
	good := f.takeBackup(t, nil)
	f.fakeBackup(t, -time.Hour, false, true)

	scratch := uniqueName("pingle_bk_drill")
	dropAfter(t, scratch)
	out := run(t, "PingleRestoreDrill.sh", f.env(nil), "--into", scratch)
	if out.code != 0 {
		t.Fatalf("the drill failed:\n%s", out.output)
	}
	if verdict := strings.TrimSpace(readFile(t, filepath.Join(f.root, "last_drill_result"))); !strings.HasSuffix(verdict, good) {
		t.Errorf("the drill tested %q, want the newest finished backup %s", verdict, good)
	}
}
