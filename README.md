# Knative Reconciler Test

This repo contains tools and frameworks to help validate and test reconciler
implementations.

## Knative End-to-End and Conformance Testing Framework

We have developed a lightweight testing framework to compose and share
Kubernetes cluster based tests for Knative. These tests are intended to be
vendor-able by other downstream components, which is useful for Conformance
tests and e2e testing for behaviors.

The testing framework is broken into two main components: Features: the small
composable steps intended to validate a function or feature, and the Test
Environment: which is highly dependent on the cluster, and the config that was
passed. The two components are intended to be fairly independent.

A feature is a contained test with phases. A feature could be anything from a
set of assertions, to waiting for or validating a condition, installing test
dependencies, or interacting with an API. One or more features are tested in an
Environment. One or more Environments are produced for test run.

### Getting Started

We will compose our test integration into two parts: 1) test entry points,
and 2) features.

#### Test Entry Points

Test entry points are the methods go sees when running `go test ./...` This
includes `TestMain` and every function with a signature of
`func Test<Name>(t *testing.T)`. Because we do not want to run these integration
style tests when running unit tests, we will tag the entry point files with
`// +build e2e`

TestMain is where the GlobalEnvironment is created. This global variable will be
used for the rest of the test run, it is a singleton.

[main_test.go](./test/example/main_test.go)

```go
// +build e2e

import (
	"knative.dev/pkg/injection"
	"knative.dev/reconciler-test/pkg/environment"
)

// global is the singleton instance of GlobalEnvironment. It is used to parse
// the testing config for the test run. The config will specify the cluster
// config as well as the parsing level and state flags.
var global environment.GlobalEnvironment

func init() {
	// environment.InitFlags registers state and level filter flags.
	environment.InitFlags(flag.CommandLine)
}

// TestMain is the first entry point for `go test`.
func TestMain(m *testing.M) {
	// We get a chance to parse flags to include the framework flags for the
	// framework as well as any additional flags included in the integration.
	flag.Parse()

	// EnableInjectionOrDie will enable client injection, this is used by the
	// testing framework for namespace management, and could be leveraged by
	// features to pull Kubernetes clients or the test environment out of the
	// context passed in the features.
	ctx, startInformers := injection.EnableInjectionOrDie(nil, nil) //nolint
	startInformers()

	// global is used to make instances of Environments, NewGlobalEnvironment
	// is passing and saving the client injection enabled context for use later.
	global = environment.NewGlobalEnvironment(ctx)

	// Run the tests.
	os.Exit(m.Run())
}
```

From the instance of GlobalEnvironment, we will test features on an instance of
the environment.

```go
// +build e2e

// TestFoo is an example simple test.
func TestFoo(t *testing.T) {
	// Signal to the go test framework that this test can be run in parallel
	// with other tests.
	t.Parallel()

	// Create an instance of an environment. The environment will be configured
	// with any relevant configuration and settings based on the global
	// environment settings. Using environment.Managed(t) will call env.Finish()
	// on test completion. If this option is not used, the test should call
	// env.Finish() to perform cleanup at the end of the test. Additional options
	// can be passed to Environment() if customization is required.
	ctx, env := global.Environment(environment.Managed(t) /*, optional environment options */)

	// With the instance of an Environment, perform one or more calls to Test().
	// Note: env.Test() is blocking until the feature completes.
	env.Test(ctx, t, FooFeature1())
	env.Test(ctx, t, FooFeature2())
}
```

The role of the `Test<Name>` methods is to control which features are tested on
environment instances. It is your responsibility to understand if it is safe to
run multiple features in an environment instance. It is reasonable to pass
additional configurations to the feature constructor, unless it is data that
should be pulled from the instance of `env`, which will be talked about in the
next [Features](#features) section.

Test Entry point files should be named "<name>\_test.go" and should be tagged
`e2e` or some other meaningful tag that will prevent them from being run on
un-excluded `go test ./...` invocations.

#### Features

Features are a series of steps that perform actions or validations. Each step is
similar to a unit test. It has a scoped objective, and if written with care, a
step can be composed and shared in downstream features.

Features have several phases that can be registered, each phase is executed in
order, but there are no order guarantees in phases that are equal. If struct
ordering is required it is recommended to break that into an independent feature
that is tested on an environment in order required.

Features have 4 phases (timing) on which steps can be composed: Setup,
Requirement, Assert, and Teardown. The step functions are run in that order.

- Setup is used to install required components or configuration in the
  environment.
- Requirement is used to validate the cluster, environment, or anything else.
  Think of this as a preflight validation for the assertions.
- Assert should assume the namespace is ready to perform or validate the test.
- Teardown should be used to do final feature cleanup, if needed. There is also
  automatic cleanup of resources and namespace for the environment.

Asserts have two additional properties that allow for filtering from within
environment.Test, State and Level.

State represents how mature the requirement is for the feature, we support
Alpha, Beta, and Stable.

Level relates to the spec language this test represents, we support Must,
MustNot, Should, ShouldNot, and May.

States and Levels are used to filter feature steps based on test parameters.

Features should be in a file with the naming pattern "<name>\_feature.go", with
no build tag on this file.

##### Composing Features

A `feature.Feature` is implemented as a builder pattern. Start with a new
`feature.Feature`:

```go
f := &feature.Feature{Name: "Meaningful Name"}
```

Then, add steps for each timing as required for the test:

```go
f.Requirement("a cluster requirement is tested", HasClusterRequirement())

f.Setup("install a dependency", InstallADependency(opts))
f.Requirement("a dependency went ready", ADependencyIsReady())

f.Alpha("some experimental feature name").
	Must("does a thing", AssertThing(opts)).
    May("could do another thing", AssertAnotherThing)

f.Beta("pretty sure this is a good feature name").
	Must("does another thing", AssertAnotherThing(opts)).
		Should("please do thing", AssertPleaseDoThing(1, 2, 3))

f.Teardown("remove a dependency", DeleteADependency(opts))
```

The step functions all have the same function signature,

```go
func AssertSomething(ctx context.Context, t *testing.T) {
	// TODO: some assert.
}
```

Step functions could return a `feature.StepFn`, allowing options or
configuration to be passed to the StepFn. For example, `AssertDelivery`:

```go
func AssertDelivery(to string, count int, interval, timeout time.Duration) feature.StepFn {
	return func(ctx context.Context, t *testing.T) {
		// TODO: some assert.
	}
}
```

The context passed to a StepFn is client injection enabled, and decorated with
environment context (and extendable based on passing `opts` callbacks in
`global.Environment(opts EnvOpts...)`).

In addition to the normal client injection `typedclient.Get(ctx)` methods, there
is

```go
env := environment.FromContext(ctx)
```

This returns the [`Environment`](./pkg/environment/interfaces.go) the feature is
being tested in.

#### Feature Sets

Sometimes it makes sense to be able to provide a set of Features all at once. We
call these "Feature Sets". They are intended to allow upstreams to bundle sets
of features together to provide a simple low code way to vender and run a
collection of Features together. `Test` and `TestSet` can be used together on an
environment.

```go
ctx, env := global.Environment(environment.Managed(t) /*, optional environment options */)

// With the instance of an Environment, perform one or more calls to Test().
env.Test(ctx, t, FooFeature1())
// Note: env.TestSet() is blocking until all features complete.
env.TestSet(ctx, t, FooFeatureSet1())
```

##### Composing Feature Sets

A `feature.FeatureSet` is implemented as a simple wrapper for a list of
feature.Features, with a name for context.

```go
fs := &feature.FeatureSet{
    Name:     "Some higher order idea",
    Features: []feature.Feature{
        *OneAspectFeature(),
        *AnotherAspectFeature(),
        *OptionalAspectFeature(),
    },
}
```

##### YAML Setup Helpers

The testing framework also enables YAML based installing of components. This
framework can help:

1. produce go-based test images via `ko publish`.
1. discover ko-based packages in local YAML file.
1. apply local YAML files with some light go templating to the test environment.

To produce images, register an interest in one or more packages with the
`environment` package in an `init` method.

```go
import "knative.dev/reconciler-test/pkg/environment"

func init() {
	environment.RegisterPackage("example.com/some/local/package", "example.com/another/package")
}
```

This can be discovered dynamically by the helper function to scan local YAML
files for `ko://` images, `manifest.ImagesLocalYaml()`

```go
import (
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/manifest"
)

func init() {
	environment.RegisterPackage(manifest.ImagesLocalYaml()...)
}
```

Images registred with the environment package will be produced on the first call
to `environment.ProduceImages()`, which happens as a byproduct of calling
`global.Environment()`. These images replace the `ko://` tags in YAML files that
are applied the the cluster with `manifest.InstallLocalYaml`.

A go file local to the target YAML can have an install step function like:

```go
func Install(message string) feature.StepFn {
	cfg := map[string]interface{}{
		"additional": "this",
		"customizations": message,
	}
	return func(ctx context.Context, t *testing.T) {
		if _, err := manifest.InstallLocalYaml(ctx, cfg); err != nil {
			t.Error(err)
		}
	}
}
```

Then given a YAML file:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: echo
  namespace: { { .namespace } }
  labels:
    additional: { { .additional } }
spec:
  backoffLimit: 0
  parallelism: 1
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: echo
          image: ko://knative.dev/reconciler-test/test/example/cmd/echo
          env:
            - name: ECHO
              value: "{{ .customizations }}"
```

Will be processed first as a go template, then applied the Environment using the
dynamic client.

### Running Tests

Running tests is nothing more than using `go test`.

```shell
go test -v -count=1 -timeout=15m -tags=e2e ./test/...
```

And normal go test filters work on test entry point names:

```shell
go test -v -count=1 -tags=e2e ./test/... -run TestKoPublish
```

#### Filters

At the moment, all features and all requirements are defaulted on. These can be
disabled using the following flags:

| Flag                    | Type    | Meaning                                           |
| ----------------------- | ------- | ------------------------------------------------- |
| --feature.alpha         | Boolean | Enable/Disable running Alpha state features.      |
| --feature.beta          | Boolean | Enable/Disable running Beta state features.       |
| --feature.stable        | Boolean | Enable/Disable running Stable state features.     |
| --feature.any           | Boolean | Enable/Disable running Any state features.        |
| --requirement.must      | Boolean | Enable/Disable running Must level features.       |
| --requirement.mustnot   | Boolean | Enable/Disable running Must Not level features.   |
| --requirement.should    | Boolean | Enable/Disable running Should level features.     |
| --requirement.shouldnot | Boolean | Enable/Disable running Should Not level features. |
| --requirement.may       | Boolean | Enable/Disable running May level features.        |
| --requirement.all       | Boolean | Enable/Disable running All level features.        |

They can be used in combination, to run only Beta state features:

```shell
go test -v -count=1 -tags=e2e ./test/... --feature.any=false --feature.beta
```

Or, only alpha and beta state features for only Must and Must Not requirements:

```shell
go test -v -count=1 -tags=e2e ./test/... --feature.any=false --feature.beta --requirement.all=false --requirement.must --requirement.mustnot
```

And normal go filters work on the go test entry point names.

```shell
go test -v -count=1 -tags=e2e ./test/... --feature.any=false --feature.beta -run TestKoPublish
```
