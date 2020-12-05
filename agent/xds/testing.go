package xds

import (
	"context"
	"sync"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	envoytype "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/mitchellh/go-testing-interface"
	"google.golang.org/grpc/metadata"
)

// TestADSStream mocks
// discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer to allow
// testing ADS handler.
type TestADSStream struct {
	ctx    context.Context
	sendCh chan *envoy.DiscoveryResponse
	recvCh chan *envoy.DiscoveryRequest
}

// NewTestADSStream makes a new TestADSStream
func NewTestADSStream(t testing.T, ctx context.Context) *TestADSStream {
	return nil
}

// Send implements ADSStream
func (s *TestADSStream) Send(r *envoy.DiscoveryResponse) error {
	return nil
}

// Recv implements ADSStream
func (s *TestADSStream) Recv() (*envoy.DiscoveryRequest, error) {
	return nil, nil
}

// SetHeader implements ADSStream
func (s *TestADSStream) SetHeader(metadata.MD) error {
	return nil
}

// SendHeader implements ADSStream
func (s *TestADSStream) SendHeader(metadata.MD) error {
	return nil
}

// SetTrailer implements ADSStream
func (s *TestADSStream) SetTrailer(metadata.MD) {
}

// Context implements ADSStream
func (s *TestADSStream) Context() context.Context {
	return s.ctx
}

// SendMsg implements ADSStream
func (s *TestADSStream) SendMsg(m interface{}) error {
	return nil
}

// RecvMsg implements ADSStream
func (s *TestADSStream) RecvMsg(m interface{}) error {
	return nil
}

type configState struct {
	lastNonce, lastVersion, acceptedVersion string
}

// TestEnvoy is a helper to simulate Envoy ADS requests.
type TestEnvoy struct {
	sync.Mutex
	stream  *TestADSStream
	proxyID string
	token   string
	state   map[string]configState
	ctx     context.Context
	cancel  func()
}

// NewTestEnvoy creates a TestEnvoy instance.
func NewTestEnvoy(t testing.T, proxyID, token string) *TestEnvoy {
	return nil
}

func hexString(v uint64) string {
	return ""
}

func stringToEnvoyVersion(vs string) (*envoytype.SemanticVersion, bool) {
	return nil, false
}

// SendReq sends a request from the test server.
func (e *TestEnvoy) SendReq(t testing.T, typeURL string, version, nonce uint64) {

}

// Close closes the client and cancels it's request context.
func (e *TestEnvoy) Close() error {
	return nil
}

// TestCheckRequest creates an envoyauth.CheckRequest with the source and
// destination service names.
func TestCheckRequest(t testing.T, source, dest string) *envoyauth.CheckRequest {
	return nil
}

func makeAttributeContextPeer(t testing.T, svc string) *envoyauth.AttributeContext_Peer {
	return nil
}
