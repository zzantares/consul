package cliplugin

import (
	"github.com/hashicorp/consul/command/flags"
	"github.com/mitchellh/cli"
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
	return flags.Usage(help, nil)
}

const synopsis = "Manage Consul CLI plugins"
const help = `
Usage: consul cli-plugin <subcommand> [options] [args]

  This command has subcommands for interacting with Consul CLI plugins.

  Install a plugin:

      $ consul cli-plugin install k8s

  List available and installed plugins:

      $ consul cli-plugin list

  Update a plugin:

      $ consul cli-plugin update k8s

  Uninstall a plugin:

      $ consul cli-plugin uninstall
`

var AllowedCLIPlugins = []string{"k8s"}

const (
	DefaultPluginDir = "~/.consul/plugins"
	PluginDirEnvVar  = "CONSUL_CLI_PLUGIN_DIR"
)
