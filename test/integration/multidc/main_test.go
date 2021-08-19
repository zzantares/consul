package multidc

import (
	"fmt"
	"os"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/require"
)

type TC struct {
	*testing.T
	client *docker.Client
}

const labelTestName = "test-name"

func TestMultiDC_FederationViaMeshGateway(t *testing.T) {
	c, err := docker.NewClientFromEnv()
	require.NoError(t, err)

	tc := TC{T: t, client: c}
	cleanup(tc)

	netDC1 := createNetwork(tc, "dc1")
	_ = createNetwork(tc, "dc2")
	netWAN := createNetwork(tc, "wan")

	cIDAgentDC1 := createAgent(tc, AgentOptions{
		// TODO: enable TLS
		Config: []string{`
datacenter = "dc1"
connect {
  enabled                            = true
  enable_mesh_gateway_wan_federation = true
}
services {
  name = "mesh-gateway"
  kind = "mesh-gateway"
  port = 4431
  meta {
    consul-wan-federation = "1"
  }
}
`},
		Networks: []string{netDC1, netWAN},
	})

	fmt.Println(cIDAgentDC1)
	// TODO: remove
	time.Sleep(15 * time.Second)
}

func createNetwork(t TC, name string) string {
	t.Helper()
	net, err := t.client.CreateNetwork(docker.CreateNetworkOptions{
		Name:   name,
		Labels: map[string]string{labelTestName: t.Name()},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := t.client.RemoveNetwork(net.ID); err != nil {
			t.Logf("failed to remove network %v: %v", net.ID, err)
		}
	})
	return net.ID
}

type AgentOptions struct {
	// Image name of the docker image. Defaults to consul:dev, which must be
	// built locally before running the tests.
	Image string

	// Config is a list of HCL configuration. Each one will be applied using
	// the -hcl command line flag.
	Config []string

	// Environment variables to pass to the container.
	Environment []string

	// Networks is a list of docker networks the container should be connected to.
	Networks []string
}

func createAgent(t TC, opts AgentOptions) string {
	t.Helper()

	if opts.Image == "" {
		opts.Image = "consul:dev"
	}
	// TODO: add license to opts.Environment

	// TODO: maybe use an explicit list of config instead of -dev here to better
	// match production environments.
	cmd := []string{"agent", "-dev", "-client=0.0.0.0"}
	for _, hcl := range opts.Config {
		cmd = append(cmd, "-hcl", hcl)
	}

	con, err := t.client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Env:    opts.Environment,
			Cmd:    cmd,
			Image:  opts.Image,
			Labels: map[string]string{labelTestName: t.Name()},
		},
	})
	require.NoError(t, err, "container created failed")
	t.Cleanup(func() {
		rmOpts := docker.RemoveContainerOptions{ID: con.ID, Force: true}
		if err := t.client.RemoveContainer(rmOpts); err != nil {
			t.Logf("failed to remove container %v: %v", con.ID, err)
		}
	})

	for _, net := range opts.Networks {
		err := t.client.ConnectNetwork(net, docker.NetworkConnectionOptions{Container: con.ID})
		require.NoError(t, err, "connect container %v to network %v", con.ID, net)
	}

	err = t.client.StartContainer(con.ID, nil)
	require.NoError(t, err, "container start failed")

	return con.ID
}

// cleanup removes all containers and networks that have a label matching the
// test name. Normally tests will clean up after themselves, but if a test panics
// or is interrupted all cleanup functions may not have run. To prevent subsequent
// test runs from failing we start each test run by removing any stray resources.
func cleanup(t TC) {
	cons, err := t.client.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"label": {labelTestName + "=" + t.Name()},
		},
	})
	require.NoError(t, err)
	for _, con := range cons {
		rmOpts := docker.RemoveContainerOptions{ID: con.ID, Force: true}
		if err := t.client.RemoveContainer(rmOpts); err != nil {
			t.Logf("failed to remove container %v: %v", con.ID, err)
		}
	}

	nets, err := t.client.FilteredListNetworks(docker.NetworkFilterOpts{
		"label": map[string]bool{labelTestName + "=" + t.Name(): true},
	})
	require.NoError(t, err)
	for _, net := range nets {
		if err := t.client.RemoveNetwork(net.ID); err != nil {
			t.Logf("failed to remove network %v: %v", net.ID, err)
		}
	}
}

type MeshGatewayOptions struct {
	Image    string
	Name     string
	AgentCID string
	Networks []string
}

var envoyVersion = os.Getenv("ENVOY_VERSION")

func runMeshGateway(t TC, opts MeshGatewayOptions) {
	if opts.Image == "" {
		// TODO: need an image with consul CLI
		opts.Image = "docker.mirror.hashicorp.services/envoyproxy/envoy:" + envoyVersion
	}

	cmd := []string{
		"consul", "connect", "envoy",
		"-proxy-id", opts.Name,
		"-envoy-version", envoyVersion,
		"-admin-bind", "0.0.0.0:19000",
	}

	con, err := t.client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Cmd:    cmd,
			Image:  opts.Image,
			Labels: map[string]string{labelTestName: t.Name()},
		},
	})
	require.NoError(t, err, "container created failed")
	t.Cleanup(func() {
		rmOpts := docker.RemoveContainerOptions{ID: con.ID, Force: true}
		if err := t.client.RemoveContainer(rmOpts); err != nil {
			t.Logf("failed to remove container %v: %v", con.ID, err)
		}
	})

	for _, net := range opts.Networks {
		err := t.client.ConnectNetwork(net, docker.NetworkConnectionOptions{Container: con.ID})
		require.NoError(t, err, "connect container %v to network %v", con.ID, net)
	}

	err = t.client.StartContainer(con.ID, nil)
	require.NoError(t, err, "container start failed")
}
