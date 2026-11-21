package fieldengapp

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The nginx configuration the field app's sign-in rests on (migration 0043),
// as shipped in scripts/deploy: the ordinary site wipes the certificate
// header and refuses the launch route, and only that - held by PacketPulsePartnerNginxCheck.sh,
// which the deploy runs - and the partner site asks for a certificate and
// passes on what it saw, for the launch route alone. Front is held to the
// same, so the suite tests what nginx does.

func deployDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../../../scripts/deploy")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func runCheck(t *testing.T, config string) (string, error) {
	t.Helper()
	command := exec.Command("/bin/bash", filepath.Join(deployDir(t), "PacketPulsePartnerNginxCheck.sh"), "-")
	command.Stdin = strings.NewReader(config)
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestTheOrdinarySiteNeverPassesOnAForgedCertificate(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(deployDir(t), "packetpulse-nginx.conf"))
	if err != nil {
		t.Fatal(err)
	}
	shipped := string(raw)
	if output, err := runCheck(t, shipped); err != nil {
		t.Fatalf("the shipped configuration fails the check:\n%s", output)
	}

	// Proved able to fail: each rule taken out is caught, by name.
	for name, broken := range map[string]string{
		"the wipe":                               strings.Replace(shipped, `proxy_set_header X-PacketPulse-Client-Cert "";`, "", 1),
		"the refusal":                            strings.Replace(shipped, "location = /api/v1/partner/launch { return 404; }", "", 1),
		"a second API location without the wipe": shipped + "\nserver {\n    location /api/ {\n        proxy_pass http://127.0.0.1:8080;\n    }\n    location = /api/v1/partner/launch { return 404; }\n}\n",
		// Refusing all of /api/v1/partner/ shut the certificate register the
		// web app reaches through the ordinary site (live, 18 November 2026).
		"the register, refused with the launch": strings.Replace(shipped, "location = /api/v1/partner/launch { return 404; }", "location ^~ /api/v1/partner/ { return 404; }", 1),
		"the register, refused as a prefix":     strings.Replace(shipped, "location = /api/v1/partner/launch { return 404; }", "location = /api/v1/partner/launch { return 404; }\n    location /api/v1/partner/ { return 404; }", 1),
		"not the site at all":                   "server { listen 80; }",
	} {
		if output, err := runCheck(t, broken); err == nil {
			t.Errorf("with %s removed the check still passed:\n%s", name, output)
		}
	}
}

func TestThePartnerSiteAsksForACertificateAndServesOnlyTheLaunch(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(deployDir(t), "packetpulse-partner-nginx.conf"))
	if err != nil {
		t.Fatal(err)
	}
	config := string(raw)
	for _, want := range []string{
		"server_name partner.packetpulse.rummaan53.com;",
		"ssl_verify_client optional_no_ca;",
		"location = " + LaunchPath + " {",
		"proxy_set_header " + ClientCertificateHeader + " $ssl_client_escaped_cert;",
		"proxy_pass http://127.0.0.1:8080;",
	} {
		if !strings.Contains(config, want) {
			t.Errorf("the partner site lacks %q", want)
		}
	}
	if strings.Count(config, "proxy_pass") != 1 {
		t.Error("the partner site passes more than the launch route on")
	}
}
