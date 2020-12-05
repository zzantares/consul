package xds

import (
	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoyendpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/consul/agent/proxycfg"
	"github.com/hashicorp/consul/agent/structs"
)

const (
	UnnamedSubset = ""
)

// endpointsFromSnapshot returns the xDS API representation of the "endpoints"
func (s *Server) endpointsFromSnapshot(_ connectionInfo, cfgSnap *proxycfg.ConfigSnapshot) ([]proto.Message, error) {
	return nil, nil
}

// endpointsFromSnapshotConnectProxy returns the xDS API representation of the "endpoints"
// (upstream instances) in the snapshot.
func (s *Server) endpointsFromSnapshotConnectProxy(cfgSnap *proxycfg.ConfigSnapshot) ([]proto.Message, error) {
	return nil, nil
}

func (s *Server) filterSubsetEndpoints(subset *structs.ServiceResolverSubset, endpoints structs.CheckServiceNodes) (structs.CheckServiceNodes, error) {
	return nil, nil
}

func (s *Server) endpointsFromSnapshotTerminatingGateway(cfgSnap *proxycfg.ConfigSnapshot) ([]proto.Message, error) {
	return nil, nil
}

func (s *Server) endpointsFromSnapshotMeshGateway(cfgSnap *proxycfg.ConfigSnapshot) ([]proto.Message, error) {
	return nil, nil
}

func (s *Server) endpointsFromServicesAndResolvers(
	cfgSnap *proxycfg.ConfigSnapshot,
	services map[structs.ServiceName]structs.CheckServiceNodes,
	resolvers map[structs.ServiceName]*structs.ServiceResolverConfigEntry) ([]proto.Message, error) {

	return nil, nil
}

func (s *Server) endpointsFromSnapshotIngressGateway(cfgSnap *proxycfg.ConfigSnapshot) ([]proto.Message, error) {
	return nil, nil
}

func makeEndpoint(host string, port int) *envoyendpoint.LbEndpoint {
	return nil
}

func (s *Server) endpointsFromDiscoveryChain(
	upstream structs.Upstream,
	chain *structs.CompiledDiscoveryChain,
	datacenter string,
	upstreamEndpoints, gatewayEndpoints map[string]structs.CheckServiceNodes,
) []proto.Message {
	return nil
}

type loadAssignmentEndpointGroup struct {
	Endpoints      structs.CheckServiceNodes
	OnlyPassing    bool
	OverrideHealth envoycore.HealthStatus
}

func makeLoadAssignment(clusterName string, endpointGroups []loadAssignmentEndpointGroup, localDatacenter string) *envoy.ClusterLoadAssignment {
	return nil
}

func makeLoadAssignmentEndpointGroup(
	targets map[string]*structs.DiscoveryTarget,
	targetHealth map[string]structs.CheckServiceNodes,
	gatewayHealth map[string]structs.CheckServiceNodes,
	targetID string,
	currentDatacenter string,
) (loadAssignmentEndpointGroup, bool) {
	return loadAssignmentEndpointGroup{}, false
}

func calculateEndpointHealthAndWeight(
	ep structs.CheckServiceNode,
	onlyPassing bool,
) (envoycore.HealthStatus, int) {
	return 0, 0
}
