package uninstall

import (
	"flag"
	"fmt"
	"github.com/hashicorp/consul/command/cliplugin"
	"github.com/hashicorp/consul/command/flags"
	"github.com/mitchellh/cli"
	"os"
	"path/filepath"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui}
	c.init()
	return c
}

type cmd struct {
	UI    cli.Ui
	flags *flag.FlagSet
	http  *flags.HTTPFlags
	help  string
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.help = flags.Usage(help, c.flags)
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 2
	}

	args = c.flags.Args()
	if len(args) != 1 {
		c.UI.Error(fmt.Sprintf("Error: command requires exactly one argument: plugin name"))
		return 1
	}

	pluginName := args[0]
	pluginDir := os.Getenv(cliplugin.PluginDirEnvVar)
	if pluginDir == "" {
		pluginDir = cliplugin.DefaultPluginDir
	}

	pluginPath := filepath.Join(pluginDir, pluginName)
	_, err := os.Stat(pluginPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.UI.Info(fmt.Sprintf("Plugin %q is not installed (could not find at %s)", pluginName, pluginPath))
			return 0
		}
		c.UI.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}

	if err = os.Remove(pluginPath); err != nil {
		c.UI.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}

	c.UI.Info("Plugin uninstalled successfully")
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return c.help
}

const (
	synopsis = "Uninstall a plugin."
	help     = `
Usage: consul cli-plugin uninstall NAME

  Uninstall a CLI plugin. The plugin will be uninstalled from the
  directory set by the CONSUL_PLUGIN_DIR environment variable
  (defaults to ~/.consul/plugins).

      $ consul cli-plugin uninstall k8s

`
)
