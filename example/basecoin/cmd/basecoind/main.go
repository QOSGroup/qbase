package main

import (
	"github.com/QOSGroup/qbase/example/basecoin/app"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/server"
	"github.com/QOSGroup/qbase/version"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"io"
)

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "basecoind",
		Short:             "basecoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	// version cmd
	rootCmd.AddCommand(version.VersionCmd)

	server.AddCommands(ctx, cdc, rootCmd, types.BaseCoinInit(),
		server.ConstructAppCreator(newApp, "basecoin"))

	executor := cli.PrepareBaseCmd(rootCmd, "basecoin", types.DefaultNodeHome)

	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, storeTracer io.Writer) abci.Application {
	return app.NewApp(logger, db, storeTracer)
}
