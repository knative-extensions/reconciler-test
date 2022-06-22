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

package logging

import (
	"context"

	"go.uber.org/zap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/signals"
)

// WithTestLogger returns a context with test logger configured.
func WithTestLogger(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = signals.NewContext()
	}
	return logging.WithLogger(ctx, logger())
}

func logger() *zap.SugaredLogger {
	if log, err := zap.NewDevelopment(); err != nil {
		panic(err)
	} else {
		return log.Named("test").Sugar()
	}
}
