package txpool

import (
	"github.com/sdesignb/polygon-edge/command/helper"
	"github.com/sdesignb/polygon-edge/command/txpool/add"
	"github.com/sdesignb/polygon-edge/command/txpool/status"
	"github.com/sdesignb/polygon-edge/command/txpool/subscribe"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	txPoolCmd := &cobra.Command{
		Use:   "txpool",
		Short: "Top level command for interacting with the transaction pool. Only accepts subcommands.",
	}

	helper.RegisterGRPCAddressFlag(txPoolCmd)

	registerSubcommands(txPoolCmd)

	return txPoolCmd
}

func registerSubcommands(baseCmd *cobra.Command) {
	baseCmd.AddCommand(
		// txpool add
		add.GetCommand(),
		// txpool status
		status.GetCommand(),
		// txpool subscribe
		subscribe.GetCommand(),
	)
}
