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

package test

import (
	"testing"

	"knative.dev/pkg/test/helpers"
	"knative.dev/reconciler-test/pkg/test/feature"
	"knative.dev/reconciler-test/pkg/test/requirement"
)

func NewContext(t *testing.T, c Config) T {
	c.ResetClients()
	return T{Config: c, T: t}
}

type T struct {
	Config
	*testing.T
}

func (t *T) Must() bool {
	return t.checkReq(requirement.Must)
}

func (t *T) MustNot() bool {
	return t.checkReq(requirement.MustNot)
}

func (t *T) Should() bool {
	return t.checkReq(requirement.Should)
}

func (t *T) ShouldNot() bool {
	return t.checkReq(requirement.ShouldNot)
}

func (t *T) May() bool {
	return t.checkReq(requirement.May)
}

func (t *T) checkReq(level requirement.Levels) bool {
	return t.Requirements()&level != 0
}

func (t *T) Alpha(name string, f func(*T)) bool {
	return t.invoke(feature.Alpha, name, f)
}

func (t *T) Beta(name string, f func(*T)) bool {
	return t.invoke(feature.Beta, name, f)
}

func (t *T) Stable(name string, f func(*T)) bool {
	return t.invoke(feature.Stable, name, f)
}

func (t *T) Run(name string, f func(*T)) bool {
	return t.invoke(feature.All, name, f)
}

func (t *T) ObjectNameForTest() string {
	return helpers.ObjectNameForTest(t)
}

func (t *T) invoke(state feature.States, name string, f func(*T)) bool {
	return t.T.Run(name, func(gotest *testing.T) {
		if t.Features()&state == 0 {
			gotest.Skipf("%s features not enabled for testing", state)
		}

		newT := *t
		newT.T = gotest
		// Each test should have it's own client
		newT.Config.ResetClients()

		f(&newT)
	})
}
