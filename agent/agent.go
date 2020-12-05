package agent

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/consul/agent/xds"
	"github.com/hashicorp/go-connlimit"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/serf/serf"
	"google.golang.org/grpc"

	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/agent/ae"
	"github.com/hashicorp/consul/agent/cache"
	"github.com/hashicorp/consul/agent/checks"
	"github.com/hashicorp/consul/agent/config"
	"github.com/hashicorp/consul/agent/consul"
	"github.com/hashicorp/consul/agent/local"
	"github.com/hashicorp/consul/agent/proxycfg"
	"github.com/hashicorp/consul/agent/rpcclient/health"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/consul/agent/token"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/consul/lib"
	"github.com/hashicorp/consul/tlsutil"
)

const (

	// maxQueryTime is used to bound the limit of a blocking query
	maxQueryTime = 600 * time.Second
)

type configSource int

const (
	ConfigSourceLocal configSource = iota
	ConfigSourceRemote
)

var configSourceToName = map[configSource]string{
	ConfigSourceLocal:  "local",
	ConfigSourceRemote: "remote",
}
var configSourceFromName = map[string]configSource{
	"local":  ConfigSourceLocal,
	"remote": ConfigSourceRemote,
	// If the value is not found in the persisted config file, then use the
	// former default.
	"": ConfigSourceLocal,
}

func (s configSource) String() string {
	return configSourceToName[s]
}

// ConfigSourceFromName will unmarshal the string form of a configSource.
func ConfigSourceFromName(name string) (configSource, bool) {
	s, ok := configSourceFromName[name]
	return s, ok
}

// delegate defines the interface shared by both
// consul.Client and consul.Server.
type delegate interface {
	GetLANCoordinate() (lib.CoordinateSet, error)
	Leave() error
	LANMembers() []serf.Member
	LANMembersAllSegments() ([]serf.Member, error)
	LANSegmentMembers(segment string) ([]serf.Member, error)
	LocalMember() serf.Member
	JoinLAN(addrs []string) (n int, err error)
	RemoveFailedNode(node string, prune bool) error
	ResolveToken(secretID string) (acl.Authorizer, error)
	ResolveTokenToIdentity(secretID string) (structs.ACLIdentity, error)
	ResolveTokenAndDefaultMeta(secretID string, entMeta *structs.EnterpriseMeta, authzContext *acl.AuthorizerContext) (acl.Authorizer, error)
	RPC(method string, args interface{}, reply interface{}) error
	UseLegacyACLs() bool
	SnapshotRPC(args *structs.SnapshotRequest, in io.Reader, out io.Writer, replyFn structs.SnapshotReplyFn) error
	Shutdown() error
	Stats() map[string]map[string]string
	ReloadConfig(config *consul.Config) error
	enterpriseDelegate
}

// notifier is called after a successful JoinLAN.
type notifier interface {
	Notify(string) error
}

// Agent is the long running process that is run on every machine.
// It exposes an RPC interface that is used by the CLI to control the
// agent. The agent runs the query interfaces like HTTP, DNS, and RPC.
// However, it can run in either a client, or server mode. In server
// mode, it runs a full Consul server. In client-only mode, it only forwards
// requests to other Consul servers.
type Agent struct {
	// TODO: remove fields that are already in BaseDeps
	baseDeps BaseDeps

	// config is the agent configuration.
	config *config.RuntimeConfig

	// Used for writing our logs
	logger hclog.InterceptLogger

	// delegate is either a *consul.Server or *consul.Client
	// depending on the configuration
	delegate delegate

	// aclMasterAuthorizer is an object that helps manage local ACL enforcement.
	aclMasterAuthorizer acl.Authorizer

	// state stores a local representation of the node,
	// services and checks. Used for anti-entropy.
	State *local.State

	// sync manages the synchronization of the local
	// and the remote state.
	sync *ae.StateSyncer

	// syncMu and syncCh are used to coordinate agent endpoints that are blocking
	// on local state during a config reload.
	syncMu sync.Mutex
	syncCh chan struct{}

	// cache is the in-memory cache for data the Agent requests.
	cache *cache.Cache

	// checkReapAfter maps the check ID to a timeout after which we should
	// reap its associated service
	checkReapAfter map[structs.CheckID]time.Duration

	// checkMonitors maps the check ID to an associated monitor
	checkMonitors map[structs.CheckID]*checks.CheckMonitor

	// checkHTTPs maps the check ID to an associated HTTP check
	checkHTTPs map[structs.CheckID]*checks.CheckHTTP

	// checkTCPs maps the check ID to an associated TCP check
	checkTCPs map[structs.CheckID]*checks.CheckTCP

	// checkGRPCs maps the check ID to an associated GRPC check
	checkGRPCs map[structs.CheckID]*checks.CheckGRPC

	// checkTTLs maps the check ID to an associated check TTL
	checkTTLs map[structs.CheckID]*checks.CheckTTL

	// checkDockers maps the check ID to an associated Docker Exec based check
	checkDockers map[structs.CheckID]*checks.CheckDocker

	// checkAliases maps the check ID to an associated Alias checks
	checkAliases map[structs.CheckID]*checks.CheckAlias

	// exposedPorts tracks listener ports for checks exposed through a proxy
	exposedPorts map[string]int

	// stateLock protects the agent state
	stateLock sync.Mutex

	// dockerClient is the client for performing docker health checks.
	dockerClient *checks.DockerClient

	// eventCh is used to receive user events
	eventCh chan serf.UserEvent

	// eventBuf stores the most recent events in a ring buffer
	// using eventIndex as the next index to insert into. This
	// is guarded by eventLock. When an insert happens, the
	// eventNotify group is notified.
	eventBuf    []*UserEvent
	eventIndex  int
	eventLock   sync.RWMutex
	eventNotify NotifyGroup

	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex

	// joinLANNotifier is called after a successful JoinLAN.
	joinLANNotifier notifier

	// retryJoinCh transports errors from the retry join
	// attempts.
	retryJoinCh chan error

	// endpoints maps unique RPC endpoint names to common ones
	// to allow overriding of RPC handlers since the golang
	// net/rpc server does not allow this.
	endpoints     map[string]string
	endpointsLock sync.RWMutex

	// dnsServer provides the DNS API
	dnsServers []*DNSServer

	// apiServers listening for connections. If any of these server goroutines
	// fail, the agent will be shutdown.
	apiServers *apiServers

	// httpHandlers provides direct access to (one of) the HTTPHandlers started by
	// this agent. This is used in tests to test HTTP endpoints without overhead
	// of TCP connections etc.
	//
	// TODO: this is a temporary re-introduction after we removed a list of
	// HTTPServers in favour of apiServers abstraction. Now that HTTPHandlers is
	// stateful and has config reloading though it's not OK to just use a
	// different instance of handlers in tests to the ones that the agent is wired
	// up to since then config reloads won't actually affect the handlers under
	// test while plumbing the external handlers in the TestAgent through bypasses
	// testing that the agent itself is actually reloading the state correctly.
	// Once we move `apiServers` to be a passed-in dependency for NewAgent, we
	// should be able to remove this and have the Test Agent create the
	// HTTPHandlers and pass them in removing the need to pull them back out
	// again.
	httpHandlers *HTTPHandlers

	// wgServers is the wait group for all HTTP and DNS servers
	// TODO: remove once dnsServers are handled by apiServers
	wgServers sync.WaitGroup

	// watchPlans tracks all the currently-running watch plans for the
	// agent.
	watchPlans []*watch.Plan

	// tokens holds ACL tokens initially from the configuration, but can
	// be updated at runtime, so should always be used instead of going to
	// the configuration directly.
	tokens *token.Store

	// proxyConfig is the manager for proxy service (Kind = connect-proxy)
	// configuration state. This ensures all state needed by a proxy registration
	// is maintained in cache and handles pushing updates to that state into XDS
	// server to be pushed out to Envoy.
	proxyConfig *proxycfg.Manager

	// serviceManager is the manager for combining local service registrations with
	// the centrally configured proxy/service defaults.
	serviceManager *ServiceManager

	// grpcServer is the server instance used currently to serve xDS API for
	// Envoy.
	grpcServer *grpc.Server

	// tlsConfigurator is the central instance to provide a *tls.Config
	// based on the current consul configuration.
	tlsConfigurator *tlsutil.Configurator

	// httpConnLimiter is used to limit connections to the HTTP server by client
	// IP.
	httpConnLimiter connlimit.Limiter

	// configReloaders are subcomponents that need to be notified on a reload so
	// they can update their internal state.
	configReloaders []ConfigReloader

	// TODO: pass directly to HTTPHandlers and DNSServer once those are passed
	// into Agent, which will allow us to remove this field.
	rpcClientHealth *health.Client

	// enterpriseAgent embeds fields that we only access in consul-enterprise builds
	enterpriseAgent
}

// New process the desired options and creates a new Agent.
// This process will
//   * parse the config given the config Flags
//   * setup logging
//      * using predefined logger given in an option
//        OR
//      * initialize a new logger from the configuration
//        including setting up gRPC logging
//   * initialize telemetry
//   * create a TLS Configurator
//   * build a shared connection pool
//   * create the ServiceManager
//   * setup the NodeID if one isn't provided in the configuration
//   * create the AutoConfig object for future use in fully
//     resolving the configuration
func New(bd BaseDeps) (*Agent, error) {
	return nil, nil
}

// GetConfig retrieves the agents config
// TODO make export the config field and get rid of this method
// This is here for now to simplify the work I am doing and make
// reviewing the final PR easier.
func (a *Agent) GetConfig() *config.RuntimeConfig {
	return nil
}

// LocalConfig takes a config.RuntimeConfig and maps the fields to a local.Config
func LocalConfig(cfg *config.RuntimeConfig) local.Config {
	lc := local.Config{}
	return lc
}

// Start verifies its configuration and runs an agent's various subprocesses.
func (a *Agent) Start(ctx context.Context) error {
	return nil
}

// Failed returns a channel which is closed when the first server goroutine exits
// with a non-nil error.
func (a *Agent) Failed() <-chan struct{} {
	return nil
}

// TODO(m1): removing this eliminates the huge block panic
func (a *Agent) listenAndServeGRPC() error {
	if len(a.config.GRPCAddrs) < 1 {
		return nil
	}

	xdsServer := &xds.Server{
		Logger:       a.logger,
		CfgMgr:       a.proxyConfig,
		ResolveToken: a.resolveToken,
		CheckFetcher: nil,
		CfgFetcher:   nil,
	}
	xdsServer.Initialize()

	var err error
	if a.config.HTTPSPort > 0 {
		// gRPC uses the same TLS settings as the HTTPS API. If HTTPS is
		// enabled then gRPC will require HTTPS as well.
		a.grpcServer, err = xdsServer.GRPCServer(a.tlsConfigurator)
	} else {
		a.grpcServer, err = xdsServer.GRPCServer(nil)
	}
	if err != nil {
		return err
	}

	ln, err := a.startListeners(a.config.GRPCAddrs)
	if err != nil {
		return err
	}

	for _, l := range ln {
		go func(innerL net.Listener) {
			a.logger.Info("Started gRPC server",
				"address", innerL.Addr().String(),
				"network", innerL.Addr().Network(),
			)
			err := a.grpcServer.Serve(innerL)
			if err != nil {
				a.logger.Error("gRPC server failed", "error", err)
			}
		}(l)
	}
	return nil
}

func (a *Agent) listenAndServeDNS() error {
	return nil
}

func (a *Agent) startListeners(addrs []net.Addr) ([]net.Listener, error) {
	return nil, nil
}

// listenHTTP binds listeners to the provided addresses and also returns
// pre-configured HTTP servers which are not yet started. The motivation is
// that in the current startup/shutdown setup we de-couple the listener
// creation from the server startup assuming that if any of the listeners
// cannot be bound we fail immediately and later failures do not occur.
// Therefore, starting a server with a running listener is assumed to not
// produce an error.
//
// The second motivation is that an HTTPS server needs to use the same TLSConfig
// on both the listener and the HTTP server. When listeners and servers are
// created at different times this becomes difficult to handle without keeping
// the TLS configuration somewhere or recreating it.
//
// This approach should ultimately be refactored to the point where we just
// start the server and any error should trigger a proper shutdown of the agent.
func (a *Agent) listenHTTP() ([]apiServer, error) {
	return nil, nil
}

func closeListeners(lns []net.Listener) {

}

// setupHTTPS adds HTTP/2 support, ConnState, and a connection handshake timeout
// to the http.Server.
func setupHTTPS(server *http.Server, connState func(net.Conn, http.ConnState), timeout time.Duration) error {
	return nil
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used so dead TCP connections eventually go away.
type tcpKeepAliveListener struct {
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	return nil, nil
}

func (a *Agent) listenSocket(path string) (net.Listener, error) {
	return nil, nil
}

// stopAllWatches stops all the currently running watches
func (a *Agent) stopAllWatches() {
}

// reloadWatches stops any existing watch plans and attempts to load the given
// set of watches.
func (a *Agent) reloadWatches(cfg *config.RuntimeConfig) error {
	return nil
}

// newConsulConfig translates a RuntimeConfig into a consul.Config.
// TODO: move this function to a different file, maybe config.go
func newConsulConfig(runtimeCfg *config.RuntimeConfig, logger hclog.Logger) (*consul.Config, error) {
	return nil, nil
}

// Setup the serf and memberlist config for any defined network segments.
func segmentConfig(config *config.RuntimeConfig) ([]consul.NetworkSegment, error) {
	return nil, nil

}

// registerEndpoint registers a handler for the consul RPC server
// under a unique name while making it accessible under the provided
// name. This allows overwriting handlers for the golang net/rpc
// service which does not allow this.
func (a *Agent) registerEndpoint(name string, handler interface{}) error {
	return nil
}

// RPC is used to make an RPC call to the Consul servers
// This allows the agent to implement the Consul.Interface
func (a *Agent) RPC(method string, args interface{}, reply interface{}) error {
	return nil
}

// Leave is used to prepare the agent for a graceful shutdown
func (a *Agent) Leave() error {
	return nil
}

// ShutdownAgent is used to hard stop the agent. Should be preceded by
// Leave to do it gracefully. Should be followed by ShutdownEndpoints to
// terminate the HTTP and DNS servers as well.
func (a *Agent) ShutdownAgent() error {
	return nil
}

// ShutdownEndpoints terminates the HTTP and DNS servers. Should be
// preceded by ShutdownAgent.
// TODO: remove this method, move to ShutdownAgent
func (a *Agent) ShutdownEndpoints() {
	return
}

// RetryJoinCh is a channel that transports errors
// from the retry join process.
func (a *Agent) RetryJoinCh() <-chan error {
	return nil
}

// ShutdownCh is used to return a channel that can be
// selected to wait for the agent to perform a shutdown.
func (a *Agent) ShutdownCh() <-chan struct{} {
	return nil
}

// JoinLAN is used to have the agent join a LAN cluster
func (a *Agent) JoinLAN(addrs []string) (n int, err error) {
	return
}

// JoinWAN is used to have the agent join a WAN cluster
func (a *Agent) JoinWAN(addrs []string) (n int, err error) {
	return 0, nil
}

// PrimaryMeshGatewayAddressesReadyCh returns a channel that will be closed
// when federation state replication ships back at least one primary mesh
// gateway (not via fallback config).
func (a *Agent) PrimaryMeshGatewayAddressesReadyCh() <-chan struct{} {
	return nil
}

// PickRandomMeshGatewaySuitableForDialing is a convenience function used for writing tests.
func (a *Agent) PickRandomMeshGatewaySuitableForDialing(dc string) string {
	return ""
}

// RefreshPrimaryGatewayFallbackAddresses is used to update the list of current
// fallback addresses for locating mesh gateways in the primary datacenter.
func (a *Agent) RefreshPrimaryGatewayFallbackAddresses(addrs []string) error {
	return nil
}

// ForceLeave is used to remove a failed node from the cluster
func (a *Agent) ForceLeave(node string, prune bool) (err error) {
	return nil
}

// LocalMember is used to return the local node
func (a *Agent) LocalMember() serf.Member {
	return serf.Member{}
}

// LANMembers is used to retrieve the LAN members
func (a *Agent) LANMembers() []serf.Member {
	return nil
}

// WANMembers is used to retrieve the WAN members
func (a *Agent) WANMembers() []serf.Member {
	return nil
}

// IsMember is used to check if a node with the given nodeName
// is a member
func (a *Agent) IsMember(nodeName string) bool {
	return false
}

// StartSync is called once Services and Checks are registered.
// This is called to prevent a race between clients and the anti-entropy routines
func (a *Agent) StartSync() {
}

// PauseSync is used to pause anti-entropy while bulk changes are made. It also
// sets state that agent-local watches use to "ride out" config reloads and bulk
// updates which might spuriously unload state and reload it again.
func (a *Agent) PauseSync() {
}

// ResumeSync is used to unpause anti-entropy after bulk changes are make
func (a *Agent) ResumeSync() {
}

// SyncPausedCh returns either a channel or nil. If nil sync is not paused. If
// non-nil, the channel will be closed when sync resumes.
func (a *Agent) SyncPausedCh() <-chan struct{} {
	return nil
}

// GetLANCoordinate returns the coordinates of this node in the local pools
// (assumes coordinates are enabled, so check that before calling).
func (a *Agent) GetLANCoordinate() (lib.CoordinateSet, error) {
	return nil, nil
}

// reapServicesInternal does a single pass, looking for services to reap.
func (a *Agent) reapServicesInternal() {
}

// reapServices is a long running goroutine that looks for checks that have been
// critical too long and deregisters their associated services.
func (a *Agent) reapServices() {
}

// persistedService is used to wrap a service definition and bundle it
// with an ACL token so we can restore both at a later agent start.
type persistedService struct {
}

// persistService saves a service definition to a JSON file in the data dir
func (a *Agent) persistService(service *structs.NodeService, source configSource) error {
	return nil
}

// purgeService removes a persisted service definition file from the data dir
func (a *Agent) purgeService(serviceID structs.ServiceID) error {
	return nil
}

// persistCheck saves a check definition to the local agent's state directory
func (a *Agent) persistCheck(check *structs.HealthCheck, chkType *structs.CheckType, source configSource) error {
	return nil
}

// purgeCheck removes a persisted check definition file from the data dir
func (a *Agent) purgeCheck(checkID structs.CheckID) error {
	return nil
}

// persistedServiceConfig is used to serialize the resolved service config that
// feeds into the ServiceManager at registration time so that it may be
// restored later on.
type persistedServiceConfig struct {
	ServiceID string
	Defaults  *structs.ServiceConfigResponse
	structs.EnterpriseMeta
}

func (a *Agent) persistServiceConfig(serviceID structs.ServiceID, defaults *structs.ServiceConfigResponse) error {
	return nil
}

func (a *Agent) purgeServiceConfig(serviceID structs.ServiceID) error {
	return nil
}

func (a *Agent) readPersistedServiceConfigs() (map[structs.ServiceID]*structs.ServiceConfigResponse, error) {
	return nil, nil
}

// AddServiceAndReplaceChecks is used to add a service entry and its check. Any check for this service missing from chkTypes will be deleted.
// This entry is persistent and the agent will make a best effort to
// ensure it is registered
func (a *Agent) AddServiceAndReplaceChecks(service *structs.NodeService, chkTypes []*structs.CheckType, persist bool, token string, source configSource) error {
	return nil
}

// AddService is used to add a service entry.
// This entry is persistent and the agent will make a best effort to
// ensure it is registered
func (a *Agent) AddService(service *structs.NodeService, chkTypes []*structs.CheckType, persist bool, token string, source configSource) error {
	return nil
}

// addServiceLocked adds a service entry to the service manager if enabled, or directly
// to the local state if it is not. This function assumes the state lock is already held.
func (a *Agent) addServiceLocked(req *addServiceRequest) error {
	return nil
}

// addServiceRequest is the union of arguments for calling both
// addServiceLocked and addServiceInternal. The overlap was significant enough
// to warrant merging them and indicating which fields are meant to be set only
// in one of the two contexts.
//
// Before using the request struct one of the fixupFor*() methods should be
// invoked to clear irrelevant fields.
//
// The ServiceManager.AddService signature is largely just a passthrough for
// addServiceLocked and should be treated as such.
type addServiceRequest struct {
	service               *structs.NodeService
	chkTypes              []*structs.CheckType
	previousDefaults      *structs.ServiceConfigResponse // just for: addServiceLocked
	waitForCentralConfig  bool                           // just for: addServiceLocked
	persistService        *structs.NodeService           // just for: addServiceInternal
	persistDefaults       *structs.ServiceConfigResponse // just for: addServiceInternal
	persist               bool
	persistServiceConfig  bool
	token                 string
	replaceExistingChecks bool
	source                configSource
	snap                  map[structs.CheckID]*structs.HealthCheck
}

func (r *addServiceRequest) fixupForAddServiceLocked() {
}

func (r *addServiceRequest) fixupForAddServiceInternal() {
}

// addServiceInternal adds the given service and checks to the local state.
func (a *Agent) addServiceInternal(req *addServiceRequest) error {
	return nil
}

// validateService validates an service and its checks, either returning an error or emitting a
// warning based on the nature of the error.
func (a *Agent) validateService(service *structs.NodeService, chkTypes []*structs.CheckType) error {
	return nil
}

// cleanupRegistration is called on  registration error to ensure no there are no
// leftovers after a partial failure
func (a *Agent) cleanupRegistration(serviceIDs []structs.ServiceID, checksIDs []structs.CheckID) {

}

// RemoveService is used to remove a service entry.
// The agent will make a best effort to ensure it is deregistered
func (a *Agent) RemoveService(serviceID structs.ServiceID) error {
	return nil
}

func (a *Agent) removeService(serviceID structs.ServiceID, persist bool) error {
	return nil
}

// removeServiceLocked is used to remove a service entry.
// The agent will make a best effort to ensure it is deregistered
func (a *Agent) removeServiceLocked(serviceID structs.ServiceID, persist bool) error {
	return nil
}

func (a *Agent) removeServiceSidecars(serviceID structs.ServiceID, persist bool) error {
	return nil
}

// AddCheck is used to add a health check to the agent.
// This entry is persistent and the agent will make a best effort to
// ensure it is registered. The Check may include a CheckType which
// is used to automatically update the check status
func (a *Agent) AddCheck(check *structs.HealthCheck, chkType *structs.CheckType, persist bool, token string, source configSource) error {
	return nil
}

func (a *Agent) addCheckLocked(check *structs.HealthCheck, chkType *structs.CheckType, persist bool, token string, source configSource) error {
	return nil
}

func (a *Agent) addCheck(check *structs.HealthCheck, chkType *structs.CheckType, service *structs.NodeService, token string, source configSource) error {
	return nil
}

// RemoveCheck is used to remove a health check.
// The agent will make a best effort to ensure it is deregistered
func (a *Agent) RemoveCheck(checkID structs.CheckID, persist bool) error {
	return nil
}

// removeCheckLocked is used to remove a health check.
// The agent will make a best effort to ensure it is deregistered
func (a *Agent) removeCheckLocked(checkID structs.CheckID, persist bool) error {
	return nil
}

// ServiceHTTPBasedChecks returns HTTP and GRPC based Checks
// for the given serviceID
func (a *Agent) ServiceHTTPBasedChecks(serviceID structs.ServiceID) []structs.CheckType {
	return nil
}

// AdvertiseAddrLAN returns the AdvertiseAddrLAN config value
func (a *Agent) AdvertiseAddrLAN() string {
	return ""
}

// resolveProxyCheckAddress returns the best address to use for a TCP check of
// the proxy's public listener. It expects the input to already have default
// values populated by applyProxyConfigDefaults. It may return an empty string
// indicating that the TCP check should not be created at all.
//
// By default this uses the proxy's bind address which in turn defaults to the
// agent's bind address. If the proxy bind address ends up being 0.0.0.0 we have
// to assume the agent can dial it over loopback which is usually true.
//
// In some topologies such as proxy being in a different container, the IP the
// agent used to dial proxy over a local bridge might not be the same as the
// container's public routable IP address so we allow a manual override of the
// check address in config "tcp_check_address" too.
//
// Finally the TCP check can be disabled by another manual override
// "disable_tcp_check" in cases where the agent will never be able to dial the
// proxy directly for some reason.
func (a *Agent) resolveProxyCheckAddress(proxyCfg map[string]interface{}) string {
	return ""
}

func (a *Agent) cancelCheckMonitors(checkID structs.CheckID) {
}

// updateTTLCheck is used to update the status of a TTL check via the Agent API.
func (a *Agent) updateTTLCheck(checkID structs.CheckID, status, output string) error {
	return nil
}

// persistCheckState is used to record the check status into the data dir.
// This allows the state to be restored on a later agent start. Currently
// only useful for TTL based checks.
func (a *Agent) persistCheckState(check *checks.CheckTTL, status, output string) error {
	return nil
}

// loadCheckState is used to restore the persisted state of a check.
func (a *Agent) loadCheckState(check *structs.HealthCheck) error {
	return nil
}

// purgeCheckState is used to purge the state of a check from the data dir
func (a *Agent) purgeCheckState(checkID structs.CheckID) error {
	return nil
}

// Stats is used to get various debugging state from the sub-systems
func (a *Agent) Stats() map[string]map[string]string {
	return nil
}

// storePid is used to write out our PID to a file if necessary
func (a *Agent) storePid() error {
	return nil
}

// deletePid is used to delete our PID on exit
func (a *Agent) deletePid() error {
	return nil
}

// loadServices will load service definitions from configuration and persisted
// definitions on disk, and load them into the local agent.
func (a *Agent) loadServices(conf *config.RuntimeConfig, snap map[structs.CheckID]*structs.HealthCheck) error {
	return nil
}

// unloadServices will deregister all services.
func (a *Agent) unloadServices() error {
	return nil
}

// loadChecks loads check definitions and/or persisted check definitions from
// disk and re-registers them with the local agent.
func (a *Agent) loadChecks(conf *config.RuntimeConfig, snap map[structs.CheckID]*structs.HealthCheck) error {
	return nil
}

// unloadChecks will deregister all checks known to the local agent.
func (a *Agent) unloadChecks() error {
	return nil
}

// snapshotCheckState is used to snapshot the current state of the health
// checks. This is done before we reload our checks, so that we can properly
// restore into the same state.
func (a *Agent) snapshotCheckState() map[structs.CheckID]*structs.HealthCheck {
	return nil
}

// loadMetadata loads node metadata fields from the agent config and
// updates them on the local agent.
func (a *Agent) loadMetadata(conf *config.RuntimeConfig) error {
	return nil
}

// unloadMetadata resets the local metadata state
func (a *Agent) unloadMetadata() {
}

// serviceMaintCheckID returns the ID of a given service's maintenance check
func serviceMaintCheckID(serviceID structs.ServiceID) structs.CheckID {
	return structs.CheckID{}
}

// EnableServiceMaintenance will register a false health check against the given
// service ID with critical status. This will exclude the service from queries.
func (a *Agent) EnableServiceMaintenance(serviceID structs.ServiceID, reason, token string) error {
	return nil
}

// DisableServiceMaintenance will deregister the fake maintenance mode check
// if the service has been marked as in maintenance.
func (a *Agent) DisableServiceMaintenance(serviceID structs.ServiceID) error {
	return nil
}

// EnableNodeMaintenance places a node into maintenance mode.
func (a *Agent) EnableNodeMaintenance(reason, token string) {
}

// DisableNodeMaintenance removes a node from maintenance mode
func (a *Agent) DisableNodeMaintenance() {
}

func (a *Agent) loadLimits(conf *config.RuntimeConfig) {

}

// ReloadConfig will atomically reload all configuration, including
// all services, checks, tokens, metadata, dnsServer configs, etc.
// It will also reload all ongoing watches.
func (a *Agent) ReloadConfig() error {
	return nil
}

// reloadConfigInternal is mainly needed for some unit tests. Instead of parsing
// the configuration using CLI flags and on disk config, this just takes a
// runtime configuration and applies it.
func (a *Agent) reloadConfigInternal(newCfg *config.RuntimeConfig) error {
	return nil
}

// LocalBlockingQuery performs a blocking query in a generic way against
// local agent state that has no RPC or raft to back it. It uses `hash` parameter
// instead of an `index`.
// `alwaysBlock` determines whether we block if the provided hash is empty.
// Callers like the AgentService endpoint will want to return the current result if a hash isn't provided.
// On the other hand, for cache notifications we always want to block. This avoids an empty first response.
func (a *Agent) LocalBlockingQuery(alwaysBlock bool, hash string, wait time.Duration,
	fn func(ws memdb.WatchSet) (string, interface{}, error)) (string, interface{}, error) {
	return "", nil, nil
}

// registerCache types on a.cache.
// This function may only be called once from New.
//
// Note: this function no longer registered all cache-types. Newer cache-types
// that do not depend on Agent are registered from registerCacheTypes.
func (a *Agent) registerCache() {
}

// LocalState returns the agent's local state
func (a *Agent) LocalState() *local.State {
	return a.State
}

// rerouteExposedChecks will inject proxy address into check targets
// Future calls to check() will dial the proxy listener
// The agent stateLock MUST be held for this to be called
func (a *Agent) rerouteExposedChecks(serviceID structs.ServiceID, proxyAddr string) error {
	return nil
}

// resetExposedChecks will set Proxy addr in HTTP checks to empty string
// Future calls to check() will use the original target c.HTTP or c.GRPC
// The agent stateLock MUST be held for this to be called
func (a *Agent) resetExposedChecks(serviceID structs.ServiceID) {
}

// listenerPort allocates a port from the configured range
// The agent stateLock MUST be held when this is called
func (a *Agent) listenerPortLocked(svcID structs.ServiceID, checkID structs.CheckID) (int, error) {
	return 0, nil
}

func listenerPortKey(svcID structs.ServiceID, checkID structs.CheckID) string {
	return ""
}

// grpcInjectAddr injects an ip and port into an address of the form: ip:port[/service]
func grpcInjectAddr(existing string, ip string, port int) string {
	return ""
}

// httpInjectAddr injects a port then an IP into a URL
func httpInjectAddr(url string, ip string, port int) string {
	return ""
}

func fixIPv6(address string) string {
	return ""
}

// defaultIfEmpty returns the value if not empty otherwise the default value.
func defaultIfEmpty(val, defaultVal string) string {
	return ""
}
