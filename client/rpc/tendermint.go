package rpc

import (
	"encoding/hex"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/gorilla/mux"
	types2 "github.com/tendermint/tendermint/types"
	"net/http"
	"strconv"
	"strings"
)

func registerTendermintRoutes(ctx context.CLIContext, m *mux.Router) {
	m.HandleFunc("/node_status", queryNodeStatusHandleFn(ctx)).Methods("GET")
	m.HandleFunc("/blocks/latest", queryLatestBlockHandleFn(ctx)).Methods("GET")
	m.HandleFunc("/blocks/{height}", queryBlockHandleFn(ctx)).Methods("GET")
	m.HandleFunc("/blocks/txs/{height}", queryBlockTxsHandleFn(ctx)).Methods("GET")
	m.HandleFunc("/validators/latest", queryLatestValidatorsHandleFn(ctx)).Methods("GET")
	m.HandleFunc("/validators/{height}", queryValidatorsHandleFn(ctx)).Methods("GET")
	m.HandleFunc("/validators/consensus/{address}", func(writer http.ResponseWriter, request *http.Request) {
		m := mux.Vars(request)
		bz, err := hex.DecodeString(m["address"])
		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		PostProcessResponseBare(writer, ctx, types.ConsAddress(bz))
	}).Methods("GET")
}

func queryBlockTxsHandleFn(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		m := mux.Vars(request)

		height, err := strconv.ParseInt(m["height"], 10, 64)
		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		bs, err := cliContext.Client.Block(&height)

		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		txs := []types2.Tx(bs.Block.Txs)
		if len(txs) == 0 {
			WriteErrorResponse(writer, http.StatusNotFound, "no txs in block")
			return
		}

		txHashes := make([]string, 0)
		for _, tx := range txs {
			txHashes = append(txHashes, strings.ToUpper(hex.EncodeToString(tx.Hash())))
		}

		PostProcessResponseBare(writer, cliContext, txHashes)
	}
}

func queryValidatorsHandleFn(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		m := mux.Vars(request)
		height, err := strconv.ParseInt(m["height"], 10, 64)

		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		vs, err := cliContext.Client.Validators(&height)

		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		PostProcessResponseBare(writer, cliContext, vs)
	}
}

func queryLatestValidatorsHandleFn(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		vs, err := cliContext.Client.Validators(nil)
		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		PostProcessResponseBare(writer, cliContext, vs)
	}
}

func queryNodeStatusHandleFn(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		rs, err := cliContext.Client.Status()
		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		PostProcessResponseBare(writer, cliContext, rs)
	}
}

func queryLatestBlockHandleFn(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		bs, err := cliContext.Client.Block(nil)
		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		PostProcessResponseBare(writer, cliContext, bs)
	}
}

func queryBlockHandleFn(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		m := mux.Vars(request)
		height, err := strconv.ParseInt(m["height"], 10, 64)
		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		bs, err := cliContext.Client.Block(&height)

		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		PostProcessResponseBare(writer, cliContext, bs)
	}
}
