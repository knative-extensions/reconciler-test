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
	"context"
	"flag"
	"fmt"
	"sync"
	"testing"

	"knative.dev/reconciler-test/pkg/test/feature"
	"knative.dev/reconciler-test/pkg/test/requirement"
	rigging "knative.dev/reconciler-test/rigging/v2"
)

func NewGlobalEnvironment() rigging.GlobalEnvironment {
	return &MagicEnvironment{}
}

type MagicEnvironment struct {
	RequirementLevel requirement.Levels
	FeatureState     feature.States

	FlagSets []rigging.FlagSetFn
}

func (mr *MagicEnvironment) Environment() rigging.Environment {
	return &MagicRunner{
		l: mr.RequirementLevel,
		s: mr.FeatureState,
	}
}

func (mr *MagicEnvironment) WithFlags(fn rigging.FlagSetFn) {
	if mr.FlagSets == nil {
		mr.FlagSets = make([]rigging.FlagSetFn, 0)
	}
	mr.FlagSets = append(mr.FlagSets, fn)
}

func (mr *MagicEnvironment) InitFlags(fs *flag.FlagSet) {
	mr.WithFlags(mr.RequirementLevel.InitFlags)
	mr.WithFlags(mr.FeatureState.InitFlags)

	for _, fn := range mr.FlagSets {
		fn(fs)
	}
}

func NewRunner(env rigging.Environment) rigging.FeatureTester {
	return &MagicRunner{
		l: env.RequirementLevel(),
		s: env.FeatureState(),
	}
}

type MagicRunner struct {
	l requirement.Levels
	s feature.States
}

func (mr *MagicRunner) RequirementLevel() requirement.Levels {
	return mr.l
}

func (mr *MagicRunner) FeatureState() feature.States {
	return mr.s
}

func (mr *MagicRunner) Namespace() string {
	panic("implement me")
}

func (mr *MagicRunner) Context() context.Context {
	panic("implement me")
}

func (mr *MagicRunner) Test(ctx context.Context, t *testing.T, f *rigging.Feature) {
	t.Helper() // Helper marks the calling function as a test helper function.

	if mr.l == 0 {
		mr.l = requirement.All
	}
	if mr.s == 0 {
		mr.s = feature.All
	}

	// do it the slow way first.
	pwg := &sync.WaitGroup{}
	for _, p := range f.Preconditions {
		pwg.Add(1)
		p := p
		t.Run(fmt.Sprintf("%s [pre] %s", f.Name, p.Name), func(t *testing.T) {
			t.Helper() // Helper marks the calling function as a test helper function.

			p.P(ctx, t)

			pwg.Done()
		})
	}
	pwg.Wait()

	awg := &sync.WaitGroup{}
	for _, a := range f.Assertions {
		a := a
		if mr.s&a.S == 0 {
			t.Skipf("%s features not enabled for testing", a.S)
		}
		if mr.l&a.L == 0 {
			t.Skipf("%s requirement not enabled for testing", a.L)
		}
		awg.Add(1)
		t.Run(fmt.Sprintf("%s [assert] %s", f.Name, a.Name), func(t *testing.T) {
			t.Helper() // Helper marks the calling function as a test helper function.

			a.A(ctx, t)

			awg.Done()
		})
	}
	awg.Done()
}
