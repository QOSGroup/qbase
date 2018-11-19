package keys

import (
	go_amino "github.com/tendermint/go-amino"
)

func RegisterCodec(cdc *go_amino.Codec) {
	cdc.RegisterInterface((*Info)(nil), nil)
	cdc.RegisterConcrete(&localInfo{}, "qbase/keys/localInfo", nil)
	cdc.RegisterConcrete(&offlineInfo{}, "qbase/keys/offlineInfo", nil)
	cdc.RegisterConcrete(&importInfo{}, "qbase/keys/importInfo", nil)
}
