package main

import (
	"io"

	"github.com/QOSGroup/qbase/example/basecoin/app"
	"github.com/QOSGroup/qbase/example/basecoin/types"
	"github.com/QOSGroup/qbase/server"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/QOSGroup/qbase/version"
	"github.com/spf13/cobra"
	go_amino "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func main() {

	addressConfig := btypes.GetAddressConfig()
	addressConfig.SetBech32PrefixForAccount("basecoin", "basecoinpub")
	addressConfig.SetBech32PrefixForConsensusNode("basecoincons", "basecoinconspub")
	addressConfig.SetBech32PrefixForValidator("basecoinval", "basecoinalpub")
	addressConfig.Seal()

	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "basecoind",
		Short:             "basecoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	// version cmd
	rootCmd.AddCommand(version.VersionCmd)

	//add init command
	rootCmd.AddCommand(server.InitCmd(ctx, cdc, genBaseCoindGenesisDoc, types.DefaultNodeHome))

	server.AddCommands(ctx, cdc, rootCmd, newApp)

	executor := cli.PrepareBaseCmd(rootCmd, "basecoin", types.DefaultNodeHome)

	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, storeTracer io.Writer) abci.Application {
	return app.NewApp(logger, db, storeTracer)
}

func genBaseCoindGenesisDoc(ctx *server.Context, cdc *go_amino.Codec, chainID string, nodeValidatorPubKey crypto.PubKey) (tmtypes.GenesisDoc, error) {

	validator := tmtypes.GenesisValidator{
		PubKey: nodeValidatorPubKey,
		Power:  10,
	}

	addr, _, err := types.GenerateCoinKey(cdc, types.DefaultCLIHome)
	if err != nil {
		return tmtypes.GenesisDoc{}, err
	}

	genTx := types.BaseCoinGenTx{addr}
	appState, err := types.BaseCoinAppGenState(cdc, genTx)
	if err != nil {
		return tmtypes.GenesisDoc{}, err
	}

	return tmtypes.GenesisDoc{
		ChainID:    chainID,
		Validators: []tmtypes.GenesisValidator{validator},
		AppState:   appState,
	}, nil

}
