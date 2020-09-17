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

package testing

import (
	"flag"
	"fmt"
	"testing"

	"github.com/octago/sflags/gen/gflag"
	"knative.dev/reconciler-test/pkg/framework"
)

type dummySuite struct {
	m *testing.M
}

func (d dummySuite) ParseFlags(def interface{}) framework.Suite {
	gflag.ParseToDef(def)
	flag.Parse()
	return d
}

func (d dummySuite) Run() {
	d.m.Run()
}

type Config struct {
	Broker string
}

func NewSuite(m *testing.M) framework.Suite {
	return &dummySuite{
		m: m,
	}
}

var config = Config{}

func TestMain(m *testing.M) {

	NewSuite(m).
		ParseFlags(&config).Run()
}

func Test1(t *testing.T) {
	fmt.Println("broker is " + config.Broker)
}
