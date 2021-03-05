package environment

import (
	"testing"

	"go.uber.org/atomic"

	"knative.dev/reconciler-test/pkg/feature"
)

// t records if the test succeded or failing
type t struct {
	*testing.T

	failed  *atomic.Bool
	skipped *atomic.Bool
}

var _ feature.T = (*t)(nil)

func wrapT(originalT *testing.T) *t {
	return &t{
		T:       originalT,
		failed:  atomic.NewBool(false),
		skipped: atomic.NewBool(false),
	}
}

func (t *t) Error(args ...interface{}) {
	t.T.Helper()
	t.failed.Store(true)
	t.T.Error(args...)
}

func (t *t) Errorf(format string, args ...interface{}) {
	t.T.Helper()
	t.failed.Store(true)
	t.T.Errorf(format, args...)
}

func (t *t) Fail() {
	t.T.Helper()
	t.failed.Store(true)
	t.T.Fail()
}

func (t *t) Fatal(args ...interface{}) {
	t.T.Helper()
	t.failed.Store(true)
	t.T.Fatal(args...)
}

func (t *t) Fatalf(format string, args ...interface{}) {
	t.T.Helper()
	t.failed.Store(true)
	t.T.Fatalf(format, args...)
}

func (t *t) FailNow() {
	t.T.Helper()
	t.failed.Store(true)
	t.T.FailNow()
}

func (t *t) Log(args ...interface{}) {
	t.T.Helper()
	t.T.Log(args...)
}

func (t *t) Logf(format string, args ...interface{}) {
	t.T.Helper()
	t.T.Logf(format, args...)
}

func (t *t) Skip(args ...interface{}) {
	t.T.Helper()
	t.skipped.Store(true)
	t.T.Skip(args...)
}

func (t *t) Skipf(format string, args ...interface{}) {
	t.T.Helper()
	t.skipped.Store(true)
	t.T.Skipf(format, args...)
}

func (t *t) SkipNow() {
	t.T.Helper()
	t.skipped.Store(true)
	t.T.SkipNow()
}
