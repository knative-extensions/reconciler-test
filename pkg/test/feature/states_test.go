/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package feature

import (
	"flag"
	"io/ioutil"
	"testing"
)

var cases = []struct {
	state States
	flag  string
}{
	{Alpha, "-feature.alpha"},
	{Beta, "-feature.beta"},
	{Stable, "-feature.stable"},
}

func TestTurnOn(t *testing.T) {
	for _, tc := range cases {
		var l States

		fs := &flag.FlagSet{}
		fs.SetOutput(ioutil.Discard)
		l.AddFlags(fs)

		if err := fs.Parse([]string{tc.flag}); err != nil {
			t.Fatal(err)
		}

		if l&tc.state == 0 {
			t.Errorf("flag %q did not enable %s", tc.flag, tc.state)
		}
	}
}

func TestTurnOff(t *testing.T) {
	for _, tc := range cases {
		l := ^States(0)

		fs := &flag.FlagSet{}
		fs.SetOutput(ioutil.Discard)
		l.AddFlags(fs)

		if err := fs.Parse([]string{tc.flag + "=false"}); err != nil {
			t.Fatal(err)
		}

		if l&tc.state != 0 {
			t.Errorf("flag %q did not disable %s", tc.flag, tc.state)
		}
	}
}
