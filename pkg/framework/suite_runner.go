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

package framework

import (
	"flag"
	"os"
	"testing"

	"github.com/octago/sflags/gen/gflag"
)

var config *BaseConfig

type suite struct {
	m *testing.M
}

func newSuite(m *testing.M) Suite {
	return &suite{m: m}
}

func (s *suite) Configure(def Config) Suite {
	// TODO: read config file

	err := gflag.ParseToDef(def)
	if err != nil {
		panic(err)
	}

	flag.Parse()

	config = def.GetBaseConfig()
	return s
}

func (s *suite) Require(component Component) Suite {
	// TODO: delegate to the component. Must first define configuration
	return s
}

func (s *suite) Run() {
	if config == nil {
		// Use default configuration
		s.Configure(&BaseConfig{})
	}

	os.Exit(s.m.Run())
}
