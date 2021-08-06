package maint

import (
	"fmt"
	"os"
)

// MainT implements agent.TestingT so that agent.TestAgent can be used from TestMain
type MainT struct {
	cleanup []func()
	failed  bool
}

func (m *MainT) Cleanup(f func()) {
	m.cleanup = append(m.cleanup, f)
}

func (m *MainT) RunCleanup() {
	for _, fn := range m.cleanup {
		// defer all the funcs so that if one panics the rest continue to run
		// and so that they execute in the correct order.
		defer fn()
	}
}

func (m *MainT) Failed() bool {
	return m.failed
}

func (m *MainT) Fatalf(format string, args ...interface{}) {
	m.Logf(format, args...)
	m.FailNow()
}

func (m *MainT) Logf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (m *MainT) Log(args ...interface{}) {
	fmt.Print(args...)
}

func (m *MainT) Name() string {
	return "main"
}

func (m *MainT) FailNow() {
	os.Exit(1)
}

func (m *MainT) Helper() {}
