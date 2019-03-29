package envoy

import (
	"bytes"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/agent"
	"github.com/hashicorp/consul/agent/xds"
	envoycmd "github.com/hashicorp/consul/command/connect/envoy"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/require"
)


func testDockerPool(t *testing.T) *dockertest.Pool {
	t.Helper()
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	return pool
}

func testRunEnvoy(t *testing.T, pool *dockertest.Pool, version string, bootstrapJSON []byte) (*dockertest.Resource, func()) {
	t.Helper()

	if pool == nil {
		pool = testDockerPool(t)
	}
	version = strings.TrimLeft(version, "v")

	// Explicitly use /tmp since this is accessible to docker on macOS while the
	// default temp files are not.
	tmp, err := ioutil.TempFile("/tmp", "envoy-bootstrap")
	require.NoError(t, err)
	// Close it immediately as we'll write it with WriteFile
	tmp.Close()

	require.NoError(t, ioutil.WriteFile(tmp.Name(), bootstrapJSON, 0644))

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       t.Name() + "_envoy_" + time.Now().Format("2006-01-02_15-04-05"),
		Repository: "envoyproxy/envoy",
		Tag:        "v" + version,
		Mounts:     []string{tmp.Name() + ":/bootstrap.json"},
		Cmd:        []string{"envoy", "-c", "/bootstrap.json", "-l", "debug"},
	})
	require.NoError(t, err)
	// Expire works even if we crash and ensures we don't leave tons of docker
	// containers around the place. We leave containers around for an hour to
	// allow debugging for failures by just uncommenting the defer.
	resource.Expire(3699)

	return resource, func() {
		pool.Purge(resource)
		os.Remove(tmp.Name())
	}
}

func testBootstrapArgs(t *testing.T, a *agent.TestAgent) *envoycmd.BootstrapTplArgs {
	agentAddr, agentPort, err := net.SplitHostPort(a.HTTPAddr())
	require.NoError(t, err)
	return &envoycmd.BootstrapTplArgs{
		ProxyCluster:          "test",
		ProxyID:               "test-sidecar-proxy",
		AgentAddress:          agentAddr,
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

	bc := envoycmd.BootstrapConfig{}
	bs, err := bc.GenerateJSON(testBootstrapArgs(t, a))
	require.NoError(err)

	pool := testDockerPool(t)

	res, cleanup := testRunEnvoy(t, pool, "1.8.0", bs)
	defer cleanup()

	var buf bytes.Buffer

	time.AfterFunc(10*time.Second, func() {
		cleanup()
	})

	logsOptions := docker.LogsOptions{
		Container:    res.Container.ID,
		OutputStream: &buf,
		ErrorStream:  &buf,
		Follow:       true,
		Stdout:       true,
		Stderr:       true,
	}
	if err := pool.Client.Logs(logsOptions); err != nil {
		panic(err)
	}


	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	// if err := pool.Retry(func() error {
	// 	resp, err := http.Get("http://" + resource.GetHostPort("10000/tcp"))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	resp.Body.Close()
	// 	return nil
	// }); err != nil {
	// 	log.Fatalf("Could not connect to docker: %s", err)
	// }
}
