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

package eventshub

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	v1 "knative.dev/pkg/apis/duck/v1"
)

func TestCertificateIsReady(t *testing.T) {
	tests := []struct {
		name       string
		conditions v1.Conditions
		want       bool
	}{
		{
			name: "ready",
			conditions: v1.Conditions{{
				Type:   apis.ConditionReady,
				Status: corev1.ConditionTrue,
				Reason: "Ready",
			}},
			want: true,
		},
		{
			name: "not ready yet",
			conditions: v1.Conditions{{
				Type:    apis.ConditionReady,
				Status:  corev1.ConditionFalse,
				Reason:  "IssuancePending",
				Message: "Waiting for Certificate to be issued",
			}},
			want: false,
		},
		{
			name:       "no conditions",
			conditions: nil,
			want:       false,
		},
		{
			name: "unrelated condition",
			conditions: v1.Conditions{{
				Type:   "Approval",
				Status: corev1.ConditionTrue,
			}},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := v1.KResource{Status: v1.Status{Conditions: tc.conditions}}
			if got := certificateIsReady(t)(obj); got != tc.want {
				t.Errorf("certificateIsReady() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestCACertsFromSecret(t *testing.T) {
	certs, err := caCertsFromSecret(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns", Name: caSecretName},
		Data:       map[string][]byte{"ca.crt": []byte("CA CERT")},
	})
	if err != nil {
		t.Fatalf("caCertsFromSecret: %v", err)
	}
	if string(certs) != "CA CERT" {
		t.Errorf("caCertsFromSecret = %q, want %q", certs, "CA CERT")
	}

	// A CA secret without the ca.crt key must be an error, not a silent nil.
	_, err = caCertsFromSecret(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns", Name: caSecretName},
		Data:       map[string][]byte{"tls.crt": []byte("server cert")},
	})
	if err == nil {
		t.Fatal("expected an error for a CA secret without the ca.crt key")
	}
}
