package client

import (
	"github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/client/block"
	"github.com/QOSGroup/qbase/client/keys"
	"github.com/QOSGroup/qbase/client/qcp"
	"github.com/QOSGroup/qbase/client/types"
	"github.com/spf13/cobra"

	go_amino "github.com/tendermint/go-amino"
)

var queryCommand = &cobra.Command{
	Use:     "query",
	Short:   "query(alias `q`) subcommands.",
	Aliases: []string{"q"},
}

var qcpCommand = &cobra.Command{
	Use:   "qcp",
	Short: "qcp subcommands",
}

var txCommand = &cobra.Command{
	Use:   "tx",
	Short: "tx subcommands",
}

var tendermintCommand = &cobra.Command{
	Use:     "tendermint",
	Short:   "tendermint(alias `t`)  subcommands",
	Aliases: []string{"t"},
}

func qcpSubCommand(cdc *go_amino.Codec) *cobra.Command {
	qcpCommand.AddCommand(
		types.GetCommands(
			qcp.QcpCommands(cdc)...,
		)...,
	)

	return qcpCommand
}

func TendermintCommand(cdc *go_amino.Codec) *cobra.Command {
	tendermintCommand.AddCommand(
		block.BlockCommand(cdc)...,
	)
	return tendermintCommand
}

//QueryCommand add query subcommand
func QueryCommand(cdc *go_amino.Codec) *cobra.Command {
	queryAccountCommand := types.GetCommands(account.QueryAccountCmd(cdc))
	queryCommand.AddCommand(queryAccountCommand[0])
	queryCommand.AddCommand(block.QueryCommand(cdc)...)
	queryCommand.AddCommand(qcpSubCommand(cdc))
	return queryCommand
}

//TxCommand add tx subcommand
func TxCommand() *cobra.Command {
	return txCommand
}

//KeysCommand add keys subcommand
func KeysCommand(cdc *go_amino.Codec) *cobra.Command {
	return keys.KeysCommand(cdc)
}
