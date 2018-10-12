package kvstore

import (
	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/store"
)

const KvMapperName = "kvmapper"

type KvMapper struct {
	*mapper.BaseMapper
}

func NewKvMapper(main store.StoreKey) *KvMapper {
	var txMapper = KvMapper{}
	txMapper.BaseMapper = mapper.NewBaseMapper(main)
	return &txMapper
}

func (mapper *KvMapper) Copy() mapper.IMapper {
	cpyMapper := &KvMapper{}
	cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
	return cpyMapper
}

func (mapper *KvMapper) Name() string {
	return KvMapperName
}

var _ mapper.IMapper = (*KvMapper)(nil)

func (mapper *KvMapper) SaveKV(key string, value string) {
	mapper.BaseMapper.Set([]byte(key), value)
}

func (mapper *KvMapper) GetKey(key string) (v string) {
	mapper.BaseMapper.Get([]byte(key), &v)
	return
}
