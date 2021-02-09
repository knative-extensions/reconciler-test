package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"sigs.k8s.io/yaml"
)

var globalConfig map[string][]byte

// ReadConfigNamed is like ReadConfig, but you can specify a file name
func ReadConfigNamed(name string) {
	// I expect the config in the main
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	configPath := path.Join(dir, name)

	log.Println("Reading config from ", configPath)

	out, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(out, &m)
	if err != nil {
		panic(err)
	}

	globalConfig = make(map[string][]byte)
	for k, v := range m {
		b, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		globalConfig[k] = b
	}
}

// ReadConfig reads the config
func ReadConfig() {
	ReadConfigNamed("config.yaml")
}

// UnmarshalConfig reads the global config and unmarshal it
func UnmarshalConfig(testName string, out interface{}) error {
	if globalConfig == nil {
		return errors.New("there's no global config!!! Make sure you invoke ReadConfig() or ReadConfigNamed(name) in your TestMain")
	}
	in := globalConfig[testName]
	if in == nil {
		return fmt.Errorf("there's no config for test name %s", testName)
	}
	return json.Unmarshal(in, out)
}
