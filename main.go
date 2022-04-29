package main

import (
	"bufio"
	"fmt"
	cliinstall "github.com/hashicorp/consul/command/cliplugin/install"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hashicorp/consul/command"
	"github.com/hashicorp/consul/command/cli"
	"github.com/hashicorp/consul/command/cliplugin"
	"github.com/hashicorp/consul/command/version"
	"github.com/hashicorp/consul/lib"
	_ "github.com/hashicorp/consul/service_os"
	cliversion "github.com/hashicorp/consul/version"
	mcli "github.com/mitchellh/cli"
)

func init() {
	lib.SeedMathRand()
}

func main() {
	os.Exit(realMain())
}

func realMain() int {
	log.SetOutput(ioutil.Discard)

	ui := &cli.BasicUI{
		BasicUi: mcli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr},
	}
	cmds := command.CommandsFromRegistry(ui)
	var names []string
	for c := range cmds {
		names = append(names, c)
	}

	cli := &mcli.CLI{
		Args:         os.Args[1:],
		Commands:     cmds,
		Autocomplete: true,
		Name:         "consul",
		HelpFunc:     mcli.FilteredHelpFunc(names, mcli.BasicHelpFunc("consul")),
		HelpWriter:   os.Stdout,
		ErrorWriter:  os.Stderr,
		InvalidCommandFunc: func(command string, args []string) (bool, error) {
			if len(args) == 0 {
				return false, nil
			}
			allowed := false
			for _, plugin := range cliplugin.AllowedCLIPlugins {
				if command == plugin {
					allowed = true
					break
				}
			}
			if !allowed {
				return false, nil
			}

			// Look to see if this plugin is installed.
			home, err := homedir.Dir()
			if err != nil {
				return false, err
			}

			pluginBinary := fmt.Sprintf("consul-%s", args[0])
			pluginPath := filepath.Join(home, ".consul", "plugins", pluginBinary)
			_, err = os.Stat(pluginPath)
			if err != nil {
				if !os.IsNotExist(err) {
					return false, err
				}

				// Prompt to install.
				fmt.Printf("Consul CLI plugin %q is not installed. Install it now? (Y/n)\n", command)
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.TrimRight(input, "\n")
				if input != "" && input != "y" && input != "Y" && input != "yes" {
					return true, nil
				}
				success, err := cliinstall.DoInstall(command, "")
				if err != nil {
					return true, err
				}
				fmt.Println(success)

				// Prompt to continue to run command.
				fmt.Printf("Continue to run \"%s %s\"? (Y/n)\n", filepath.Base(os.Args[0]), strings.Join(args, " "))
				input, _ = reader.ReadString('\n')
				input = strings.TrimRight(input, "\n")
				if input != "" && input != "y" && input != "Y" && input != "yes" {
					return true, nil
				}
			}

			env := append(os.Environ(), fmt.Sprintf("CONSUL_CLI_VERSION=%s", cliversion.Version))
			execErr := syscall.Exec(pluginPath, append([]string{pluginBinary}, args[1:]...), env)
			if execErr != nil {
				return false, err
			}
			return true, nil
		},
	}

	if cli.IsVersion() {
		cmd := version.New(ui)
		return cmd.Run(nil)
	}

	exitCode, err := cli.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %v\n", err)
		return 1
	}

	return exitCode
}
