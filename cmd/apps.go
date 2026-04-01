package cmd

import (
	"github.com/osmargm1202/orgm/apps/bt"
	"github.com/osmargm1202/orgm/apps/calc"
	"github.com/osmargm1202/orgm/apps/env"
	"github.com/osmargm1202/orgm/apps/org"
	"github.com/osmargm1202/orgm/apps/proy"
	"github.com/osmargm1202/orgm/apps/san"
)

func init() {
	// Register all app subcommands
	RootCmd.AddCommand(env.Cmd())
	RootCmd.AddCommand(calc.Cmd())
	RootCmd.AddCommand(proy.Cmd())
	RootCmd.AddCommand(org.Cmd())
	RootCmd.AddCommand(san.Cmd())
	RootCmd.AddCommand(bt.Cmd())
}
