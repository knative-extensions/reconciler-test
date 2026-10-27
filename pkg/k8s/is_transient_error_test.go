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

package k8s_test

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/reconciler-test/pkg/k8s"
)

// The samples below are real transient errors observed in CI, see also
// https://github.com/knative/pkg/issues/1509.
func TestIsTransientError(t *testing.T) {
	issuers := schema.GroupResource{Group: "cert-manager.io", Resource: "issuers"}
	tests := []struct {
		name      string
		err       error
		transient bool
	}{
		{
			name:      "internal error on create",
			err:       apierrors.NewInternalError(errors.New(`Internal Server Error: "/apis/cert-manager.io/v1/namespaces/test-swrgmkeu/issuers/selfsigned-issuer": the server is currently unable to handle the request`)),
			transient: true,
		},
		{
			name:      "service unavailable",
			err:       apierrors.NewServiceUnavailable("the server is currently unable to handle the request"),
			transient: true,
		},
		{
			name:      "server timeout",
			err:       apierrors.NewServerTimeout(issuers, "create", 1),
			transient: true,
		},
		{
			name:      "request timeout",
			err:       apierrors.NewTimeoutError("Timeout: request did not complete within requested timeout - context deadline exceeded", 5),
			transient: true,
		},
		{
			name:      "too many requests",
			err:       apierrors.NewTooManyRequestsError("quota exceeded"),
			transient: true,
		},
		{
			name:      "unauthorized (credential refresh)",
			err:       apierrors.NewUnauthorized("token expired"),
			transient: true,
		},
		{
			name:      "webhook no agent available",
			err:       errors.New(`Internal error occurred: failed calling webhook "webhook.cert-manager.io": failed to call webhook: Post "https://webhook.cert-manager.svc:443/validate?timeout=10s": remote error: tls: internal error: "No agent available"`),
			transient: true,
		},
		{
			name:      "webhook EOF",
			err:       errors.New(`Internal error occurred: failed calling webhook "defaulting.webhook.kafka.eventing.knative.dev": Post "https://kafka-webhook-eventing.knative-eventing.svc:443/defaulting?timeout=2s": EOF`),
			transient: true,
		},
		{
			name:      "webhook deadline exceeded",
			err:       errors.New(`Internal error occurred: failed calling webhook "inmemorychannel.eventing.knative.dev": failed to call webhook: Post "https://inmemorychannel-webhook.knative-eventing-9zkohswwf9.svc:443/defaulting?timeout=10s": context deadline exceeded`),
			transient: true,
		},
		{
			// Older apiserver wording without the "failed calling webhook" prefix.
			name:      "webhook failed to call only",
			err:       errors.New(`Internal error occurred: failed to call webhook "inmemorychannel.eventing.knative.dev": Post "https://inmemorychannel-webhook.knative-eventing-9zkohswwf9.svc:443/defaulting?timeout=10s": EOF`),
			transient: true,
		},
		{
			name:      "connection reset",
			err:       &net.OpError{Op: "read", Err: os.NewSyscallError("read", syscall.ECONNRESET)},
			transient: true,
		},
		{
			// Callers wrap API errors before checking them.
			name:      "wrapped internal error",
			err:       fmt.Errorf("failed to create resource %w - Resource:\n%s", apierrors.NewInternalError(errors.New("boom")), "yaml"),
			transient: true,
		},
		{
			name:      "invalid",
			err:       apierrors.NewInvalid(schema.GroupKind{Group: "cert-manager.io", Kind: "Issuer"}, "selfsigned-issuer", nil),
			transient: false,
		},
		{
			name:      "bad request",
			err:       apierrors.NewBadRequest("malformed object"),
			transient: false,
		},
		{
			name:      "forbidden",
			err:       apierrors.NewForbidden(issuers, "selfsigned-issuer", errors.New("rbac")),
			transient: false,
		},
		{
			name:      "not found",
			err:       apierrors.NewNotFound(issuers, "selfsigned-issuer"),
			transient: false,
		},
		{
			name:      "already exists",
			err:       apierrors.NewAlreadyExists(issuers, "selfsigned-issuer"),
			transient: false,
		},
		{
			name:      "plain error",
			err:       errors.New("some test bug"),
			transient: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := k8s.IsTransientError(tc.err); got != tc.transient {
				t.Errorf("IsTransientError(%v) = %t, want %t", tc.err, got, tc.transient)
			}
		})
	}
}
