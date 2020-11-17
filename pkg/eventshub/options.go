/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

        https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eventshub

import (
	"context"
	"encoding/json"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"knative.dev/pkg/network"

	"knative.dev/reconciler-test/pkg/environment"
)

// EventsHubOption is used to define an env for the eventshub image
type EventsHubOption = func(context.Context, map[string]string) error

// StartReceiver starts the receiver in the eventshub
// This can be used together with EchoEvent, ReplyWithTransformedEvent, ReplyWithAppendedData
var StartReceiver EventsHubOption = envAdditive("EVENT_GENERATORS", "receiver")

// StartSender starts the sender in the eventshub
// This can be used together with InputEvent, AddTracing, EnableIncrementalId, InputEncoding and InputHeader options
func StartSender(sinkSvc string) EventsHubOption {
	return compose(envAdditive("EVENT_GENERATORS", "sender"), func(ctx context.Context, envs map[string]string) error {
		envs["SINK"] = "http://" + network.GetServiceHostname(sinkSvc, environment.FromContext(ctx).Namespace())
		return nil
	})
}

// EchoEvent is an option to let the eventshub reply with the received event
var EchoEvent EventsHubOption = envOption("REPLY", "true")

// ReplyWithTransformedEvent is an option to let the eventshub reply with the transformed event
func ReplyWithTransformedEvent(replyEventType string, replyEventSource string, replyEventData string) EventsHubOption {
	return compose(
		envOption("REPLY", "true"),
		envOptionalOpt("REPLY_EVENT_TYPE", replyEventType),
		envOptionalOpt("REPLY_EVENT_SOURCE", replyEventSource),
		envOptionalOpt("REPLY_EVENT_DATA", replyEventData),
	)
}

// ReplyWithAppendedData is an option to let the eventshub reply with the transformed event with appended data
func ReplyWithAppendedData(appendData string) EventsHubOption {
	return compose(
		envOption("REPLY", "true"),
		envOptionalOpt("REPLY_APPEND_DATA", appendData),
	)
}

// InputEvent is an option to provide the event to send when deploying the event sender
func InputEvent(event cloudevents.Event) EventsHubOption {
	encodedEvent, err := json.Marshal(event)
	if err != nil {
		return func(ctx context.Context, envs map[string]string) error {
			return err
		}
	}
	return envOption("INPUT_EVENT", string(encodedEvent))
}

// AddTracing adds tracing headers when sending events.
func AddTracing() EventsHubOption {
	return envOption("ADD_TRACING", "true")
}

// EnableIncrementalId creates a new incremental id for each sent event.
func EnableIncrementalId() EventsHubOption {
	return envOption("INCREMENTAL_ID", "true")
}

// InputEncoding forces the encoding of the event for each sent event.
func InputEncoding(encoding cloudevents.Encoding) EventsHubOption {
	return envOption("EVENT_ENCODING", encoding.String())
}

// InputHeader adds the following header to the sent headers.
func InputHeader(k, v string) EventsHubOption {
	return envAdditive("INPUT_HEADERS", k+":"+v)
}

func noop(context.Context, map[string]string) error {
	return nil
}

func compose(options ...EventsHubOption) EventsHubOption {
	return func(ctx context.Context, envs map[string]string) error {
		for _, opt := range options {
			if err := opt(ctx, envs); err != nil {
				return err
			}
		}
		return nil
	}
}

func envOptionalOpt(key, value string) EventsHubOption {
	if value != "" {
		return func(ctx context.Context, envs map[string]string) error {
			envs[key] = value
			return nil
		}
	} else {
		return noop
	}
}

func envOption(key, value string) EventsHubOption {
	return func(ctx context.Context, envs map[string]string) error {
		envs[key] = value
		return nil
	}
}

func envAdditive(key, value string) EventsHubOption {
	return func(ctx context.Context, m map[string]string) error {
		if containedValue, ok := m[key]; ok {
			m[key] = containedValue + "," + value
		} else {
			m[key] = value
		}
		return nil
	}
}
