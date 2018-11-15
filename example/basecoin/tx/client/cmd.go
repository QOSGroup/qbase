package client

import (
	"github.com/QOSGroup/qbase/client/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
)

func AddCommands(cmd *cobra.Command, cdc *amino.Codec) {
	cmd.AddCommand(types.GetCommands(
		SendTxCmd(cdc),
		SendQCPTxCmd(cdc))...,
	)
}
