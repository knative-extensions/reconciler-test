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
	"strings"
	"testing"
	"time"

	"go.uber.org/atomic"
	corev1 "k8s.io/api/core/v1"

	"knative.dev/reconciler-test/pkg/feature"
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
	mr.DeleteNamespaceIfNeeded()
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

func (mr *MagicEnvironment) Test(ctx context.Context, t *testing.T, f *feature.Feature) {
	t.Helper() // Helper marks the calling function as a test helper function.

	steps := feature.ReorderSteps(f.Steps)

	featureTestName := strings.ReplaceAll(f.Name, " ", "_")
	for _, s := range steps {
		stepName := strings.ReplaceAll(s.TestName(), " ", "_")
		fmt.Printf("=== %s   %s/[%s]%s\n", run, t.Name(), featureTestName, stepName)

		skipped, failed, duration := mr.safeExecuteStep(ctx, t, s)

		if skipped {
			fmt.Printf("    --- %s: %s/[%s]%s (%.2fs) \n\n", skip, t.Name(), featureTestName, stepName, duration.Seconds())
		} else if failed {
			fmt.Printf("    --- %s: %s/[%s]%s (%.2fs) \n\n", fail, t.Name(), featureTestName, stepName, duration.Seconds())
			t.FailNow() // Here we can have different policies, depending on feature level etc
		} else {
			fmt.Printf("    --- %s: %s/[%s]%s (%.2fs) \n\n", pass, t.Name(), featureTestName, stepName, duration.Seconds())
		}
	}
}

const (
	run  = "RUN"
	pass = "PASS"
	fail = "FAIL"
	skip = "SKIP"
)

func (mr *MagicEnvironment) safeExecuteStep(ctx context.Context, testingT *testing.T, step feature.Step) (skipped bool, failed bool, duration time.Duration) {
	testingT.Helper()

	if mr.s&step.S == 0 || mr.l&step.L == 0 {
		if mr.s&step.S == 0 {
			testingT.Logf("%s features not enabled for testing", step.S)
		}
		if mr.l&step.L == 0 {
			testingT.Logf("%s requirement not enabled for testing", step.L)
		}
		skipped = true
		failed = false
		return
	}

	t := t{
		t:       testingT,
		failed:  atomic.NewBool(false),
		skipped: atomic.NewBool(false),
	}

	var cancelFn context.CancelFunc
	deadLine, ok := testingT.Deadline()
	if ok {
		ctx, cancelFn = context.WithDeadline(ctx, deadLine)
	} else {
		ctx, cancelFn = context.WithCancel(ctx)
	}
	var panicValue interface{}

	start := time.Now()
	go func() {
		testingT.Helper()
		defer func() {
			// A panic might happen while executing the test.
			// In this case, mark the test as failed
			if panicValue = recover(); panicValue != nil {
				t.failed.Store(true)
			}
			cancelFn()
		}()
		step.Fn(ctx, &t)
	}()

	<-ctx.Done()
	duration = time.Now().Sub(start)
	if ctx.Err() == context.DeadlineExceeded {
		testingT.Logf("Timeout exceeded")
		t.failed.Store(true)
	}

	skipped = t.skipped.Load()
	failed = t.failed.Load()

	if panicValue != nil {
		testingT.Logf("Panic while executing the test: %v", panicValue)
	}

	return
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
