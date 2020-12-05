package main

import (
	"github.com/hashicorp/consul/agent"
	"github.com/hashicorp/consul/command/flags"
	"github.com/mitchellh/cli"
)

func main() { println("issue 42738") }

func init() {
	registry["agent"] = func(ui cli.Ui) (cli.Command, error) {
		return (*cmd)(nil), nil
	}
}

type cmd struct {
	_ *flags.HTTPFlags
}

func (*cmd) Run(args []string) int {
	agent.NewBaseDeps(nil, nil)
	return 1
}

func (*cmd) Synopsis() string { return "" }

func (*cmd) Help() string { return "" }

type Factory func(cli.Ui) (cli.Command, error)

var registry = make(map[string]Factory)
