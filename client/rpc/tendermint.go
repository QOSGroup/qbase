package rpc

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/gorilla/mux"
	types2 "github.com/tendermint/tendermint/types"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type TxsSearchEvent struct {
	Key string `json:"key"`
	Value string `json:"value"`
	Op string `json:"op"`
}

type TxsSearchRequest struct {
	Events []TxsSearchEvent `json:"events"`
	Proof bool `json:"proof"`
	Page int `json:"page"`
	Limit int `json:"limit"`
}

type TxsSearchResponse struct {
	Txs []TxsSearchItem `json:"txs"`
	TotalCount int `json:"total_count"`
}

type TxsSearchItem struct {
	Hash string  `json:"hash"`
	Height int64  `json:"height"`
	Index uint32  `json:"index"`
	GasUsed int64  `json:"gas_used"`
}

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

	m.HandleFunc("/txs/search", queryTxsByCondition(ctx)).Methods("POST")
}

func queryTxsByCondition(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			Write40XErrorResponse(writer, err)
			return
		}

		var searchRequest TxsSearchRequest
		if err = json.Unmarshal(body, &searchRequest); err != nil {
			Write40XErrorResponse(writer, err)
			return
		}

		if len(searchRequest.Events) == 0 {
			Write40XErrorResponse(writer, errors.New("miss query condition"))
			return
		}

		queryCnd := make([]string, 0,len(searchRequest.Events))
		for _ , e := range searchRequest.Events {
			op := "="
			if e.Op != "" {
				op = e.Op
			}
			queryCnd = append(queryCnd, strings.Join([]string{e.Key,e.Value} , op))
		}

		query := strings.Join(queryCnd, " AND ")

		page := 1
		limit := 10

		if searchRequest.Page != 0 {
			page = searchRequest.Page
		}

		if searchRequest.Limit != 0 && searchRequest.Limit <= 100 {
			limit = searchRequest.Limit
		}


		searchResult, err := cliContext.Client.TxSearch(query, searchRequest.Proof, page, limit)
		if err != nil {
			Write40XErrorResponse(writer, err)
			return
		}

		var response TxsSearchResponse
		response.TotalCount = searchResult.TotalCount
		for _ , tx := range searchResult.Txs {
			response.Txs = append(response.Txs , TxsSearchItem{
				Hash:    tx.Hash.String(),
				Height:  tx.Height,
				Index:   tx.Index,
				GasUsed: tx.TxResult.GasUsed,
			})
		}

		PostProcessResponseBare(writer, cliContext, response)
	}
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
