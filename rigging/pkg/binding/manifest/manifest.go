package manifest

import (
	"context"

	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/reconciler-test/rigging/pkg/manifest"
)

const (
	MountPath = "/var/bindings/manifests" // filepath.Join isn't const.
)

// New
func New(ctx context.Context) (manifest.Manifest, error) {
	return manifest.NewYamlManifest(MountPath, true, dynamicclient.Get(ctx))
}
