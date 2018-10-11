package main

import (
	"github.com/QOSGroup/qbase/example/inittest"
	"github.com/QOSGroup/qbase/server"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"io"
	"os"
)

func main() {
	cdc := inittest.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "inittestd",
		Short:             "inittest Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, inittest.InitTestAppInit(),
		server.ConstructAppCreator(newApp, "inittest"))

	rootDir := os.ExpandEnv("$HOME/.inittestd")
	executor := cli.PrepareBaseCmd(rootCmd, "inittest", rootDir)

	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, storeTracer io.Writer) abci.Application {
	return inittest.NewApp(logger, db, storeTracer)
}
