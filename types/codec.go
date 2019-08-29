package types

import (
	go_amino "github.com/tendermint/go-amino"
)

func RegisterCodec(cdc *go_amino.Codec) {
	cdc.RegisterConcrete(&AccAddress{}, "qbase/types/AccAddress", nil)
	cdc.RegisterConcrete(&ValAddress{}, "qbase/types/ValAddress", nil)
	cdc.RegisterConcrete(&ConsAddress{}, "qbase/types/ConsAddress", nil)
	cdc.RegisterInterface((*Address)(nil), nil)
}
