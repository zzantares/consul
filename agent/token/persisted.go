package token

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/consul/lib/file"
	"github.com/hashicorp/go-hclog"
)

type persistedTokens struct {
	Replication string `json:"replication,omitempty"`
	AgentMaster string `json:"agent_master,omitempty"`
	Default     string `json:"default,omitempty"`
	Agent       string `json:"agent,omitempty"`
}

type Config struct {
	EnablePersistence bool
	DataDir           string

	ACLToken            string
	ACLAgentToken       string
	ACLAgentMasterToken string
	ACLReplicationToken string
}

type PersistedStore struct {
	*Store
	cfg    Config
	logger hclog.Logger
	// lock is used to synchronize access to the persisted token store within
	// the data directory. This will prevent loading while writing as well as
	// multiple concurrent writes.
	lock sync.RWMutex
}

func NewPersistenceStore(cfg Config, logger hclog.Logger) (*PersistedStore, error) {
	p := &PersistedStore{
		Store:  new(Store),
		cfg:    cfg,
		logger: logger,
		lock:   sync.RWMutex{},
	}

	tokens, err := p.loadFromFile()
	if err != nil {
		p.logger.Warn("unable to load persisted tokens", "error", err)
	}

	if tokens.Default != "" {
		p.Store.UpdateUserToken(tokens.Default, TokenSourceAPI)

		if cfg.ACLToken != "" {
			p.logger.Warn("\"default\" token present in both the cfg.guration and persisted token store, using the persisted token")
		}
	} else {
		p.Store.UpdateUserToken(cfg.ACLToken, TokenSourceConfig)
	}

	if tokens.Agent != "" {
		p.Store.UpdateAgentToken(tokens.Agent, TokenSourceAPI)

		if cfg.ACLAgentToken != "" {
			p.logger.Warn("\"agent\" token present in both the cfg.guration and persisted token store, using the persisted token")
		}
	} else {
		p.Store.UpdateAgentToken(cfg.ACLAgentToken, TokenSourceConfig)
	}

	if tokens.AgentMaster != "" {
		p.Store.UpdateAgentMasterToken(tokens.AgentMaster, TokenSourceAPI)

		if cfg.ACLAgentMasterToken != "" {
			p.logger.Warn("\"agent_master\" token present in both the cfg.guration and persisted token store, using the persisted token")
		}
	} else {
		p.Store.UpdateAgentMasterToken(cfg.ACLAgentMasterToken, TokenSourceConfig)
	}

	if tokens.Replication != "" {
		p.Store.UpdateReplicationToken(tokens.Replication, TokenSourceAPI)

		if cfg.ACLReplicationToken != "" {
			p.logger.Warn("\"replication\" token present in both the cfg.guration and persisted token store, using the persisted token")
		}
	} else {
		p.Store.UpdateReplicationToken(cfg.ACLReplicationToken, TokenSourceConfig)
	}

	return p, err
}

var tokensPath = "acl-tokens.json"

// TODO: test case
func (p *PersistedStore) loadFromFile() (*persistedTokens, error) {
	tokens := &persistedTokens{}
	if !p.cfg.EnablePersistence {
		return tokens, nil
	}

	p.lock.RLock()
	defer p.lock.RUnlock()

	tokensFullPath := filepath.Join(p.cfg.DataDir, tokensPath)
	buf, err := ioutil.ReadFile(tokensFullPath)
	switch {
	case os.IsNotExist(err):
		// non-existence is not an error we care about
		return tokens, nil
	case err != nil:
		return tokens, fmt.Errorf("failed reading tokens file %q: %w", tokensFullPath, err)
	}

	if err := json.Unmarshal(buf, tokens); err != nil {
		return tokens, fmt.Errorf("failed to decode tokens file %q: %w", tokensFullPath, err)
	}

	return tokens, nil
}

// WithPersistenceLock executes f while hold a lock. If f returns without an
// error, the tokens in Store will be persisted to the tokens file.
//
// The lock is held so that the writes are persisted before some other thread
// can change the value.
func (p *PersistedStore) WithPersistenceLock(f func() error) error {
	if !p.cfg.EnablePersistence {
		return f()
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	if err := f(); err != nil {
		return err
	}

	return p.saveToFile()
}

// TODO: test case
func (p *PersistedStore) saveToFile() error {
	tokens := persistedTokens{}
	if tok, source := p.Store.UserTokenAndSource(); tok != "" && source == TokenSourceAPI {
		tokens.Default = tok
	}

	if tok, source := p.Store.AgentTokenAndSource(); tok != "" && source == TokenSourceAPI {
		tokens.Agent = tok
	}

	if tok, source := p.Store.AgentMasterTokenAndSource(); tok != "" && source == TokenSourceAPI {
		tokens.AgentMaster = tok
	}

	if tok, source := p.Store.ReplicationTokenAndSource(); tok != "" && source == TokenSourceAPI {
		tokens.Replication = tok
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		p.logger.Warn("failed to persist tokens", "error", err)
		return fmt.Errorf("Failed to marshal tokens for persistence: %v", err)
	}

	if err := file.WriteAtomicWithPerms(filepath.Join(p.cfg.DataDir, tokensPath), data, 0700, 0600); err != nil {
		p.logger.Warn("failed to persist tokens", "error", err)
		return fmt.Errorf("Failed to persist tokens - %v", err)
	}
	return nil
}
