package consul

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/serf/serf"

	"github.com/hashicorp/consul/agent/metadata"
	"github.com/hashicorp/consul/agent/structs"
	libserf "github.com/hashicorp/consul/lib/serf"
	"github.com/hashicorp/consul/logging"
	"github.com/hashicorp/consul/types"
)

const (
	// serfEventBacklog is the maximum number of unprocessed Serf Events
	// that will be held in queue before new serf events block.  A
	// blocking serf event queue is a bad thing.
	serfEventBacklog = 256

	// serfEventBacklogWarning is the threshold at which point log
	// warnings will be emitted indicating a problem when processing serf
	// events.
	serfEventBacklogWarning = 200
)

func newClientGossipFromConsulConfig(config *Config, logger hclog.InterceptLogger) (*clientGossip, error) {
	c := config.SerfLANConfig
	c.Init()
	chEvent := make(chan serf.Event, serfEventBacklog)
	c.EventCh = chEvent
	c.ReconnectTimeoutOverride = libserf.NewReconnectOverride(logger)
	serfConfigAddFromConsulConfig(c, config)
	serfConfigAddClient(c, config)
	serfConfigAddLogger(c, logger, logging.LAN)
	serfConfigAddEnterpriseTags(c.Tags)
	if err := serfConfigAddSnapshotPath(c, config.DataDir, serfLANSnapshot); err != nil {
		return nil, err
	}

	serf, err := serf.Create(c)
	if err != nil {
		return nil, err
	}

	return &clientGossip{eventCh: chEvent, serf: serf}, nil
}

type clientGossip struct {
	eventCh chan serf.Event
	serf    *serf.Serf
}

func serfConfigAddFromConsulConfig(conf *serf.Config, cc *Config) {
	conf.NodeName = cc.NodeName
	conf.Tags["dc"] = cc.Datacenter
	conf.Tags["segment"] = cc.Segment
	conf.Tags["id"] = string(cc.NodeID)
	conf.Tags["vsn"] = fmt.Sprintf("%d", cc.ProtocolVersion)
	conf.Tags["vsn_min"] = fmt.Sprintf("%d", ProtocolVersionMin)
	conf.Tags["vsn_max"] = fmt.Sprintf("%d", ProtocolVersionMax)
	conf.Tags["build"] = cc.Build

	if cc.ACLsEnabled {
		// we start in legacy mode and then transition to normal
		// mode once we know the cluster can handle it.
		conf.Tags["acls"] = string(structs.ACLModeLegacy)
	} else {
		conf.Tags["acls"] = string(structs.ACLModeDisabled)
	}

	conf.ProtocolVersion = protocolVersionMap[cc.ProtocolVersion]
	conf.RejoinAfterLeave = cc.RejoinAfterLeave
	conf.Merge = &lanMergeDelegate{
		dc:       cc.Datacenter,
		nodeID:   cc.NodeID,
		nodeName: cc.NodeName,
		segment:  cc.Segment,
	}
}

func serfConfigAddClient(conf *serf.Config, cc *Config) {
	conf.Tags["role"] = "node"
	if cc.AdvertiseReconnectTimeout != 0 {
		conf.Tags[libserf.ReconnectTimeoutTag] = cc.AdvertiseReconnectTimeout.String()
	}
}

func serfConfigAddLogger(conf *serf.Config, logger hclog.InterceptLogger, name string) {
	conf.Logger = logger.
		NamedIntercept(logging.Serf).
		NamedIntercept(name).
		StandardLoggerIntercept(&hclog.StandardLoggerOptions{InferLevels: true})
	conf.MemberlistConfig.Logger = logger.
		NamedIntercept(logging.Memberlist).
		NamedIntercept(name).
		StandardLoggerIntercept(&hclog.StandardLoggerOptions{InferLevels: true})
}

func serfConfigAddSnapshotPath(conf *serf.Config, root, relative string) error {
	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	conf.SnapshotPath = path
	return nil
}

// lanEventHandler is used to handle events from the lan Serf cluster
func (c *Client) lanEventHandler() {
	var numQueuedEvents int
	for {
		numQueuedEvents = len(c.eventCh)
		if numQueuedEvents > serfEventBacklogWarning {
			c.logger.Warn("number of queued serf events above warning threshold",
				"queued_events", numQueuedEvents,
				"warning_threshold", serfEventBacklogWarning,
			)
		}

		select {
		case e := <-c.eventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				c.nodeJoin(e.(serf.MemberEvent))
			case serf.EventMemberLeave, serf.EventMemberFailed, serf.EventMemberReap:
				c.nodeFail(e.(serf.MemberEvent))
			case serf.EventUser:
				c.localEvent(e.(serf.UserEvent))
			case serf.EventMemberUpdate: // Ignore
				c.nodeUpdate(e.(serf.MemberEvent))
			case serf.EventQuery: // Ignore
			default:
				c.logger.Warn("unhandled LAN Serf Event", "event", e)
			}
		case <-c.shutdownCh:
			return
		}
	}
}

// nodeJoin is used to handle join events on the serf cluster
func (c *Client) nodeJoin(me serf.MemberEvent) {
	for _, m := range me.Members {
		ok, parts := metadata.IsConsulServer(m)
		if !ok {
			continue
		}
		if parts.Datacenter != c.config.Datacenter {
			c.logger.Warn("server has joined the wrong cluster: wrong datacenter",
				"server", m.Name,
				"datacenter", parts.Datacenter,
			)
			continue
		}
		c.logger.Info("adding server", "server", parts)
		c.router.AddServer(types.AreaLAN, parts)

		// Trigger the callback
		if c.config.ServerUp != nil {
			c.config.ServerUp()
		}
	}
}

// nodeUpdate is used to handle update events on the serf cluster
func (c *Client) nodeUpdate(me serf.MemberEvent) {
	for _, m := range me.Members {
		ok, parts := metadata.IsConsulServer(m)
		if !ok {
			continue
		}
		if parts.Datacenter != c.config.Datacenter {
			c.logger.Warn("server has joined the wrong cluster: wrong datacenter",
				"server", m.Name,
				"datacenter", parts.Datacenter,
			)
			continue
		}
		c.logger.Info("updating server", "server", parts.String())
		c.router.AddServer(types.AreaLAN, parts)
	}
}

// nodeFail is used to handle fail events on the serf cluster
func (c *Client) nodeFail(me serf.MemberEvent) {
	for _, m := range me.Members {
		ok, parts := metadata.IsConsulServer(m)
		if !ok {
			continue
		}
		c.logger.Info("removing server", "server", parts.String())
		c.router.RemoveServer(types.AreaLAN, parts)
	}
}

// localEvent is called when we receive an event on the local Serf
func (c *Client) localEvent(event serf.UserEvent) {
	// Handle only consul events
	if !strings.HasPrefix(event.Name, "consul:") {
		return
	}

	switch name := event.Name; {
	case name == newLeaderEvent:
		c.logger.Info("New leader elected", "payload", string(event.Payload))

		// Trigger the callback
		if c.config.ServerUp != nil {
			c.config.ServerUp()
		}
	case isUserEvent(name):
		event.Name = rawUserEventName(name)
		c.logger.Debug("user event", "name", event.Name)

		// Trigger the callback
		if c.config.UserEventHandler != nil {
			c.config.UserEventHandler(event)
		}
	default:
		if !c.handleEnterpriseUserEvents(event) {
			c.logger.Warn("Unhandled local event", "event", event)
		}
	}
}
