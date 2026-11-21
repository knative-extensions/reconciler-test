/*
Copyright 2026 The Knative Authors

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
	"errors"
	"testing"
	"time"

	"knative.dev/reconciler-test/pkg/feature"
)

// recordingT records FailNow without terminating the goroutine.
type recordingT struct {
	failNowCalled bool
}

var _ feature.T = (*recordingT)(nil)

func (r *recordingT) Name() string                              { return "recordingT" }
func (r *recordingT) Log(args ...interface{})                   {}
func (r *recordingT) Logf(format string, args ...interface{})   {}
func (r *recordingT) Error(args ...interface{})                 {}
func (r *recordingT) Errorf(format string, args ...interface{}) {}
func (r *recordingT) Fail()                                     {}
func (r *recordingT) Fatal(args ...interface{})                 { r.FailNow() }
func (r *recordingT) Fatalf(format string, args ...interface{}) { r.FailNow() }
func (r *recordingT) FailNow()                                  { r.failNowCalled = true }
func (r *recordingT) Skip(args ...interface{})                  {}
func (r *recordingT) Skipf(format string, args ...interface{})  {}
func (r *recordingT) SkipNow()                                  {}
func (r *recordingT) Failed() bool                              { return r.failNowCalled }
func (r *recordingT) Skipped() bool                             { return false }
func (r *recordingT) Cleanup(f func())                          {}
func (r *recordingT) Deadline() (time.Time, bool)               { return time.Time{}, false }

func TestMagicEnvironment_failSetup_scopesToManagedTest(t *testing.T) {
	rt := &recordingT{}
	env := &MagicEnvironment{managedT: rt}

	env.failSetup(context.Background(), errors.New("boom"))

	if !rt.failNowCalled {
		t.Fatal("expected failSetup to FailNow the managed test so a setup error cannot os.Exit the binary")
	}
}
