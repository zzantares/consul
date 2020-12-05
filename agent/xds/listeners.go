package xds

import (
	envoyhttp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
)

type listenerFilterOpts struct {
	useRDS     bool
	protocol   string
	filterName string
	routeName  string
	cluster    string
	statPrefix string
	routePath  string
	ingress    bool

	// TODO(m1) removing this line eliminates the problem
	httpAuthzFilter *envoyhttp.HttpFilter
}
