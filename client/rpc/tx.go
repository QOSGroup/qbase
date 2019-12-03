package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/gorilla/mux"
	amino "github.com/tendermint/go-amino"
	atypes "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"net/http"
	"strings"
	"time"
)

type TxResponse struct {
	Height    int64    `json:"height"`
	TxHash    string   `json:"txhash"`
	Code      uint32   `json:"code"`
	Data      string   `json:"data,omitempty"`
	RawLog    string   `json:"raw_log,omitempty"`
	Info      string   `json:"info,omitempty"`
	GasWanted int64    `json:"gas_wanted,omitempty"`
	GasUsed   int64    `json:"gas_used,omitempty"`
	Codespace string   `json:"codespace,omitempty"`
	Tx        types.Tx `json:"tx,omitempty"`
	Timestamp string   `json:"timestamp,omitempty"`
	Events    []events `JSON:"event,omitempty"`
}

type event struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type events struct {
	Type       string  `json:"type"`
	Attributes []event `json:"attributes"`
}

func NewResponseResultTx(res *ctypes.ResultTx, tx types.Tx, timestamp string) TxResponse {
	if res == nil {
		return TxResponse{}
	}

	es, _ := ParseEvents(res.TxResult.Events)

	return TxResponse{
		TxHash:    res.Hash.String(),
		Height:    res.Height,
		Code:      res.TxResult.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.TxResult.Data)),
		RawLog:    res.TxResult.Log,
		Info:      res.TxResult.Info,
		GasWanted: res.TxResult.GasWanted,
		GasUsed:   res.TxResult.GasUsed,
		Tx:        tx,
		Timestamp: timestamp,
		Events:    es,
	}
}

func ParseEvents(rte []atypes.Event) ([]events, error) {
	es := make([]events, 0)
	for _, ev := range rte {
		var e events
		e.Type = ev.Type

		for _, av := range ev.Attributes {
			fmt.Println(string(av.Key))
			fmt.Println(string(av.Value))
			e.Attributes = append(e.Attributes, event{string(av.Key), string(av.Value)})
		}

		es = append(es, e)
	}

	return es, nil
}

func registerTxRoutes(ctx context.CLIContext, m *mux.Router) {
	m.HandleFunc("/txs/{hash}", queryTxHashHandleFn(ctx)).Methods("GET")
}

func queryTxHashHandleFn(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {

		m := mux.Vars(request)

		br, _ := ParseRequestForm(request)
		ctx := br.Setup(cliContext)

		response, err := queryTx(ctx, m["hash"])
		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		PostProcessResponseBare(writer, ctx, response)
	}
}

func queryTx(cliCtx context.CLIContext, hashHexStr string) (TxResponse, error) {
	hash, err := hex.DecodeString(hashHexStr)
	if err != nil {
		return TxResponse{}, err
	}

	node, err := cliCtx.GetNode()
	if err != nil {
		return TxResponse{}, err
	}

	resTx, err := node.Tx(hash, false)
	if err != nil {
		return TxResponse{}, err
	}

	resBlocks, err := getBlocksForTxResults(cliCtx, []*ctypes.ResultTx{resTx})
	if err != nil {
		return TxResponse{}, err
	}

	out, err := formatTxResult(cliCtx.Codec, resTx, resBlocks[resTx.Height])
	if err != nil {
		return out, err
	}

	return out, nil
}

func getBlocksForTxResults(cliCtx context.CLIContext, resTxs []*ctypes.ResultTx) (map[int64]*ctypes.ResultBlock, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return nil, err
	}

	resBlocks := make(map[int64]*ctypes.ResultBlock)

	for _, resTx := range resTxs {
		if _, ok := resBlocks[resTx.Height]; !ok {
			resBlock, err := node.Block(&resTx.Height)
			if err != nil {
				return nil, err
			}

			resBlocks[resTx.Height] = resBlock
		}
	}

	return resBlocks, nil
}

func formatTxResults(cdc *amino.Codec, resTxs []*ctypes.ResultTx, resBlocks map[int64]*ctypes.ResultBlock) ([]TxResponse, error) {
	var err error
	out := make([]TxResponse, len(resTxs))
	for i := range resTxs {
		out[i], err = formatTxResult(cdc, resTxs[i], resBlocks[resTxs[i].Height])
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func formatTxResult(cdc *amino.Codec, resTx *ctypes.ResultTx, resBlock *ctypes.ResultBlock) (TxResponse, error) {
	tx, err := parseTx(cdc, resTx.Tx)
	if err != nil {
		return TxResponse{}, err
	}

	return NewResponseResultTx(resTx, tx, resBlock.Block.Time.Format(time.RFC3339)), nil
}

func parseTx(cdc *amino.Codec, txBytes []byte) (types.Tx, error) {
	var tx types.Tx

	err := cdc.UnmarshalBinaryBare(txBytes, &tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
