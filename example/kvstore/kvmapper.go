package kvstore

import (
	"github.com/QOSGroup/qbase/mapper"
)

const KVMapperName = "kv"

type KvMapper struct {
	*mapper.BaseMapper
}

func NewKvMapper() *KvMapper {
	var txMapper = KvMapper{}
	txMapper.BaseMapper = mapper.NewBaseMapper(nil, KVMapperName)
	return &txMapper
}

func (mapper *KvMapper) Copy() mapper.IMapper {
	cpyMapper := &KvMapper{}
	cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
	return cpyMapper
}

var _ mapper.IMapper = (*KvMapper)(nil)

func (mapper *KvMapper) SaveKV(key string, value string) {
	mapper.BaseMapper.Set([]byte(key), value)
}

func (mapper *KvMapper) GetKey(key string) (v string) {
	mapper.BaseMapper.Get([]byte(key), &v)
	return
}
