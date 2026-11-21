package fieldengapp

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Partner is the app's SERVER talking to PacketPulse: it asks the partner
// address for a link over mutual TLS, presenting the app's certificate, and
// reads results back from the ordinary address with a read-only API key.
// Neither the certificate's key nor the API key ever reaches a phone.
type Partner struct {
	partnerURL     string
	packetPulseURL string
	apiKey         string
	mutual         *http.Client
	plain          *http.Client
}

// LaunchLink is PacketPulse's answer to a launch.
type LaunchLink struct {
	Url       string `json:"url"`
	ExpiresIn int    `json:"expires_in"`
	Engineer  struct {
		UserId  string `json:"user_id"`
		Name    string `json:"name"`
		Created bool   `json:"created"`
	} `json:"engineer"`
}

// Refusal is PacketPulse's error envelope: a message written for the
// engineer, which the app shows as it is.
type Refusal struct {
	Status  int
	Code    string `json:"code"`
	Message string `json:"message"`
	Action  string `json:"action"`
}

func (r *Refusal) Error() string {
	return fmt.Sprintf("PacketPulse answered %d %s: %s", r.Status, r.Code, r.Message)
}

// unexplained words an answer that is not PacketPulse's error envelope - a
// proxy's own page, or a PacketPulse from before the field app's sign-in -
// so the engineer is told what it means, not shown a bare status.
func unexplained(status int) string {
	if status == http.StatusNotFound {
		return "PacketPulse does not answer field app sign-ins at this address. " +
			"Check the address the app calls, and that PacketPulse's field app sign-in is switched on."
	}
	return fmt.Sprintf("PacketPulse answered %d without saying why. Try again in a minute.", status)
}

// NewPartner builds the app server's clients from the certificate and key
// files, trusting the system's authorities and, when given, one more
// certificate (a local Front's own).
func NewPartner(config Config) (*Partner, error) {
	certificate, err := tls.LoadX509KeyPair(config.CertificateFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("the app's certificate and key: %w", err)
	}
	roots, err := x509.SystemCertPool()
	if err != nil {
		roots = x509.NewCertPool()
	}
	if config.TrustFile != "" {
		extra, err := os.ReadFile(config.TrustFile)
		if err != nil {
			return nil, fmt.Errorf("the certificate to trust: %w", err)
		}
		if !roots.AppendCertsFromPEM(extra) {
			return nil, fmt.Errorf("%s holds no PEM certificate", config.TrustFile)
		}
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{certificate}, RootCAs: roots, MinVersion: tls.VersionTLS12}
	return &Partner{
		partnerURL:     config.PartnerURL,
		packetPulseURL: config.PacketPulseURL,
		apiKey:         config.ApiKey,
		mutual:         &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{TLSClientConfig: tlsConfig}},
		plain:          &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}}},
	}, nil
}

// Launch asks PacketPulse for a link that opens Run tests for engineer,
// signed in, with ticket filled in.
//
// returnURL, when not empty, is the app's own page for the ticket: Run tests
// offers to send the engineer back to it once they are done.
func (p *Partner) Launch(ctx context.Context, engineer Engineer, ticket Ticket, returnURL string) (*LaunchLink, error) {
	fields := map[string]string{
		"engineer_id": engineer.Id, "email": engineer.Email, "name": engineer.Name,
		"tt_number": ticket.TtNumber, "customer_id": ticket.CustomerId,
	}
	if returnURL != "" {
		fields["return_url"] = returnURL
	}
	body, _ := json.Marshal(fields)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.partnerURL+"/api/v1/partner/launch", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", DefaultAppName+" server")
	var link LaunchLink
	if err := p.do(p.mutual, request, http.StatusCreated, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// Result reads a ticket's latest results back by TT number.
func (p *Partner) Result(ctx context.Context, ttNumber string) (map[string]any, error) {
	if p.apiKey == "" || p.packetPulseURL == "" {
		return nil, fmt.Errorf("no Results API key or PacketPulse address is set for this app")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		p.packetPulseURL+"/api/v1/result/bytt/"+url.PathEscape(ttNumber), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+p.apiKey)
	result := map[string]any{}
	if err := p.do(p.plain, request, http.StatusOK, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// do sends a request and reads the answer it wants, or PacketPulse's
// refusal.
func (p *Partner) do(client *http.Client, request *http.Request, want int, into any) error {
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("PacketPulse could not be reached: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if response.StatusCode != want {
		envelope := struct {
			Error Refusal `json:"error"`
		}{}
		_ = json.Unmarshal(raw, &envelope)
		envelope.Error.Status = response.StatusCode
		if envelope.Error.Message == "" {
			envelope.Error.Message = unexplained(response.StatusCode)
		}
		return &envelope.Error
	}
	if err := json.Unmarshal(raw, into); err != nil {
		return fmt.Errorf("PacketPulse's answer could not be read: %w", err)
	}
	return nil
}
