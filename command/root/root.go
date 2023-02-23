package root

import (
	"fmt"
	"github.com/sdesignb/polygon-edge/command/backup"
	"github.com/sdesignb/polygon-edge/command/genesis"
	"github.com/sdesignb/polygon-edge/command/helper"
	"github.com/sdesignb/polygon-edge/command/ibft"
	"github.com/sdesignb/polygon-edge/command/license"
	"github.com/sdesignb/polygon-edge/command/loadbot"
	"github.com/sdesignb/polygon-edge/command/monitor"
	"github.com/sdesignb/polygon-edge/command/peers"
	"github.com/sdesignb/polygon-edge/command/secrets"
	"github.com/sdesignb/polygon-edge/command/server"
	"github.com/sdesignb/polygon-edge/command/status"
	"github.com/sdesignb/polygon-edge/command/txpool"
	"github.com/sdesignb/polygon-edge/command/version"
	"github.com/spf13/cobra"
	"os"
)

type RootCommand struct {
	baseCmd *cobra.Command
}

func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{
		baseCmd: &cobra.Command{
			Short: "Polygon Edge is a framework for building Ethereum-compatible Blockchain networks",
		},
	}

	helper.RegisterJSONOutputFlag(rootCommand.baseCmd)

	rootCommand.registerSubCommands()

	return rootCommand
}

func (rc *RootCommand) registerSubCommands() {
	rc.baseCmd.AddCommand(
		version.GetCommand(),
		txpool.GetCommand(),
		status.GetCommand(),
		secrets.GetCommand(),
		peers.GetCommand(),
		monitor.GetCommand(),
		loadbot.GetCommand(),
		ibft.GetCommand(),
		backup.GetCommand(),
		genesis.GetCommand(),
		server.GetCommand(),
		license.GetCommand(),
	)
}

func (rc *RootCommand) Execute() {
	if err := rc.baseCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
