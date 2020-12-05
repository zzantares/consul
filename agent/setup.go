package agent

import (
	"io"
	"net/http"

	autoconf "github.com/hashicorp/consul/agent/auto-config"
	"github.com/hashicorp/consul/agent/cache"
	"github.com/hashicorp/consul/agent/config"
	"github.com/hashicorp/consul/agent/consul"
)

// TODO: BaseDeps should be renamed in the future once more of Agent.Start
// has been moved out in front of Agent.New, and we can better see the setup
// dependencies.
type BaseDeps struct {
	consul.Deps // TODO: un-embed

	RuntimeConfig  *config.RuntimeConfig
	MetricsHandler MetricsHandler
	AutoConfig     *autoconf.AutoConfig // TODO: use an interface
	Cache          *cache.Cache
}

// MetricsHandler provides an http.Handler for displaying metrics.
type MetricsHandler interface {
	DisplayMetrics(resp http.ResponseWriter, req *http.Request) (interface{}, error)
}

type ConfigLoader func(source config.Source) (cfg *config.RuntimeConfig, warnings []string, err error)

func NewBaseDeps(configLoader ConfigLoader, logOut io.Writer) (BaseDeps, error) {
	d := BaseDeps{}
	return d, nil
}
