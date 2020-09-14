package grpc

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/agent/grpc/internal/testservice"
	"github.com/hashicorp/consul/agent/grpc/resolver"
	"github.com/hashicorp/consul/agent/metadata"
	"github.com/hashicorp/consul/sdk/testutil/retry"
	"github.com/stretchr/testify/require"
)

// TODO(streaming): TestNewDialer with TLS enabled

func TestClientConnPool_IntegrationWithGRPCResolver_Failover(t *testing.T) {
	count := 4
	cfg := resolver.Config{Scheme: newScheme(t.Name())}
	res := resolver.NewServerResolverBuilder(cfg)
	resolver.RegisterWithGRPC(res)
	pool := NewClientConnPool(res, nil)

	for i := 0; i < count; i++ {
		name := fmt.Sprintf("server-%d", i)
		srv := newTestServer(t, name, "dc1")
		res.AddServer(srv.Metadata())
		t.Cleanup(srv.shutdown)
	}

	conn, err := pool.ClientConn("dc1")
	require.NoError(t, err)
	client := testservice.NewSimpleClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	first, err := client.Something(ctx, &testservice.Req{})
	require.NoError(t, err)

	res.RemoveServer(&metadata.Server{ID: first.ServerName, Datacenter: "dc1"})

	resp, err := client.Something(ctx, &testservice.Req{})
	require.NoError(t, err)
	require.NotEqual(t, resp.ServerName, first.ServerName)
}

func newScheme(n string) string {
	s := strings.Replace(n, "/", "", -1)
	s = strings.Replace(s, "_", "", -1)
	return strings.ToLower(s)
}

func TestClientConnPool_IntegrationWithGRPCResolver_Rebalance(t *testing.T) {
	count := 4
	cfg := resolver.Config{Scheme: newScheme(t.Name())}
	res := resolver.NewServerResolverBuilder(cfg)
	resolver.RegisterWithGRPC(res)
	pool := NewClientConnPool(res, nil)

	for i := 0; i < count; i++ {
		name := fmt.Sprintf("server-%d", i)
		srv := newTestServer(t, name, "dc1")
		res.AddServer(srv.Metadata())
		t.Cleanup(srv.shutdown)
	}

	conn, err := pool.ClientConn("dc1")
	require.NoError(t, err)
	client := testservice.NewSimpleClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	first, err := client.Something(ctx, &testservice.Req{})
	require.NoError(t, err)

	// Rebalance is random, but if we repeat it a few times it should give us a
	// new server.
	retry.RunWith(fastRetry, t, func(r *retry.R) {
		res.Rebalance()

		resp, err := client.Something(ctx, &testservice.Req{})
		require.NoError(t, err)
		require.NotEqual(t, resp.ServerName, first.ServerName)
	})
}

func TestClientConnPool_IntegrationWithGRPCResolver_MultiDC(t *testing.T) {
	dcs := []string{"dc1", "dc2", "dc3"}

	cfg := resolver.Config{Scheme: newScheme(t.Name())}
	res := resolver.NewServerResolverBuilder(cfg)
	resolver.RegisterWithGRPC(res)
	pool := NewClientConnPool(res, nil)

	for _, dc := range dcs {
		name := "server-0-" + dc
		srv := newTestServer(t, name, dc)
		res.AddServer(srv.Metadata())
		t.Cleanup(srv.shutdown)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	for _, dc := range dcs {
		conn, err := pool.ClientConn(dc)
		require.NoError(t, err)
		client := testservice.NewSimpleClient(conn)

		resp, err := client.Something(ctx, &testservice.Req{})
		require.NoError(t, err)
		require.Equal(t, resp.Datacenter, dc)
	}
}
