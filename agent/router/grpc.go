package router

import "github.com/hashicorp/consul/agent/metadata"

// ServerTracker is called when Router is notified of a server being added or
// removed.
type ServerTracker interface {
	Rebalancer
	AddServer(*metadata.Server)
	RemoveServer(*metadata.Server)
}

// Rebalancer is called periodically to re-order the servers so that the load on the
// servers is evenly balanced.
type Rebalancer interface {
	Rebalance()
}

// NoOpServerTracker is a ServerTracker that does nothing. Used when gRPC is not
// enabled.
type NoOpServerTracker struct{}

// Rebalance does nothing
func (NoOpServerTracker) Rebalance() {}

// AddServer does nothing
func (NoOpServerTracker) AddServer(*metadata.Server) {}

// RemoveServer does nothing
func (NoOpServerTracker) RemoveServer(*metadata.Server) {}
