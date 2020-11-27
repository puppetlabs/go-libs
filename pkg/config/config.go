package config

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"

	"github.com/spf13/viper"
)

func setupViperConfig(cfg interface{}) error {
	t := reflect.TypeOf(cfg)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() == reflect.Struct {
			err := setupViperConfig(reflect.ValueOf(cfg).Field(i).Interface())
			if err != nil {
				return err
			}
			continue
		}

		//The env tag is mandatory. It not present or we fail to set it up then error.
		if envTag, ok := f.Tag.Lookup("env"); ok {
			if err := viper.BindEnv(f.Name, envTag); err != nil {
				return fmt.Errorf("unable to bind %s to environment variable %s", f.Name, envTag)
			}
		} else {
			return fmt.Errorf("empty env tag for field %s", f.Name)
		}

		if defaultTag, ok := f.Tag.Lookup("default"); ok {
			viper.SetDefault(f.Name, defaultTag)
		}
	}
	return nil

}

//LoadViperConfig will return an interface with a populated struct of config with the type being the same as that passed
// in. See unit tests for usage.
func LoadViperConfig(cfg interface{}) (interface{}, error) {

	err := setupViperConfig(cfg)
	if err != nil {
		return nil, err
	}
	cfgInt := cfg
	err = viper.Unmarshal(&cfgInt, func(config *mapstructure.DecoderConfig) {
		config.Squash = true
	})
	if err != nil {
		return nil, err
	}
	return cfgInt, nil

}
