package config

import (
	"io/ioutil"
	"launchpad.net/goyaml"
)

type Config map[string]interface{}

// Of the configurations, provided option, return the value as a bool
func (c *Config) GetBool(option string) (value bool) {
	conf := Config{}
	conf = *c
	switch conf[option] {
	default:
		value = false
	case "yes", "on", "true":
		value = true
	}
	return
}

// Of the configurations, provided option, return the value as a string
func (c *Config) GetString(option string) (value string) {
	conf := Config{}
	conf = *c
	value, _ = conf[option].(string)
	return
}

// Given a filename to a YAML file, unmarshal it, and return a Config
func ReadConfigFile(filename string) (config Config, err error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = goyaml.Unmarshal(bytes, &config)
	if err != nil {
		return
	}

	return config, nil
}

/*
func WriteConfigFile(filename string, data []byte) (err error) {
	return
}
*/
