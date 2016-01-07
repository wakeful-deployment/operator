package logger

func ExampleNotVerbose() {
	Log("FOO: ", "bar")
	// Output:
	//
}

func ExampleInfoLog() {
	Verbose = true
	defer func() {
		Verbose = false
	}()

	Info("foo")
	// Output:
	// INFO: foo
}

func ExampleErrorLog() {
	Verbose = true
	defer func() {
		Verbose = false
	}()

	Error("foo")
	// Output:
	// ERROR: foo
}
