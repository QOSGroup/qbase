package rpc

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
	"github.com/gorilla/mux"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type BaseRequest struct {
	From    string `json:"from"`
	ChainId string `json:"chain_id"`
	Nonce   int64  `json:"nonce"`
	MaxGas  int64  `json:"max_gas"`
	Height  int64  `json:"height"`
	Indent  bool   `json:"indent"`
	Mode    string `json:"mode"`
}

type TxGenerateResponse struct {
	Code      int              `json:"code"`
	Tx        string           `json:"tx"`
	Signer    types.AccAddress `json:"signer"`
	PubKey    crypto.PubKey    `json:"pubkey"`
	Nonce     int64            `json:"nonce"`
	SignBytes string           `json:"sign_bytes"`
}

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func NewBaseRequest(from, chainId string, nonce, maxGas, height int64, indent bool, mode string) BaseRequest {
	return BaseRequest{
		From:    strings.TrimSpace(from),
		ChainId: strings.TrimSpace(chainId),
		Nonce:   nonce,
		MaxGas:  maxGas,
		Height:  height,
		Indent:  indent,
		Mode:    strings.TrimSpace(mode),
	}
}

func (br BaseRequest) Sanitize() BaseRequest {
	return NewBaseRequest(br.From, br.ChainId, br.Nonce, br.MaxGas, br.Height, br.Indent, br.Mode)
}

func (br BaseRequest) ValidateBasic() error {
	if len(br.From) == 0 {
		return errors.New("from address is empty")
	}

	_, err := types.AccAddressFromBech32(br.From)
	if err != nil {
		return err
	}

	if br.Nonce < int64(0) {
		return errors.New("nonce is less than zero")
	}

	if br.MaxGas < int64(0) {
		return errors.New("maxGas is less than zero")
	}

	if br.Height < int64(0) {
		return errors.New("height is less than zero")
	}

	return nil
}

func (br BaseRequest) Setup(ctx context.CLIContext) context.CLIContext {

	if br.ChainId != "" {
		ctx = ctx.WithChainID(br.ChainId)
	}

	if br.Height >= 0 {
		ctx = ctx.WithHeight(br.Height)
	}

	if br.MaxGas > 0 {
		ctx = ctx.WithMaxGas(br.MaxGas)
	}

	ctx = ctx.WithIndent(br.Indent)
	ctx = ctx.WithBroadcastMode(br.Mode)

	return ctx
}

func BuildStdTxAndResponse(writer http.ResponseWriter, request *http.Request, cliCtx context.CLIContext, typ reflect.Type, fn func(req interface{}, from types.AccAddress, vars map[string]string) (txs.ITx, error)) {
	vars := mux.Vars(request)
	rv := reflect.New(typ)

	if !readRESTRequest(writer, request, cliCtx.Codec, rv.Interface(), cliCtx.Logger) {
		return
	}

	brRv := rv.Elem().FieldByName("BaseRequest")
	if !brRv.IsValid() {
		WriteErrorResponse(writer, http.StatusBadRequest, fmt.Sprintf("find BaseRequest field error. check type: %+v ", typ))
		return
	}

	br, ok := brRv.Interface().(BaseRequest)
	if !ok {
		WriteErrorResponse(writer, http.StatusBadRequest, "covert BaseRequest type error")
		return
	}

	br, from, err := sanitizeBaseRequest(br)
	if err != nil {
		WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
		return
	}

	brRv.Set(reflect.ValueOf(br))

	itx, err := fn(rv.Elem().Interface(), from, vars)
	if err != nil {
		if err == context.RecordsNotFoundError {
			WriteErrorResponse(writer, http.StatusNotFound, err.Error())
		} else {
			WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
		}

		return
	}

	WriteGenStdTxResponse(writer, cliCtx, br, itx)
}

func sanitizeBaseRequest(br BaseRequest) (BaseRequest, types.AccAddress, error) {
	br = br.Sanitize()
	err := br.ValidateBasic()
	if err != nil {
		return br, nil, err
	}

	fromAddr, err := types.AccAddressFromBech32(br.From)
	if err != nil {
		return br, nil, err
	}

	return br, fromAddr, nil
}

func ParseURIPathValue(request *http.Request, pathName string) (string, error) {
	vars := mux.Vars(request)
	return vars[pathName], nil
}

func ParseURIPathAddress(request *http.Request, pathName string) (types.AccAddress, error) {
	vars := mux.Vars(request)
	addr, err := types.AccAddressFromBech32(vars[pathName])
	return addr, err
}

func readRESTRequest(w http.ResponseWriter, r *http.Request, cdc *amino.Codec, req interface{}, logger log.Logger) bool {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return false
	}

	logger.Debug("readRESTRequest", "body", string(body))
	err = cdc.UnmarshalJSON(body, req)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to decode JSON payload: %s", err))
		return false
	}

	return true
}

func NewErrorResponse(code int, err string) ErrorResponse {
	return ErrorResponse{Code: code, Error: err}
}

func WriteErrorResponse(w http.ResponseWriter, status int, err string) {
	cdc := amino.NewCodec()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(cdc.MustMarshalJSON(NewErrorResponse(status, err)))
}

func Write40XErrorResponse(w http.ResponseWriter, err error) {
	if err == context.RecordsNotFoundError {
		WriteErrorResponse(w, http.StatusNotFound, err.Error())
	} else {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
	}
}

func PostProcessResponseBare(w http.ResponseWriter, ctx context.CLIContext, body interface{}) {

	var (
		resp []byte
		err  error
	)

	cdc, err := ctx.GetCodec()
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	switch b := body.(type) {
	case []byte:
		resp = b

	default:
		if ctx.IsJSONIndent() {
			resp, err = cdc.MarshalJSONIndent(body, "", "  ")
		} else {
			resp, err = cdc.MarshalJSON(body)
		}

		if err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	ctx.Logger.Debug("PostProcessResponseBare", "data", string(resp))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}

func WriteGenStdTxResponse(writer http.ResponseWriter, cliCtx context.CLIContext, req BaseRequest, tx txs.ITx) {
	err := req.ValidateBasic()
	if err != nil {
		WriteErrorResponse(writer, http.StatusBadRequest, err.Error())
		return
	}

	cliCtx = req.Setup(cliCtx)

	maxGas := req.MaxGas
	if maxGas <= int64(0) {
		maxGas = cliCtx.MaxGas
	}

	signer, _ := types.AccAddressFromBech32(req.From)
	acc, _ := account.GetAccount(cliCtx, signer)

	nonce := int64(0)
	if req.Nonce != 0 {
		nonce = req.Nonce
	} else if acc != nil {
		nonce = acc.GetNonce()
	}

	nonce = nonce + 1

	var pubkey crypto.PubKey
	if acc != nil {
		pubkey = acc.GetPublicKey()
	}

	stdTx := txs.NewTxStd(tx, cliCtx.ChainID, types.NewInt(maxGas))
	stdTx.Signature = []txs.Signature{
		txs.Signature{
			Pubkey:    pubkey,
			Signature: nil,
			Nonce:     nonce,
		},
	}

	bz, err := cliCtx.JSONResult(stdTx)
	if err != nil {
		WriteErrorResponse(writer, http.StatusInternalServerError, err.Error())
		return
	}

	signBz := stdTx.BuildSignatureBytes(nonce, cliCtx.ChainID)
	signStr := base64.StdEncoding.EncodeToString(signBz)

	response := TxGenerateResponse{
		Code:      http.StatusOK,
		Tx:        string(bz),
		Signer:    signer,
		PubKey:    pubkey,
		Nonce:     nonce,
		SignBytes: signStr,
	}

	bz, err = cliCtx.JSONResult(response)
	if err != nil {
		WriteErrorResponse(writer, http.StatusInternalServerError, err.Error())
		return
	}

	cliCtx.Logger.Debug("WriteGenStdTxResponse result", "data", string(bz))
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(bz)
}

//form submit

func ParseRequestForm(r *http.Request) (br BaseRequest, err error) {
	height := int64(0)
	nonce := int64(0)
	maxGas := int64(0)
	indent := false

	from := r.FormValue("from")
	nonceStr := r.FormValue("nonce")
	if nonceStr != "" {
		if nonce, err = strconv.ParseInt(nonceStr, 10, 64); err != nil {
			return BaseRequest{}, errors.New("invalid nonce")
		}
	}
	maxGasStr := r.FormValue("max_gas")
	if maxGasStr != "" {
		if maxGas, err = strconv.ParseInt(maxGasStr, 10, 64); err != nil {
			return BaseRequest{}, errors.New("invalid max gas")
		}
	}

	chainId := r.FormValue("chain_id")
	mode := r.FormValue("mode")
	heightStr := r.FormValue("height")
	if heightStr != "" {
		height, _ = strconv.ParseInt(heightStr, 10, 64)
		if height < 0 {
			return BaseRequest{}, errors.New("invalid height")
		}
	}

	indentStr := r.FormValue("indent")
	if indentStr != "" {
		indent = true
	}

	br = NewBaseRequest(from, chainId, nonce, maxGas, height, indent, mode)
	br = br.Sanitize()
	err = br.ValidateBasic()

	return
}
