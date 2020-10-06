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
	"flag"
	"os"
	"os/exec"
	"testing"

	"knative.dev/reconciler-test/pkg/test/feature"
	"knative.dev/reconciler-test/pkg/test/requirement"
)

type mockContext struct {
	BaseContext
}

func (c *mockContext) Copy() Context {
	return c
}

func TestFlags(t *testing.T) {
	ctx := mockContext{}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	ctx.AddFlags(fs)

	if err := fs.Parse(nil); err != nil {
		t.Fatal("failed to parse", err)
	}

	if got, want := ctx.RequirementLevels, requirement.All; got != want {
		t.Errorf("wrong requirement level - got: %s want: %s", got, want)
	}

	if got, want := ctx.FeatureStates, feature.All; got != want {
		t.Errorf("wrong requirement level - got: %s want: %s", got, want)
	}
}

func TestRunInvocation(t *testing.T) {
	ctx := &mockContext{}
	ctx.Setup(ctx, t)

	invoked := false
	ctx.Run("subtest", func(m *mockContext) {
		invoked = true
	})

	if !invoked {
		t.Error("Run() did not invoke the subtest")
	}
}

func TestLevelInvocation(t *testing.T) {
	cases := []struct {
		name  string
		level requirement.Levels
		f     func(*mockContext, string, interface{}) bool
	}{
		{"Must", requirement.Must, (*mockContext).Must},
		{"MustNot", requirement.MustNot, (*mockContext).MustNot},
		{"Should", requirement.Should, (*mockContext).Should},
		{"ShouldNot", requirement.ShouldNot, (*mockContext).ShouldNot},
		{"May", requirement.May, (*mockContext).May},

		{"Must", requirement.All, (*mockContext).Must},
		{"MustNot", requirement.All, (*mockContext).MustNot},
		{"Should", requirement.All, (*mockContext).Should},
		{"ShouldNot", requirement.All, (*mockContext).ShouldNot},
		{"May", requirement.All, (*mockContext).May},
	}

	for _, c := range cases {
		t.Run(c.level.String(), func(t *testing.T) {
			ctx := &mockContext{}
			ctx.Setup(ctx, t)
			ctx.RequirementLevels = c.level

			invoked := false
			c.f(ctx, "subtest", func(m *mockContext) {
				invoked = true
			})

			if !invoked {
				t.Errorf("level %s did not invoke %s", c.level, c.name)
			}
		})
	}
}

func TestStateInvocation(t *testing.T) {
	cases := []struct {
		name  string
		state feature.States
		f     func(*mockContext, string, interface{}) bool
	}{
		{"Alpha", feature.Alpha, (*mockContext).Alpha},
		{"Beta", feature.Beta, (*mockContext).Beta},
		{"Stable", feature.Stable, (*mockContext).Stable},

		{"Alpha", feature.All, (*mockContext).Alpha},
		{"Beta", feature.All, (*mockContext).Beta},
		{"Stable", feature.All, (*mockContext).Stable},
	}

	for _, c := range cases {
		t.Run(c.state.String(), func(t *testing.T) {
			ctx := &mockContext{}
			ctx.Setup(ctx, t)
			ctx.FeatureStates = c.state

			invoked := false
			c.f(ctx, "subtest", func(m *mockContext) {
				invoked = true
			})

			if !invoked {
				t.Errorf("state %s did not invoke %s", c.state, c.name)
			}
		})
	}
}

func TestBadCallback(t *testing.T) {
	if os.Getenv("CRASH") == "1" {
		t.Run("nil", func(t *testing.T) {
			ctx := &mockContext{}
			ctx.Setup(ctx, t)
			ctx.Run("subtest", nil)
		})

		t.Run("non-func", func(t *testing.T) {
			ctx := &mockContext{}
			ctx.Setup(ctx, t)
			ctx.Run("subtest", 1)
		})

		t.Run("bad-type", func(t *testing.T) {
			ctx := &mockContext{}
			ctx.Setup(ctx, t)
			ctx.Run("subtest", func(int) {})
		})
		return
	}

	for _, test := range []string{"nil", "non-func", "bad-type"} {
		t.Run(test, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=TestBadCallback/"+test)
			cmd.Env = append(os.Environ(), "CRASH=1")
			err := cmd.Run()
			if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
				return
			}
			// Anything but an exit code 1 is abnormal
			// ie. 2 = something paniced
			t.Fatalf("process ran with err %v, want exit status 1", err)
		})
	}
}
