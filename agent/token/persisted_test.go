package token

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/hashicorp/consul/agent/config"
	"github.com/stretchr/testify/require"
)

func TestAgent_loadTokens(t *testing.T) {
	t.Parallel()
	a := NewTestAgent(t, `
		acl = {
			enabled = true
			tokens = {
				agent = "alfa"
				agent_master = "bravo",
				default = "charlie"
				replication = "delta"
			}
		}

	`)
	defer a.Shutdown()
	require := require.New(t)

	tokensFullPath := filepath.Join(a.config.DataDir, tokensPath)

	t.Run("original-configuration", func(t *testing.T) {
		require.Equal("alfa", a.tokens.AgentToken())
		require.Equal("bravo", a.tokens.AgentMasterToken())
		require.Equal("charlie", a.tokens.UserToken())
		require.Equal("delta", a.tokens.ReplicationToken())
	})

	t.Run("updated-configuration", func(t *testing.T) {
		cfg := &config.RuntimeConfig{
			ACLToken:            "echo",
			ACLAgentToken:       "foxtrot",
			ACLAgentMasterToken: "golf",
			ACLReplicationToken: "hotel",
		}
		// ensures no error for missing persisted tokens file
		require.NoError(a.loadTokens(cfg))
		require.Equal("echo", a.tokens.UserToken())
		require.Equal("foxtrot", a.tokens.AgentToken())
		require.Equal("golf", a.tokens.AgentMasterToken())
		require.Equal("hotel", a.tokens.ReplicationToken())
	})

	t.Run("persisted-tokens", func(t *testing.T) {
		cfg := &config.RuntimeConfig{
			ACLToken:            "echo",
			ACLAgentToken:       "foxtrot",
			ACLAgentMasterToken: "golf",
			ACLReplicationToken: "hotel",
		}

		tokens := `{
			"agent" : "india",
			"agent_master" : "juliett",
			"default": "kilo",
			"replication" : "lima"
		}`

		require.NoError(ioutil.WriteFile(tokensFullPath, []byte(tokens), 0600))
		require.NoError(a.loadTokens(cfg))

		// no updates since token persistence is not enabled
		require.Equal("echo", a.tokens.UserToken())
		require.Equal("foxtrot", a.tokens.AgentToken())
		require.Equal("golf", a.tokens.AgentMasterToken())
		require.Equal("hotel", a.tokens.ReplicationToken())

		a.config.ACLEnableTokenPersistence = true
		require.NoError(a.loadTokens(cfg))

		require.Equal("india", a.tokens.AgentToken())
		require.Equal("juliett", a.tokens.AgentMasterToken())
		require.Equal("kilo", a.tokens.UserToken())
		require.Equal("lima", a.tokens.ReplicationToken())
	})

	t.Run("persisted-tokens-override", func(t *testing.T) {
		tokens := `{
			"agent" : "mike",
			"agent_master" : "november",
			"default": "oscar",
			"replication" : "papa"
		}`

		cfg := &config.RuntimeConfig{
			ACLToken:            "quebec",
			ACLAgentToken:       "romeo",
			ACLAgentMasterToken: "sierra",
			ACLReplicationToken: "tango",
		}

		require.NoError(ioutil.WriteFile(tokensFullPath, []byte(tokens), 0600))
		require.NoError(a.loadTokens(cfg))

		require.Equal("mike", a.tokens.AgentToken())
		require.Equal("november", a.tokens.AgentMasterToken())
		require.Equal("oscar", a.tokens.UserToken())
		require.Equal("papa", a.tokens.ReplicationToken())
	})

	t.Run("partial-persisted", func(t *testing.T) {
		tokens := `{
			"agent" : "uniform",
			"agent_master" : "victor"
		}`

		cfg := &config.RuntimeConfig{
			ACLToken:            "whiskey",
			ACLAgentToken:       "xray",
			ACLAgentMasterToken: "yankee",
			ACLReplicationToken: "zulu",
		}

		require.NoError(ioutil.WriteFile(tokensFullPath, []byte(tokens), 0600))
		require.NoError(a.loadTokens(cfg))

		require.Equal("uniform", a.tokens.AgentToken())
		require.Equal("victor", a.tokens.AgentMasterToken())
		require.Equal("whiskey", a.tokens.UserToken())
		require.Equal("zulu", a.tokens.ReplicationToken())
	})

	t.Run("persistence-error-not-json", func(t *testing.T) {
		cfg := &config.RuntimeConfig{
			ACLToken:            "one",
			ACLAgentToken:       "two",
			ACLAgentMasterToken: "three",
			ACLReplicationToken: "four",
		}

		require.NoError(ioutil.WriteFile(tokensFullPath, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, 0600))
		err := a.loadTokens(cfg)
		require.Error(err)

		require.Equal("one", a.tokens.UserToken())
		require.Equal("two", a.tokens.AgentToken())
		require.Equal("three", a.tokens.AgentMasterToken())
		require.Equal("four", a.tokens.ReplicationToken())
	})

	t.Run("persistence-error-wrong-top-level", func(t *testing.T) {
		cfg := &config.RuntimeConfig{
			ACLToken:            "alfa",
			ACLAgentToken:       "bravo",
			ACLAgentMasterToken: "charlie",
			ACLReplicationToken: "foxtrot",
		}

		require.NoError(ioutil.WriteFile(tokensFullPath, []byte("[1,2,3]"), 0600))
		err := a.loadTokens(cfg)
		require.Error(err)

		require.Equal("alfa", a.tokens.UserToken())
		require.Equal("bravo", a.tokens.AgentToken())
		require.Equal("charlie", a.tokens.AgentMasterToken())
		require.Equal("foxtrot", a.tokens.ReplicationToken())
	})
}
