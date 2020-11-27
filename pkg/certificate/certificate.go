package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// KeyPair stores a PEM encoded certificate and
// a PEM encoded RSA private key
type KeyPair struct {
	Certificate []byte
	PrivateKey  []byte
}

var subject = pkix.Name{
	Organization:       []string{"Puppet, Inc"},
	OrganizationalUnit: []string{"Estate Reporting internal"},
	Country:            []string{"US"},
	Province:           []string{"Oregon"},
	Locality:           []string{"Portland"},
}

//HostNames contains the list of hosts the cert will be generated for
type HostNames []string

func (h *HostNames) String() string {
	var output string
	for _, host := range *h {
		output = fmt.Sprintf("%s %s", output, host)
	}
	return output
}

//Set will add the hostname to the hostname array
func (h *HostNames) Set(value string) error {
	*h = append(*h, value)
	return nil
}

// GenerateCA will generate a new CA key/cert pair
func GenerateCA() (*KeyPair, error) {
	// choose a random number between 0 and 999999999999999999
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(999999999999999999))
	// multiply by 100 to get it up to 20 digits. hard coding it overflows int64
	serialNum := randomNum.Mul(randomNum, big.NewInt(100))

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("can't create private key because: %s", err)
	}

	marshalPublicKey, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return nil, fmt.Errorf("can't marshal public key because: %s", err)
	}

	subjectKeyID := sha256.Sum256(marshalPublicKey)

	template := &x509.Certificate{
		SerialNumber:          serialNum,
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 8760),
		IsCA:                  true,
		SubjectKeyId:          subjectKeyID[:],
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	template.Subject.CommonName = "Puppet Estate Reporting SelfSign CA"

	certificate, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("can't create certificate because: %s", err)
	}

	keyPair := KeyPair{
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}),
	}

	return &keyPair, nil
}

// GenerateSignedCert will generate a new signed certificate signed by the input CA key/cert pair with one of multiple hostnames
// and with the given CN.
func GenerateSignedCert(ca *KeyPair, hostnames HostNames, commonName string) (*KeyPair, error) {
	if ca == nil {
		return nil, fmt.Errorf("nil pointer for root CA key pair")
	}

	tlsKeyPair, err := tls.X509KeyPair(ca.Certificate, ca.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("can't convert to X509KeyPair because: %s", err)
	}

	caCert, err := x509.ParseCertificate(tlsKeyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("can't parse ca certificate because: %s", err)
	}

	// choose a random number between 0 and 999999999999999999
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(999999999999999999))
	// multiply by 100 to get it up to 20 digits. hard coding it overflows int64
	serialNum := randomNum.Mul(randomNum, big.NewInt(100))

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("can't create private key because: %s", err)
	}

	marshalPublicKey, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return nil, fmt.Errorf("can't marshal public key because: %s", err)
	}

	subjectKeyID := sha256.Sum256(marshalPublicKey)

	var dnsNames []string
	var ips []net.IP
	if len(hostnames) > 0 {
		for _, hostname := range hostnames {
			ipList, err := net.LookupHost(hostname)
			if err == nil {
				for _, ip := range ipList {
					ips = append(ips, net.ParseIP(ip))
				}
			} else {
				fmt.Printf("Could not resolve hostname %s\n", hostname)
			}
			dnsNames = append(dnsNames, hostname)
		}
	}

	template := &x509.Certificate{
		SerialNumber: serialNum,
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 8760),
		SubjectKeyId: subjectKeyID[:],
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     dnsNames,
		IPAddresses:  ips,
	}

	if commonName == "" {
		template.Subject.CommonName = "localhost"
	} else {
		template.Subject.CommonName = commonName
	}

	certificate, err := x509.CreateCertificate(rand.Reader, template, caCert, &privateKey.PublicKey, tlsKeyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("can't create certificate because: %s", err)
	}

	keyPair := KeyPair{
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}),
	}

	return &keyPair, nil
}
