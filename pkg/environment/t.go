package environment

import (
	"runtime"
	"testing"

	"go.uber.org/atomic"
)

type t struct {
	t *testing.T

	failed  *atomic.Bool
	skipped *atomic.Bool
}

func (t *t) Error(args ...interface{}) {
	t.t.Helper()
	t.t.Log(args...)
}

func (t *t) Errorf(format string, args ...interface{}) {
	t.t.Helper()
	t.t.Logf(format, args...)
	t.Fail()
}

func (t *t) Fail() {
	t.failed.Store(true)
}

func (t *t) Fatal(args ...interface{}) {
	t.t.Helper()
	t.t.Log(args...)
	t.FailNow()
}

func (t *t) Fatalf(format string, args ...interface{}) {
	t.t.Helper()
	t.t.Logf(format, args...)
	t.FailNow()
}

func (t *t) FailNow() {
	t.failed.Store(true)
	runtime.Goexit()
}

func (t *t) Log(args ...interface{}) {
	t.t.Helper()
	t.t.Log(args...)
}

func (t *t) Logf(format string, args ...interface{}) {
	t.t.Helper()
	t.t.Logf(format, args...)
}

func (t *t) Skip(args ...interface{}) {
	t.t.Helper()
	t.t.Log(args...)
	t.SkipNow()
}

func (t *t) Skipf(format string, args ...interface{}) {
	t.t.Helper()
	t.t.Logf(format, args...)
	t.SkipNow()
}

func (t *t) SkipNow() {
	t.skipped.Store(true)
	runtime.Goexit()
}
