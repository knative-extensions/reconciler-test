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

	"knative.dev/reconciler-test/pkg/images/ko"
)

var (
	// CurrentImageProducer is the function that will be used to produce the
	// container images. By default, it is ko.Publish, but can be overridden.
	CurrentImageProducer = ImageProducer(ko.Publish)
)

// ImageProducer is a function that will be used to produce the container images.
type ImageProducer func(ctx context.Context, pack string) (string, error)

// RegisterPackage registers an interest in producing an image based on the
// provided package.
// Can be called multiple times with the same package.
// A package will be used to produce the image and used
// like `image: ko://<package>` inside test yaml.
func RegisterPackage(pack ...string) EnvOpts {
	ri := func(ctx context.Context, env Environment) (context.Context, error) {
		rk := registeredPackagesKey{}
		rk.register(ctx, pack)
		store := rk.get(ctx)
		return context.WithValue(ctx, rk, store), nil
	}
	return UnionOpts(ri, WithImages(map[string]string{}))
}

// WithImages will bypass ProduceImages() and use the provided image set
// instead. Should be called before ProduceImages(), if used, likely in an
// init() method. An images value should be a container registry image. The
// images map is presented to the templates on the field `images`, and used
// like `image: <key>` inside test yaml.
func WithImages(given map[string]string) EnvOpts {
	return func(ctx context.Context, env Environment) (context.Context, error) {
		ik := imageStoreKey{}
		store := ik.get(ctx)
		store.withImages(given)
		return context.WithValue(ctx, ik, store), nil
	}
}

// ProduceImages returns back the packages that have been added.
// Will produce images once, can be called many times.
func ProduceImages(ctx context.Context) (map[string]string, error) {
	rk := registeredPackagesKey{}
	ik := imageStoreKey{}
	store := ik.get(ctx)
	for _, pack := range rk.packages(ctx) {
		koPack := fmt.Sprintf("ko://%s", pack)
		if store.refs[koPack] != "" {
			continue
		}
		image, err := CurrentImageProducer(ctx, pack)
		if err != nil {
			return nil, err
		}
		store.refs[koPack] = strings.TrimSpace(image)
	}
	return store.refs, nil
}

type registeredPackagesKey struct{}

type packagesStore struct {
	refs []string
}

func (k registeredPackagesKey) get(ctx context.Context) *packagesStore {
	if registered, ok := ctx.Value(k).(*packagesStore); ok {
		return registered
	}
	return &packagesStore{}
}

func (k registeredPackagesKey) packages(ctx context.Context) []string {
	return k.get(ctx).refs
}

func (k registeredPackagesKey) register(ctx context.Context, packs []string) {
	toRegister := make([]string, 0, len(packs))
	toRegister = append(toRegister, k.packages(ctx)...)
	for _, pack := range packs {
		pack = strings.TrimPrefix(pack, "ko://")
		if !k.contains(toRegister, pack) {
			toRegister = append(toRegister, pack)
		}
	}
	store := k.get(ctx)
	store.refs = toRegister
}

func (k registeredPackagesKey) contains(packs []string, pack string) bool {
	for _, i := range packs {
		if i == pack {
			return true
		}
	}
	return false
}

type imageStoreKey struct{}

func (k imageStoreKey) get(ctx context.Context) *imageStore {
	if i, ok := ctx.Value(k).(*imageStore); ok {
		return i
	}
	return &imageStore{
		refs: make(map[string]string),
	}
}

type imageStore struct {
	refs map[string]string
}

func (i *imageStore) withImages(given map[string]string) {
	if i.refs == nil {
		i.refs = make(map[string]string, len(given))
	}
	for k, v := range given {
		i.refs[k] = v
	}
}
