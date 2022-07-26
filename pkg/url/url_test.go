package url

import (
	"testing"

	"gotest.tools/assert"
)

type urlTest struct {
	input  string
	scheme string
	port   int
	want   string
	err    string
}

func TestURLValidation(t *testing.T) {
	expectedOutput := "https://my-host.delivery.puppetlabs.net:8081"

	tests := []urlTest{
		{input: "https://my-host.delivery.puppetlabs.net:8081", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "https://my-host.delivery.puppetlabs.net", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "my-host.delivery.puppetlabs.net:8081", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "my-host.delivery.puppetlabs.net", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "https://my-host.delivery.puppetlabs.net/", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "https://my-host.delivery.puppetlabs.net?xyz=abc123", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "my-host.delivery.puppetlabs.net/test?xyz=abc123", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "my-host.delivery.puppetlabs.net/test/something", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "https://my-host.delivery.puppetlabs.net/test/something", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "my-host.delivery.puppetlabs.net/:8888", port: 8081, scheme: "https", want: expectedOutput, err: ""},
		{input: "my-host.delivery.puppetlabs.net:9999/test?xyz=abc123", port: 8081, scheme: "http", want: "http://my-host.delivery.puppetlabs.net:9999", err: ""},
		{input: "http://my-host.delivery.puppetlabs.net:8888", port: 8081, scheme: "", want: "http://my-host.delivery.puppetlabs.net:8888", err: ""},
		{input: "my-host.delivery.puppetlabs.net", port: 0, scheme: "https", want: "https://my-host.delivery.puppetlabs.net", err: ""},
		{input: "my-host.delivery.puppetlabs.net", port: 0, scheme: "", want: "", err: "no scheme available"},
		{input: "", port: 9999, scheme: "http", want: "", err: "input cannot be empty"},
	}

	for _, test := range tests {
		output, err := BuildURL(test.input, test.scheme, test.port)
		if len(test.err) > 0 {
			assert.Error(t, err, test.err)
		} else {
			assert.NilError(t, err)
		}
		assert.Equal(t, test.want, output)
	}
}
