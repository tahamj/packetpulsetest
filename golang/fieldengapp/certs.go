package fieldengapp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"
)

// KeyPair is a certificate and its private key, PEM.
type KeyPair struct {
	CertificatePEM []byte
	KeyPEM         []byte
}

// NewClientCertificate makes what the operator makes for its app's server
// with `openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 -days
// 397`: a self-signed EC P-256 certificate, valid from now for validity.
func NewClientCertificate(commonName string, validity time.Duration) (KeyPair, error) {
	return newCertificate(commonName, validity, nil)
}

// NewServerCertificate is a self-signed certificate for a local Front to
// serve TLS with, for localhost.
func NewServerCertificate() (KeyPair, error) {
	return newCertificate("localhost", 24*time.Hour, func(template *x509.Certificate) {
		template.DNSNames = []string{"localhost"}
		template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	})
}

func newCertificate(commonName string, validity time.Duration, change func(*x509.Certificate)) (KeyPair, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return KeyPair{}, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	if err != nil {
		return KeyPair{}, err
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{Organization: []string{DefaultAppName}, CommonName: commonName},
		NotBefore:    now.Add(-time.Minute), NotAfter: now.Add(validity),
	}
	if change != nil {
		change(template)
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return KeyPair{}, err
	}
	keyDer, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return KeyPair{}, err
	}
	return KeyPair{
		CertificatePEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		KeyPEM:         pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDer}),
	}, nil
}

// Fingerprint is a PEM certificate's SHA-256 as `openssl x509 -fingerprint
// -sha256` prints it, and as PacketPulse shows it after the certificate is
// registered: the two are compared by eye.
func Fingerprint(certificatePEM []byte) (string, error) {
	block, _ := pem.Decode(certificatePEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("not a PEM certificate")
	}
	sum := sha256.Sum256(block.Bytes)
	parts := make([]string, len(sum))
	for index, b := range sum {
		parts[index] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, ":"), nil
}
