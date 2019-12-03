package main

import (
	bcli "github.com/QOSGroup/qbase/client"
	"github.com/QOSGroup/qbase/client/config"
	ctypes "github.com/QOSGroup/qbase/client/types"
	"github.com/QOSGroup/qbase/example/basecoin/app"
	"github.com/QOSGroup/qbase/example/basecoin/tx/client"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	btypes "github.com/QOSGroup/qbase/types"
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

	addressConfig := btypes.GetAddressConfig()
	addressConfig.SetBech32PrefixForAccount("basecoin", "basecoinpub")
	addressConfig.SetBech32PrefixForConsensusNode("basecoincons", "basecoinconspub")
	addressConfig.SetBech32PrefixForValidator("basecoinval", "basecoinalpub")
	addressConfig.Seal()

	cobra.EnableCommandSorting = false
	cdc := app.MakeCodec()

	//tx
	txCommand := bcli.TxCommand()
	txCommand.AddCommand(ctypes.PostCommands(client.Commands(cdc)...)...)

	rootCmd.AddCommand(
		config.Cmd(types.DefaultCLIHome),
		txCommand,
		bcli.KeysCommand(cdc),
		bcli.QueryCommand(cdc),
		bcli.TendermintCommand(cdc),
		version.VersionCmd,
	)

	executor := cli.PrepareBaseCmd(rootCmd, "BC", types.DefaultCLIHome)

	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}
