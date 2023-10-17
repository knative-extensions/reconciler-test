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
	"testing"
	"time"

	cetest "github.com/cloudevents/sdk-go/v2/test"

	"knative.dev/pkg/system"
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
		knative.WithKnativeNamespace(system.Namespace()),
		knative.WithLoggingConfig,
		knative.WithTracingConfig,
		k8s.WithEventListener,
	)

	env.Test(ctx, t, validToken())
	env.Test(ctx, t, invalidAudience())
	env.Test(ctx, t, expiredToken())
	env.Test(ctx, t, corruptedSignatureToken())
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

func invalidAudience() *feature.Feature {
	f := feature.NewFeatureNamed("OIDC - invalid audience")

	sinkName := feature.MakeRandomK8sName("sink")
	sourceName := feature.MakeRandomK8sName("source")

	event := cetest.FullEvent()

	f.Setup("deploy receiver", eventshub.Install(sinkName,
		eventshub.StartReceiver,
		eventshub.OIDCReceiverAudience("my-sink-audience")))

	f.Requirement("deploy sender", func(ctx context.Context, t feature.T) {
		eventshub.Install(sourceName,
			eventshub.StartSenderToResource(eventshub.ReceiverGVR(ctx), sinkName),
			eventshub.InputEvent(event),
			eventshub.OIDCSinkAudience("some-other-audience"),
		)(ctx, t)
	})

	f.Assert("Sink does not receive event", assert.OnStore(sinkName).
		MatchReceivedEvent(cetest.HasId(event.ID())).
		Not(),
	)

	f.Assert("Source sent event", assert.OnStore(sourceName).
		MatchSentEvent(cetest.HasId(event.ID())).
		Exact(1),
	)

	f.Assert("Source gets 401 response", assert.OnStore(sourceName).
		Match(assert.MatchStatusCode(401)).
		Exact(1),
	)

	return f
}

func expiredToken() *feature.Feature {
	f := feature.NewFeatureNamed("OIDC - expried token")

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
			eventshub.OIDCExpiredToken(),
		)(ctx, t)
	})

	f.Assert("Sink does not receive event", assert.OnStore(sinkName).
		MatchReceivedEvent(cetest.HasId(event.ID())).
		Not(),
	)

	f.Assert("Source sent event", assert.OnStore(sourceName).
		MatchSentEvent(cetest.HasId(event.ID())).
		Exact(1),
	)

	f.Assert("Source gets 401 response", assert.OnStore(sourceName).
		Match(assert.MatchStatusCode(401)).
		Exact(1),
	)

	return f
}

func corruptedSignatureToken() *feature.Feature {
	f := feature.NewFeatureNamed("OIDC - expried token")

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
			eventshub.OIDCCorruptedSignature(),
		)(ctx, t)
	})

	f.Assert("Sink does not receive event", assert.OnStore(sinkName).
		MatchReceivedEvent(cetest.HasId(event.ID())).
		Not(),
	)

	f.Assert("Source sent event", assert.OnStore(sourceName).
		MatchSentEvent(cetest.HasId(event.ID())).
		Exact(1),
	)

	f.Assert("Source gets 401 response", assert.OnStore(sourceName).
		Match(assert.MatchStatusCode(401)).
		Exact(1),
	)

	return f
}
