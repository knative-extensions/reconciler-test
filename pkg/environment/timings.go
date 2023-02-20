/*
Copyright 2021 The Knative Authors

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
	"time"
)

const (
	DefaultPollInterval = 3 * time.Second
	DefaultPollTimeout  = 2 * time.Minute
)

type timingsKey struct{}
type timingsType struct {
	interval time.Duration
	timeout  time.Duration
}

// PollTimingsFromContext will get the previously set poll timing from context,
// or return the defaults if not found.
// - values from from context.
// - defaults.
func ContextWithPollTimings(ctx context.Context, interval, timeout time.Duration) context.Context {
	return context.WithValue(ctx, timingsKey{}, timingsType{
		interval: interval,
		timeout:  timeout,
	})
}

// PollTimingsFromContext will get the previously set poll timing from context,
// or return the defaults if not found.
// - values from from context.
// - defaults.
func PollTimingsFromContext(ctx context.Context) (time.Duration, time.Duration) {
	if t, ok := ctx.Value(timingsKey{}).(timingsType); ok {
		return t.interval, t.timeout
	}
	return DefaultPollInterval, DefaultPollTimeout
}

// PollTimings will find the correct timings based on priority:
// - passed timing slice [interval, timeout].
// - values from from context.
// - defaults.
func PollTimings(ctx context.Context, timings []time.Duration) (time.Duration /*interval*/, time.Duration /*timeout*/) {
	// Use the passed timing first, but it could be nil or a strange length.
	if len(timings) >= 2 {
		return timings[0], timings[1]
	}

	var interval *time.Duration

	// Use the passed timings if only interval is provided.
	if len(timings) == 1 {
		interval = &timings[0]
	}

	di, timeout := PollTimingsFromContext(ctx)
	if interval == nil {
		interval = &di
	}

	return *interval, timeout
}
