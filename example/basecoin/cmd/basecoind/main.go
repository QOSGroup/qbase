package main

import (
	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/example/basecoin/app"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/server"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"io"
	"os"
)

func main() {
	cdc := makeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "basecoind",
		Short:             "basecoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, types.BaseCoinInit(),
		server.ConstructAppCreator(newApp, "basecoin"))

	rootDir := os.ExpandEnv("$HOME/.basecoind")
	executor := cli.PrepareBaseCmd(rootCmd, "basecoin", rootDir)

	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, storeTracer io.Writer) abci.Application {
	return app.NewApp(logger, db, storeTracer)
}

func makeCodec() *amino.Codec {
	cdc := baseabci.MakeQBaseCodec()

	types.RegisterCodec(cdc)
	tx.RegisterCodec(cdc)

	return cdc
}
