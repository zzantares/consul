package token

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/hashicorp/consul/lib/file"
)

type persistedTokens struct {
	Replication string `json:"replication,omitempty"`
	AgentMaster string `json:"agent_master,omitempty"`
	Default     string `json:"default,omitempty"`
	Agent       string `json:"agent,omitempty"`
}

type fileStore struct {
	filename string
	logger   Logger
	// lock is used to synchronize access to the persisted token store within
	// the data directory. This will prevent loading while writing as well as
	// multiple concurrent writes.
	lock sync.RWMutex
}

func (p *fileStore) load(s *Store, cfg Config) error {
	p.lock.RLock()
	defer p.lock.RUnlock()
	tokens, err := readPersistedFromFile(p.filename)
	if err != nil {
		p.logger.Warn("unable to load persisted tokens", "error", err)
	}
	loadTokens(s, cfg, tokens, p.logger)
	return err
}

func loadTokens(s *Store, cfg Config, tokens persistedTokens, logger Logger) {
	if tokens.Default != "" {
		s.UpdateUserToken(tokens.Default, TokenSourceAPI)

		if cfg.ACLDefaultToken != "" {
			logger.Warn("\"default\" token present in both the configuration and persisted token store, using the persisted token")
		}
	} else {
		s.UpdateUserToken(cfg.ACLDefaultToken, TokenSourceConfig)
	}

	if tokens.Agent != "" {
		s.UpdateAgentToken(tokens.Agent, TokenSourceAPI)

		if cfg.ACLAgentToken != "" {
			logger.Warn("\"agent\" token present in both the configuration and persisted token store, using the persisted token")
		}
	} else {
		s.UpdateAgentToken(cfg.ACLAgentToken, TokenSourceConfig)
	}

	if tokens.AgentMaster != "" {
		s.UpdateAgentMasterToken(tokens.AgentMaster, TokenSourceAPI)

		if cfg.ACLAgentMasterToken != "" {
			logger.Warn("\"agent_master\" token present in both the configuration and persisted token store, using the persisted token")
		}
	} else {
		s.UpdateAgentMasterToken(cfg.ACLAgentMasterToken, TokenSourceConfig)
	}

	if tokens.Replication != "" {
		s.UpdateReplicationToken(tokens.Replication, TokenSourceAPI)

		if cfg.ACLReplicationToken != "" {
			logger.Warn("\"replication\" token present in both the configuration and persisted token store, using the persisted token")
		}
	} else {
		s.UpdateReplicationToken(cfg.ACLReplicationToken, TokenSourceConfig)
	}

	loadEnterpriseTokens(s, cfg)
}

func readPersistedFromFile(filename string) (persistedTokens, error) {
	tokens := persistedTokens{}

	buf, err := ioutil.ReadFile(filename)
	switch {
	case os.IsNotExist(err):
		// non-existence is not an error we care about
		return tokens, nil
	case err != nil:
		return tokens, fmt.Errorf("failed reading tokens file %q: %w", filename, err)
	}

	if err := json.Unmarshal(buf, &tokens); err != nil {
		return tokens, fmt.Errorf("failed to decode tokens file %q: %w", filename, err)
	}

	return tokens, nil
}

func (p *fileStore) withPersistenceLock(s *Store, f func() error) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if err := f(); err != nil {
		return err
	}

	return p.saveToFile(s)
}

// TODO: test case
func (p *fileStore) saveToFile(s *Store) error {
	tokens := persistedTokens{}
	if tok, source := s.UserTokenAndSource(); tok != "" && source == TokenSourceAPI {
		tokens.Default = tok
	}

	if tok, source := s.AgentTokenAndSource(); tok != "" && source == TokenSourceAPI {
		tokens.Agent = tok
	}

	if tok, source := s.AgentMasterTokenAndSource(); tok != "" && source == TokenSourceAPI {
		tokens.AgentMaster = tok
	}

	if tok, source := s.ReplicationTokenAndSource(); tok != "" && source == TokenSourceAPI {
		tokens.Replication = tok
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		p.logger.Warn("failed to persist tokens", "error", err)
		return fmt.Errorf("Failed to marshal tokens for persistence: %v", err)
	}

	if err := file.WriteAtomicWithPerms(p.filename, data, 0700, 0600); err != nil {
		p.logger.Warn("failed to persist tokens", "error", err)
		return fmt.Errorf("Failed to persist tokens - %v", err)
	}
	return nil
}
