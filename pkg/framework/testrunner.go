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
	"context"
	"fmt"
	"strings"
	"testing"

	// TODO: remove dependencies because of flags definition
	pkgtest "knative.dev/pkg/test"
	"knative.dev/pkg/test/helpers"

	"github.com/onsi/gomega"
)

type test struct {
	t           *testing.T
	feature     string
	maturity    string
	requirement string
}

func newTest(t *testing.T) Test {
	return &test{t: t}
}

func (t *test) Feature(name string) Test {
	if t.feature != "" {
		panic("Test.Feature called multiple times")
	}
	t.feature = name
	return t
}

func (t *test) Alpha() Test {
	if t.maturity != "" {
		panic("Test.Alpha/Beta/Stable called multiple times")
	}
	t.maturity = "alpha"
	return t
}

func (t *test) Beta() Test {
	if t.maturity != "" {
		panic("Test.Alpha/Beta/Stable called multiple times")
	}
	t.maturity = "beta"
	return t
}

func (t *test) Stable() Test {
	if t.maturity != "" {
		panic("Test.Alpha/Beta/Stable called multiple times")
	}
	t.maturity = "stable"
	return t
}

func (t *test) Must(name string) Test {
	if t.requirement != "" {
		panic("Test.Must/Should/May called multiple times")
	}
	t.requirement = "must"
	return t
}

func (t *test) Should(name string) Test {
	if t.requirement != "" {
		panic("Test.Must/Should/May called multiple times")
	}
	t.requirement = "should"
	return t
}

func (t *test) May(name string) Test {
	if t.requirement != "" {
		panic("Test.Must/Should/May called multiple times")
	}
	t.requirement = "may"
	return t
}

func (t *test) Run(fn func(TestContext)) {
	if t.requirement == "must" && !baseConfig.Requirements.Must {
		t.t.Skip("skipping test marked as Must")
	}
	if t.requirement == "should" && !baseConfig.Requirements.Should {
		t.t.Skip("skipping test marked as Should")
	}
	if t.requirement == "may" && !baseConfig.Requirements.May {
		t.t.Skip("skipping test marked as May")
	}

	if testing.Verbose() && (t.feature != "" || t.requirement != "" || t.maturity != "") {
		decorate := ""
		sep := ""
		if t.feature != "" {
			decorate += "✨" + t.feature
			sep = " "
		}
		if t.maturity != "" {
			decorate += sep + "[" + t.maturity + "]"
			sep = ""
		}
		if t.requirement != "" {
			decorate += sep + "[" + t.requirement + "]"

		}
		t.t.Log(decorate)
	}

	// TODO: validate feature to match DNS-1123 label
	namespace := helpers.AppendRandomString(strings.ToLower(t.t.Name()))
	ctx := withInjection(context.Background())

	tc := &testContextImpl{
		resourceContextImpl: resourceContextImpl{
			context:   ctx,
			namespace: namespace,
		},
		t:     t.t,
		WithT: gomega.NewGomegaWithT(t.t),
	}

	nsspec := fmt.Sprintf(namespaceTemplate, namespace)
	tc.CreateFromYAMLOrFail(nsspec)

	cleanup := func() {
		if err := tc.DeleteFromYAML(nsspec); err != nil {
			t.t.Logf("warning: failed to delete namespace %s (%v)", namespace, err)
		}
	}

	// Clean up resources if the test is interrupted in the middle.
	pkgtest.CleanupOnInterrupt(cleanup, t.t.Logf)

	t.t.Logf("namespace is %s", namespace)

	// Finally run user-code
	fn(tc)

	cleanup()
}

const namespaceTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`
