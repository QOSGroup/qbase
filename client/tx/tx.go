package tx

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"
	go_amino "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Result struct {
	Hash   []byte                 `json:"hash"`
	Height int64                  `json:"height"`
	Tx     types.Tx               `json:"tx"`
	Result abci.ResponseDeliverTx `json:"result"`
}

func formatTxResult(cdc *go_amino.Codec, res *ctypes.ResultTx) (Result, error) {

	var tx types.Tx
	err := cdc.UnmarshalBinaryBare(res.Tx, &tx)

	if err != nil {
		return Result{}, err
	}

	return Result{
		Hash:   res.Hash,
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
