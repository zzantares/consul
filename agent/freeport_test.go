package agent

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/consul/sdk/freeport"
)

func TestMain(m *testing.M) {
	code := m.Run()
	fmt.Println("freeport elapsed time", time.Duration(atomic.LoadInt64(&freeport.Elapsed)))
	os.Exit(code)
}
