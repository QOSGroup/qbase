package block

import (
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
)

var qcpCommand = &cobra.Command{
	Use:   "query",
	Short: "query subcommands",
}

//AddCommands block info commands
func AddCommands(cmd *cobra.Command, cdc *go_amino.Codec) {
	qcpCommand.AddCommand(
		statusCommand(cdc),
		validatorsCommand(cdc),
		blockCommand(cdc),
	)

	cmd.AddCommand(qcpCommand)
}
