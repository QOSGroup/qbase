package inittest

import (
	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/store"
)

const TX_MAPPER_NAME = "inittestmapper"

type TxMapper struct {
	*mapper.BaseMapper
}

func NewTxMapper(main store.StoreKey) *TxMapper {
	var txMapper = TxMapper{}
	txMapper.BaseMapper = mapper.NewBaseMapper(main)
	return &txMapper
}

func (mapper *TxMapper) Copy() mapper.IMapper {
	cpyMapper := &TxMapper{}
	cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
	return cpyMapper
}

func (mapper *TxMapper) Name() string {
	return TX_MAPPER_NAME
}

var _ mapper.IMapper = (*TxMapper)(nil)

func (mapper *TxMapper) SaveKV(key string, value string) {
	mapper.BaseMapper.SetObject([]byte(key), value)
}

func (mapper *TxMapper) GetKey(key string) (v string) {
	mapper.BaseMapper.GetObject([]byte(key), &v)
	return
}
