package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/consul/agent/config"
	"github.com/hashicorp/consul/agent/token"
)

func (a *Agent) loadTokens(conf *config.RuntimeConfig) error {
	persistedTokens, persistenceErr := a.getPersistedTokens()

	if persistenceErr != nil {
		a.logger.Warn("unable to load persisted tokens", "error", persistenceErr)
	}

	if persistedTokens.Default != "" {
		a.tokens.UpdateUserToken(persistedTokens.Default, token.TokenSourceAPI)

		if conf.ACLToken != "" {
			a.logger.Warn("\"default\" token present in both the configuration and persisted token store, using the persisted token")
		}
	} else {
		a.tokens.UpdateUserToken(conf.ACLToken, token.TokenSourceConfig)
	}

	if persistedTokens.Agent != "" {
		a.tokens.UpdateAgentToken(persistedTokens.Agent, token.TokenSourceAPI)

		if conf.ACLAgentToken != "" {
			a.logger.Warn("\"agent\" token present in both the configuration and persisted token store, using the persisted token")
		}
	} else {
		a.tokens.UpdateAgentToken(conf.ACLAgentToken, token.TokenSourceConfig)
	}

	if persistedTokens.AgentMaster != "" {
		a.tokens.UpdateAgentMasterToken(persistedTokens.AgentMaster, token.TokenSourceAPI)

		if conf.ACLAgentMasterToken != "" {
			a.logger.Warn("\"agent_master\" token present in both the configuration and persisted token store, using the persisted token")
		}
	} else {
		a.tokens.UpdateAgentMasterToken(conf.ACLAgentMasterToken, token.TokenSourceConfig)
	}

	if persistedTokens.Replication != "" {
		a.tokens.UpdateReplicationToken(persistedTokens.Replication, token.TokenSourceAPI)

		if conf.ACLReplicationToken != "" {
			a.logger.Warn("\"replication\" token present in both the configuration and persisted token store, using the persisted token")
		}
	} else {
		a.tokens.UpdateReplicationToken(conf.ACLReplicationToken, token.TokenSourceConfig)
	}

	loadEnterpriseTokens(a.tokens, conf)
	return persistenceErr
}

type persistedTokens struct {
	Replication string `json:"replication,omitempty"`
	AgentMaster string `json:"agent_master,omitempty"`
	Default     string `json:"default,omitempty"`
	Agent       string `json:"agent,omitempty"`
}

func (a *Agent) getPersistedTokens() (*persistedTokens, error) {
	persistedTokens := &persistedTokens{}
	if !a.config.ACLEnableTokenPersistence {
		return persistedTokens, nil
	}

	a.persistedTokensLock.RLock()
	defer a.persistedTokensLock.RUnlock()

	tokensFullPath := filepath.Join(a.config.DataDir, tokensPath)

	buf, err := ioutil.ReadFile(tokensFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// non-existence is not an error we care about
			return persistedTokens, nil
		}
		return persistedTokens, fmt.Errorf("failed reading tokens file %q: %s", tokensFullPath, err)
	}

	if err := json.Unmarshal(buf, persistedTokens); err != nil {
		return persistedTokens, fmt.Errorf("failed to decode tokens file %q: %s", tokensFullPath, err)
	}

	return persistedTokens, nil
}
