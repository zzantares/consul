package testutil

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/go-hclog"
)

func Logger(t TestingTB) hclog.InterceptLogger {
	return LoggerWithOutput(t, NewLogBuffer(t))
}

func LoggerWithOutput(t TestingTB, output io.Writer) hclog.InterceptLogger {
	return hclog.NewInterceptLogger(&hclog.LoggerOptions{
		Name:   t.Name(),
		Level:  hclog.Trace,
		Output: output,
	})
}

var (
	sendTestLogsToStdout  = os.Getenv("NOLOGBUFFER") == "1"
	sendTestLogsToDevNull = os.Getenv("NOAGENTLOGS") == "1"
	sendTestLogsToFile    = os.Getenv("AGENTLOGS_FILEPATH")
)

// NewLogBuffer returns an io.Writer which buffers all writes. When the test
// ends, t.Failed is checked. If the test has failed or has been run in verbose
// mode all log output is printed to stdout.
//
// Set the env var NOLOGBUFFER=1 to disable buffering, resulting in all log
// output being written immediately to stdout.
func NewLogBuffer(t TestingTB) io.Writer {
	switch {
	case sendTestLogsToStdout:
		return os.Stdout
	case sendTestLogsToDevNull:
		return ioutil.Discard
	case sendTestLogsToFile != "":
		fh, err := os.Create(sendTestLogsToFile)
		if err != nil {
			t.Fatalf("failed to open file for logs: %v", err)
		}
		t.Cleanup(func() { fh.Close() })
		return fh
	}

	buf := &logBuffer{buf: new(bytes.Buffer)}
	t.Cleanup(func() {
		if t.Failed() || testing.Verbose() {
			buf.Lock()
			defer buf.Unlock()
			buf.buf.WriteTo(os.Stdout)
		}
	})
	return buf
}

type logBuffer struct {
	buf *bytes.Buffer
	sync.Mutex
}

func (lb *logBuffer) Write(p []byte) (n int, err error) {
	lb.Lock()
	defer lb.Unlock()
	return lb.buf.Write(p)
}
