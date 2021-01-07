package url

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	schemeSeparator = "://"
	pathSeparator   = "/"
	querySeparator  = "?"
	portSeparator   = ":"
)

// BuildURL assumes input is a url and converts it into the format: scheme://host[:<port>]
func BuildURL(input string, defaultScheme string, defaultPort int) (string, error) {
	var result string

	if len(input) == 0 {
		return "", fmt.Errorf("Input cannot be empty")
	}

	// Save the scheme
	var scheme string
	if strings.Contains(input, schemeSeparator) {
		parts := strings.Split(input, schemeSeparator)
		scheme = parts[0]
		result = parts[1]
	} else {
		if len(defaultScheme) == 0 {
			return "", fmt.Errorf("No scheme available")
		}
		scheme = defaultScheme
		result = input

	}

	// Remove path
	if strings.Contains(result, pathSeparator) {
		result = strings.Split(result, pathSeparator)[0]
	}

	// Remove query string
	if strings.Contains(result, querySeparator) {
		result = strings.Split(result, querySeparator)[0]
	}

	// Add the default port
	if defaultPort != 0 && !strings.Contains(result, portSeparator) {
		result = fmt.Sprintf("%s%s%d", result, portSeparator, defaultPort)
	}

	// Put back the scheme
	result = fmt.Sprintf("%s%s%s", scheme, schemeSeparator, result)

	// Parse the URL and check for errors
	_, err := url.Parse(result)
	if err != nil {
		return "", fmt.Errorf("URL parse error: %w", err)
	}

	return result, nil
}
