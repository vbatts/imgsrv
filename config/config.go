package config

import (
	"io/ioutil"
	"launchpad.net/goyaml"
)

type Config struct {
	Server        bool   // Run as server, if different than false (server)
	Ip            string // Bind address, if different than 0.0.0.0 (server)
	Port          string // listen port, if different than '7777' (server)
	MongoHost     string // mongoDB host, if different than 'localhost' (server)
	MongoDB       string // mongoDB db name, if different than 'filesrv' (server)
	MongoUsername string // mongoDB username, if any (server)
	MongoPassword string // mongoDB password, if any (server)

	RemoteHost string // imgsrv server to push files to (client)

	Map map[string]interface{} // key/value options (not used currently)
}

// Of the configurations, provided option, return the value as a bool
func (c Config) GetBool(option string) (value bool) {
	switch c.Map[option] {
	default:
		value = false
	case "yes", "on", "true":
		value = true
	}
	return
}

// Of the configurations, provided option, return the value as a string
func (c Config) GetString(option string) (value string) {
	value, _ = c.Map[option].(string)
	return
}

func (c *Config) Merge(other *Config) error {
  if other == nil {
    return nil
  }
  return nil
}

// Given a filename to a YAML file, unmarshal it, and return a Config
func ReadConfigFile(filename string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := Config{}
	if err = goyaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

/*
func WriteConfigFile(filename string, data []byte) (err error) {
	return
}
*/
