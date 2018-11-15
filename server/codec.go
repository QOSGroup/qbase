package server

import (
	"bytes"
	"encoding/json"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
)

type Codec = amino.Codec

func New() *Codec {
	cdc := amino.NewCodec()
	return cdc
}

func RegisterCrypto(cdc *Codec) {
	cryptoAmino.RegisterAmino(cdc)
}

func MarshalJSONIndent(cdc *Codec, obj interface{}) ([]byte, error) {
	bz, err := cdc.MarshalJSON(obj)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	err = json.Indent(&out, bz, "", "  ")
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
