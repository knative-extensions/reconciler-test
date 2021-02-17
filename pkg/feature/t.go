package feature

type T interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	FailNow()

	Log(args ...interface{})
	Logf(format string, args ...interface{})

	Skip(args ...interface{})
	Skipf(format string, args ...interface{})
	SkipNow()
}
