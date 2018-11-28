package block

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/types"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	go_amino "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	flagTags  = "tag"
	flagPage  = "page"
	flagLimit = "limit"
)

func searchTxCmd(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Search for all transactions that match the given tags.",
		Long: strings.TrimSpace(`
Search for all transactions that match the given tags.

		example:
$ basecli txs --tag test1,test2
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			tags := viper.GetStringSlice(flagTags)
			page := viper.GetInt(flagPage)
			limit := viper.GetInt(flagLimit)

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txs, err := searchTxs(cliCtx, tags, page, limit)
			if err != nil {
				return err
			}

			return cliCtx.PrintResult(txs)
		},
	}

	cmd.Flags().StringP(types.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(types.FlagNode, cmd.Flags().Lookup(types.FlagNode))
	cmd.Flags().String(types.FlagChainID, "", "Chain ID of Tendermint node")
	viper.BindPFlag(types.FlagChainID, cmd.Flags().Lookup(types.FlagChainID))
	cmd.Flags().Bool(types.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	viper.BindPFlag(types.FlagTrustNode, cmd.Flags().Lookup(types.FlagTrustNode))
	cmd.Flags().StringSlice(flagTags, nil, "Comma-separated list of tags that must match")
	cmd.Flags().Int(flagPage, 1, "search page")
	cmd.Flags().Int(flagLimit, 100, "per page limit for result")
	cmd.Flags().Bool(types.FlagJSONIndet, false, "print indent result json")

	cmd.MarkFlagRequired(flagTags)
	return cmd
}

func queryTxCmd(cdc *go_amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [hash]",
		Short: "Query match hash tx in all commit block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hashHexStr := args[0]
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			output, err := queryTx(cliCtx, hashHexStr)
			if err != nil {
				return err
			}
			return cliCtx.PrintResult(output)
		},
	}

	cmd.Flags().StringP(types.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(types.FlagNode, cmd.Flags().Lookup(types.FlagNode))
	cmd.Flags().String(types.FlagChainID, "", "Chain ID of Tendermint node")
	viper.BindPFlag(types.FlagChainID, cmd.Flags().Lookup(types.FlagChainID))
	cmd.Flags().Bool(types.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	viper.BindPFlag(types.FlagTrustNode, cmd.Flags().Lookup(types.FlagTrustNode))
	cmd.Flags().Bool(types.FlagJSONIndet, false, "print indent result json")
	return cmd
}

type Result struct {
	Hash   string                 `json:"hash"`
	Height int64                  `json:"height"`
	Tx     btypes.Tx              `json:"tx"`
	Result abci.ResponseDeliverTx `json:"result"`
}

func formatTxResult(cdc *go_amino.Codec, res *ctypes.ResultTx) (Result, error) {

	var tx btypes.Tx
	err := cdc.UnmarshalBinaryBare(res.Tx, &tx)

	if err != nil {
		return Result{}, err
	}

	return Result{
		Hash:   hex.EncodeToString(res.Hash),
		Height: res.Height,
		Tx:     tx,
		Result: res.TxResult,
	}, nil
}

func searchTxs(cliCtx context.CLIContext, tags []string, page, limit int) ([]Result, error) {
	if len(tags) == 0 {
		return nil, errors.New("tags is empty")
	}

	query := strings.Join(tags, " AND ")
	node, err := cliCtx.GetNode()
	if err != nil {
		return nil, err
	}

	prove := !cliCtx.TrustNode
	res, err := node.TxSearch(query, prove, page, limit)
	if err != nil {
		return nil, err
	}
	//todo 校验tx

	results := make([]Result, len(res.Txs))
	for i, k := range res.Txs {
		results[i], err = formatTxResult(cliCtx.Codec, k)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func queryTx(cliCtx context.CLIContext, hashHexStr string) (Result, error) {
	hash, err := hex.DecodeString(hashHexStr)
	if err != nil {
		return Result{}, err
	}

	node, err := cliCtx.GetNode()
	if err != nil {
		return Result{}, err
	}

	res, err := node.Tx(hash, cliCtx.TrustNode)
	if err != nil {
		return Result{}, err
	}

	result, err := formatTxResult(cliCtx.Codec, res)
	if err != nil {
		return Result{}, err
	}

	return result, nil
}
