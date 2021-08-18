package multidc

import (
	"fmt"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/require"
)

func TestMultiDC_FederationViaMeshGateway(t *testing.T) {
	c, err := docker.NewClientFromEnv()
	require.NoError(t, err)

	netDC1 := createNetwork(t, c, "dc1")
	_ = createNetwork(t, c, "dc2")
	netWAN := createNetwork(t, c, "wan")

	cIDAgentDC1 := createAgent(t, c, AgentOptions{
		Config:   []string{`datacenter = "dc1"`},
		Networks: []string{netDC1, netWAN},
	})

	fmt.Println(cIDAgentDC1)
}

func createNetwork(t *testing.T, c *docker.Client, name string) string {
	t.Helper()
	net, err := c.CreateNetwork(docker.CreateNetworkOptions{Name: name})
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := c.RemoveNetwork(net.ID); err != nil {
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

func createAgent(t *testing.T, c *docker.Client, opts AgentOptions) string {
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

	con, err := c.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Env:   opts.Environment,
			Cmd:   cmd,
			Image: opts.Image,
		},
	})
	require.NoError(t, err, "container created failed")
	t.Cleanup(func() {
		rmOpts := docker.RemoveContainerOptions{ID: con.ID, Force: true}
		if err := c.RemoveContainer(rmOpts); err != nil {
			t.Logf("failed to remove container %v: %v", con.ID, err)
		}
	})

	for _, net := range opts.Networks {
		err := c.ConnectNetwork(net, docker.NetworkConnectionOptions{Container: con.ID})
		require.NoError(t, err, "connect container %v to network %v", con.ID, net)
	}

	err = c.StartContainer(con.ID, nil)
	require.NoError(t, err, "container start failed")

	return con.ID
}
