package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/puppetlabs/go-libs/pkg/certificate"
)

var hostnames certificate.HostNames

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current working directory %s", err)
		os.Exit(1)
	}

	flag.Var(&hostnames, "hostname", "Hostname(s) for the certificate.")
	commonName := flag.String("cn", "localhost", "common name for certificate")
	directory := flag.String("directory", wd, "output to generate certs to")
	generateCAFiles := flag.Bool("cafiles", false, "whether to output generated CA certs or not")
	caCert := flag.String("cacertfile", "", "The location of the CA certificate file.")
	caKey := flag.String("cakeyfile", "", "The location of the CA key file.")

	var CAKeyPair *certificate.KeyPair
	flag.Parse()

	if caCert != nil && len(*caCert) > 0 && caKey != nil && len(*caKey) > 0 {
		certBytes, err := ioutil.ReadFile(*caCert)
		if err != nil {
			fmt.Printf("Failed to read CA certificate file %s", err)
			os.Exit(1)
		}

		keyBytes, err := ioutil.ReadFile(*caKey)
		if err != nil {
			fmt.Printf("Failed to read CA private key file %s", err)
			os.Exit(1)
		}
		CAKeyPair = &certificate.KeyPair{Certificate: certBytes, PrivateKey: keyBytes}
	} else {
		CAKeyPair, err = certificate.GenerateCA()
		if err != nil {
			fmt.Printf("Failed to generate CA cerificate :%s", err)
			os.Exit(1)
		}
	}

	if len(hostnames) == 0 {
		hostnames = []string{"localhost"}
	}

	certKeyPair, err := certificate.GenerateSignedCert(CAKeyPair, hostnames, *commonName)
	if err != nil {
		fmt.Printf("Failed to generate TLS cerificate :%s", err)
		os.Exit(3)
	}

	if generateCAFiles != nil && *generateCAFiles {
		err = ioutil.WriteFile(fmt.Sprintf(filepath.Join(filepath.Clean(*directory), "ca.crt")), CAKeyPair.Certificate, 0o600)
		if err != nil {
			fmt.Printf("Failed to write CA certificate file to disk :%s", err)
			os.Exit(3)
		}

		err = ioutil.WriteFile(fmt.Sprintf(filepath.Join(filepath.Clean(*directory), "ca.key")), CAKeyPair.PrivateKey, 0o600)
		if err != nil {
			fmt.Printf("Failed to write CA key file to disk: %s.", err)
			os.Exit(3)
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(filepath.Join(filepath.Clean(*directory), "tls.crt")), certKeyPair.Certificate, 0o600)
	if err != nil {
		fmt.Printf("Failed to write TLS cert file to disk: %s.", err)
		os.Exit(3)
	}

	err = ioutil.WriteFile(fmt.Sprintf(filepath.Join(filepath.Clean(*directory), "tls.key")), certKeyPair.PrivateKey, 0o600)
	if err != nil {
		fmt.Printf("Failed to write TLS key file to disk: %s.", err)
		os.Exit(3)
	}
}
