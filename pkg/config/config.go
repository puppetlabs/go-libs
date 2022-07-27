// Package config provides facilities for application config.
package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	errEmptyEnvTagForField           = errors.New("empty env tag for field")
	errInvalidConfigType             = errors.New("config type must be either a pointer to a struct or a struct")
	errMissingFileExtension          = errors.New("missing file extension")
	errViperConfigNonPointerArgument = errors.New("reading Viper config requires a pointer argument")
)

// nolint:cyclop // naturally complex, we may be able to simplify this function
func setUpViperConfig(cfg interface{}, v *viper.Viper) error {
	// Get the type from the pointer or the struct itself - N.B. It will be a struct when called recursively.
	var t reflect.Type
	cfgType := reflect.TypeOf(cfg).Kind()
	switch cfgType {
	case reflect.Ptr:
		t = reflect.TypeOf(cfg).Elem()
	case reflect.Struct:
		t = reflect.TypeOf(cfg)
	default:
		return errInvalidConfigType
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() == reflect.Struct {
			var val interface{}
			if cfgType == reflect.Ptr {
				val = reflect.ValueOf(cfg).Elem().Field(i).Interface()
			} else {
				val = reflect.ValueOf(cfg).Field(i).Interface()
			}
			err := setUpViperConfig(val, v)
			if err != nil {
				return err
			}

			continue
		}

		// The env tag is mandatory. It not present or we fail to set it up then error.
		if envTag, ok := f.Tag.Lookup("env"); ok {
			if err := v.BindEnv(f.Name, envTag); err != nil {
				return fmt.Errorf("unable to bind %s to environment variable %s: %w", f.Name, envTag, err)
			}
		} else {
			return fmt.Errorf("%w: %s", errEmptyEnvTagForField, f.Name)
		}

		var fileDefaultSet bool
		if fileTag, ok := f.Tag.Lookup("file"); ok {
			// It is not mandatory to have a default set, so we will log this and move on.

			//#nosec gosec picks this up. It is build time injection, so it is assumed this will be tested first.
			fileBytes, err := ioutil.ReadFile(fileTag)
			if err != nil {
				logrus.Warnf("Unable to read file %s due to error %s.", fileTag, err)
			}

			v.SetDefault(f.Name, fileBytes)
			fileDefaultSet = true
		}

		if !fileDefaultSet {
			if defaultTag, ok := f.Tag.Lookup("default"); ok {
				v.SetDefault(f.Name, defaultTag)
			}
		}
	}

	return nil
}

// LoadViperConfig populates the cfg structure passed in (it must be the address passed in).
func LoadViperConfig(cfg interface{}) error {
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		return errViperConfigNonPointerArgument
	}
	v := viper.New()
	err := setUpViperConfig(cfg, v)
	if err != nil {
		return err
	}

	err = v.Unmarshal(cfg, func(config *mapstructure.DecoderConfig) {
		config.Squash = true
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// flattenCfgMap will take a map of any depth and flatten it down so there is only one level. N.B. The key will
// always remain the same.
func flattenCfgMap(cfgMap map[string]interface{}) (map[string]interface{}, error) {
	flatMap := make(map[string]interface{})
	for k, v := range cfgMap {
		if innerMap, ok := v.(map[string]interface{}); ok {
			flatInnerMap, err := flattenCfgMap(innerMap)
			if err != nil {
				return nil, err
			}
			err = mergo.Merge(&flatMap, flatInnerMap)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}
		} else {
			flatMap[k] = v
		}
	}

	return flatMap, nil
}

// LoadViperConfigFromFile populates the cfg structure passed in (it must be the address passed in).
// The config will be read from a file but environment variables override and default values can still be set.
// The file can be anything that Viper supports. N.B. This only supports anonymous nested structs.
func LoadViperConfigFromFile(filename string, cfg interface{}) error {
	if len(filepath.Ext(filename)) < 1 {
		return errMissingFileExtension
	}

	reader, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = LoadViperConfigFromReader(reader, cfg, filepath.Ext(filename)[1:])

	// N.B. Deferring a file close is bad practise so making it key part of function.
	fileCloseErr := reader.Close()
	if fileCloseErr != nil {
		return fmt.Errorf("%w", fileCloseErr)
	}

	return err
}

// LoadViperConfigFromReader will populate the cfg structure passed in (it must be the address passed in).
// The config will be read from a reader which must be setup prior to the call.
// N.B. The cfgType can be anything supported by viper i.e. yaml, json, env, ini, toml.
func LoadViperConfigFromReader(in io.Reader, cfg interface{}, cfgType string) error {
	v := viper.New()
	v.SetConfigType(cfgType)
	err := v.ReadConfig(in)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Reading in config from a source with multiple levels (usually a config file) will produce a multidimensional map
	// whereas reading from a struct produces a flattened map. Flattening the map from the reader and merging with the
	// already flattened struct map aligns this.
	flatCfgMap, err := flattenCfgMap(v.AllSettings())
	if err != nil {
		return err
	}
	err = v.MergeConfigMap(flatCfgMap)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = setUpViperConfig(cfg, v)
	if err != nil {
		return err
	}

	err = v.Unmarshal(cfg, func(config *mapstructure.DecoderConfig) {
		config.Squash = true
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
