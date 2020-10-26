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

package example

import (
	"time"

	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/test/example/config/producer"
	"knative.dev/reconciler-test/test/example/config/recorder"
)

func RecorderFeature() *feature.Feature {
	to := feature.MakeRandomK8sName("recorder")
	count := 5

	f := new(feature.Feature)

	f.Precondition("install recorder", recorder.Install(to))
	f.Precondition("install producer", producer.Install(count, to))

	f.Alpha("direct sending between a producer and a recorder").
		Must("the recorder received all sent events within the time",
			recorder.AssertDelivery(to, count, 3*time.Second, 30*time.Second))

	return f
}
