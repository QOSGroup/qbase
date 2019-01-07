package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	go_amino "github.com/tendermint/go-amino"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/types"
)

const (
	flagOverwrite  = "overwrite"
	flagClientHome = "home-client"
	flagMoniker    = "moniker"
	flagChainID    = "chain-id"
)

type printInfo struct {
	Moniker    string          `json:"moniker"`
	ChainID    string          `json:"chain_id"`
	NodeID     string          `json:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message"`
}

//自定义生成GenesisDoc
type CustomGenGenesisDocFunc func(ctx *Context, cdc *go_amino.Codec, chainID string, nodeValidatorPubKey crypto.PubKey) (types.GenesisDoc, error)

// nolint: errcheck
func displayInfo(cdc *go_amino.Codec, info printInfo) error {
	out, err := cdc.MarshalJSONIndent(info, "", " ")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s\n", string(out))
	return nil
}

// get cmd to initialize all files for tendermint and application
// nolint
func InitCmd(ctx *Context, cdc *go_amino.Codec, genGenesisDocFun CustomGenGenesisDocFunc, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			chainID := viper.GetString(flagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}

			nodeID, valPubkey, err := InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			config.Moniker = viper.GetString(flagMoniker)

			genFile := config.GenesisFile()

			overwrite := viper.GetBool(flagOverwrite)
			if !overwrite && common.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			genesisDoc, err := genGenesisDocFun(ctx, cdc, chainID, valPubkey)
			if err != nil {
				return err
			}

			if err = SaveGenDoc(genFile, genesisDoc); err != nil {
				return err
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", genesisDoc.AppState)

			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			return displayInfo(cdc, toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(flagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(flagMoniker, "", "set the validator's moniker")
	cmd.MarkFlagRequired(flagMoniker)

	return cmd
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string,
	appMessage json.RawMessage) printInfo {

	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}
