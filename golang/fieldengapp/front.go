package fieldengapp

import (
	"crypto/tls"
	"encoding/pem"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// ClientCertificateHeader is the header nginx's partner site passes the
// caller's certificate on in, as PacketPulse reads it.
const ClientCertificateHeader = "X-PacketPulse-Client-Cert"

// LaunchPath is the one path the partner site serves.
const LaunchPath = "/api/v1/partner/launch"

// FrontHandler stands in for nginx's partner site where there is no nginx:
// on a developer's machine and in the test suite. Behind the TLS FrontTLS
// configures, it does what scripts/deploy/packetpulse-partner-nginx.conf
// does:
//
//   - it serves the launch route and nothing else;
//   - it passes the certificate the caller showed on in
//     ClientCertificateHeader, escaped as nginx escapes
//     $ssl_client_escaped_cert, and empty when none was shown;
//   - it replaces any such header the caller sent, as proxy_set_header does;
//   - it appends the caller's address to X-Forwarded-For.
func FrontHandler(upstream *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != LaunchPath {
			http.NotFound(w, r)
			return
		}
		presented := ""
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			presented = nginxEscape(string(pem.EncodeToMemory(&pem.Block{
				Type: "CERTIFICATE", Bytes: r.TLS.PeerCertificates[0].Raw})))
		}
		r.Header.Set(ClientCertificateHeader, presented)
		proxy.ServeHTTP(w, r)
	})
}

// FrontTLS is the partner site's TLS: it serves with server and asks every
// caller for a certificate without asking who signed it (nginx's
// ssl_verify_client optional_no_ca). The handshake still proves the caller
// holds the key for the certificate it shows; PacketPulse decides whether
// that certificate is one it trusts.
func FrontTLS(server tls.Certificate) *tls.Config {
	return &tls.Config{Certificates: []tls.Certificate{server}, ClientAuth: tls.RequestClientCert, MinVersion: tls.VersionTLS12}
}

// nginxEscape escapes as nginx's URI-component escaping does: a space as
// %20, and '+', '/' and '=' escaped too.
func nginxEscape(text string) string {
	return strings.ReplaceAll(url.QueryEscape(text), "+", "%20")
}
