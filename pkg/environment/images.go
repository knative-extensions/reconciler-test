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
	"strings"
	"sync"

	"knative.dev/pkg/logging"
	"knative.dev/reconciler-test/pkg/images/ko"
)

var packages = []string(nil)
var images = make(map[string]string)

// RegisterPackage registers an interest in producing an image based on the
// provide package.
// Can be called multiple times with the same package.
// A package will be used to produce the image and used
// like `image: ko://<package>` inside test yaml.
func RegisterPackage(pack ...string) {
	for _, p := range pack {
		exists := false
		for _, k := range packages {
			if p == k {
				exists = true
				break
			}
		}
		if !exists {
			packages = append(packages, p)
		}
	}
}

// WithImages will bypass ProduceImages() and use the provided image set
// instead. Should be called before ProduceImages(), if used, likely in an
// init() method. An images value should be a container registry image. The
// images map is presented to the templates on the field `images`, and used
// like `image: <key>` inside test yaml.
func WithImages(given map[string]string) {
	images = given
}

// ProduceImages returns back the packages that have been added.
// Will produce images once, can be called many times.
func ProduceImages(ctx context.Context) (map[string]string, error) {
	for _, pack := range packages {
		image, err := ko.Publish(ctx, pack)
		if err != nil {
			log := logging.FromContext(ctx)
			log.Error("error attempting to ko publish: ", err)
			return nil, err
		}
		images["ko://"+pack] = strings.TrimSpace(image)
	}
	return images, nil
}

type imageStore struct {
	refs map[string]string
	err  error
	once sync.Once
}

func (i *imageStore) Get(ctx context.Context) (map[string]string, error) {
	i.once.Do(func() {
		i.refs, i.err = ProduceImages(ctx)
	})
	return i.refs, i.err
}
