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

	cetest "github.com/cloudevents/sdk-go/v2/test"
	"github.com/google/uuid"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/eventshub/assert"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/knative"
	"knative.dev/reconciler-test/pkg/resources/secret"
)

func TestEventsHubReceiverTLS(t *testing.T) {
	t.Parallel()

	ctx, env := global.Environment(
		environment.Managed(t), // Will call env.Finish() when the test exits.
		eventshub.WithTLS(t),
		knative.WithKnativeNamespace("knative-reconciler-test"),
		knative.WithLoggingConfig,
		knative.WithTracingConfig,
		k8s.WithEventListener,
	)

	env.Prerequisite(ctx, t, ensureCACerts())
	env.Test(ctx, t, receiverTLS())
}

func ensureCACerts() *feature.Feature {
	f := feature.NewFeatureNamed("ensure CA certs created")

	f.Assert("CA secret is present", secret.IsPresent("eventshub-ca"))
	f.Assert("CA certs is in context", func(ctx context.Context, t feature.T) {
		caCerts := eventshub.GetCaCerts(ctx)
		if caCerts == nil || len(*caCerts) == 0 {
			t.Errorf("expected non empty CA certs")
			return
		}
	})

	return f
}

func receiverTLS() *feature.Feature {
	f := feature.NewFeatureNamed("Receiver TLS")

	sinkName := feature.MakeRandomK8sName("sink")
	sourceName := feature.MakeRandomK8sName("source")
	secretName := fmt.Sprintf("server-tls-%s", sinkName)

	event := cetest.FullEvent()
	event.SetID(uuid.New().String())

	f.Setup("deploy TLS receiver", eventshub.Install(sinkName, eventshub.StartReceiverTLS))

	f.Requirement("deploy TLS sender", func(ctx context.Context, t feature.T) {
		eventshub.Install(sourceName,
			eventshub.StartSenderToResourceTLS(eventshub.ReceiverGVR(ctx), sinkName, eventshub.GetCaCerts(ctx)),
			eventshub.InputEvent(event),
		)(ctx, t)
	})

	f.Assert("TLS certificate pair secret is present", secret.IsPresent(secretName))

	f.Assert("Receive event", assert.OnStore(sinkName).
		MatchReceivedEvent(cetest.HasId(event.ID())).
		AtLeast(1),
	)
	f.Assert("Sent event", assert.OnStore(sourceName).
		MatchSentEvent(cetest.HasId(event.ID())).
		AtLeast(1),
	)
	f.Assert("Sender received expected peer certificate", assert.OnStore(sourceName).
		MatchPeerCertificatesReceived(assert.MatchPeerCertificatesFromSecret(secretName, "tls.crt")).
		AtLeast(1),
	)

	return f
}
