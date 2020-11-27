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
	CAKeypair, err := certificate.GenerateCA()
	if err != nil {
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	flag.Var(&hostnames, "hostname", "Hostname(s) for the certificate.")
	commonName := flag.String("cn", "localhost", "common name for certificate")
	directory := flag.String("directory", wd, "output to generate certs to")
	generateCAFiles := flag.Bool("cafiles", false, "whether to output generated CA certs or not")

	flag.Parse()

	if len(hostnames) == 0 {
		hostnames = []string{"localhost"}
	}

	certKeyPair, err := certificate.GenerateSignedCert(CAKeypair, hostnames, *commonName)
	if err != nil {
		fmt.Println("Unable to generate siged cert.")
		os.Exit(3)
	}

	if generateCAFiles != nil && *generateCAFiles {
		err = ioutil.WriteFile(fmt.Sprintf(filepath.Join(filepath.Clean(*directory), "ca.crt")), CAKeypair.Certificate, 0600)
		if err != nil {
			fmt.Println("Failed to write CA certificate")
			os.Exit(3)
		}
		err = ioutil.WriteFile(fmt.Sprintf(filepath.Join(filepath.Clean(*directory), "ca.key")), CAKeypair.Certificate, 0600)
		if err != nil {
			fmt.Println("Failed to write CA key.")
			os.Exit(3)
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(filepath.Join(filepath.Clean(*directory), "tls.crt")), certKeyPair.Certificate, 0600)
	if err != nil {
		fmt.Println("Failed to write TLS certificate")
		os.Exit(3)
	}
	err = ioutil.WriteFile(fmt.Sprintf(filepath.Join(filepath.Clean(*directory), "tls.key")), certKeyPair.Certificate, 0600)
	if err != nil {
		fmt.Println("Failed to write TLS key.")
		os.Exit(3)
	}

}
