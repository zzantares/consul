package k8s

import (
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
	"os"
	"os/exec"
	"path/filepath"
)

func New() *cmd {
	return &cmd{}
}

type cmd struct{}

func (c *cmd) Run(args []string) int {
	return cli.RunResultHelp
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	installed, err := k8sCLIInstalled()
	if err != nil {
		return fmt.Sprintf("Error: could not determine if k8s CLI plugin is installed: %s", err)
	}

	if !installed {
		return fmt.Sprintf("Error: k8s CLI plugin not installed. To install, run: consul cli-plugin install k8s")
	}

	path, err := cliPath()
	if err != nil {
		return fmt.Sprintf("Error: could not determine k8s CLI plugin path: %s", err)
	}
	cmd := exec.Command(path, "help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 127 {
			return string(out)
		}
		return fmt.Sprintf("Error: %s", err)
	}
	return string(out)
}

const synopsis = "Manage Consul on Kubernetes"

func cliPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".consul", "plugins", "consul-k8s"), nil
}

func k8sCLIInstalled() (bool, error) {
	path, err := cliPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func execK8sCLI() {

}
