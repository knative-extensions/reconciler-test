//go:build e2e
// +build e2e

/*
Copyright 2023 The Knative Authors

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

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeclient "knative.dev/pkg/client/injection/kube/client"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/knative"
	"knative.dev/reconciler-test/pkg/resources/cronjob"
	"knative.dev/reconciler-test/resources/certificate"
)

func TestCronJobInstall(t *testing.T) {

	ctx, env := global.Environment(
		knative.WithKnativeNamespace("knative-reconciler-test"),
		knative.WithLoggingConfig,
		knative.WithTracingConfig,
		k8s.WithEventListener,
		environment.Managed(t),
	)
	t.Cleanup(env.Finish)

	name := feature.MakeRandomK8sName("cron")
	sink := feature.MakeRandomK8sName("sink")

	env.Test(ctx, t, eventshub.Install(sink, eventshub.StartReceiver).AsFeature())
	env.Test(ctx, t, cronjob.Install(
		name,
		"gcr.io/knative-nightly/knative.dev/eventing/cmd/heartbeats",
		cronjob.WitEnvs(map[string]string{
			"POD_NAME":      "heartbeats",
			"POD_NAMESPACE": environment.FromContext(ctx).Namespace(),
			"K_SINK":        fmt.Sprintf("%s://%s.%s.svc", "http", sink, environment.FromContext(ctx).Namespace()),
			"ONE_SHOT":      "true",
		})).
		AsFeature(),
	)
	env.Test(ctx, t, cronjob.AtLeastOneIsSucceeded(name).AsFeature())
}

func TestRotateCertificates(t *testing.T) {

	ctx, env := global.Environment()

	ns := "knative-reconciler-test"

	secret := types.NamespacedName{
		Namespace: ns,
		Name:      "test-certificate-tls",
	}

	rc := certificate.RotateCertificate{
		Certificate: types.NamespacedName{
			Namespace: ns,
			Name:      "test-certificate",
		},
	}

	f := feature.NewFeatureNamed("Rotate certificates")

	var before *corev1.Secret

	f.Setup("get secret", func(ctx context.Context, t feature.T) {
		var err error
		before, err = kubeclient.Get(ctx).
			CoreV1().
			Secrets(secret.Namespace).
			Get(ctx, secret.Name, metav1.GetOptions{})
		if err != nil {
			t.Errorf("Failed to get secret %s/%s: %v", secret.Namespace, secret.Name, err)
		}
	})

	f.Requirement("rotate certs", certificate.Rotate(rc))

	f.Assert("verify different certificates", func(ctx context.Context, t feature.T) {
		after, err := kubeclient.Get(ctx).
			CoreV1().
			Secrets(secret.Namespace).
			Get(ctx, secret.Name, metav1.GetOptions{})
		if err != nil {
			t.Errorf("Failed to get secret %s/%s: %v", secret.Namespace, secret.Name, err)
		}

		if isEqualKey(before, after, "tls.crt") {
			t.Errorf("Certificates rotation didn't happen tls.crt is equal")
		}
		if isEqualKey(before, after, "tls.key") {
			t.Errorf("Certificates rotation didn't happen tls.key is equal")
		}
	})

	env.Test(ctx, t, f)
}

func isEqualKey(s1, s2 *corev1.Secret, key string) bool {
	return bytes.Equal(s1.Data[key], s2.Data[key])
}
