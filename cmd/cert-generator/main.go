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

const (
	errorExitCode             = 1
	fileModeUserReadWriteOnly = 0o600
)

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
			os.Exit(errorExitCode)
		}

		keyBytes, err := ioutil.ReadFile(*caKey)
		if err != nil {
			fmt.Printf("Failed to read CA private key file %s", err)
			os.Exit(errorExitCode)
		}
		CAKeyPair = &certificate.KeyPair{Certificate: certBytes, PrivateKey: keyBytes}
	} else {
		CAKeyPair, err = certificate.GenerateCA()
		if err != nil {
			fmt.Printf("Failed to generate CA cerificate :%s", err)
			os.Exit(errorExitCode)
		}
	}

	if len(hostnames) == 0 {
		hostnames = []string{"localhost"}
	}

	certKeyPair, err := certificate.GenerateSignedCert(CAKeyPair, hostnames, *commonName)
	if err != nil {
		fmt.Printf("Failed to generate TLS cerificate :%s", err)
		os.Exit(errorExitCode)
	}

	if generateCAFiles != nil && *generateCAFiles {
		err = ioutil.WriteFile(fmt.Sprint(filepath.Join(filepath.Clean(*directory), "ca.crt")), CAKeyPair.Certificate,
			fileModeUserReadWriteOnly)
		if err != nil {
			fmt.Printf("Failed to write CA certificate file to disk :%s", err)
			os.Exit(errorExitCode)
		}

		err = ioutil.WriteFile(fmt.Sprint(filepath.Join(filepath.Clean(*directory), "ca.key")), CAKeyPair.PrivateKey,
			fileModeUserReadWriteOnly)
		if err != nil {
			fmt.Printf("Failed to write CA key file to disk: %s.", err)
			os.Exit(errorExitCode)
		}
	}

	err = ioutil.WriteFile(fmt.Sprint(filepath.Join(filepath.Clean(*directory), "tls.crt")), certKeyPair.Certificate,
		fileModeUserReadWriteOnly)
	if err != nil {
		fmt.Printf("Failed to write TLS cert file to disk: %s.", err)
		os.Exit(errorExitCode)
	}

	err = ioutil.WriteFile(fmt.Sprint(filepath.Join(filepath.Clean(*directory), "tls.key")), certKeyPair.PrivateKey,
		fileModeUserReadWriteOnly)
	if err != nil {
		fmt.Printf("Failed to write TLS key file to disk: %s.", err)
		os.Exit(errorExitCode)
	}
}
