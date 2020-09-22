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
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	configFilename = "test-config.yaml"
)

// Config is the test suite configuration
type Config interface {
}

// GetConfig navigates the configuration tree and
// returns the node selected by path
func GetConfig(cfg interface{}, path string) interface{} {
	if path == "" {
		return cfg
	}
	return getConfig(cfg, strings.Split(path, "/"))
}

func getConfig(s interface{}, path []string) interface{} {
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

	name := path[0]
	rest := path[1:]
	last := len(rest) == 0

	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)

		if strings.ToLower(f.Name) == strings.ToLower(name) {
			if last {
				return sv.Field(i).Interface()
			}
			return getConfig(sv.Field(i).Interface(), rest)
		}

		if f.Anonymous {
			subcfg := getConfig(sv.Field(i).Interface(), path)
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
