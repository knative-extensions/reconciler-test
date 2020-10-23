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

package test_test

//
//import (
//	"flag"
//	"os"
//	"os/exec"
//	"testing"
//
//	"knative.dev/reconciler-test/pkg/test"
//	"knative.dev/reconciler-test/pkg/test/feature"
//	"knative.dev/reconciler-test/pkg/test/requirement"
//)
//
//type mockContext struct {
//	test.MagicRunner
//}
//
//func TestFlags(t *testing.T) {
//	ctx := mockContext{}
//
//	fs := flag.NewFlagSet("test", flag.ContinueOnError)
//	ctx.InitFlags(fs)
//
//	if err := fs.Parse(nil); err != nil {
//		t.Fatal("failed to parse", err)
//	}
//
//	if got, want := ctx.RequirementLevel, requirement.Levels(0); got != want {
//		t.Errorf("wrong requirement level - got: %s want: %s", got, want)
//	}
//
//	if got, want := ctx.FeatureState, feature.States(0); got != want {
//		t.Errorf("wrong requirement level - got: %s want: %s", got, want)
//	}
//
//	// defaulting
//	test.Init(&ctx, t)
//
//	if got, want := ctx.RequirementLevel, requirement.All; got != want {
//		t.Errorf("wrong requirement level - got: %s want: %s", got, want)
//	}
//
//	if got, want := ctx.FeatureState, feature.All; got != want {
//		t.Errorf("wrong requirement level - got: %s want: %s", got, want)
//	}
//}
//
//func TestRunInvocation(t *testing.T) {
//	ctx := &mockContext{}
//	test.Init(ctx, t)
//
//	invoked := false
//	ctx.Run("subtest", func(m *mockContext) {
//		invoked = true
//	})
//
//	if !invoked {
//		t.Error("Run() did not invoke the subtest")
//	}
//}
//
//func TestLevelInvocation(t *testing.T) {
//	cases := []struct {
//		name  string
//		level requirement.Levels
//		f     func(*mockContext, string, interface{}) bool
//	}{
//		{"Must", requirement.Must, (*mockContext).Must},
//		{"MustNot", requirement.MustNot, (*mockContext).MustNot},
//		{"Should", requirement.Should, (*mockContext).Should},
//		{"ShouldNot", requirement.ShouldNot, (*mockContext).ShouldNot},
//		{"May", requirement.May, (*mockContext).May},
//
//		{"All Must", requirement.All, (*mockContext).Must},
//		{"All MustNot", requirement.All, (*mockContext).MustNot},
//		{"All Should", requirement.All, (*mockContext).Should},
//		{"All ShouldNot", requirement.All, (*mockContext).ShouldNot},
//		{"All May", requirement.All, (*mockContext).May},
//	}
//
//	for _, c := range cases {
//		t.Run(c.level.String(), func(t *testing.T) {
//			ctx := &mockContext{}
//			test.Init(ctx, t)
//
//			invoked := false
//			subtest := func(m *mockContext) { invoked = true }
//
//			ctx.RequirementLevel = ^c.level
//			c.f(ctx, "off", subtest)
//			if invoked {
//				t.Errorf("unexpected invocation of %s method when invoked with requirements %s",
//					c.level, ctx.RequirementLevel)
//			}
//
//			invoked = false
//			ctx.RequirementLevel = c.level
//			c.f(ctx, "on", subtest)
//			if !invoked {
//				t.Errorf("level %s did not invoke %s", c.level, c.name)
//			}
//		})
//	}
//}
//
//func TestStateInvocation(t *testing.T) {
//	cases := []struct {
//		name  string
//		state feature.States
//		f     func(*mockContext, string, interface{}) bool
//	}{
//		{"Alpha", feature.Alpha, (*mockContext).Alpha},
//		{"Beta", feature.Beta, (*mockContext).Beta},
//		{"Stable", feature.Stable, (*mockContext).Stable},
//
//		{"All Alpha", feature.All, (*mockContext).Alpha},
//		{"All Beta", feature.All, (*mockContext).Beta},
//		{"All Stable", feature.All, (*mockContext).Stable},
//	}
//
//	for _, c := range cases {
//		t.Run(c.state.String(), func(t *testing.T) {
//			ctx := &mockContext{}
//			test.Init(ctx, t)
//
//			invoked := false
//			subtest := func(m *mockContext) { invoked = true }
//
//			ctx.FeatureState = ^c.state
//			c.f(ctx, "off", subtest)
//			if invoked {
//				t.Errorf("unexpected invocation of %s method when invoked with states %s",
//					c.state, ctx.FeatureState)
//			}
//
//			invoked = false
//			ctx.FeatureState = c.state
//			c.f(ctx, "on", subtest)
//			if !invoked {
//				t.Errorf("state %s did not invoke %s", c.state, c.name)
//			}
//		})
//	}
//}
//
//func TestBadCallback(t *testing.T) {
//	if os.Getenv("CRASH") == "1" {
//		t.Run("nil", func(t *testing.T) {
//			ctx := &mockContext{}
//			test.Init(ctx, t)
//			ctx.Run("subtest", nil)
//		})
//
//		t.Run("non-func", func(t *testing.T) {
//			ctx := &mockContext{}
//			test.Init(ctx, t)
//			ctx.Run("subtest", 1)
//		})
//
//		t.Run("bad-type", func(t *testing.T) {
//			ctx := &mockContext{}
//			test.Init(ctx, t)
//			ctx.Run("subtest", func(int) {})
//		})
//		return
//	}
//
//	for _, test := range []string{"nil", "non-func", "bad-type"} {
//		t.Run(test, func(t *testing.T) {
//			cmd := exec.Command(os.Args[0], "-test.run=TestBadCallback/"+test)
//			cmd.Env = append(os.Environ(), "CRASH=1")
//			err := cmd.Run()
//			if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
//				return
//			}
//			// Anything but an exit code 1 is abnormal
//			// ie. 2 = something panicked
//			t.Fatalf("process ran with err %v, want exit status 1", err)
//		})
//	}
//}
//
//type noEmbeddedT struct{}
//
//func TestBadInitParam(t *testing.T) {
//	defer func() {
//		if err := recover(); err == nil {
//			t.Error("test.Init should panic when passed a type that doesn't embed test.MagicRunner")
//		}
//	}()
//	test.Init(noEmbeddedT{}, t)
//}
//
//type customSetup struct {
//	test.MagicRunner
//	setupArg *testing.T
//}
//
//func (c *customSetup) Setup(t *testing.T) {
//	c.setupArg = t
//}
//
//func TestCustomSetup(t *testing.T) {
//	cs := &customSetup{}
//	test.Init(cs, t)
//
//	if cs.setupArg != t {
//		t.Error("expected Setup to be called")
//	}
//}
