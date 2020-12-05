package xds

import (
	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/consul/agent/proxycfg"
	"github.com/hashicorp/consul/agent/structs"
)

// routesFromSnapshot returns the xDS API representation of the "routes" in the
// snapshot.
func (s *Server) routesFromSnapshot(cInfo connectionInfo, cfgSnap *proxycfg.ConfigSnapshot) ([]proto.Message, error) {
	return nil, nil
}

// routesFromSnapshotConnectProxy returns the xDS API representation of the
// "routes" in the snapshot.
func routesForConnectProxy(
	cInfo connectionInfo,
	upstreams structs.Upstreams,
	chains map[string]*structs.CompiledDiscoveryChain,
) ([]proto.Message, error) {
	return nil, nil
}

// routesFromSnapshotTerminatingGateway returns the xDS API representation of the "routes" in the snapshot.
// For any HTTP service we will return a default route.
func (s *Server) routesFromSnapshotTerminatingGateway(_ connectionInfo, cfgSnap *proxycfg.ConfigSnapshot) ([]proto.Message, error) {
	return nil, nil
}

func makeNamedDefaultRouteWithLB(clusterName string, lb *structs.LoadBalancer) (*envoy.RouteConfiguration, error) {
	return nil, nil
}

// routesForIngressGateway returns the xDS API representation of the
// "routes" in the snapshot.
func routesForIngressGateway(
	cInfo connectionInfo,
	upstreams map[proxycfg.IngressListenerKey]structs.Upstreams,
	chains map[string]*structs.CompiledDiscoveryChain,
) ([]proto.Message, error) {
	return nil, nil
}

func generateUpstreamIngressDomains(listenerKey proxycfg.IngressListenerKey, u structs.Upstream) []string {
	return nil
}

func makeUpstreamRouteForDiscoveryChain(
	cInfo connectionInfo,
	routeName string,
	chain *structs.CompiledDiscoveryChain,
	serviceDomains []string,
) (*envoyroute.VirtualHost, error) {
	return nil, nil
}

func makeRouteMatchForDiscoveryRoute(_ connectionInfo, discoveryRoute *structs.DiscoveryRoute) *envoyroute.RouteMatch {
	return nil
}

func makeDefaultRouteMatch() *envoyroute.RouteMatch {
	return nil
}

func makeRouteActionForChainCluster(targetID string, chain *structs.CompiledDiscoveryChain) *envoyroute.Route_Route {
	return nil
}

func makeRouteActionFromName(clusterName string) *envoyroute.Route_Route {
	return nil
}

func makeRouteActionForSplitter(splits []*structs.DiscoverySplit, chain *structs.CompiledDiscoveryChain) (*envoyroute.Route_Route, error) {
	return nil, nil
}

func injectLBToRouteAction(lb *structs.LoadBalancer, action *envoyroute.RouteAction) error {
	return nil
}
