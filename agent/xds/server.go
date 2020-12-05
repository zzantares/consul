package xds

import (
	"context"
	"time"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoydisco "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/agent/proxycfg"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/consul/tlsutil"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
)

// ADSStream is a shorter way of referring to this thing...
type ADSStream = envoydisco.AggregatedDiscoveryService_StreamAggregatedResourcesServer

const (
	// Resource types in xDS v2. These are copied from
	// envoyproxy/go-control-plane/pkg/cache/resource.go since we don't need any of
	// the rest of that package.
	typePrefix = "type.googleapis.com/envoy.api.v2."

	// EndpointType is the TypeURL for Endpoint discovery responses.
	EndpointType = typePrefix + "ClusterLoadAssignment"

	// ClusterType is the TypeURL for Cluster discovery responses.
	ClusterType = typePrefix + "Cluster"

	// RouteType is the TypeURL for Route discovery responses.
	RouteType = typePrefix + "RouteConfiguration"

	// ListenerType is the TypeURL for Listener discovery responses.
	ListenerType = typePrefix + "Listener"

	// PublicListenerName is the name we give the public listener in Envoy config.
	PublicListenerName = "public_listener"

	// LocalAppClusterName is the name we give the local application "cluster" in
	// Envoy config. Note that all cluster names may collide with service names
	// since we want cluster names and service names to match to enable nice
	// metrics correlation without massaging prefixes on cluster names.
	//
	// We should probably make this more unlikely to collide however changing it
	// potentially breaks upgrade compatibility without restarting all Envoy's as
	// it will no longer match their existing cluster name. Changing this will
	// affect metrics output so could break dashboards (for local app traffic).
	//
	// We should probably just make it configurable if anyone actually has
	// services named "local_app" in the future.
	LocalAppClusterName = "local_app"

	// LocalAgentClusterName is the name we give the local agent "cluster" in
	// Envoy config. Note that all cluster names may collide with service names
	// since we want cluster names and service names to match to enable nice
	// metrics correlation without massaging prefixes on cluster names.
	//
	// We should probably make this more unlikely to collied however changing it
	// potentially breaks upgrade compatibility without restarting all Envoy's as
	// it will no longer match their existing cluster name. Changing this will
	// affect metrics output so could break dashboards (for local agent traffic).
	//
	// We should probably just make it configurable if anyone actually has
	// services named "local_agent" in the future.
	LocalAgentClusterName = "local_agent"

	// DefaultAuthCheckFrequency is the default value for
	// Server.AuthCheckFrequency to use when the zero value is provided.
	DefaultAuthCheckFrequency = 5 * time.Minute
)

// ACLResolverFunc is a shim to resolve ACLs. Since ACL enforcement is so far
// entirely agent-local and all uses private methods this allows a simple shim
// to be written in the agent package to allow resolving without tightly
// coupling this to the agent.
type ACLResolverFunc func(id string) (acl.Authorizer, error)

// ServiceChecks is the interface the agent needs to expose
// for the xDS server to fetch a service's HTTP check definitions
type HTTPCheckFetcher interface {
	ServiceHTTPBasedChecks(serviceID structs.ServiceID) []structs.CheckType
}

// ConfigFetcher is the interface the agent needs to expose
// for the xDS server to fetch agent config, currently only one field is fetched
type ConfigFetcher interface {
	AdvertiseAddrLAN() string
}

// ConfigManager is the interface xds.Server requires to consume proxy config
// updates. It's satisfied normally by the agent's proxycfg.Manager, but allows
// easier testing without several layers of mocked cache, local state and
// proxycfg.Manager.
type ConfigManager interface {
	Watch(proxyID structs.ServiceID) (<-chan *proxycfg.ConfigSnapshot, proxycfg.CancelFunc)
}

// Server represents a gRPC server that can handle xDS requests from Envoy. All
// of it's public members must be set before the gRPC server is started.
//
// A full description of the XDS protocol can be found at
// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol
type Server struct {
	Logger       hclog.Logger
	CfgMgr       ConfigManager
	ResolveToken ACLResolverFunc
	// AuthCheckFrequency is how often we should re-check the credentials used
	// during a long-lived gRPC Stream after it has been initially established.
	// This is only used during idle periods of stream interactions (i.e. when
	// there has been no recent DiscoveryRequest).
	AuthCheckFrequency time.Duration
	CheckFetcher       HTTPCheckFetcher
	CfgFetcher         ConfigFetcher
}

// Initialize will finish configuring the Server for first use.
func (s *Server) Initialize() {

}

// StreamAggregatedResources implements
// envoydisco.AggregatedDiscoveryServiceServer. This is the ADS endpoint which is
// the only xDS API we directly support for now.
func (s *Server) StreamAggregatedResources(stream ADSStream) error {
	return nil
}

const (
	stateInit int = iota
	statePendingInitialConfig
	stateRunning
)

func (s *Server) process(stream ADSStream, reqCh <-chan *envoy.DiscoveryRequest) error {
	return nil
}

type xDSType struct {
	typeURL string
	stream  ADSStream
	req     *envoy.DiscoveryRequest
	node    *envoycore.Node

	lastNonce string
	// lastVersion is the version that was last sent to the proxy. It is needed
	// because we don't want to send the same version more than once.
	// req.VersionInfo may be an older version than the most recent once sent in
	// two cases: 1) if the ACK wasn't received yet and `req` still points to the
	// previous request we already responded to and 2) if the proxy rejected the
	// last version we sent with a Nack then req.VersionInfo will be the older
	// version it's hanging on to.
	lastVersion  uint64
	resources    func(cInfo connectionInfo, cfgSnap *proxycfg.ConfigSnapshot) ([]proto.Message, error)
	allowEmptyFn func(cfgSnap *proxycfg.ConfigSnapshot) bool
}

// connectionInfo represents details specific to this connection
type connectionInfo struct {
	Token string
}

func (t *xDSType) Recv(req *envoy.DiscoveryRequest, node *envoycore.Node) {
}

func (t *xDSType) SendIfNew(cfgSnap *proxycfg.ConfigSnapshot, version uint64, nonce *uint64) error {
	return nil
}

func tokenFromContext(ctx context.Context) string {
	return ""
}

// DeltaAggregatedResources implements envoydisco.AggregatedDiscoveryServiceServer
func (s *Server) DeltaAggregatedResources(_ envoydisco.AggregatedDiscoveryService_DeltaAggregatedResourcesServer) error {
	return nil
}

// GRPCServer returns a server instance that can handle xDS requests.
func (s *Server) GRPCServer(tlsConfigurator *tlsutil.Configurator) (*grpc.Server, error) {
	return nil, nil
}
