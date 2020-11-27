package config

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

//StructNoTags holds one field with no tags against the struct
type StructNoTags struct {
	DummyString string
}

//StructMissingEnvTag holds one field with no env tag
type StructMissingEnvTag struct {
	TestVal string `default:"xyz"`
}

//StructMissingDefaultTags holds one field with no default tag
type StructMissingDefaultTags struct {
	TestString string `env:"TEST_STRING"`
	TestInt    int    `env:"TEST_INT"`
	TestBool   bool   `env:"TEST_BOOL"`
}

//StructEmptyDefaultTag holds one field with no default tag
type StructEmptyDefaultTag struct {
	TestVal string `env:"TEST_VAL" default:""`
}

//StructAllTagged holds multiple fields with tags.
type StructAllTagged struct {
	TestVal  string `env:"TEST_VAL" default:"xyz"`
	TestVal2 string `env:"TEST_VAL2" default:"abc"`
	TestInt  int    `env:"TEST_INT" default:"2"`
}

//InnerStructError is the nested struct
type InnerStructError struct {
	TestNestInner string `default:"inner"`
}

//OuterStructInnerError is the nested struct
type OuterStructInnerError struct {
	TestNestOuter string `env:"TEST_NEST_OUTER" default:"outer"`
	InnerStructError
}

//InnerStruct is the nested struct
type InnerStruct struct {
	TestNestInner string `env:"TEST_NEST_INNER" default:"inner"`
}

//OuterStruct is the struct containing the nested struct
type OuterStruct struct {
	TestNestOuter string `env:"TEST_NEST_OUTER" default:"outer"`
	InnerStruct
}

func getNestedErrorConfig() (*OuterStructInnerError, error) {
	cfgInt, err := LoadViperConfig(OuterStructInnerError{})
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgInt.(OuterStructInnerError)
	if !ok {
		return nil, fmt.Errorf("viper returned wrong type for config")
	}
	return &cfg, nil
}

func getNestedConfig() (*OuterStruct, error) {
	cfgInt, err := LoadViperConfig(OuterStruct{})
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgInt.(OuterStruct)
	if !ok {
		return nil, fmt.Errorf("viper returned wrong type for config")
	}
	return &cfg, nil
}

func getAllTaggedConfig() (*StructAllTagged, error) {
	cfgInt, err := LoadViperConfig(StructAllTagged{})
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgInt.(StructAllTagged)
	if !ok {
		return nil, fmt.Errorf("viper returned wrong type for config")
	}
	return &cfg, nil
}

func getMissingDefaultsConfig() (*StructMissingDefaultTags, error) {
	cfgInt, err := LoadViperConfig(StructMissingDefaultTags{})
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgInt.(StructMissingDefaultTags)
	if !ok {
		return nil, fmt.Errorf("viper returned wrong type for config")
	}
	return &cfg, nil
}

func TestNoTagsErrors(t *testing.T) {
	var noTags StructNoTags
	_, err := LoadViperConfig(noTags)
	if err == nil {
		t.Error("Config struct with no default or env tag should error")
	}
}

func TestMissingEnvTagErrors(t *testing.T) {
	var noEnvTag StructMissingEnvTag
	_, err := LoadViperConfig(noEnvTag)
	if err == nil {
		t.Error("Config struct with no default or env tag should error")
	}
}

func TestNoDefaultsOk(t *testing.T) {
	expected := StructMissingDefaultTags{TestString: "", TestInt: 0, TestBool: false}
	viperCfg, err := getMissingDefaultsConfig()
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if !reflect.DeepEqual(expected, *viperCfg) {
		t.Errorf("Viper config %v is not equal to expected config %v", viperCfg, expected)
	}
}

func TestEmptyDefaultTagIsOk(t *testing.T) {
	var emptyDefaultTag StructEmptyDefaultTag
	_, err := LoadViperConfig(emptyDefaultTag)
	if err != nil {
		t.Errorf("Empty default tag should be ok. Error: %s", err)
	}
}

func TestDefaultValues(t *testing.T) {
	allTaggeddefault := StructAllTagged{TestVal: "xyz", TestVal2: "abc", TestInt: 2}
	viperCfg, err := getAllTaggedConfig()
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if !reflect.DeepEqual(allTaggeddefault, *viperCfg) {
		t.Errorf("Viper config %v is not equal to expected config %v", viperCfg, allTaggeddefault)
	}
}

func TestEnvironmentValues(t *testing.T) {
	allTaggedEnvironment := StructAllTagged{TestVal: "env1", TestVal2: "env2", TestInt: 3}

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

	viperCfg, err := getAllTaggedConfig()
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}

	if !reflect.DeepEqual(allTaggedEnvironment, *viperCfg) {
		t.Errorf("Viper config %v is not equal to expected config %v", viperCfg, allTaggedEnvironment)
	}
}

func TestNestedStruct(t *testing.T) {
	nestedStruct := OuterStruct{TestNestOuter: "outer", InnerStruct: InnerStruct{TestNestInner: "inner"}}
	viperCfg, err := getNestedConfig()
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if !reflect.DeepEqual(nestedStruct, *viperCfg) {
		t.Errorf("Viper config %v is not equal to expected config %v", viperCfg, nestedStruct)
	}
}

func TestNestedStructWithError(t *testing.T) {
	_, err := getNestedErrorConfig()
	fmt.Println(err)
	if err == nil {
		t.Errorf("Nested struct with no env tag should error.")
	}
}
