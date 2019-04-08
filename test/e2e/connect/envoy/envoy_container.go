package envoy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/banks/lennut"
	"github.com/mitchellh/go-testing-interface"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/require"
)

type Container struct {
	pool    *dockertest.Pool
	res     *dockertest.Resource
	shim    *dockertest.Resource
	lc      *lennut.Client
	tmpFile string
}

func TestContainer(t testing.T, name, version, agentGRPCAddr string, bootstrapJSON []byte) *Container {
	t.Helper()

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	version = strings.TrimLeft(version, "v")

	// Explicitly use /tmp since this is accessible to docker on macOS while the
	// default temp files are not.
	tmp, err := ioutil.TempFile("/tmp", "envoy-bootstrap")
	require.NoError(t, err)
	// Close it immediately as we'll write it with WriteFile
	tmp.Close()

	require.NoError(t, ioutil.WriteFile(tmp.Name(), bootstrapJSON, 0644))

	containerName := "consul_e2e_" + t.Name() + "_" + name
	shimName := containerName + "_agent_shim"

	// Remove it if one is left over from last run
	require.NoError(t, pool.RemoveContainerByName(containerName))
	require.NoError(t, pool.RemoveContainerByName(shimName))

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name: containerName,
		Labels: map[string]string{
			"consul_test": t.Name(),
		},
		ExposedPorts: []string{
			"3001/tcp",                            // This is for the reverse tunnel shim
			"19000/tcp",                           // For Envoy Admin server
			"21000/tcp", "22000/tcp", "22001/tcp", // For Envoy listeners
		},
		Repository: "envoyproxy/envoy",
		Tag:        "v" + version,
		Mounts:     []string{tmp.Name() + ":/bootstrap.json"},
		Cmd:        []string{"envoy", "-c", "/bootstrap.json", "-l", "debug"},
	}, func(hc *docker.HostConfig) {
		hc.NetworkMode = "host"
	})
	require.NoError(t, err)

	// Would be cleaner to have the shim share the netns but dockertest makes that
	// hard because it has no way to override the baked in EXPOSE ports
	// (explicitly empty list here is ignored) so Docker daemon rejects the
	// exposed ports and container network mode.
	shim, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name: shimName,
		Labels: map[string]string{
			"consul_test": t.Name(),
		},
		ExposedPorts: []string{"3001/tcp"},
		Repository:   "pbanks/lennut",
		// Proxy gRPC (8502) back to the TestAgent via reverse tunnel
		Cmd: []string{"-bind-proxy", "localhost:8502"},
	}, func(hc *docker.HostConfig) {
		hc.NetworkMode = "host"
	})
	if err != nil {
		// Cleanup the other container
		pool.Purge(resource)
	}
	require.NoError(t, err)

	// Finally run, a lennut client to proxy connections back to the test agent
	lc := lennut.Client{
		DialAddr: resource.GetHostPort("3001/tcp"), // Dial the shim proxy container
		ProxyTo:  agentGRPCAddr,
	}
	go func() {
		require.NoError(t, lc.Run())
	}()

	return &Container{
		pool:    pool,
		res:     resource,
		shim:    shim,
		lc:      &lc,
		tmpFile: tmp.Name(),
	}
}

func (c *Container) HTTPEndpoint(port int) string {
	return "http://" + c.res.GetHostPort(fmt.Sprintf("%d/tcp", port))
}

func (c *Container) Logs(t testing.T) string {
	if c.pool == nil || c.res == nil {
		return ""
	}
	var buf bytes.Buffer
	logsOptions := docker.LogsOptions{
		Container:    c.res.Container.ID,
		OutputStream: &buf,
		ErrorStream:  &buf,
		Follow:       true,
		Stdout:       true,
		Stderr:       true,
	}
	require.NoError(t, c.pool.Client.Logs(logsOptions))
	return buf.String()
}

func (c *Container) Stop() {
	if c.pool == nil {
		return
	}
	if c.res != nil {
		c.pool.Purge(c.res)
		c.res = nil
	}
	if c.shim != nil {
		c.pool.Purge(c.shim)
		c.shim = nil
	}
	if c.lc != nil {
		c.lc.Close()
		c.lc = nil
	}
	if c.tmpFile != "" {
		os.Remove(c.tmpFile)
		c.tmpFile = ""
	}
}
