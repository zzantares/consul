package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/consul/agent/config"
	"github.com/hashicorp/consul/agent/dns"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/types"
)

// addServiceRequest is the union of arguments for calling both
// addServiceLocked and addServiceInternal. The overlap was significant enough
// to warrant merging them and indicating which fields are meant to be set only
// in one of the two contexts.
//
// Before using the request struct one of the fixupFor*() methods should be
// invoked to clear irrelevant fields.
//
// The ServiceManager.AddService signature is largely just a passthrough for
// addServiceLocked and should be treated as such.
type addServiceRequest struct {
	service               *structs.NodeService
	chkTypes              []*structs.CheckType
	previousDefaults      *structs.ServiceConfigResponse // just for: addServiceLocked
	waitForCentralConfig  bool                           // just for: addServiceLocked
	persistService        *structs.NodeService           // just for: addServiceInternal
	persistDefaults       *structs.ServiceConfigResponse // just for: addServiceInternal
	persist               bool
	persistServiceConfig  bool
	token                 string
	replaceExistingChecks bool
	source                configSource
	snap                  map[structs.CheckID]*structs.HealthCheck
}

func (r *addServiceRequest) fixupForAddServiceLocked() {
	r.persistService = nil
	r.persistDefaults = nil
}

func (r *addServiceRequest) fixupForAddServiceInternal() {
	r.previousDefaults = nil
	r.waitForCentralConfig = false
}

// AddServiceAndReplaceChecks is used to add a service entry and its check. Any check for this service missing from chkTypes will be deleted.
// This entry is persistent and the agent will make a best effort to
// ensure it is registered
func (a *Agent) AddServiceAndReplaceChecks(service *structs.NodeService, chkTypes []*structs.CheckType, persist bool, token string, source configSource) error {
	a.stateLock.Lock()
	defer a.stateLock.Unlock()
	return a.addServiceLocked(&addServiceRequest{
		service:               service,
		chkTypes:              chkTypes,
		previousDefaults:      nil,
		waitForCentralConfig:  true,
		persist:               persist,
		persistServiceConfig:  true,
		token:                 token,
		replaceExistingChecks: true,
		source:                source,
		snap:                  a.State.Checks(structs.WildcardEnterpriseMeta()),
	})
}

// AddService is used to add a service entry.
// This entry is persistent and the agent will make a best effort to
// ensure it is registered
func (a *Agent) AddService(service *structs.NodeService, chkTypes []*structs.CheckType, persist bool, token string, source configSource) error {
	a.stateLock.Lock()
	defer a.stateLock.Unlock()
	return a.addServiceLocked(&addServiceRequest{
		service:               service,
		chkTypes:              chkTypes,
		previousDefaults:      nil,
		waitForCentralConfig:  true,
		persist:               persist,
		persistServiceConfig:  true,
		token:                 token,
		replaceExistingChecks: false,
		source:                source,
		snap:                  a.State.Checks(structs.WildcardEnterpriseMeta()),
	})
}

// addServiceLocked adds a service entry to the service manager if enabled, or directly
// to the local state if it is not. This function assumes the state lock is already held.
func (a *Agent) addServiceLocked(req *addServiceRequest) error {
	req.fixupForAddServiceLocked()

	req.service.EnterpriseMeta.Normalize()

	if err := a.validateService(req.service, req.chkTypes); err != nil {
		return err
	}

	if a.config.EnableCentralServiceConfig {
		return a.serviceManager.AddService(req)
	}

	// previousDefaults are ignored here because they are only relevant for central config.
	req.persistService = nil
	req.persistDefaults = nil
	req.persistServiceConfig = false

	return a.addServiceInternal(req)
}

// validateService validates an service and its checks, either returning an error or emitting a
// warning based on the nature of the error.
func (a *Agent) validateService(service *structs.NodeService, chkTypes []*structs.CheckType) error {
	if service.Service == "" {
		return fmt.Errorf("Service name missing")
	}
	if service.ID == "" && service.Service != "" {
		service.ID = service.Service
	}
	for _, check := range chkTypes {
		if err := check.Validate(); err != nil {
			return fmt.Errorf("Check is not valid: %v", err)
		}
	}

	// Set default weights if not specified. This is important as it ensures AE
	// doesn't consider the service different since it has nil weights.
	if service.Weights == nil {
		service.Weights = &structs.Weights{Passing: 1, Warning: 1}
	}

	// Warn if the service name is incompatible with DNS
	if dns.InvalidNameRe.MatchString(service.Service) {
		a.logger.Warn("Service name will not be discoverable "+
			"via DNS due to invalid characters. Valid characters include "+
			"all alpha-numerics and dashes.",
			"service", service.Service,
		)
	} else if len(service.Service) > dns.MaxLabelLength {
		a.logger.Warn("Service name will not be discoverable "+
			"via DNS due to it being too long. Valid lengths are between "+
			"1 and 63 bytes.",
			"service", service.Service,
		)
	}

	// Warn if any tags are incompatible with DNS
	for _, tag := range service.Tags {
		if dns.InvalidNameRe.MatchString(tag) {
			a.logger.Debug("Service tag will not be discoverable "+
				"via DNS due to invalid characters. Valid characters include "+
				"all alpha-numerics and dashes.",
				"tag", tag,
			)
		} else if len(tag) > dns.MaxLabelLength {
			a.logger.Debug("Service tag will not be discoverable "+
				"via DNS due to it being too long. Valid lengths are between "+
				"1 and 63 bytes.",
				"tag", tag,
			)
		}
	}

	// Check IPv4/IPv6 tagged addresses
	if service.TaggedAddresses != nil {
		if sa, ok := service.TaggedAddresses[structs.TaggedAddressLANIPv4]; ok {
			ip := net.ParseIP(sa.Address)
			if ip == nil || ip.To4() == nil {
				return fmt.Errorf("Service tagged address %q must be a valid ipv4 address", structs.TaggedAddressLANIPv4)
			}
		}
		if sa, ok := service.TaggedAddresses[structs.TaggedAddressWANIPv4]; ok {
			ip := net.ParseIP(sa.Address)
			if ip == nil || ip.To4() == nil {
				return fmt.Errorf("Service tagged address %q must be a valid ipv4 address", structs.TaggedAddressWANIPv4)
			}
		}
		if sa, ok := service.TaggedAddresses[structs.TaggedAddressLANIPv6]; ok {
			ip := net.ParseIP(sa.Address)
			if ip == nil || ip.To4() != nil {
				return fmt.Errorf("Service tagged address %q must be a valid ipv6 address", structs.TaggedAddressLANIPv6)
			}
		}
		if sa, ok := service.TaggedAddresses[structs.TaggedAddressLANIPv6]; ok {
			ip := net.ParseIP(sa.Address)
			if ip == nil || ip.To4() != nil {
				return fmt.Errorf("Service tagged address %q must be a valid ipv6 address", structs.TaggedAddressLANIPv6)
			}
		}
	}

	return nil
}

// addServiceInternal adds the given service and checks to the local state.
func (a *Agent) addServiceInternal(req *addServiceRequest) error {
	req.fixupForAddServiceInternal()
	var (
		service               = req.service
		chkTypes              = req.chkTypes
		persistService        = req.persistService
		persistDefaults       = req.persistDefaults
		persist               = req.persist
		persistServiceConfig  = req.persistServiceConfig
		token                 = req.token
		replaceExistingChecks = req.replaceExistingChecks
		source                = req.source
		snap                  = req.snap
	)

	// Pause the service syncs during modification
	a.PauseSync()
	defer a.ResumeSync()

	// Set default tagged addresses
	serviceIP := net.ParseIP(service.Address)
	serviceAddressIs4 := serviceIP != nil && serviceIP.To4() != nil
	serviceAddressIs6 := serviceIP != nil && serviceIP.To4() == nil
	if service.TaggedAddresses == nil {
		service.TaggedAddresses = map[string]structs.ServiceAddress{}
	}
	if _, ok := service.TaggedAddresses[structs.TaggedAddressLANIPv4]; !ok && serviceAddressIs4 {
		service.TaggedAddresses[structs.TaggedAddressLANIPv4] = structs.ServiceAddress{Address: service.Address, Port: service.Port}
	}
	if _, ok := service.TaggedAddresses[structs.TaggedAddressWANIPv4]; !ok && serviceAddressIs4 {
		service.TaggedAddresses[structs.TaggedAddressWANIPv4] = structs.ServiceAddress{Address: service.Address, Port: service.Port}
	}
	if _, ok := service.TaggedAddresses[structs.TaggedAddressLANIPv6]; !ok && serviceAddressIs6 {
		service.TaggedAddresses[structs.TaggedAddressLANIPv6] = structs.ServiceAddress{Address: service.Address, Port: service.Port}
	}
	if _, ok := service.TaggedAddresses[structs.TaggedAddressWANIPv6]; !ok && serviceAddressIs6 {
		service.TaggedAddresses[structs.TaggedAddressWANIPv6] = structs.ServiceAddress{Address: service.Address, Port: service.Port}
	}

	var checks []*structs.HealthCheck

	// all the checks must be associated with the same enterprise meta of the service
	// so this map can just use the main CheckID for indexing
	existingChecks := map[structs.CheckID]bool{}
	for _, check := range a.State.ChecksForService(service.CompoundServiceID(), false) {
		existingChecks[check.CompoundCheckID()] = false
	}

	// Create an associated health check
	for i, chkType := range chkTypes {
		checkID := string(chkType.CheckID)
		if checkID == "" {
			checkID = fmt.Sprintf("service:%s", service.ID)
			if len(chkTypes) > 1 {
				checkID += fmt.Sprintf(":%d", i+1)
			}
		}

		cid := structs.NewCheckID(types.CheckID(checkID), &service.EnterpriseMeta)
		existingChecks[cid] = true

		name := chkType.Name
		if name == "" {
			name = fmt.Sprintf("Service '%s' check", service.Service)
		}
		check := &structs.HealthCheck{
			Node:           a.config.NodeName,
			CheckID:        types.CheckID(checkID),
			Name:           name,
			Status:         api.HealthCritical,
			Notes:          chkType.Notes,
			ServiceID:      service.ID,
			ServiceName:    service.Service,
			ServiceTags:    service.Tags,
			Type:           chkType.Type(),
			EnterpriseMeta: service.EnterpriseMeta,
		}
		if chkType.Status != "" {
			check.Status = chkType.Status
		}

		// Restore the fields from the snapshot.
		prev, ok := snap[cid]
		if ok {
			check.Output = prev.Output
			check.Status = prev.Status
		}

		checks = append(checks, check)
	}

	// cleanup, store the ids of services and checks that weren't previously
	// registered so we clean them up if something fails halfway through the
	// process.
	var cleanupServices []structs.ServiceID
	var cleanupChecks []structs.CheckID

	sid := service.CompoundServiceID()
	if s := a.State.Service(sid); s == nil {
		cleanupServices = append(cleanupServices, sid)
	}

	for _, check := range checks {
		cid := check.CompoundCheckID()
		if c := a.State.Check(cid); c == nil {
			cleanupChecks = append(cleanupChecks, cid)
		}
	}

	err := a.State.AddServiceWithChecks(service, checks, token)
	if err != nil {
		a.cleanupRegistration(cleanupServices, cleanupChecks)
		return err
	}

	for i := range checks {
		if err := a.addCheck(checks[i], chkTypes[i], service, token, source); err != nil {
			a.cleanupRegistration(cleanupServices, cleanupChecks)
			return err
		}

		if persist && a.config.DataDir != "" {
			if err := a.persistCheck(checks[i], chkTypes[i], source); err != nil {
				a.cleanupRegistration(cleanupServices, cleanupChecks)
				return err

			}
		}
	}

	// If a proxy service wishes to expose checks, check targets need to be rerouted to the proxy listener
	// This needs to be called after chkTypes are added to the agent, to avoid being overwritten
	psid := structs.NewServiceID(service.Proxy.DestinationServiceID, &service.EnterpriseMeta)

	if service.Proxy.Expose.Checks {
		err := a.rerouteExposedChecks(psid, service.Address)
		if err != nil {
			a.logger.Warn("failed to reroute L7 checks to exposed proxy listener")
		}
	} else {
		// Reset check targets if proxy was re-registered but no longer wants to expose checks
		// If the proxy is being registered for the first time then this is a no-op
		a.resetExposedChecks(psid)
	}

	if persistServiceConfig && a.config.DataDir != "" {
		var err error
		if persistDefaults != nil {
			err = a.persistServiceConfig(service.CompoundServiceID(), persistDefaults)
		} else {
			err = a.purgeServiceConfig(service.CompoundServiceID())
		}

		if err != nil {
			a.cleanupRegistration(cleanupServices, cleanupChecks)
			return err
		}
	}

	// Persist the service to a file
	if persist && a.config.DataDir != "" {
		if persistService == nil {
			persistService = service
		}

		if err := a.persistService(persistService, source); err != nil {
			a.cleanupRegistration(cleanupServices, cleanupChecks)
			return err
		}
	}

	if replaceExistingChecks {
		for checkID, keep := range existingChecks {
			if !keep {
				a.removeCheckLocked(checkID, persist)
			}
		}
	}

	return nil
}

// cleanupRegistration is called on  registration error to ensure no there are no
// leftovers after a partial failure
func (a *Agent) cleanupRegistration(serviceIDs []structs.ServiceID, checksIDs []structs.CheckID) {
	for _, s := range serviceIDs {
		if err := a.State.RemoveService(s); err != nil {
			a.logger.Error("failed to remove service during cleanup",
				"service", s.String(),
				"error", err,
			)
		}
		if err := a.purgeService(s); err != nil {
			a.logger.Error("failed to purge service file during cleanup",
				"service", s.String(),
				"error", err,
			)
		}
		if err := a.purgeServiceConfig(s); err != nil {
			a.logger.Error("failed to purge service config file during cleanup",
				"service", s,
				"error", err,
			)
		}
		if err := a.removeServiceSidecars(s, true); err != nil {
			a.logger.Error("service registration: cleanup: failed remove sidecars for", "service", s, "error", err)
		}
	}

	for _, c := range checksIDs {
		a.cancelCheckMonitors(c)
		if err := a.State.RemoveCheck(c); err != nil {
			a.logger.Error("failed to remove check during cleanup",
				"check", c.String(),
				"error", err,
			)
		}
		if err := a.purgeCheck(c); err != nil {
			a.logger.Error("failed to purge check file during cleanup",
				"check", c.String(),
				"error", err,
			)
		}
	}
}

// reapServicesInternal does a single pass, looking for services to reap.
func (a *Agent) reapServicesInternal() {
	reaped := make(map[structs.ServiceID]bool)
	for checkID, cs := range a.State.CriticalCheckStates(structs.WildcardEnterpriseMeta()) {
		serviceID := cs.Check.CompoundServiceID()

		// There's nothing to do if there's no service.
		if serviceID.ID == "" {
			continue
		}

		// There might be multiple checks for one service, so
		// we don't need to reap multiple times.
		if reaped[serviceID] {
			continue
		}

		// See if there's a timeout.
		// todo(fs): this looks fishy... why is there another data structure in the agent with its own lock?
		a.stateLock.Lock()
		timeout := a.checkReapAfter[checkID]
		a.stateLock.Unlock()

		// Reap, if necessary. We keep track of which service
		// this is so that we won't try to remove it again.
		if timeout > 0 && cs.CriticalFor() > timeout {
			reaped[serviceID] = true
			if err := a.RemoveService(serviceID); err != nil {
				a.logger.Error("unable to deregister service after check has been critical for too long",
					"service", serviceID.String(),
					"check", checkID.String(),
					"error", err)
			} else {
				a.logger.Info("Check for service has been critical for too long; deregistered service",
					"service", serviceID.String(),
					"check", checkID.String(),
				)
			}
		}
	}
}

// reapServices is a long running goroutine that looks for checks that have been
// critical too long and deregisters their associated services.
func (a *Agent) reapServices() {
	for {
		select {
		case <-time.After(a.config.CheckReapInterval):
			a.reapServicesInternal()

		case <-a.shutdownCh:
			return
		}
	}
}

// RemoveService is used to remove a service entry.
// The agent will make a best effort to ensure it is deregistered
func (a *Agent) RemoveService(serviceID structs.ServiceID) error {
	return a.removeService(serviceID, true)
}

func (a *Agent) removeService(serviceID structs.ServiceID, persist bool) error {
	a.stateLock.Lock()
	defer a.stateLock.Unlock()
	return a.removeServiceLocked(serviceID, persist)
}

// removeServiceLocked is used to remove a service entry.
// The agent will make a best effort to ensure it is deregistered
func (a *Agent) removeServiceLocked(serviceID structs.ServiceID, persist bool) error {
	// Validate ServiceID
	if serviceID.ID == "" {
		return fmt.Errorf("ServiceID missing")
	}

	// Shut down the config watch in the service manager if enabled.
	if a.config.EnableCentralServiceConfig {
		a.serviceManager.RemoveService(serviceID)
	}

	// Reset the HTTP check targets if they were exposed through a proxy
	// If this is not a proxy or checks were not exposed then this is a no-op
	svc := a.State.Service(serviceID)

	if svc != nil {
		psid := structs.NewServiceID(svc.Proxy.DestinationServiceID, &svc.EnterpriseMeta)
		a.resetExposedChecks(psid)
	}

	checks := a.State.ChecksForService(serviceID, false)
	var checkIDs []structs.CheckID
	for id := range checks {
		checkIDs = append(checkIDs, id)
	}

	// Remove service immediately
	if err := a.State.RemoveServiceWithChecks(serviceID, checkIDs); err != nil {
		a.logger.Warn("Failed to deregister service",
			"service", serviceID.String(),
			"error", err,
		)
		return nil
	}

	// Remove the service from the data dir
	if persist {
		if err := a.purgeService(serviceID); err != nil {
			return err
		}
		if err := a.purgeServiceConfig(serviceID); err != nil {
			return err
		}
	}

	// Deregister any associated health checks
	for checkID := range checks {
		if err := a.removeCheckLocked(checkID, persist); err != nil {
			return err
		}
	}

	a.logger.Debug("removed service", "service", serviceID.String())

	// If any Sidecar services exist for the removed service ID, remove them too.
	return a.removeServiceSidecars(serviceID, persist)
}

func (a *Agent) removeServiceSidecars(serviceID structs.ServiceID, persist bool) error {
	sidecarSID := structs.NewServiceID(a.sidecarServiceID(serviceID.ID), &serviceID.EnterpriseMeta)
	if sidecar := a.State.Service(sidecarSID); sidecar != nil {
		// Double check that it's not just an ID collision and we actually added
		// this from a sidecar.
		if sidecar.LocallyRegisteredAsSidecar {
			// Remove it!
			err := a.removeServiceLocked(sidecarSID, persist)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// loadServices will load service definitions from configuration and persisted
// definitions on disk, and load them into the local agent.
func (a *Agent) loadServices(conf *config.RuntimeConfig, snap map[structs.CheckID]*structs.HealthCheck) error {
	// Load any persisted service configs so we can feed those into the initial
	// registrations below.
	persistedServiceConfigs, err := a.readPersistedServiceConfigs()
	if err != nil {
		return err
	}

	// Register the services from config
	for _, service := range conf.Services {
		ns := service.NodeService()
		chkTypes, err := service.CheckTypes()
		if err != nil {
			return fmt.Errorf("Failed to validate checks for service %q: %v", service.Name, err)
		}

		// Grab and validate sidecar if there is one too
		sidecar, sidecarChecks, sidecarToken, err := a.sidecarServiceFromNodeService(ns, service.Token)
		if err != nil {
			return fmt.Errorf("Failed to validate sidecar for service %q: %v", service.Name, err)
		}

		// Remove sidecar from NodeService now it's done it's job it's just a config
		// syntax sugar and shouldn't be persisted in local or server state.
		ns.Connect.SidecarService = nil

		sid := ns.CompoundServiceID()
		err = a.addServiceLocked(&addServiceRequest{
			service:               ns,
			chkTypes:              chkTypes,
			previousDefaults:      persistedServiceConfigs[sid],
			waitForCentralConfig:  false, // exclusively use cached values
			persist:               false, // don't rewrite the file with the same data we just read
			persistServiceConfig:  false, // don't rewrite the file with the same data we just read
			token:                 service.Token,
			replaceExistingChecks: false, // do default behavior
			source:                ConfigSourceLocal,
			snap:                  snap,
		})
		if err != nil {
			return fmt.Errorf("Failed to register service %q: %v", service.Name, err)
		}

		// If there is a sidecar service, register that too.
		if sidecar != nil {
			sidecarServiceID := sidecar.CompoundServiceID()
			err = a.addServiceLocked(&addServiceRequest{
				service:               sidecar,
				chkTypes:              sidecarChecks,
				previousDefaults:      persistedServiceConfigs[sidecarServiceID],
				waitForCentralConfig:  false, // exclusively use cached values
				persist:               false, // don't rewrite the file with the same data we just read
				persistServiceConfig:  false, // don't rewrite the file with the same data we just read
				token:                 sidecarToken,
				replaceExistingChecks: false, // do default behavior
				source:                ConfigSourceLocal,
				snap:                  snap,
			})
			if err != nil {
				return fmt.Errorf("Failed to register sidecar for service %q: %v", service.Name, err)
			}
		}
	}

	// Load any persisted services
	svcDir := filepath.Join(a.config.DataDir, servicesDir)
	files, err := ioutil.ReadDir(svcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("Failed reading services dir %q: %s", svcDir, err)
	}
	for _, fi := range files {
		// Skip all dirs
		if fi.IsDir() {
			continue
		}

		// Skip all partially written temporary files
		if strings.HasSuffix(fi.Name(), "tmp") {
			a.logger.Warn("Ignoring temporary service file", "file", fi.Name())
			continue
		}

		// Read the contents into a buffer
		file := filepath.Join(svcDir, fi.Name())
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed reading service file %q: %s", file, err)
		}

		// Try decoding the service definition
		var p persistedService
		if err := json.Unmarshal(buf, &p); err != nil {
			// Backwards-compatibility for pre-0.5.1 persisted services
			if err := json.Unmarshal(buf, &p.Service); err != nil {
				a.logger.Error("Failed decoding service file",
					"file", file,
					"error", err,
				)
				continue
			}
		}

		// Restore LocallyRegisteredAsSidecar, see persistedService.LocallyRegisteredAsSidecar
		p.Service.LocallyRegisteredAsSidecar = p.LocallyRegisteredAsSidecar

		serviceID := p.Service.CompoundServiceID()

		source, ok := ConfigSourceFromName(p.Source)
		if !ok {
			a.logger.Warn("service exists with invalid source, purging",
				"service", serviceID.String(),
				"source", p.Source,
			)
			if err := a.purgeService(serviceID); err != nil {
				return fmt.Errorf("failed purging service %q: %s", serviceID, err)
			}
			if err := a.purgeServiceConfig(serviceID); err != nil {
				return fmt.Errorf("failed purging service config %q: %s", serviceID, err)
			}
			continue
		}

		if a.State.Service(serviceID) != nil {
			// Purge previously persisted service. This allows config to be
			// preferred over services persisted from the API.
			a.logger.Debug("service exists, not restoring from file",
				"service", serviceID.String(),
				"file", file,
			)
			if err := a.purgeService(serviceID); err != nil {
				return fmt.Errorf("failed purging service %q: %s", serviceID.String(), err)
			}
			if err := a.purgeServiceConfig(serviceID); err != nil {
				return fmt.Errorf("failed purging service config %q: %s", serviceID.String(), err)
			}
		} else {
			a.logger.Debug("restored service definition from file",
				"service", serviceID.String(),
				"file", file,
			)
			err = a.addServiceLocked(&addServiceRequest{
				service:               p.Service,
				chkTypes:              nil,
				previousDefaults:      persistedServiceConfigs[serviceID],
				waitForCentralConfig:  false, // exclusively use cached values
				persist:               false, // don't rewrite the file with the same data we just read
				persistServiceConfig:  false, // don't rewrite the file with the same data we just read
				token:                 p.Token,
				replaceExistingChecks: false, // do default behavior
				source:                source,
				snap:                  snap,
			})
			if err != nil {
				return fmt.Errorf("failed adding service %q: %s", serviceID, err)
			}
		}
	}

	for serviceID := range persistedServiceConfigs {
		if a.State.Service(serviceID) == nil {
			// This can be cleaned up now.
			if err := a.purgeServiceConfig(serviceID); err != nil {
				return fmt.Errorf("failed purging service config %q: %s", serviceID, err)
			}
		}
	}

	return nil
}

// unloadServices will deregister all services.
func (a *Agent) unloadServices() error {
	for id := range a.State.Services(structs.WildcardEnterpriseMeta()) {
		if err := a.removeServiceLocked(id, false); err != nil {
			return fmt.Errorf("Failed deregistering service '%s': %v", id, err)
		}
	}
	return nil
}
