package list

import (
	"flag"
	"fmt"
	"github.com/hashicorp/consul/command/cliplugin"
	"github.com/hashicorp/consul/command/flags"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	if len(args) != 0 {
		c.UI.Error(fmt.Sprintf("Error: command accepts no arguments"))
		return 1
	}

	plugins, err := listPlugins()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if len(plugins) == 0 {
		c.UI.Info("No plugins installed")
		return 0
	}
	c.UI.Info(strings.Join(plugins, "\n"))
	return 0
}

func listPlugins() ([]string, error) {
	pluginDir := os.Getenv(cliplugin.PluginDirEnvVar)
	if pluginDir == "" {
		pluginDir = cliplugin.DefaultPluginDir
	}
	pluginDir, err := homedir.Expand(pluginDir)
	if err != nil {
		return nil, err
	}

	var pluginNames []string
	err = filepath.WalkDir(pluginDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// This is the plugin root dir, so just skip this.
			if pluginDir == path {
				return nil
			}
			// Otherwise we're inside a directory inside the plugin dir in which case we should skip this whole
			// directory since plugins are only in the root.
			return filepath.SkipDir
		}
		f, err := d.Info()
		if err != nil {
			return err
		}
		pluginNames = append(pluginNames, f.Name())
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pluginNames, nil
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return c.help
}

const (
	synopsis = "Install a plugin."
	help     = `
Usage: consul cli-plugin install NAME [options]

  Install a specific CLI plugin. The plugin will be installed into the
  directory set by the CONSUL_PLUGIN_DIR environment variable
  (defaults to ~/.consul/plugins).

      $ consul cli-plugin install k8s
      $ consul cli-plugin install k8s -version 9.9.9

`
)
