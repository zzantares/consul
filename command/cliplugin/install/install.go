package install

import (
	"context"
	"flag"
	"fmt"
	"github.com/hashicorp/consul/command/cliplugin"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"os"
	"strings"

	"github.com/hashicorp/consul/command/flags"
	"github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/src"
	"github.com/mitchellh/cli"
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

	// flags
	flagVersion string
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.flagVersion, "version", "",
		"Version to install. Defaults to the latest version.")
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
	allowed := false
	for _, allowedPlugin := range cliplugin.AllowedCLIPlugins {
		if pluginName == allowedPlugin {
			allowed = true
			break
		}
	}
	if !allowed {
		c.UI.Error(fmt.Sprintf("Error: %q is not a supported plugin. Supported plugins are: %s", pluginName, strings.Join(cliplugin.AllowedCLIPlugins, ",")))
		return 1
	}

	var pluginVersion *version.Version
	if c.flagVersion != "" {
		var err error
		pluginVersion, err = version.NewVersion(c.flagVersion)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error: unable to parse version %q: %s", c.flagVersion, err))
			return 1
		}
	}

	err := DoInstall(pluginName, pluginVersion)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}

	versionStr := "latest"
	if pluginVersion != nil {
		versionStr = pluginVersion.String()
	}
	c.UI.Info(fmt.Sprintf("Installed %s plugin (version %s) successfully. To use, run \"consul %s\"",
		pluginName, versionStr, pluginName))
	return 0
}

func DoInstall(plugin string, pluginVersion *version.Version) error {
	pluginDir := os.Getenv(cliplugin.PluginDirEnvVar)
	if pluginDir == "" {
		pluginDir = cliplugin.DefaultPluginDir
	}

	ctx := context.Background()
	if err := os.MkdirAll(pluginDir, 0700); err != nil {
		return fmt.Errorf("unable to create plugin dir: %s", err)
	}

	installer := install.NewInstaller()
	pluginProduct := product.Product{
		Name:              fmt.Sprintf("consul-%s", plugin),
		BinaryName:        func() string { return fmt.Sprintf("consul-%s", plugin) },
		GetVersion:        nil,
		BuildInstructions: nil,
	}
	var installable src.Installable
	if pluginVersion == nil {
		installable = &releases.LatestVersion{
			Product:            pluginProduct,
			InstallDir:         pluginDir,
			IncludePrereleases: false,
		}
	} else {
		installable = &releases.ExactVersion{
			Product:    pluginProduct,
			Version:    pluginVersion,
			InstallDir: pluginDir,
		}
	}

	_, err := installer.Install(ctx, []src.Installable{installable})
	if err != nil {
		return fmt.Errorf("unable to install: %s", err)
	}
	return nil
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return c.help
}

const (
	synopsis = "Install a plugin."
)

var help = fmt.Sprintf(`
Usage: consul cli-plugin install NAME [options]

  Install a CLI plugin. Installs into the directory set by the %s environment
  variable (defaults to %s). If installing into a non-default directory, the
  the %s environment variable must also be set when executing
  the plugins.

      $ consul cli-plugin install <plugin name>
      $ consul cli-plugin install <plugin name> -version 1.0.0

`, cliplugin.PluginDirEnvVar, cliplugin.DefaultPluginDir, cliplugin.PluginDirEnvVar)
