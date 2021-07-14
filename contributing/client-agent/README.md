# Agent


- command/agent/agent.go
- config.Load (link to config section)
- Agent.start
- delegate (Client/Server)


## Client Agent responsibilities (not actually entirely in client agents)

- agent/cache - is a read cache with duplicate date when a server agent
- agent/local (local state)
- agent/ae - anti-entropy sync, sync agent/local to the catalog
- link to service-discovery/health-checks

- also runs in server agents
- agent/local and AE do need to exist in servers, because agent/local is the canonical
  source of truth for service registrations.
