package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/puppetlabs/go-libs/pkg/certificate"
	"github.com/puppetlabs/go-libs/pkg/util"
)

const (
	fileModeUserReadWriteOnly                       = 0o600
	fileModeUserReadWriteExecuteGroupReadOthersRead = 0o744
)

var (
	serviceDir            string
	name                  string
	listenAddress         string
	listenPort            string
	tlsSetup              string
	corsEnabled           string
	readinessCheckEnabled string
	rateLimit             string
	rateInterval          string
	metricsEnabled        string

	certPrefix = "server"
	certSuffix = "crt"
	keySuffix  = "key"

	boolMap = map[string]bool{
		"y": true,
		"n": false,
	}
)

// Substitution will hold the values to be substituted in the templates.
type Substitution struct {
	Name                  string
	Port                  string
	ListenAddress         string
	CertFile              string
	KeyFile               string
	CorsEnabled           bool
	ReadinessCheckEnabled bool
	MetricsEnabled        bool
	RateLimit             int
	RateInterval          int
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func writeOutputFile(inputFile string, subst Substitution, outputFile string) error {
	tmpl, err := template.ParseFiles(inputFile)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	var tmplOutput bytes.Buffer
	err = tmpl.Execute(&tmplOutput, subst)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err = os.WriteFile(outputFile, tmplOutput.Bytes(), fileModeUserReadWriteOnly)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func generateCerts(filepath string) error {
	CAKeypair, err := certificate.GenerateCA()
	if err != nil {
		os.Exit(1)
	}

	certKeyPair, err := certificate.GenerateSignedCert(CAKeypair, []string{"localhost"}, "localhost")
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err = os.WriteFile(fmt.Sprintf("%s.%s", filepath, certSuffix), certKeyPair.Certificate,
		fileModeUserReadWriteOnly)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err = os.WriteFile(fmt.Sprintf("%s.%s", filepath, keySuffix), certKeyPair.PrivateKey,
		fileModeUserReadWriteOnly)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func main() {
	dir, err := os.Getwd()
	checkError(err)

	rateLimitInt, err := strconv.Atoi(rateLimit)
	if err != nil {
		rateLimitInt = 0
	}
	rateIntervalInt, err := strconv.Atoi(rateInterval)
	if err != nil {
		rateIntervalInt = 0
	}

	var tlsCertFile string
	var tlsKeyFile string

	checkError(os.MkdirAll(filepath.Join(serviceDir, "cmd", name), fileModeUserReadWriteExecuteGroupReadOthersRead))
	checkError(os.MkdirAll(filepath.Join(serviceDir, "pkg", "config"), fileModeUserReadWriteExecuteGroupReadOthersRead))
	checkError(os.MkdirAll(filepath.Join(serviceDir, "pkg", "handlers"),
		fileModeUserReadWriteExecuteGroupReadOthersRead))
	checkError(util.FileCopy(filepath.Join(dir, "internal", "tmpl", "handlers.go"),
		filepath.Join(serviceDir, "pkg", "handlers", "handlers.go")))
	checkError(util.FileCopy(filepath.Join(dir, "internal", "tmpl", "config_test.go.tmpl"),
		filepath.Join(serviceDir, "pkg", "config", "config_test.go")))
	if boolMap[tlsSetup] {
		checkError(generateCerts(filepath.Join(serviceDir, certPrefix)))
		tlsCertFile = fmt.Sprintf("%s.%s", certPrefix, certSuffix)
		tlsKeyFile = fmt.Sprintf("%s.%s", certPrefix, keySuffix)
	}

	substitution := Substitution{
		Name:                  name,
		Port:                  listenPort,
		ListenAddress:         listenAddress,
		CertFile:              tlsCertFile,
		KeyFile:               tlsKeyFile,
		MetricsEnabled:        boolMap[metricsEnabled],
		CorsEnabled:           boolMap[corsEnabled],
		ReadinessCheckEnabled: boolMap[readinessCheckEnabled],
		RateLimit:             rateLimitInt,
		RateInterval:          rateIntervalInt,
	}

	checkError(writeOutputFile(filepath.Join(dir, "internal", "tmpl", "config.go.tmpl"), substitution,
		filepath.Join(serviceDir, "pkg", "config", "config.go")))
	checkError(writeOutputFile(filepath.Join(dir, "internal", "tmpl", "main.go.tmpl"), substitution,
		filepath.Join(serviceDir, "cmd", name, "main.go")))
	checkError(writeOutputFile(filepath.Join(dir, "internal", "tmpl", "docker-compose.yml.tmpl"), substitution,
		filepath.Join(serviceDir, "docker-compose.yml")))
	checkError(writeOutputFile(filepath.Join(dir, "internal", "tmpl", "go.mod.tmpl"), substitution,
		filepath.Join(serviceDir, "go.mod")))
	checkError(writeOutputFile(filepath.Join(dir, "internal", "tmpl", "Dockerfile.tmpl"), substitution,
		filepath.Join(serviceDir, "Dockerfile")))
	checkError(writeOutputFile(filepath.Join(dir, "internal", "tmpl", "Makefile.tmpl"), substitution,
		filepath.Join(serviceDir, "Makefile")))
}
