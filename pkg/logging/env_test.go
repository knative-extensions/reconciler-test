/*
Copyright 2022 The Knative Authors
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

package logging_test

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"go.uber.org/zap/zapcore"
	"knative.dev/reconciler-test/pkg/logging"
)

func TestLevelFromEnvironment(t *testing.T) {
	const want = zapcore.WarnLevel
	t.Setenv("TEST_LOGGER_LEVEL", want.String())

	got := logging.LevelFromEnvironment(context.TODO())
	if got != want {
		t.Errorf("want: %+v, got: %+v", want, got)
	}
}

func TestLevelFromEnvironment_invalidFallsBackToDebug(t *testing.T) {
	if os.Getenv("RECONCILER_TEST_INVALID_LEVEL_CHILD") == "1" {
		got := logging.LevelFromEnvironment(context.TODO())
		if got != zapcore.DebugLevel {
			os.Exit(2)
		}
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestLevelFromEnvironment_invalidFallsBackToDebug$")
	cmd.Env = append(os.Environ(), "RECONCILER_TEST_INVALID_LEVEL_CHILD=1", "TEST_LOGGER_LEVEL=not-a-level")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("invalid TEST_LOGGER_LEVEL must not os.Exit the process (got %v)", err)
	}
}
