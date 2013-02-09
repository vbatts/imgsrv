package main

import (
  "launchpad.net/goyaml"
  "io/ioutil"
)

type Config map[string]interface{}

func (c *Config) GetBool(option string) (value bool) {
  conf := Config{}
  conf = *c
  switch conf[option] {
  default: value = false
  case "yes", "on", "true": value = true
  }
  return 
}

func (c *Config) GetString(option string) (value string) {
  conf := Config{}
  conf = *c
  value, _ = conf[option].(string)
  return
}

func ReadConfigFile(filename string) (config Config, err error) {
  bytes, err := ioutil.ReadFile(filename)
  if (err != nil) {
    return
  }

  err = goyaml.Unmarshal(bytes, &config)
  if (err != nil) {
    return
  }

  return config, nil
}

func WriteConfigFile(filename string, data []byte) (err error) {
  return
}

