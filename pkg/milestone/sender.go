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

package milestone

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
)

const (
	enableRetry  = true
	retryBackoff = 10 * time.Millisecond
	maxTries     = 3
)

type envConfig struct {
	MilestoneEventsTarget string `envconfig:"MILESTONE_EVENTS_TARGET"`
}

// Sender sends milestone events.
type Sender interface {
	// Send will send CloudEvents and return the result.
	// Send must be implemented in a nil safe way.
	Send(ctx context.Context, event cloudevents.Event)
}

// NilSafeClient is a simple wrapper around a cloudevent client that implements
// Sender to provide nil check safety.
type NilSafeClient struct {
	Client cloudevents.Client
}

// Send implements Sender.Send.
func (n *NilSafeClient) Send(ctx context.Context, event cloudevents.Event) {
	if n == nil || n.Client == nil {
		return
	}
	if enableRetry {
		// Adds retry to the outbound send attempt.
		ctx = cloudevents.ContextWithRetriesExponentialBackoff(ctx, retryBackoff, maxTries)
	}
	if result := n.Client.Send(ctx, event); cloudevents.IsUndelivered(result) {
		logging.FromContext(ctx).Info("failed to deliver milestone event", result.Error())
		logging.FromContext(ctx).Errorw("failed to deliver milestone event", zap.Error(result))
	} else if cloudevents.IsNACK(result) {
		logging.FromContext(ctx).Error("milestone event target returned NACK", result, event.Type())
		logging.FromContext(ctx).Errorw("milestone event target returned NACK", zap.Error(result), zap.String("event", event.Type()))
	}
}

// NewMilestoneEventSenderFromEnv will attempt to pull the env var
// `MILESTONE_EVENTS_TARGET` as the target uri for sending milestone events.
func NewMilestoneEventSenderFromEnv() (Sender, error) {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		return nil, err
	}
	if len(env.MilestoneEventsTarget) > 0 {
		fmt.Printf("milestone events target: %s\n\n", env.MilestoneEventsTarget)
		return NewMilestoneEventSender(env.MilestoneEventsTarget)
	}
	return &NilSafeClient{}, nil
}

// NewMilestoneEventSender will convert target uri to a milestone event sender and return it.
//
func NewMilestoneEventSender(uri string) (Sender, error) {
	target, err := apis.ParseURL(uri)
	if err != nil {
		return nil, err
	}
	switch target.Scheme {
	case "http", "https":
		p, err := cloudevents.NewHTTP(cloudevents.WithTarget(target.String()))
		if err != nil {
			return nil, err
		}
		client, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		return &NilSafeClient{Client: client}, err
	default:
		return nil, fmt.Errorf("unsupported milestone event target uri: %q", target.String())
	}
}
