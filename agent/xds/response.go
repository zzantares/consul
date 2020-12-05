package xds

import (
	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoymatcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
)

func createResponse(typeURL string, version, nonce string, resources []proto.Message) (*envoy.DiscoveryResponse, error) {
	return nil, nil
}

func makeAddress(ip string, port int) *envoycore.Address {
	return nil
}

func makeUint32Value(n int) *wrappers.UInt32Value {
	return nil
}

func makeBoolValue(n bool) *wrappers.BoolValue {
	return nil
}

func makeEnvoyRegexMatch(patt string) *envoymatcher.RegexMatcher {
	return nil
}
