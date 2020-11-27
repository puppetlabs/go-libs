package certificate

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"
)

func TestGenerateCAWorks(t *testing.T) {
	keyPair, err := GenerateCA()
	if err != nil {
		t.Errorf("Unable to generate cert due to %s", err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(keyPair.Certificate))
	if !ok {
		t.Error("failed to parse root certificate")
	}
}

func TestGenerateCertLocalhostWorks(t *testing.T) {
	rootKeyPair, err := GenerateCA()
	if err != nil {
		t.Errorf("Unable to generate cert due to %s", err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootKeyPair.Certificate))
	if !ok {
		t.Error("failed to parse root certificate")
	}

	keyPair, err := GenerateSignedCert(rootKeyPair, []string{"localhost"}, "localhost")

	block, _ := pem.Decode([]byte(keyPair.Certificate))
	if block == nil {
		t.Error("failed to decode certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Errorf("failed to parse certificate: %s", err)
	}

	opts := x509.VerifyOptions{
		DNSName: "localhost",
		Roots:   roots,
	}

	if _, err := cert.Verify(opts); err != nil {
		t.Errorf("failed to verify certificate: %s", err)
	}
}

func TestGenerateCertNonLocalhostWorks(t *testing.T) {
	rootKeyPair, err := GenerateCA()
	if err != nil {
		t.Errorf("Unable to generate cert due to %s", err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootKeyPair.Certificate))
	if !ok {
		t.Error("failed to parse root certificate")
	}

	keyPair, err := GenerateSignedCert(rootKeyPair, []string{"externalhost"}, "externalhost")

	block, _ := pem.Decode([]byte(keyPair.Certificate))
	if block == nil {
		t.Error("failed to decode certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Errorf("failed to parse certificate: %s", err)
	}

	opts := x509.VerifyOptions{
		DNSName: "externalhost",
		Roots:   roots,
	}

	if _, err := cert.Verify(opts); err != nil {
		t.Errorf("failed to verify certificate: %s", err.Error())
	}
}

func TestNilRootCertFailsCertGeneration(t *testing.T) {
	var rootKeyPair *KeyPair
	_, err := GenerateSignedCert(rootKeyPair, []string{"externalhost"}, "externalhost")
	if err == nil {
		t.Error("expected error when using nil for root CA pair")
	}
}

func TestInvalidRootCertFailsCertGeneration(t *testing.T) {
	rootKeyPair := &KeyPair{Certificate: []byte{},
		PrivateKey: []byte{}}
	_, err := GenerateSignedCert(rootKeyPair, []string{"externalhost"}, "externalhost")
	fmt.Println(err)
	if err == nil {
		t.Error("expected error when using invalid root CA pair")
	}
}
