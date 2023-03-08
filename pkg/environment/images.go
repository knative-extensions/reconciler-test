/*
Copyright 2020 The Knative Authors

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
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/logging"

	"knative.dev/reconciler-test/pkg/images/ko"
)

var (
	// defaultImageProducer is the function that will be used to produce the
	// container images by default.
	//
	// To use a different image producer pass a different ImageProducer
	// when creating an Environment through GlobalEnvironment
	// see WithImageProducer.
	defaultImageProducer = ImageProducer(ko.Publish)

	// produceImagesLock is used to ensure that ProduceImages is only called
	// once at the time.
	produceImagesLock = sync.Mutex{}
)

// ImageProducer is a function that will be used to produce the container images.
//
// pack is a Go main package reference like `knative.dev/reconciler-test/cmd/eventshub`.
type ImageProducer func(ctx context.Context, pack string) (string, error)

// RegisterPackage registers an interest in producing an image based on the
// provided package.
// Can be called multiple times with the same package.
// A package will be used to produce the image and used
// like `image: ko://<package>` inside test yaml.
func RegisterPackage(packs ...string) EnvOpts {
	return func(ctx context.Context, _ Environment) (context.Context, error) {
		ps := getPackagesStore(ctx)
		ps.register(packs)
		return withPackagesStore(ctx, ps), nil
	}
}

// WithImages will bypass ProduceImages() and use the provided image set
// instead. Should be called before ProduceImages(), if used, likely in an
// init() method. An images value should be a container registry image. The
// images map is presented to the templates on the field `images`, and used
// like `image: <key>` inside test yaml.
func WithImages(given map[string]string) EnvOpts {
	return func(ctx context.Context, _ Environment) (context.Context, error) {
		store := getImageStore(ctx)
		store.register(given)
		return withImageStore(ctx, store), nil
	}
}

// ProduceImages returns back the packages that have been added.
// Will produce images once, can be called many times.
func ProduceImages(ctx context.Context) (map[string]string, error) {
	produceImagesLock.Lock()
	defer produceImagesLock.Unlock()

	store := getImageStore(ctx)
	ip := GetImageProducer(ctx)

	eg := errgroup.Group{}

	for _, pack := range getRegisteredPackages(ctx) {
		if store.has(pack) {
			continue
		}

		eg.Go(func() error {
			image, err := ip(ctx, pack)
			if err != nil {
				return err
			}
			store.register(map[string]string{pack: strings.TrimSpace(image)})
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to produce images: %w", err)
	}
	return store.copyRefs(), nil
}

func initializeImageStores(ctx context.Context) context.Context {
	var emptyPkgs []string
	emptyImgs := make(map[string]string)
	mctx, err := UnionOpts(
		RegisterPackage(emptyPkgs...),
		WithImages(emptyImgs),
	)(ctx, nil)
	if err != nil {
		logging.FromContext(ctx).
			Fatal("Failed to initialize image stores: ", err)
	}
	return mctx
}

// packagesStore aggregates packages registered and to be resolved by the ImageProducer.
type packagesStore struct {
	refs   sets.String
	refsMu *sync.Mutex
}

func (ps *packagesStore) register(packs []string) {
	ps.refsMu.Lock()
	defer ps.refsMu.Unlock()

	ps.refs.Insert(packs...)
}

func (ps *packagesStore) copyPackages() []string {
	ps.refsMu.Lock()
	defer ps.refsMu.Unlock()

	return ps.refs.List()
}

// packagesStoreKey is the key for packagesStore registered and to be resolved by the
// ImageProducer.
type packagesStoreKey struct{}

// getPackagesStore gets the packagesStore stored in the given context.
func getPackagesStore(ctx context.Context) *packagesStore {
	if ps, ok := ctx.Value(packagesStoreKey{}).(*packagesStore); ok {
		return ps
	}
	return &packagesStore{
		refs:   sets.NewString(),
		refsMu: &sync.Mutex{},
	}
}

// withPackagesStore put the given packageStore in the context.
func withPackagesStore(ctx context.Context, store *packagesStore) context.Context {
	return context.WithValue(ctx, packagesStoreKey{}, store)
}

// getRegisteredPackages get the registered packages from the packagesStore stored in the given
// context.
func getRegisteredPackages(ctx context.Context) []string {
	ps := getPackagesStore(ctx)
	return ps.copyPackages()
}

// imageStore stores a mapping between Go packages and container images
type imageStore struct {
	refs   map[string]string
	refsMu *sync.Mutex
}

// imageStoreKey is the context key for the imageStore.
type imageStoreKey struct{}

func getImageStore(ctx context.Context) *imageStore {
	if is, ok := ctx.Value(imageStoreKey{}).(*imageStore); ok {
		return is
	}
	return newImageStore(make(map[string]string))
}

func newImageStore(refs map[string]string) *imageStore {
	return &imageStore{
		refs:   refs,
		refsMu: &sync.Mutex{},
	}
}

func withImageStore(ctx context.Context, store *imageStore) context.Context {
	return context.WithValue(ctx, imageStoreKey{}, store)
}

func (is *imageStore) copyRefs() map[string]string {
	is.refsMu.Lock()
	defer is.refsMu.Unlock()

	refs := make(map[string]string, len(is.refs))
	for k, v := range is.refs {
		refs[k] = v
	}
	return refs
}

func (is *imageStore) has(key string) bool {
	is.refsMu.Lock()
	defer is.refsMu.Unlock()

	_, ok := is.refs[key]
	return ok
}

func (is *imageStore) register(images map[string]string) {
	is.refsMu.Lock()
	defer is.refsMu.Unlock()

	for k, v := range images {
		// Overrides existing keys
		is.refs[k] = v
	}
}

// imageProducerKey is the key for the ImageProducer context value.
type imageProducerKey struct{}

// WithImageProducer allows using a different ImageProducer
// when creating an Environment through GlobalEnvironment.
// Example usage:
// GlobalEnvironment.Environment(WithImageProducer(file.ImageProducer("images.yaml")))
func WithImageProducer(producer ImageProducer) EnvOpts {
	return func(ctx context.Context, env Environment) (context.Context, error) {
		return withImageProducer(ctx, producer), nil
	}
}

func withImageProducer(ctx context.Context, producer ImageProducer) context.Context {
	return context.WithValue(ctx, imageProducerKey{}, producer)
}

// GetImageProducer extracts an ImageProducer from the given context.
func GetImageProducer(ctx context.Context) ImageProducer {
	p := ctx.Value(imageProducerKey{})
	if p == nil {
		return defaultImageProducer
	}
	return p.(ImageProducer)
}
