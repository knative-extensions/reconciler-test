package rigging

import (
	"context"
	"flag"
	"fmt"
	"knative.dev/reconciler-test/pkg/test/feature"
	"knative.dev/reconciler-test/pkg/test/requirement"

	"testing"
)

type T interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Logf(format string, args ...interface{})
	Skipf(format string, args ...interface{})
}

type GlobalEnvironment interface {
	Environment() Environment

	WithFlags(fn FlagSetFn)
	InitFlags(flagset *flag.FlagSet)
}

type Environment interface {
	RequirementLevel() requirement.Levels
	FeatureState() feature.States
	Namespace() string
	Context() context.Context
}

type FlagSetFn func(flagset *flag.FlagSet)

type Feature struct {
	Name          string
	Preconditions []Precondition
	Assertions    []Assertion
}

type AssertFn func(ctx context.Context, t T)

type PreConFn func(ctx context.Context, t T)

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
	s    feature.States
}

func (a *Asserter) Must(name string, fn AssertFn) Assertable {
	a.Assert(requirement.Must, name, fn)
	return a
}

func (a *Asserter) Should(name string, fn AssertFn) Assertable {
	a.Assert(requirement.Should, name, fn)
	return a
}

func (a *Asserter) May(name string, fn AssertFn) Assertable {
	a.Assert(requirement.May, name, fn)
	return a
}

func (a *Asserter) MustNot(name string, fn AssertFn) Assertable {
	a.Assert(requirement.MustNot, name, fn)
	return a
}

func (a *Asserter) ShouldNot(name string, fn AssertFn) Assertable {
	a.Assert(requirement.ShouldNot, name, fn)
	return a
}

func (f *Feature) Alpha(name string) Assertable {
	return f.asserter(feature.Alpha, name)
}

func (f *Feature) Beta(name string) Assertable {
	return f.asserter(feature.Beta, name)
}

func (f *Feature) Stable(name string) Assertable {
	return f.asserter(feature.Stable, name)
}

func (f *Feature) asserter(s feature.States, name string) Assertable {
	return &Asserter{
		f:    f,
		name: name,
		s:    s,
	}
}

func (a *Asserter) Assert(l requirement.Levels, name string, fn AssertFn) {
	a.f.Assertions = append(a.f.Assertions, Assertion{
		Name: fmt.Sprintf("%s %s", a.name, name),
		A:    fn,
		S:    a.s,
		L:    l,
	})
}

type Assertion struct {
	Name string
	S    feature.States
	L    requirement.Levels
	A    AssertFn
}

type FeatureTester interface {
	Test(ctx context.Context, t *testing.T, f *Feature)
}

type FilterFlagMagic struct{}

func (magic FilterFlagMagic) Test(ctx context.Context, t *testing.T, f *Feature) {
	panic("implement me")
}

func GlobalRunner() FeatureTester {
	return &FilterFlagMagic{}
}

// --------------------------

func AssertSomething(ctx context.Context, t T) {
	//todo
}

func RequireSomething(ctx context.Context, t T) {
	// todo
}

//
//func FeatureBar() *Feature {
//	f := new(Feature)
//
//	f.Precondition("needs a foo", RequireSomething)
//
//	f.Stable("for a real feature").
//		Must("feature a", AssertSomething).
//		Must("feature b", AssertSomething).
//		MustNot("scope creep", AssertSomething).
//		May("optional feature", AssertSomething).
//		Should("add an inline function", func(ctx context.Context, t T) {
//			// todo
//		}).
//		ShouldNot("do this other bad thing but we will ignore it", func(t PT, e Environment) {
//			// todo: more
//		})
//
//	return f
//}

//
//func TestBar(t *testing.T) {
//	GlobalRunner().Test(t, FeatureBar())
//}
//
//// In another repo...
//func TestBar_Alpha(t *testing.T) {
//	f := FeatureBar()
//
//	f.Alpha("in-progress feature X").
//		May("return some data", AssertSomething)
//
//	GlobalRunner().Test(t, f)
//}
