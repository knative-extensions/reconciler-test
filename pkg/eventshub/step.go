package eventshub

import (
	"context"
	"testing"

	"knative.dev/reconciler-test/pkg/feature"
)

type StoreIdentifier string

func OnStore(name string) StoreIdentifier {
	return StoreIdentifier(name)
}

type AtLeastMatch struct {
	StoreIdentifier
	min int
}

func (name StoreIdentifier) AtLeast(min int) AtLeastMatch {
	return AtLeastMatch{name, min}
}

// Match
// OnStore(store).AtLeast(min).Match(matchers) is equivalent to StoreFromContext(ctx, store).AssertAtLeast(min, matchers)
func (m AtLeastMatch) Match(matchers ...EventInfoMatcher) feature.StepFn {
	return func(ctx context.Context, t *testing.T) {
		StoreFromContext(ctx, string(m.StoreIdentifier)).AssertAtLeast(m.min, matchers...)
	}
}

type InRangeMatch struct {
	StoreIdentifier
	min int
	max int
}

func (name StoreIdentifier) InRange(min int, max int) InRangeMatch {
	return InRangeMatch{name, min, max}
}

// Match
// OnStore(store).InRange(min, max).Match(matchers) is equivalent to StoreFromContext(ctx, store).AssertInRange(min, max, matchers)
func (m InRangeMatch) Match(matchers ...EventInfoMatcher) feature.StepFn {
	return func(ctx context.Context, t *testing.T) {
		StoreFromContext(ctx, string(m.StoreIdentifier)).AssertInRange(m.min, m.max, matchers...)
	}
}

type ExactMatch struct {
	StoreIdentifier
	n int
}

func (name StoreIdentifier) Exact(n int) ExactMatch {
	return ExactMatch{name, n}
}

// Match
// OnStore(store).Exact(n).Match(matchers) is equivalent to StoreFromContext(ctx, store).AssertExact(n, matchers)
func (m ExactMatch) Match(matchers ...EventInfoMatcher) feature.StepFn {
	return func(ctx context.Context, t *testing.T) {
		StoreFromContext(ctx, string(m.StoreIdentifier)).AssertExact(m.n, matchers...)
	}
}

type NotMatch struct {
	StoreIdentifier
}

func (name StoreIdentifier) Not() NotMatch {
	return NotMatch{name}
}

// Match
// OnStore(store).Not().Match(matchers) is equivalent to StoreFromContext(ctx, store).AssertNot(matchers)
func (m NotMatch) Match(matchers ...EventInfoMatcher) feature.StepFn {
	return func(ctx context.Context, t *testing.T) {
		StoreFromContext(ctx, string(m.StoreIdentifier)).AssertNot(matchers...)
	}
}
