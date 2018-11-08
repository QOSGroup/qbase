package main

import (
	bcli "github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/example/basecoin/app"
	"github.com/QOSGroup/qbase/example/basecoin/tx/client"
	"github.com/QOSGroup/qbase/version"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
	"os"
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

	rootCmd.AddCommand(
		bcli.GetCommands(account.QueryAccountCmd(cdc))...)

	rootCmd.AddCommand(bcli.LineBreak)

	client.AddCommands(rootCmd, cdc)

	rootCmd.AddCommand(bcli.LineBreak)

	rootCmd.AddCommand(
		version.VersionCmd,
	)

	executor := cli.PrepareMainCmd(rootCmd, "BC", os.ExpandEnv("$HOME/.basecli"))
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}
