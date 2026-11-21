// Command fieldengapp runs "Field Eng App", the mock of an operator's own
// field engineer app that opens PacketPulse's Run tests page without a
// second sign-in (see package fieldengapp).
//
//	fieldengapp keygen --cert app.crt --key app.key   make the app server's certificate
//	fieldengapp front --listen 127.0.0.1:18443 --upstream http://127.0.0.1:8080 --trust-out front.crt
//	                                                    stand in for nginx's partner site, locally
//	fieldengapp serve                                   run the app (FIELD_ENG_* settings)
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tahamj/packetpulsetest/fieldengapp"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(args []string, out io.Writer) int {
	command := "serve"
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}
	var err error
	switch command {
	case "keygen":
		err = keygen(args, out)
	case "front":
		err = front(args)
	case "serve":
		err = serve()
	default:
		err = fmt.Errorf("unknown command %q: use keygen, front or serve", command)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fieldengapp:", err)
		return 1
	}
	return 0
}

// keygen makes the app server's certificate and key, as the operator would
// with OpenSSL, and prints the fingerprint the person registering it compares with the one
// PacketPulse shows after the upload. The key is written readable by its
// owner alone.
func keygen(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("keygen", flag.ContinueOnError)
	certPath := flags.String("cert", "field-eng-app.crt", "where to write the certificate (PEM)")
	keyPath := flags.String("key", "field-eng-app.key", "where to write the private key (PEM)")
	name := flags.String("cn", fieldengapp.DefaultAppName+" PacketPulse launch", "the certificate's common name")
	days := flags.Int("days", 397, "how many days it is valid for")
	if err := flags.Parse(args); err != nil {
		return err
	}
	pair, err := fieldengapp.NewClientCertificate(*name, time.Duration(*days)*24*time.Hour)
	if err != nil {
		return err
	}
	if err := os.WriteFile(*keyPath, pair.KeyPEM, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(*certPath, pair.CertificatePEM, 0o644); err != nil {
		return err
	}
	fingerprint, _ := fieldengapp.Fingerprint(pair.CertificatePEM)
	fmt.Fprintf(out, "certificate: %s\nkey:         %s (keep it on this server only)\nSHA-256:     %s\n", *certPath, *keyPath, fingerprint)
	return nil
}

// front stands in for nginx's partner site on a machine without nginx.
func front(args []string) error {
	flags := flag.NewFlagSet("front", flag.ContinueOnError)
	listen := flags.String("listen", "127.0.0.1:18443", "where to serve TLS")
	upstream := flags.String("upstream", "http://127.0.0.1:8080", "PacketPulse's own address")
	trustOut := flags.String("trust-out", "front.crt", "where to write this front's certificate, for the app to trust")
	if err := flags.Parse(args); err != nil {
		return err
	}
	target, err := url.Parse(*upstream)
	if err != nil || target.Host == "" {
		return fmt.Errorf("--upstream must be an address such as http://127.0.0.1:8080")
	}
	pair, err := fieldengapp.NewServerCertificate()
	if err != nil {
		return err
	}
	if err := os.WriteFile(*trustOut, pair.CertificatePEM, 0o644); err != nil {
		return err
	}
	certificate, err := tls.X509KeyPair(pair.CertificatePEM, pair.KeyPEM)
	if err != nil {
		return err
	}
	server := &http.Server{Addr: *listen, Handler: fieldengapp.FrontHandler(target),
		TLSConfig: fieldengapp.FrontTLS(certificate), ReadHeaderTimeout: 10 * time.Second}
	slog.Info("partner site stand-in", "listen", "https://"+*listen, "upstream", *upstream, "trust", *trustOut)
	return until(server, func() error { return server.ListenAndServeTLS("", "") })
}

// serve runs the app itself.
func serve() error {
	config, err := fieldengapp.ConfigFromEnvironment()
	if err != nil {
		return err
	}
	partner, err := fieldengapp.NewPartner(config)
	if err != nil {
		return err
	}
	listen := os.Getenv("FIELD_ENG_LISTEN")
	if listen == "" {
		listen = "127.0.0.1:18090"
	}
	server := &http.Server{Addr: listen, Handler: fieldengapp.NewApp(config, partner), ReadHeaderTimeout: 10 * time.Second}
	slog.Info(config.AppName+" ready", "listen", "http://"+listen, "partner", config.PartnerURL,
		"results", config.PacketPulseURL != "" && config.ApiKey != "")
	return until(server, server.ListenAndServe)
}

// until serves until the process is told to stop, then stops gracefully.
func until(server *http.Server, listen func() error) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	failed := make(chan error, 1)
	go func() { failed <- listen() }()
	select {
	case err := <-failed:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
}
