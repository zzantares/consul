package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/consul/agent/checks"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/consul/lib/file"
)

const (
	servicesDir      = "services"
	serviceConfigDir = "services/configs"
)

// persistedService is used to wrap a service definition and bundle it
// with an ACL token so we can restore both at a later agent start.
type persistedService struct {
	Token   string
	Service *structs.NodeService
	Source  string
	// whether this service was registered as a sidecar, see structs.NodeService
	// we store this field here because it is excluded from json serialization
	// to exclude it from API output, but we need it to properly deregister
	// persisted sidecars.
	LocallyRegisteredAsSidecar bool `json:",omitempty"`
}

// persistService saves a service definition to a JSON file in the data dir
func (a *Agent) persistService(service *structs.NodeService, source configSource) error {
	svcID := service.CompoundServiceID()
	svcPath := filepath.Join(a.config.DataDir, servicesDir, svcID.StringHash())

	wrapped := persistedService{
		Token:                      a.State.ServiceToken(service.CompoundServiceID()),
		Service:                    service,
		Source:                     source.String(),
		LocallyRegisteredAsSidecar: service.LocallyRegisteredAsSidecar,
	}
	encoded, err := json.Marshal(wrapped)
	if err != nil {
		return err
	}

	return file.WriteAtomic(svcPath, encoded)
}

// purgeService removes a persisted service definition file from the data dir
func (a *Agent) purgeService(serviceID structs.ServiceID) error {
	svcPath := filepath.Join(a.config.DataDir, servicesDir, serviceID.StringHash())
	if _, err := os.Stat(svcPath); err == nil {
		return os.Remove(svcPath)
	}
	return nil
}

const (
	// Path to save local agent checks
	checksDir     = "checks"
	checkStateDir = "checks/state"
)

// persistCheck saves a check definition to the local agent's state directory
func (a *Agent) persistCheck(check *structs.HealthCheck, chkType *structs.CheckType, source configSource) error {
	cid := check.CompoundCheckID()
	checkPath := filepath.Join(a.config.DataDir, checksDir, cid.StringHash())

	// Create the persisted check
	wrapped := persistedCheck{
		Check:   check,
		ChkType: chkType,
		Token:   a.State.CheckToken(check.CompoundCheckID()),
		Source:  source.String(),
	}

	encoded, err := json.Marshal(wrapped)
	if err != nil {
		return err
	}

	return file.WriteAtomic(checkPath, encoded)
}

// purgeCheck removes a persisted check definition file from the data dir
func (a *Agent) purgeCheck(checkID structs.CheckID) error {
	checkPath := filepath.Join(a.config.DataDir, checksDir, checkID.StringHash())
	if _, err := os.Stat(checkPath); err == nil {
		return os.Remove(checkPath)
	}
	return nil
}

// persistCheckState is used to record the check status into the data dir.
// This allows the state to be restored on a later agent start. Currently
// only useful for TTL based checks.
func (a *Agent) persistCheckState(check *checks.CheckTTL, status, output string) error {
	// Create the persisted state
	state := persistedCheckState{
		CheckID:        check.CheckID.ID,
		Status:         status,
		Output:         output,
		Expires:        time.Now().Add(check.TTL).Unix(),
		EnterpriseMeta: check.CheckID.EnterpriseMeta,
	}

	// Encode the state
	buf, err := json.Marshal(state)
	if err != nil {
		return err
	}

	// Create the state dir if it doesn't exist
	dir := filepath.Join(a.config.DataDir, checkStateDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed creating check state dir %q: %s", dir, err)
	}

	// Write the state to the file
	file := filepath.Join(dir, check.CheckID.StringHash())

	// Create temp file in same dir, to make more likely atomic
	tempFile := file + ".tmp"

	// persistCheckState is called frequently, so don't use writeFileAtomic to avoid calling fsync here
	if err := ioutil.WriteFile(tempFile, buf, 0600); err != nil {
		return fmt.Errorf("failed writing temp file %q: %s", tempFile, err)
	}
	if err := os.Rename(tempFile, file); err != nil {
		return fmt.Errorf("failed to rename temp file from %q to %q: %s", tempFile, file, err)
	}

	return nil
}

// loadCheckState is used to restore the persisted state of a check.
func (a *Agent) loadCheckState(check *structs.HealthCheck) error {
	cid := check.CompoundCheckID()
	// Try to read the persisted state for this check
	file := filepath.Join(a.config.DataDir, checkStateDir, cid.StringHash())
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed reading file %q: %s", file, err)
	}

	// Decode the state data
	var p persistedCheckState
	if err := json.Unmarshal(buf, &p); err != nil {
		a.logger.Error("failed decoding check state", "error", err)
		return a.purgeCheckState(cid)
	}

	// Check if the state has expired
	if time.Now().Unix() >= p.Expires {
		a.logger.Debug("check state expired, not restoring", "check", cid.String())
		return a.purgeCheckState(cid)
	}

	// Restore the fields from the state
	check.Output = p.Output
	check.Status = p.Status
	return nil
}

// purgeCheckState is used to purge the state of a check from the data dir
func (a *Agent) purgeCheckState(checkID structs.CheckID) error {
	file := filepath.Join(a.config.DataDir, checkStateDir, checkID.StringHash())
	err := os.Remove(file)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// persistedServiceConfig is used to serialize the resolved service config that
// feeds into the ServiceManager at registration time so that it may be
// restored later on.
type persistedServiceConfig struct {
	ServiceID string
	Defaults  *structs.ServiceConfigResponse
	structs.EnterpriseMeta
}

func (a *Agent) persistServiceConfig(serviceID structs.ServiceID, defaults *structs.ServiceConfigResponse) error {
	// Create the persisted config.
	wrapped := persistedServiceConfig{
		ServiceID:      serviceID.ID,
		Defaults:       defaults,
		EnterpriseMeta: serviceID.EnterpriseMeta,
	}

	encoded, err := json.Marshal(wrapped)
	if err != nil {
		return err
	}

	dir := filepath.Join(a.config.DataDir, serviceConfigDir)
	configPath := filepath.Join(dir, serviceID.StringHash())

	// Create the config dir if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed creating service configs dir %q: %s", dir, err)
	}

	return file.WriteAtomic(configPath, encoded)
}

func (a *Agent) purgeServiceConfig(serviceID structs.ServiceID) error {
	configPath := filepath.Join(a.config.DataDir, serviceConfigDir, serviceID.StringHash())
	if _, err := os.Stat(configPath); err == nil {
		return os.Remove(configPath)
	}
	return nil
}

func (a *Agent) readPersistedServiceConfigs() (map[structs.ServiceID]*structs.ServiceConfigResponse, error) {
	out := make(map[structs.ServiceID]*structs.ServiceConfigResponse)

	configDir := filepath.Join(a.config.DataDir, serviceConfigDir)
	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Failed reading service configs dir %q: %s", configDir, err)
	}

	for _, fi := range files {
		// Skip all dirs
		if fi.IsDir() {
			continue
		}

		// Skip all partially written temporary files
		if strings.HasSuffix(fi.Name(), "tmp") {
			a.logger.Warn("Ignoring temporary service config file", "file", fi.Name())
			continue
		}

		// Read the contents into a buffer
		file := filepath.Join(configDir, fi.Name())
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed reading service config file %q: %s", file, err)
		}

		// Try decoding the service config definition
		var p persistedServiceConfig
		if err := json.Unmarshal(buf, &p); err != nil {
			a.logger.Error("Failed decoding service config file",
				"file", file,
				"error", err,
			)
			continue
		}
		out[structs.NewServiceID(p.ServiceID, &p.EnterpriseMeta)] = p.Defaults
	}

	return out, nil
}
