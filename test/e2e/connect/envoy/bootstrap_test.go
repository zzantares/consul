package envoy

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/hashicorp/consul/agent"
	"github.com/hashicorp/consul/agent/xds"
	"github.com/hashicorp/consul/api"
	envoycmd "github.com/hashicorp/consul/command/connect/envoy"
	"github.com/hashicorp/consul/sdk/testutil/retry"

	"github.com/stretchr/testify/require"
)

func testBootstrapArgs(t *testing.T, a *agent.TestAgent) *envoycmd.BootstrapTplArgs {
	_, agentPort, err := net.SplitHostPort(a.HTTPAddr())
	require.NoError(t, err)
	return &envoycmd.BootstrapTplArgs{
		ProxyCluster:          "web",
		ProxyID:               "web-sidecar-proxy",
		AgentAddress:          "127.0.0.1",
		AgentPort:             agentPort,
		AdminBindAddress:      "0.0.0.0",
		AdminBindPort:         "19000",
		LocalAgentClusterName: xds.LocalAgentClusterName,
	}
}

func TestEnvoyBootstrap(t *testing.T) {
	require := require.New(t)

	a := agent.NewTestAgent(t, "envoy-bootstrap-agent", "")
	defer a.Shutdown()

	// Start a simple HTTP service to proxy to
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(err)
	serverPort, err := strconv.Atoi(serverURL.Port())
	require.NoError(err)

	// Register our service with a sidecar proxy
	agent := a.Client().Agent()
	reg := &api.AgentServiceRegistration{
		Name: "web",
		Port: serverPort,
		Connect: &api.AgentServiceConnect{
			SidecarService: &api.AgentServiceRegistration{},
		},
	}
	require.NoError(agent.ServiceRegister(reg))

	// Register another service with web as an upstream so we can connect through
	// it.
	reg = &api.AgentServiceRegistration{
		Name: "ingress",
		Port: serverPort, // doesn't matter we won't be testing inbound traffic on this.
		Connect: &api.AgentServiceConnect{
			SidecarService: &api.AgentServiceRegistration{
				Proxy: &api.AgentServiceConnectProxyConfig{
					Upstreams: []api.Upstream{
						api.Upstream{
							DestinationType: "service",
							DestinationName: "web",
							LocalBindPort:   22000,
						},
					},
				},
			},
		},
	}
	require.NoError(agent.ServiceRegister(reg))

	bc := envoycmd.BootstrapConfig{}
	bs, err := bc.GenerateJSON(testBootstrapArgs(t, a))
	require.NoError(err)

	c := TestContainer(t, "web-sidecar", "1.8.0", a.HTTPAddr(), bs)
	defer c.Stop()

	retry.Run(t, func(r *retry.R) {
		resp, err := http.Get(c.HTTPEndpoint(19000))
		require.NoError(err)
		resp.Body.Close()
	})

	c.Stop()

	// Dump Envoy logs on fail
	t.Log(c.Logs(t))
}
