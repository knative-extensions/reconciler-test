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
	"context"
	"fmt"
	"testing"
)

type Feature struct {
	Name          string
	Preconditions []Precondition
	Assertions    []Assertion
}

type AssertFn func(ctx context.Context, t *testing.T)

type PreConFn func(ctx context.Context, t *testing.T)

type Precondition struct {
	Name string
	P    PreConFn
}

func (f *Feature) Precondition(name string, fn PreConFn) {
	if f.Preconditions == nil {
		f.Preconditions = make([]Precondition, 0)
	}
	f.Preconditions = append(f.Preconditions, Precondition{
		Name: name,
		P:    fn,
	})
}

type Assertable interface {
	Must(name string, fn AssertFn) Assertable
	Should(name string, fn AssertFn) Assertable
	May(name string, fn AssertFn) Assertable
	MustNot(name string, fn AssertFn) Assertable
	ShouldNot(name string, fn AssertFn) Assertable
}

type Asserter struct {
	f    *Feature
	name string
	s    States
}

func (a *Asserter) Must(name string, fn AssertFn) Assertable {
	a.Assert(Must, name, fn)
	return a
}

func (a *Asserter) Should(name string, fn AssertFn) Assertable {
	a.Assert(Should, name, fn)
	return a
}

func (a *Asserter) May(name string, fn AssertFn) Assertable {
	a.Assert(May, name, fn)
	return a
}

func (a *Asserter) MustNot(name string, fn AssertFn) Assertable {
	a.Assert(MustNot, name, fn)
	return a
}

func (a *Asserter) ShouldNot(name string, fn AssertFn) Assertable {
	a.Assert(ShouldNot, name, fn)
	return a
}

func (f *Feature) Alpha(name string) Assertable {
	return f.asserter(Alpha, name)
}

func (f *Feature) Beta(name string) Assertable {
	return f.asserter(Beta, name)
}

func (f *Feature) Stable(name string) Assertable {
	return f.asserter(Stable, name)
}

func (f *Feature) asserter(s States, name string) Assertable {
	return &Asserter{
		f:    f,
		name: name,
		s:    s,
	}
}

func (a *Asserter) Assert(l Levels, name string, fn AssertFn) {
	a.f.Assertions = append(a.f.Assertions, Assertion{
		Name: fmt.Sprintf("%s %s", a.name, name),
		A:    fn,
		S:    a.s,
		L:    l,
	})
}

type Assertion struct {
	Name string
	S    States
	L    Levels
	A    AssertFn
}
