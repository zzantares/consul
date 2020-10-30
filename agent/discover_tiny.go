// +build tiny

package agent

import (
	"github.com/hashicorp/go-hclog"
)

type disco struct {
}

func (disco) Names() []string {
	return nil
}

func newDiscover() (disco, error) {
	return disco{}, nil
}

func retryJoinAddrs(disco disco, variant, cluster string, retryJoin []string, logger hclog.Logger) []string {
	return retryJoin
}
