//go:build e2e
// +build e2e

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

package eventshub_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	cetest "github.com/cloudevents/sdk-go/v2/test"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/eventshub/assert"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/knative"
)

func TestEventsHubOIDCAuth(t *testing.T) {
	t.Parallel()

	ctx, env := global.Environment(
		environment.Managed(t),
		environment.WithPollTimings(4*time.Second, 12*time.Minute),
		knative.WithKnativeNamespace("knative-reconciler-test"),
		knative.WithLoggingConfig,
		knative.WithTracingConfig,
		k8s.WithEventListener,
	)

	env.ParallelTest(ctx, t, validToken())
	env.ParallelTestSet(ctx, t, invalidTokens())
}

func validToken() *feature.Feature {
	f := feature.NewFeatureNamed("OIDC - valid token")

	sinkName := feature.MakeRandomK8sName("sink")
	sourceName := feature.MakeRandomK8sName("source")
	audience := "my-sink-audience"

	event := cetest.FullEvent()

	f.Setup("deploy receiver", eventshub.Install(sinkName,
		eventshub.StartReceiver,
		eventshub.OIDCReceiverAudience(audience)))

	f.Requirement("deploy sender", func(ctx context.Context, t feature.T) {
		eventshub.Install(sourceName,
			eventshub.StartSenderToResource(eventshub.ReceiverGVR(ctx), sinkName),
			eventshub.InputEvent(event),
			eventshub.OIDCSinkAudience(audience),
		)(ctx, t)
	})

	f.Assert("Receive event", assert.OnStore(sinkName).
		MatchReceivedEvent(cetest.HasId(event.ID())).
		AtLeast(1),
	)

	f.Assert("Sent event", assert.OnStore(sourceName).
		MatchSentEvent(cetest.HasId(event.ID())).
		AtLeast(1),
	)

	return f
}

func invalidTokens() *feature.FeatureSet {
	tests := []struct {
		name   string
		option eventshub.EventsHubOption
	}{
		{
			name:   "invalid audience",
			option: eventshub.OIDCInvalidAudience(),
		},
		{
			name:   "invalid signature",
			option: eventshub.OIDCCorruptedSignature(),
		},
		{
			name:   "expired token",
			option: eventshub.OIDCExpiredToken(),
		},
	}

	fs := &feature.FeatureSet{
		Name: "OIDC: invalid tokens",
	}
	for i, test := range tests {
		f := feature.NewFeatureNamed(test.name)
		opt := test.option

		sinkName := feature.MakeRandomK8sName(fmt.Sprintf("sink-%d", i))
		sourceName := feature.MakeRandomK8sName(fmt.Sprintf("source-%d", i))
		sinkAudience := fmt.Sprintf("my-sink-audience-%d", i)

		event := cetest.FullEvent()

		f.Setup("deploy receiver", eventshub.Install(sinkName,
			eventshub.StartReceiver,
			eventshub.OIDCReceiverAudience(sinkAudience)))

		f.Requirement("deploy sender", func(ctx context.Context, t feature.T) {
			eventshub.Install(sourceName,
				eventshub.StartSenderToResource(eventshub.ReceiverGVR(ctx), sinkName),
				eventshub.InputEvent(event),
				eventshub.OIDCSinkAudience(sinkAudience),
				opt,
			)(ctx, t)
		})

		f.Assert("Source sends event", assert.OnStore(sourceName).
			MatchSentEvent(cetest.HasId(event.ID())).
			Exact(1),
		)

		f.Assert("Sink rejects event", assert.OnStore(sinkName).
			MatchRejectedEvent(cetest.HasId(event.ID())).
			Exact(1),
		)

		f.Assert("Source gets 401 response", assert.OnStore(sourceName).
			Match(assert.MatchStatusCode(401)).
			Exact(1),
		)

		fs.Features = append(fs.Features, f)
	}

	return fs
}
