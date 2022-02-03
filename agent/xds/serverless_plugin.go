package xds

import (
	"fmt"
	"strings"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_lambda_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/aws_lambda/v3"
	envoy_http_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/golang/protobuf/proto"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/hashicorp/consul/agent/connect"
	"github.com/hashicorp/consul/agent/proxycfg"
	"github.com/hashicorp/consul/agent/structs"
)

const (
	lambdaEnabledTag      string = "serverless/lambda/enabled"
	arnTag                string = "serverless/lambda/arn"
	payloadPassthroughTag string = "serverless/lambda/payload-passhthrough"
	regionTag             string = "serverless/lambda/region"
)

type mutateConfiguration struct {
	serviceMeta map[structs.ServiceName]map[string]string
	kind        structs.ServiceKind
	datacenter  string
	trustDomain string
}

func makeMutateConfiguration(cfgSnap *proxycfg.ConfigSnapshot) mutateConfiguration {
	serviceMeta := make(map[structs.ServiceName]map[string]string)
	switch cfgSnap.Kind {
	case structs.ServiceKindTerminatingGateway:
		for serviceName, c := range cfgSnap.TerminatingGateway.ServiceConfigs {
			serviceMeta[serviceName] = c.Meta
		}
	case structs.ServiceKindConnectProxy:
		for serviceName, c := range cfgSnap.ConnectProxy.ServiceConfigs {
			serviceMeta[serviceName] = c.Meta
		}
	}

	trustDomain := ""
	if cfgSnap.Roots != nil {
		trustDomain = cfgSnap.Roots.TrustDomain
	}

	return mutateConfiguration{
		serviceMeta: serviceMeta,
		kind:        cfgSnap.Kind,
		datacenter:  cfgSnap.Datacenter,
		trustDomain: trustDomain,
	}
}

func MutateIndexedResources(resources *IndexedResources, config mutateConfiguration) (*IndexedResources, error) {
	switch config.kind {
	case structs.ServiceKindConnectProxy:
		serviceNamesToHack := makeServiceNamesToHack(config.serviceMeta)
		// hack listeners
		for name, msg := range resources.Index[ListenerType] {
			listener, ok := msg.(*envoy_listener_v3.Listener)

			if !ok {
				return resources, fmt.Errorf("error decoding listener")
			}

			// TODO there has to be a better way to do this.
			serviceName := ""
			if i := strings.IndexByte(name, ':'); i != -1 {
				serviceName = name[:i]
			}

			meta, ok := serviceNamesToHack[serviceName]
			if !ok || len(meta) == 0 || serviceName == "" {
				continue
			}

			newListener, err := hackConnectProxyListener(listener, meta)
			if err != nil {
				return resources, fmt.Errorf("Error hacking listener %s", name)
			}
			resources.Index[ListenerType][name] = newListener
		}

		snisToHack := makeSnisToHack(config)
		// hack clusters
		for name, msg := range resources.Index[ClusterType] {
			cluster, ok := msg.(*envoy_cluster_v3.Cluster)

			if !ok {
				return resources, fmt.Errorf("error decoding cluster")
			}

			if meta, ok := snisToHack[name]; ok {
				newCluster, err := hackTerminatingGatewayCluster(cluster, meta)
				if err != nil {
					return resources, fmt.Errorf("Error hacking cluster %s", name)
				}
				resources.Index[ClusterType][name] = newCluster
			}
		}
	case structs.ServiceKindTerminatingGateway:
		snisToHack := makeSnisToHack(config)

		// hack clusters
		for name, msg := range resources.Index[ClusterType] {
			cluster, ok := msg.(*envoy_cluster_v3.Cluster)

			if !ok {
				return resources, fmt.Errorf("error decoding cluster")
			}

			if meta, ok := snisToHack[name]; ok {
				newCluster, err := hackTerminatingGatewayCluster(cluster, meta)
				if err != nil {
					return resources, fmt.Errorf("Error hacking cluster %s", name)
				}
				resources.Index[ClusterType][name] = newCluster
			}
		}

		// hack listeners
		for name, msg := range resources.Index[ListenerType] {
			listener, ok := msg.(*envoy_listener_v3.Listener)

			if !ok {
				return resources, fmt.Errorf("error decoding listener")
			}

			newListener, err := hackTerminatingGatewayListener(listener, snisToHack)
			if err != nil {
				return resources, fmt.Errorf("Error hacking listener %s", name)
			}
			resources.Index[ListenerType][name] = newListener
		}
	}
	return resources, nil
}

func hackTerminatingGatewayCluster(c *envoy_cluster_v3.Cluster, meta map[string]string) (proto.Message, error) {
	transportSocket, err := makeUpstreamTLSTransportSocket(&envoy_tls_v3.UpstreamTlsContext{
		Sni: "*.amazonaws.com",
	})

	if err != nil {
		return c, fmt.Errorf("failed to make transport socket")
	}

	cluster := &envoy_cluster_v3.Cluster{
		Name:                 c.Name,
		ConnectTimeout:       c.ConnectTimeout,
		ClusterDiscoveryType: &envoy_cluster_v3.Cluster_Type{Type: envoy_cluster_v3.Cluster_LOGICAL_DNS},
		DnsLookupFamily:      envoy_cluster_v3.Cluster_V4_ONLY,
		LbPolicy:             envoy_cluster_v3.Cluster_ROUND_ROBIN,
		Metadata: &envoy_core_v3.Metadata{
			FilterMetadata: map[string]*_struct.Struct{
				"com.amazonaws.lambda": {
					Fields: map[string]*_struct.Value{
						"egress_gateway": {Kind: &_struct.Value_BoolValue{BoolValue: true}},
					},
				},
			},
		},
		LoadAssignment: &envoy_endpoint_v3.ClusterLoadAssignment{
			ClusterName: c.Name,
			Endpoints: []*envoy_endpoint_v3.LocalityLbEndpoints{
				{
					LbEndpoints: []*envoy_endpoint_v3.LbEndpoint{
						{
							HostIdentifier: &envoy_endpoint_v3.LbEndpoint_Endpoint{
								Endpoint: &envoy_endpoint_v3.Endpoint{
									Address: &envoy_core_v3.Address{
										Address: &envoy_core_v3.Address_SocketAddress{
											SocketAddress: &envoy_core_v3.SocketAddress{
												Address: fmt.Sprintf("lambda.%s.amazonaws.com", meta[regionTag]),
												PortSpecifier: &envoy_core_v3.SocketAddress_PortValue{
													PortValue: 443,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		TransportSocket: transportSocket,
	}
	return cluster, nil
}

func getSni(chain *envoy_listener_v3.FilterChain) string {
	var sni string

	if chain == nil {
		return sni
	}

	if chain.FilterChainMatch == nil {
		return sni
	}

	if len(chain.FilterChainMatch.ServerNames) == 0 {
		return sni
	}

	return chain.FilterChainMatch.ServerNames[0]
}

func hackTerminatingGatewayListener(l *envoy_listener_v3.Listener, snisToHack map[string]map[string]string) (proto.Message, error) {
	for _, filterChain := range l.FilterChains {
		sni := getSni(filterChain)
		if len(snisToHack[sni]) == 0 {
			continue
		}
		meta := snisToHack[sni]

		var filters []*envoy_listener_v3.Filter

		for _, filter := range filterChain.Filters {
			if filter.Name != "envoy.filters.network.http_connection_manager" {
				fmt.Println("IN HERE WEIRDO")
				filters = append(filters, filter)
				continue
			}
			if typedConfig := filter.GetTypedConfig(); typedConfig != nil {
				config := envoy_resource_v3.GetHTTPConnectionManager(filter)
				if config == nil {
					return l, fmt.Errorf("error unemarshaling filter.")
				}
				httpFilter, err := makeEnvoyHTTPFilter(
					"envoy.filters.http.aws_lambda",
					&envoy_lambda_v3.Config{
						Arn:                meta[arnTag],
						PayloadPassthrough: meta[payloadPassthroughTag] == "true",
						// TODO add invokation. Do we need to PR go-control-plane for this?
						// It looks like it is at least partially implemented there.
					},
				)
				if err != nil {
					return nil, err
				}

				config.HttpFilters = []*envoy_http_v3.HttpFilter{
					// TODO This should be more selective and replace
					// envoy.filters.http.router so it doesn't override RBAC
					httpFilter,
					{Name: "envoy.filters.http.router"},
				}
				newFilter, err := makeFilter("envoy.filters.network.http_connection_manager", config)
				if err != nil {
					return l, fmt.Errorf("error making new filter")
				}
				filters = append(filters, newFilter)
			}
		}
		filterChain.Filters = filters
	}

	return l, nil
}

func makeSnisToHack(config mutateConfiguration) map[string]map[string]string {
	snisToHack := make(map[string]map[string]string)

	for svc, meta := range config.serviceMeta {
		if meta[lambdaEnabledTag] != "" {
			sni := connect.ServiceSNI(svc.Name, "", svc.NamespaceOrDefault(), svc.PartitionOrDefault(), config.datacenter, config.trustDomain)
			snisToHack[sni] = meta
		}
	}

	return snisToHack
}

func hackConnectProxyListener(l *envoy_listener_v3.Listener, meta map[string]string) (*envoy_listener_v3.Listener, error) {
	for _, filterChain := range l.FilterChains {
		var filters []*envoy_listener_v3.Filter

		for _, filter := range filterChain.Filters {
			if filter.Name != "envoy.filters.network.http_connection_manager" {
				filters = append(filters, filter)
				continue
			}
			if typedConfig := filter.GetTypedConfig(); typedConfig != nil {
				config := envoy_resource_v3.GetHTTPConnectionManager(filter)
				if config == nil {
					return l, fmt.Errorf("error unemarshaling filter.")
				}
				httpFilter, err := makeEnvoyHTTPFilter(
					"envoy.filters.http.aws_lambda",
					&envoy_lambda_v3.Config{
						Arn:                meta[arnTag],
						PayloadPassthrough: meta[payloadPassthroughTag] == "true",
						// TODO add invokation. Do we need to PR go-control-plane for this?
						// It looks like it is at least partially implemented there.
					},
				)
				if err != nil {
					return nil, err
				}

				config.HttpFilters = []*envoy_http_v3.HttpFilter{
					// TODO This should be more selective and replace
					// envoy.filters.http.router so it doesn't override RBAC
					httpFilter,
					{Name: "envoy.filters.http.router"},
				}
				newFilter, err := makeFilter("envoy.filters.network.http_connection_manager", config)
				if err != nil {
					return l, fmt.Errorf("error making new filter")
				}
				filters = append(filters, newFilter)
			}
		}
		filterChain.Filters = filters
	}

	return l, nil
}

func makeServiceNamesToHack(serviceConfigs map[structs.ServiceName]map[string]string) map[string]map[string]string {
	serviceNamesToHack := make(map[string]map[string]string)

	for svc, meta := range serviceConfigs {
		if meta[lambdaEnabledTag] != "" {
			serviceNamesToHack[svc.Name] = meta
		}
	}

	return serviceNamesToHack
}
