package recorder

import (
	"flag"
	"os"

	_ "knative.dev/pkg/system/testing"

	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/reconciler-test/pkg/observer"
	recorder_vent "knative.dev/reconciler-test/pkg/observer/recorder-vent"
	writer_vent "knative.dev/reconciler-test/pkg/observer/writer-vent"
)

func main() {
	flag.Parse()
	ctx := sharedmain.EnableInjectionOrDie(nil, nil) //nolint

	obs := observer.New(
		writer_vent.NewEventLog(ctx, os.Stdout),
		recorder_vent.NewFromEnv(ctx),
	)

	if err := obs.Start(ctx); err != nil {
		panic(err)
	}
}
