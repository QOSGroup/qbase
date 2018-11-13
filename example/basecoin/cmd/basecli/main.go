package main

import (
	"github.com/QOSGroup/qbase/client/block"
	bcli "github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/client/keys"
	"github.com/QOSGroup/qbase/client/qcp"
	"github.com/QOSGroup/qbase/example/basecoin/app"
	"github.com/QOSGroup/qbase/example/basecoin/tx/client"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/version"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

var (
	rootCmd = &cobra.Command{
		Use:   "basecli",
		Short: "Basecoin light-client",
	}
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()

	rootCmd.AddCommand(bcli.LineBreak)

	// account
	rootCmd.AddCommand(
		bcli.GetCommands(account.QueryAccountCmd(cdc))...)
	rootCmd.AddCommand(bcli.LineBreak)

	// keys
	rootCmd.AddCommand(
		bcli.GetCommands(keys.Commands(cdc))...)
	rootCmd.AddCommand(bcli.LineBreak)

	// basecoin
	client.AddCommands(rootCmd, cdc)
	rootCmd.AddCommand(bcli.LineBreak)

	// qcp
	qcp.AddCommands(rootCmd, cdc)
	rootCmd.AddCommand(bcli.LineBreak)

	//query
	block.AddCommands(rootCmd, cdc)
	rootCmd.AddCommand(bcli.LineBreak)

	// version
	rootCmd.AddCommand(
		version.VersionCmd,
	)


	executor := cli.PrepareMainCmd(rootCmd, "BC", types.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}
