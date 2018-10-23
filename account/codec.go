package account

import (
	go_amino "github.com/tendermint/go-amino"
)

// 为包内定义结构注册codec
func RegisterCodec(cdc *go_amino.Codec) {
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "qbase/account/BaseAccount", nil)
}
