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
	"flag"
	"fmt"
	"sync"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/reconciler-test/pkg/feature"
)

func NewGlobalEnvironment(ctx context.Context) GlobalEnvironment {
	return &MagicGlobalEnvironment{
		c: ctx,
	}
}

type MagicGlobalEnvironment struct {
	c context.Context

	RequirementLevel feature.Levels
	FeatureState     feature.States

	FlagSets []FlagSetFn
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
	if mr.refs == nil {
		mr.refs = make([]corev1.ObjectReference, 0)
	}
	mr.refs = append(mr.refs, ref...)
}

func (mr *MagicEnvironment) References() []corev1.ObjectReference {
	if mr.refs == nil {
		return []corev1.ObjectReference{} // return an empty list.
	}
	return mr.refs
}

func (mr *MagicEnvironment) Finish() {
	mr.DeleteNamespaceIfNeeded()
}

func (mr *MagicGlobalEnvironment) WithFlags(fn FlagSetFn) {
	if mr.FlagSets == nil {
		mr.FlagSets = make([]FlagSetFn, 0)
	}
	mr.FlagSets = append(mr.FlagSets, fn)
}

func (mr *MagicGlobalEnvironment) InitFlags(fs *flag.FlagSet) {
	mr.WithFlags(mr.RequirementLevel.InitFlags)
	mr.WithFlags(mr.FeatureState.InitFlags)

	for _, fn := range mr.FlagSets {
		fn(fs)
	}
}

func (mr *MagicGlobalEnvironment) Environment() (context.Context, Environment) {
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

	if err := env.CreateNamespaceIfNeeded(); err != nil {
		panic(err)
	}

	ctx := ContextWith(mr.c, env)

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
	cfg["images"] = mr.images
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

func (mr *MagicEnvironment) Test(ctx context.Context, t *testing.T, f *feature.Feature) {
	t.Helper() // Helper marks the calling function as a test helper function.

	if mr.l == 0 {
		mr.l = feature.All
	}
	if mr.s == 0 {
		mr.s = feature.Any
	}

	// do it the slow way first.
	pwg := &sync.WaitGroup{}
	pwg.Add(1)

	t.Run("preconditions", func(t *testing.T) {
		t.Helper() // Helper marks the calling function as a test helper function.
		t.Log(len(f.Preconditions), " preconditions.")
		defer pwg.Done() // Outer wait.

		for _, p := range f.Preconditions {
			pwg.Add(1)
			p := p
			t.Run(p.Name, func(t *testing.T) {
				t.Helper() // Helper marks the calling function as a test helper function.

				p.P(ctx, t)

				pwg.Done()
			})
		}
	})

	pwg.Wait()

	awg := &sync.WaitGroup{}
	awg.Add(1)

	t.Run("assertions", func(t *testing.T) {
		t.Helper() // Helper marks the calling function as a test helper function.
		t.Log(len(f.Assertions), " assertions.")
		defer awg.Done() // Outer wait.

		for _, a := range f.Assertions {
			a := a
			if mr.s&a.S == 0 {
				t.Skipf("%s features not enabled for testing", a.S)
			}
			if mr.l&a.L == 0 {
				t.Skipf("%s requirement not enabled for testing", a.L)
			}
			awg.Add(1)
			t.Run(fmt.Sprintf("[%s/%s]%s", a.S, a.L, a.Name), func(t *testing.T) {
				t.Helper() // Helper marks the calling function as a test helper function.

				a.A(ctx, t)

				awg.Done()
			})
		}
	})

	awg.Wait()
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
