package backupconformance

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The backups are encrypted to a key whose private half never touches the
// server, so a compromised box can write backups but not read them. That
// holds only while the private half stays out of every place one machine
// reaches, and a git repository is the worst of those: anyone who can clone
// reads its history, and once pushed it cannot be taken back.
func TestNoPrivateKeyMaterialIsInTheRepository(t *testing.T) {
	root := repoRoot(t)
	skip := map[string]bool{".git": true, "node_modules": true, "build": true, ".dart_tool": true, "Pods": true}
	// Assembled at runtime, so this file never matches itself.
	markers := []string{"PGP PRIVATE KEY BLOCK", "OPENSSH PRIVATE KEY"}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			if skip[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Size() > 2<<20 {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, marker := range markers {
			if strings.Contains(string(content), "-----BEGIN "+marker) {
				t.Errorf("%s holds a private key (%s)", strings.TrimPrefix(path, root+"/"), marker)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// The shipped key must be able to encrypt (an encryption subkey) and must be
// public only. Its fingerprint is recorded in the deployment record, where
// the person holding the private half can check it is theirs.
func TestTheShippedBackupKeyIsPublicAndCanEncrypt(t *testing.T) {
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg is not installed")
	}
	out, err := exec.Command("gpg", "--show-keys", "--with-colons", scriptPath(t, "packetpulse-backup-pub.asc")).Output()
	if err != nil {
		t.Fatalf("gpg cannot read the shipped key: %v", err)
	}
	listing := string(out)
	if regexp.MustCompile(`(?m)^(sec|ssb):`).MatchString(listing) {
		t.Fatal("packetpulse-backup-pub.asc contains SECRET key material")
	}
	// Field 12 of a colon listing holds a key's capabilities; e is encrypt.
	canEncrypt := false
	for _, line := range strings.Split(listing, "\n") {
		fields := strings.Split(line, ":")
		if len(fields) > 11 && fields[0] == "sub" && strings.Contains(fields[11], "e") {
			canEncrypt = true
		}
	}
	if !canEncrypt {
		t.Errorf("the shipped key has no encryption subkey:\n%s", listing)
	}
	primary := regexp.MustCompile(`(?m)^fpr:+([0-9A-F]{40}):`).FindStringSubmatch(listing)
	if primary == nil {
		t.Fatal("no fingerprint in the key listing")
	}
	record := readFile(t, filepath.Join(repoRoot(t), "docs", "PacketPulseDeploymentRecord.md"))
	if !strings.Contains(record, primary[1]) {
		t.Errorf("the deployment record does not state the backup key's fingerprint %s", primary[1])
	}
}

// Every job in the schedule runs a script the installer actually installs,
// names the .env explicitly, and the two jobs keep their cadence: a backup
// every night, a drill every week.
func TestTheScheduleRunsWhatTheInstallerInstalls(t *testing.T) {
	cron := readFile(t, scriptPath(t, "packetpulse-backup.cron"))
	installer := readFile(t, scriptPath(t, "PacketPulseBackupInstall.sh"))
	job := regexp.MustCompile(`(?m)^(\S+ \S+ \S+ \S+ \S+)\s+root\s+(.*)$`)
	jobs := job.FindAllStringSubmatch(cron, -1)
	if len(jobs) != 2 {
		t.Fatalf("want exactly two jobs (backup and drill), found %d", len(jobs))
	}
	schedules := map[string]string{}
	for _, entry := range jobs {
		command := entry[2]
		if !strings.Contains(command, "PACKETPULSE_ENV_FILE=/opt/packetpulse/.env ") {
			t.Errorf("job does not name the .env explicitly: %s", command)
		}
		script := regexp.MustCompile(`/opt/packetpulse/backup/(\S+\.sh)`).FindStringSubmatch(command)
		if script == nil {
			t.Errorf("job runs nothing from /opt/packetpulse/backup: %s", command)
			continue
		}
		if !exists(scriptPath(t, script[1])) {
			t.Errorf("the schedule runs %s, which is not in the repository", script[1])
		}
		if !strings.Contains(installer, `/`+script[1]) {
			t.Errorf("the schedule runs %s, which the installer does not install", script[1])
		}
		schedules[script[1]] = entry[1]
	}
	if fields := strings.Fields(schedules["PacketPulseBackup.sh"]); len(fields) != 5 || fields[2] != "*" || fields[4] != "*" {
		t.Errorf("the backup does not run every night: %q", schedules["PacketPulseBackup.sh"])
	}
	if fields := strings.Fields(schedules["PacketPulseRestoreDrill.sh"]); len(fields) != 5 || fields[4] == "*" {
		t.Errorf("the drill does not run weekly: %q", schedules["PacketPulseRestoreDrill.sh"])
	}
}
