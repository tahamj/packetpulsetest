// Package fieldengapp is "Field Eng App": a mock of the operator's own field
// engineer app, for testing and demonstrating how it opens PacketPulse's Run
// tests page for an engineer it has already signed in (migration 0043).
//
// It is what the operator's app and its server do, in one small program:
//
//   - a web app the engineer uses on their phone: they sign in to the app,
//     see their tickets, and tap "Test this line in PacketPulse";
//   - the app's SERVER, which holds the client certificate and asks
//     PacketPulse's partner address for a link over mutual TLS, naming the
//     engineer, then sends the phone to the link it gets back;
//   - the result read back by TT number through the read-only Results API.
//
// It runs the same against a local PacketPulse (with Front standing in for
// nginx's partner site) and against the live one on the Oracle host, through
// real DNS, real TLS and the real nginx.
package fieldengapp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DefaultAppName is the mock's name, as the customer sees it in the demo.
const DefaultAppName = "Field Eng App"

// Engineer is one of the app's own engineers: the app's id for them, their
// email and name as the app sends them to PacketPulse, and the mock's own
// password for signing in to the app itself.
type Engineer struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Ticket is a trouble ticket an engineer works in the app.
type Ticket struct {
	TtNumber   string `json:"tt_number"`
	CustomerId string `json:"customer_id"`
	Customer   string `json:"customer"`
	Address    string `json:"address"`
	Fault      string `json:"fault"`

	// What a field app shows of a job beside the ticket itself. All are
	// optional in a tickets file: a job without them shows what it has.

	// Priority is P1, P2 or P3: P1 first.
	Priority string `json:"priority,omitempty"`
	// Area is the locality, for the jobs list; Address is the full one.
	Area string `json:"area,omitempty"`
	// Plan is the customer's broadband plan.
	Plan string `json:"plan,omitempty"`
	// Window is the appointment the customer agreed, such as 10:00-12:00.
	Window string `json:"window,omitempty"`
	// DueInMinutes is how long after the app starts the job's SLA runs out.
	DueInMinutes int `json:"due_in_minutes,omitempty"`
	// Phone is the customer's number as the app shows it, masked.
	Phone string `json:"phone,omitempty"`
	// Ont is the serial of the customer's optical terminal.
	Ont string `json:"ont,omitempty"`
	// Opened is when the customer reported the fault, in words.
	Opened string `json:"opened,omitempty"`
}

// Config is everything the mock needs. Only the partner address and the
// certificate are required to ask for links; the API key only to read
// results back.
type Config struct {
	AppName string
	// PartnerURL is PacketPulse's partner address, the only one that asks
	// for a certificate: https://partner.packetpulse.rummaan53.com live.
	PartnerURL string
	// PacketPulseURL is PacketPulse's ordinary address, for the Results API.
	PacketPulseURL string
	// PublicURL is this app's own https address, as phones reach it, for
	// the way back from PacketPulse. Empty takes it from each request that
	// came in over https; over plain http there is no way back to offer, as
	// PacketPulse takes only an https address.
	PublicURL string
	// CertificateFile and KeyFile are the app server's client certificate
	// and its private key, PEM.
	CertificateFile string
	KeyFile         string
	// TrustFile is an extra certificate to trust for the partner address,
	// PEM: a local Front's own. Empty trusts the system's authorities only.
	TrustFile string
	// ApiKey is a read-only Results API key of the same organisation.
	ApiKey string
	// SessionSecret signs the mock's own sign-in cookie.
	SessionSecret []byte

	Engineers []Engineer
	Tickets   []Ticket
}

// DefaultEngineers are the mock's engineers. Their addresses are under
// .example (RFC 2606): PacketPulse never mails an engineer the app signs in.
func DefaultEngineers() []Engineer {
	return []Engineer{
		{Id: "FE-1001", Email: "asha.rao@field-eng-app.example", Name: "Asha Rao", Password: "field1001"},
		{Id: "FE-1002", Email: "vikram.shah@field-eng-app.example", Name: "Vikram Shah", Password: "field1002"},
		{Id: "FE-1003", Email: "meera.iyer@field-eng-app.example", Name: "Meera Iyer", Password: "field1003"},
	}
}

// DefaultTickets are the mock's open jobs: one engineer's day in one city,
// as a field app would hand it out. Every number and serial is made up.
func DefaultTickets() []Ticket {
	return []Ticket{
		{TtNumber: "TT-2026-01458", CustomerId: "CUST-40217", Customer: "R. Menon",
			Address: "Flat 1204, Lodha Splendora, Andheri West, Mumbai 400053", Area: "Andheri West",
			Fault: "Slow internet in the evenings", Priority: "P2", Plan: "Fiber 300 Mbps",
			Window: "10:00–12:00", DueInMinutes: 150, Phone: "+91 98••• ••412", Ont: "ZTEG4F2A91C7", Opened: "Yesterday, 21:40"},
		{TtNumber: "TT-2026-00752", CustomerId: "CUST-77230", Customer: "P. Nair",
			Address: "B-302, Hiranandani Gardens, Powai, Mumbai 400076", Area: "Powai",
			Fault: "Connection drops every few minutes", Priority: "P1", Plan: "Fiber 500 Mbps",
			Window: "12:00–14:00", DueInMinutes: 50, Phone: "+91 99••• ••230", Ont: "HWTC8C1D0E44", Opened: "Today, 07:15"},
		{TtNumber: "TT-2026-00744", CustomerId: "CUST-90115", Customer: "S. Kulkarni",
			Address: "14 Hill Road, Bandra West, Mumbai 400050", Area: "Bandra West",
			Fault: "Video calls freeze", Priority: "P2", Plan: "Fiber 150 Mbps",
			Window: "14:00–16:00", DueInMinutes: 240, Phone: "+91 97••• ••115", Ont: "ZTEG77B3C2A1", Opened: "Yesterday, 18:05"},
		{TtNumber: "TT-2026-00768", CustomerId: "CUST-81904", Customer: "A. Gupta",
			Address: "Oberoi Splendor, Goregaon East, Mumbai 400060", Area: "Goregaon East",
			Fault: "Online games lag", Priority: "P3", Plan: "Fiber 100 Mbps",
			Window: "16:00–18:00", DueInMinutes: 420, Phone: "+91 90••• ••904", Ont: "HWTC2F90A7B3", Opened: "2 days ago"},
	}
}

// ConfigFromEnvironment reads FIELD_ENG_* settings. A file named by
// FIELD_ENG_ENGINEERS_FILE or FIELD_ENG_TICKETS_FILE replaces the defaults.
func ConfigFromEnvironment() (Config, error) {
	config := Config{
		AppName:         firstSet(os.Getenv("FIELD_ENG_APP_NAME"), DefaultAppName),
		PartnerURL:      strings.TrimRight(os.Getenv("FIELD_ENG_PARTNER_URL"), "/"),
		PacketPulseURL:  strings.TrimRight(os.Getenv("FIELD_ENG_PACKETPULSE_URL"), "/"),
		PublicURL:       strings.TrimRight(os.Getenv("FIELD_ENG_PUBLIC_URL"), "/"),
		CertificateFile: os.Getenv("FIELD_ENG_CERT_FILE"),
		KeyFile:         os.Getenv("FIELD_ENG_KEY_FILE"),
		TrustFile:       os.Getenv("FIELD_ENG_TRUST_FILE"),
		ApiKey:          os.Getenv("FIELD_ENG_API_KEY"),
		SessionSecret:   []byte(os.Getenv("FIELD_ENG_SESSION_SECRET")),
		Engineers:       DefaultEngineers(),
		Tickets:         DefaultTickets(),
	}
	if path := os.Getenv("FIELD_ENG_ENGINEERS_FILE"); path != "" {
		if err := readJSON(path, &config.Engineers); err != nil {
			return Config{}, err
		}
	}
	if path := os.Getenv("FIELD_ENG_TICKETS_FILE"); path != "" {
		if err := readJSON(path, &config.Tickets); err != nil {
			return Config{}, err
		}
	}
	if config.PartnerURL == "" || config.CertificateFile == "" || config.KeyFile == "" {
		return Config{}, fmt.Errorf("FIELD_ENG_PARTNER_URL, FIELD_ENG_CERT_FILE and FIELD_ENG_KEY_FILE must be set")
	}
	return config, nil
}

func readJSON(path string, into any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, into); err != nil {
		return fmt.Errorf("%s is not the JSON expected: %w", path, err)
	}
	return nil
}

func firstSet(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
