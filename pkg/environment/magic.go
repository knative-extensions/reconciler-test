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

package environment

import (
	"context"
	"fmt"
	"sync"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/state"
)

func NewGlobalEnvironment(ctx context.Context) GlobalEnvironment {

	fmt.Printf("level %s, state %s\n\n", l, s)

	return &MagicGlobalEnvironment{
		c:                ctx,
		RequirementLevel: *l,
		FeatureState:     *s,
	}
}

type MagicGlobalEnvironment struct {
	c context.Context

	RequirementLevel feature.Levels
	FeatureState     feature.States
}

type MagicEnvironment struct {
	c context.Context
	l feature.Levels
	s feature.States

	images           map[string]string
	namespace        string
	namespaceCreated bool
	refs             []corev1.ObjectReference
}

func (mr *MagicEnvironment) Reference(ref ...corev1.ObjectReference) {
	mr.refs = append(mr.refs, ref...)
}

func (mr *MagicEnvironment) References() []corev1.ObjectReference {
	return mr.refs
}

func (mr *MagicEnvironment) Finish() {
	if err := mr.DeleteNamespaceIfNeeded(); err != nil {
		panic(err)
	}
}

func (mr *MagicGlobalEnvironment) Environment(opts ...EnvOpts) (context.Context, Environment) {
	images, err := ProduceImages()
	if err != nil {
		panic(err)
	}

	namespace := feature.MakeK8sNamePrefix(feature.AppendRandomString("rekt"))

	env := &MagicEnvironment{
		c:         mr.c,
		l:         mr.RequirementLevel,
		s:         mr.FeatureState,
		images:    images,
		namespace: namespace,
	}

	ctx := ContextWith(mr.c, env)

	for _, opt := range opts {
		if ctx, err = opt(ctx, env); err != nil {
			panic(err)
		}
	}

	if err := env.CreateNamespaceIfNeeded(); err != nil {
		panic(err)
	}

	return ctx, env
}

func (mr *MagicEnvironment) Images() map[string]string {
	return mr.images
}

func (mr *MagicEnvironment) TemplateConfig(base map[string]interface{}) map[string]interface{} {
	cfg := make(map[string]interface{})
	for k, v := range base {
		cfg[k] = v
	}
	cfg["namespace"] = mr.namespace
	return cfg
}

func (mr *MagicEnvironment) RequirementLevel() feature.Levels {
	return mr.l
}

func (mr *MagicEnvironment) FeatureState() feature.States {
	return mr.s
}

func (mr *MagicEnvironment) Namespace() string {
	return mr.namespace
}

func (mr *MagicEnvironment) Prerequisite(ctx context.Context, t *testing.T, f *feature.Feature) {
	t.Helper() // Helper marks the calling function as a test helper function.
	t.Run("Prerequisite", func(t *testing.T) {
		mr.Test(ctx, t, f)
	})
}

// Test implements Environment.Test.
// In the MagicEnvironment implementation, the Store that is inside of the
// Feature will be assigned to the context. If no Store is set on Feature,
// Test will create a new store.KVStore and set it on the feature and then
// apply it to the Context.
func (mr *MagicEnvironment) Test(ctx context.Context, originalT *testing.T, f *feature.Feature) {
	originalT.Helper() // Helper marks the calling function as a test helper function.

	if f.State == nil {
		f.State = &state.KVStore{}
	}
	ctx = state.ContextWith(ctx, f.State)

	steps := categorizeSteps(f.Steps)

	skipAssertions := false
	skipRequirements := false
	skipReason := ""

	for _, s := range steps[feature.Setup] {
		s := s

		wg := &sync.WaitGroup{}
		wg.Add(1)

		var internalT *testing.T
		originalT.Run(s.T.String()+"/"+s.TestName(), func(st *testing.T) {
			internalT = st
			st.Helper()
			st.Cleanup(wg.Done) // Make sure wg.Done() is always invoked, no matter what

			if mr.s&s.S == 0 {
				st.Skipf("%s features not enabled for testing", s.S)
			}
			if mr.l&s.L == 0 {
				st.Skipf("%s requirement not enabled for testing", s.L)
			}

			// Perform step.
			s.Fn(ctx, st)
		})

		// Wait for the test to execute before spawning the next one
		wg.Wait()

		// Failed setup fails everything, so just run the teardown
		if internalT.Failed() {
			skipAssertions = true
			skipRequirements = true // No need to test other requirements
		}

	}

	for _, s := range steps[feature.Requirement] {
		s := s

		if skipRequirements {
			break
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)

		var internalT *requirementT

		originalT.Run(s.T.String()+"/"+s.TestName(), func(st *testing.T) {
			st.Helper()
			st.Cleanup(wg.Done) // Make sure wg.Done() is always invoked, no matter what
			internalT = createRequirementT(st)

			if mr.s&s.S == 0 {
				st.Skipf("%s features not enabled for testing", s.S)
			}
			if mr.l&s.L == 0 {
				st.Skipf("%s requirement not enabled for testing", s.L)
			}

			// Perform step.
			s.Fn(ctx, internalT)
		})

		// Wait for the test to execute before spawning the next one
		wg.Wait()

		if internalT.failed.Load() {
			skipAssertions = true
			skipRequirements = true // No need to test other requirements
			skipReason = fmt.Sprintf("requirement %q failed", s.Name)
		}

	}

	for _, s := range steps[feature.Assert] {
		s := s

		if skipAssertions {
			break
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)

		originalT.Run(s.T.String()+"/"+s.TestName(), func(st *testing.T) {
			st.Helper()
			st.Cleanup(wg.Done) // Make sure wg.Done() is always invoked, no matter what

			// TODO here we should run all the asserts and not filter them
			//  and then we record which one failed and which not, computing the actual compliance
			if mr.s&s.S == 0 {
				st.Skipf("%s features not enabled for testing", s.S)
			}
			if mr.l&s.L == 0 {
				st.Skipf("%s requirement not enabled for testing", s.L)
			}

			// Perform step.
			s.Fn(ctx, st)
		})

		// Wait for the test to execute before spawning the next one
		wg.Wait()

		// TODO record compliance level and fail parent test accordingly
	}

	for _, s := range steps[feature.Teardown] {
		s := s

		wg := &sync.WaitGroup{}
		wg.Add(1)

		originalT.Run(s.T.String()+"/"+s.TestName(), func(st *testing.T) {
			st.Helper()
			st.Cleanup(wg.Done) // Make sure wg.Done() is always invoked, no matter what

			if mr.s&s.S == 0 {
				st.Skipf("%s features not enabled for testing", s.S)
			}
			if mr.l&s.L == 0 {
				st.Skipf("%s requirement not enabled for testing", s.L)
			}

			// Perform step.
			s.Fn(ctx, st)
		})

		// Wait for the test to execute before spawning the next one
		wg.Wait()
	}

	if skipReason != "" {
		originalT.Skipf("Skipping feature '%s' assertions because %s", f.Name, skipReason)
	}
}

// TestSet implements Environment.TestSet
func (mr *MagicEnvironment) TestSet(ctx context.Context, t *testing.T, fs *feature.FeatureSet) {
	t.Helper() // Helper marks the calling function as a test helper function

	wg := &sync.WaitGroup{}
	for _, f := range fs.Features {
		wg.Add(1)
		t.Run(fs.Name, func(t *testing.T) {
			t.Cleanup(wg.Done)
			// FeatureSets should be run in parellel.
			mr.Test(ctx, t, &f)
		})
	}

	wg.Wait()
}

type envKey struct{}

func ContextWith(ctx context.Context, e Environment) context.Context {
	return context.WithValue(ctx, envKey{}, e)
}

func FromContext(ctx context.Context) Environment {
	if e, ok := ctx.Value(envKey{}).(Environment); ok {
		return e
	}
	panic("no Environment found in context")
}
