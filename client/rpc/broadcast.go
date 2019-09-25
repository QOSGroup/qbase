package rpc

import (
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/types"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

type BroadcastReq struct {
	Tx   types.Tx `json:"tx"`
	Mode string   `json:"mode"`
}

func registerTxsRoutes(ctx context.CLIContext, m *mux.Router) {
	m.HandleFunc("/txs", broadcastTxRequest(ctx)).Methods("POST")
}

func broadcastTxRequest(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		var req BroadcastReq
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		err = cliContext.Codec.UnmarshalJSON(body, &req)
		if err != nil {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
			return
		}

		ctx := cliContext.WithBroadcastMode(req.Mode)
		txBytes, err := ctx.Codec.MarshalBinaryBare(req.Tx)
		if err != nil {
			WriteErrorResponse(writer, http.StatusInternalServerError, err.Error())
			return
		}

		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			WriteErrorResponse(writer, http.StatusInternalServerError, err.Error())
			return
		}

		PostProcessResponseBare(writer, cliContext, res)
	}
}
