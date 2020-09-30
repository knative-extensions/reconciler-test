/*
 * Copyright 2020 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/util/yaml"
	"knative.dev/pkg/apis"
)

const (
	configFilename = "test-config.yaml"
)

// Config is the test suite configuration
type Config interface {

	// Validate checks the configuration is valid
	// Does not traverse the configuration tree (shallow validation)
	Validate() *apis.FieldError
}

// GetConfig navigates the configuration tree and
// returns the node selected by path.
func GetConfig(cfg Config, path string) Config {
	if path == "" {
		return cfg
	}
	return getConfig(cfg, "", strings.Split(path, "/"))
}

func getConfig(s Config, parent string, path []string) Config {
	if len(path) == 0 {
		return s
	}

	st := reflect.TypeOf(s)
	sv := reflect.ValueOf(s)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
		sv = sv.Elem()
	}

	if st.Kind() != reflect.Struct {
		return nil // not found
	}

	name := strings.ToLower(path[0])
	parent += "/" + name
	rest := path[1:]
	last := len(rest) == 0

	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)

		if strings.ToLower(f.Name) == name {
			if last {
				return subconfig(sv.Field(i), parent)
			}
			return getConfig(subconfig(sv.Field(i), parent), parent, rest)
		}

		if f.Anonymous {
			subcfg := getConfig(subconfig(sv.Field(i), parent), parent, path)
			if subcfg != nil {
				return subcfg
			}
		}
	}
	return nil
}

// ParseConfigFile locates the configuration file starting
// from the current working directory, up to the project root
// directory. If found, parses it,otherwise panic.
func ParseConfigFile(def Config) {
	cw, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	filename := findConfigFile(cw)
	if filename == "" {
		panic(configFilename + " not found in " + cw + " and in parent directories")
	}

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := yaml.NewYAMLToJSONDecoder(file)
	err = decoder.Decode(def)
	if err != nil {
		panic(err)
	}
}

func findConfigFile(dir string) string {
	filename := filepath.Join(dir, configFilename)
	info, err := os.Stat(filename)
	if os.IsNotExist(err) || info.IsDir() {
		info, err = os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil {
			return ""
		}

		// Try parent directory
		parent := filepath.Dir(dir)
		if parent == "." {
			// not found
			return ""
		}

		return findConfigFile(parent)
	}
	return filename
}

func subconfig(v reflect.Value, path string) Config {
	if !v.CanAddr() {
		panic(fmt.Sprintf("%+v at %s is unaddressable", v, path))
	}
	a := v.Addr()
	if !a.CanInterface() {
		panic(fmt.Sprintf("%+v at %s is not an interface", v, path))
	}
	i := a.Interface()
	cfg, ok := i.(Config)
	if !ok {
		panic(fmt.Sprintf("%+v at %s does not implement config.Config", v, path))
	}
	return cfg
}
