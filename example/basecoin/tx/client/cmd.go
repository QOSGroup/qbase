package client

import (
	bcli "github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/tx"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
)

func AddCommands(cmd *cobra.Command, cdc *amino.Codec) {
	tx.AddCommands(cmd, cdc)
	cmd.AddCommand(bcli.GetCommands(
		SendTxCmd(cdc),
		SendQCPTxCmd(cdc))...
	)
}
