package config

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var yamlExample = []byte(`db:
  dbName: mydb
  host: abc
  port: 5432
  sslMode: disable
  user: postgres
webserver:
  apiReadTimeout: 120s
  apiWriteTimeout: 120s`)

// StructNoTags holds one field with no tags against the struct
type StructNoTags struct {
	DummyString string
}

// StructMissingEnvTag holds one field with no env tag
type StructMissingEnvTag struct {
	TestVal string `default:"xyz"`
}

// StructMissingDefaultTags holds one field with no default tag
type StructMissingDefaultTags struct {
	TestString string `env:"TEST_STRING"`
	TestInt    int    `env:"TEST_INT"`
	TestBool   bool   `env:"TEST_BOOL"`
}

// StructEmptyDefaultTag holds one field with no default tag
type StructEmptyDefaultTag struct {
	TestVal string `default:"" env:"TEST_VAL"`
}

// StructAllTagged holds multiple fields with tags.
type StructAllTagged struct {
	TestVal  string `default:"xyz" env:"TEST_VAL"`
	TestVal2 string `default:"abc" env:"TEST_VAL2"`
	TestInt  int    `default:"2"   env:"TEST_INT"`
}

// InnerStructError is the nested struct
type InnerStructError struct {
	TestNestInner string `default:"inner"`
}

// OuterStructInnerError is the nested struct
type OuterStructInnerError struct {
	TestNestOuter string `default:"outer" env:"TEST_NEST_OUTER"`
	InnerStructError
}

// InnerStruct is the nested struct
type InnerStruct struct {
	TestNestInner string `default:"inner" env:"TEST_NEST_INNER"`
	InnerMostStruct
}

// InnerMostStruct is the nested struct
type InnerMostStruct struct {
	TestNestInnerMost string `default:"innermost" env:"TEST_NEST_INNER_MOST"`
}

// OuterStruct is the struct containing the nested struct
type OuterStruct struct {
	TestNestOuter string `default:"outer" env:"TEST_NEST_OUTER"`
	InnerStruct
}

// Database contains the postgres config.
type Database struct {
	DBName   string `default:"nottest"  env:"UT_DB_NAME"`
	Host     string `default:"xyz"      env:"UT_DB_HOST"`
	Password string `default:"12345"    env:"UT_DB_PASSWORD"`
	Port     uint   `default:"5"        env:"UT_DB_PORT"`
	SSLMode  string `default:"disabled" env:"UT_DB_SSL_MODE"`
	User     string `default:"abc"      env:"UT_DB_USER"`
}

// WebServer contains webserver configuration.
type WebServer struct {
	APIReadTimeout  time.Duration `default:"60s" env:"UT_WS_API_READ_TIMEOUT"`
	APIWriteTimeout time.Duration `default:"60s" env:"UT_WS_API_WRITE_TIMEOUT"`
}

// AppConfig The configuration.
type AppConfig struct {
	Database
	WebServer
}

// MandatorySet has config with a value that is mandatory
type MandatorySet struct {
	TestVal string `env:"MANDATORY_TEST_VAL" mandatory:"true"`
}

func TestNoTagsErrors(t *testing.T) {
	var noTags StructNoTags
	err := LoadViperConfig(noTags)
	if err == nil {
		t.Error("Config struct with no default or env tag should error")
	}
}

func TestMissingEnvTagErrors(t *testing.T) {
	var noEnvTag StructMissingEnvTag
	err := LoadViperConfig(noEnvTag)
	if err == nil {
		t.Error("Config struct with no default or env tag should error")
	}
}

func TestNoDefaultsOk(t *testing.T) {
	expected := StructMissingDefaultTags{TestString: "", TestInt: 0, TestBool: false}
	var actual StructMissingDefaultTags
	err := LoadViperConfig(&actual)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Viper config %v is not equal to expected config %v", actual, expected)
	}
}

func TestEmptyDefaultTagIsOk(t *testing.T) {
	var emptyDefaultTag StructEmptyDefaultTag
	err := LoadViperConfig(&emptyDefaultTag)
	if err != nil {
		t.Errorf("Empty default tag should be ok. Error: %s", err)
	}
}

func TestDefaultValues(t *testing.T) {
	expected := StructAllTagged{TestVal: "xyz", TestVal2: "abc", TestInt: 2}
	var actual StructAllTagged
	err := LoadViperConfig(&actual)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Viper config %v is not equal to expected config %v", actual, expected)
	}
}

func TestEnvironmentValues(t *testing.T) {
	expected := StructAllTagged{TestVal: "env1", TestVal2: "env2", TestInt: 3}

	err := os.Setenv("TEST_VAL", "env1")
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	err = os.Setenv("TEST_VAL2", "env2")
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	err = os.Setenv("TEST_INT", "3")
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}

	var actual StructAllTagged
	err = LoadViperConfig(&actual)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Viper config %v is not equal to expected config %v", actual, expected)
	}
}

func TestNestedStruct(t *testing.T) {
	expected := OuterStruct{TestNestOuter: "outer", InnerStruct: InnerStruct{TestNestInner: "inner", InnerMostStruct: InnerMostStruct{TestNestInnerMost: "innermost"}}}
	var actual OuterStruct
	err := LoadViperConfig(&actual)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Viper config %v is not equal to expected config %v", actual, expected)
	}
}

func TestNestedStructWithError(t *testing.T) {
	var actual OuterStructInnerError
	err := LoadViperConfig(&actual)
	fmt.Println(err)
	if err == nil {
		t.Errorf("Nested struct with no env tag should error.")
	}
}

// The password is default
func TestReadYamlConfigWithSomeDefaults(t *testing.T) {
	os.Clearenv()
	expected := AppConfig{
		Database{DBName: "mydb", Host: "abc", Port: 5432, SSLMode: "disable", Password: "12345", User: "postgres"},
		WebServer{APIReadTimeout: time.Second * time.Duration(120), APIWriteTimeout: time.Second * time.Duration(120)},
	}
	reader := bytes.NewReader(yamlExample)
	var actual AppConfig
	err := LoadViperConfigFromReader(reader, &actual, "yaml")
	if err != nil {
		t.Errorf("Unexpected error occurred %s.", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Viper config %v is not equal to expected config %v", actual, expected)
	}
}

func TestReadYamlConfigWithEnvOverride(t *testing.T) {
	os.Clearenv()
	expected := AppConfig{
		Database{DBName: "mydb", Host: "abc", Port: 5432, SSLMode: "disable", Password: "environment", User: "postgres"},
		WebServer{APIReadTimeout: time.Second * time.Duration(120), APIWriteTimeout: time.Second * time.Duration(120)},
	}
	var actual AppConfig
	err := os.Setenv("UT_DB_PASSWORD", "environment")
	if err != nil {
		t.Errorf("Unexpected error occurred %s.", err)
	}
	reader := bytes.NewReader(yamlExample)
	err = LoadViperConfigFromReader(reader, &actual, "yaml")
	if err != nil {
		t.Errorf("Unexpected error occurred %s.", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Viper config %v is not equal to expected config %v", actual, expected)
	}
}

func TestLoadViperConfigFromFileNoFileExtension(t *testing.T) {
	os.Clearenv()
	var actual AppConfig
	err := LoadViperConfigFromFile("./nofile", &actual)
	if err == nil {
		t.Error("File with no extension should error")
	}
}

func TestMandatorySet(t *testing.T) {
	os.Clearenv()

	err := os.Setenv("MANDATORY_TEST_VAL", "blah")
	if err != nil {
		t.Errorf("Unexpected error occurred %s.", err)
	}

	var actual MandatorySet
	err = LoadViperConfig(&actual)
	assert.NoError(t, err)

	os.Clearenv()
	err = LoadViperConfig(&actual)
	assert.Error(t, err)
}
