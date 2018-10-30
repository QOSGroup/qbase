package mapper

import (
	"reflect"

	"github.com/QOSGroup/qbase/store"
	go_amino "github.com/tendermint/go-amino"
)

type IMapper interface {
	Copy() IMapper

	//BaseMapper implement below methods
	MapperName() string
	GetStoreKey() store.StoreKey

	SetStore(store store.KVStore)
	SetCodec(cdc *go_amino.Codec)
}

type BaseMapper struct {
	cdc   *go_amino.Codec //:( 世界唯一
	key   store.StoreKey  //:) 确保唯一
	store store.KVStore   //Important:注意在不同的context中要覆盖该值
}

func NewBaseMapper(cdc *go_amino.Codec, mapperName string) *BaseMapper {
	return &BaseMapper{cdc: cdc, key: store.NewKVStoreKey(mapperName)}
}

func (baseMapper *BaseMapper) Copy() *BaseMapper {
	return &BaseMapper{
		cdc:   baseMapper.cdc,
		key:   baseMapper.key,
		store: baseMapper.store,
	}
}

func (baseMapper *BaseMapper) MapperName() string {
	return baseMapper.key.Name()
}

func (baseMapper *BaseMapper) isRegistered() bool {
	return baseMapper.cdc != nil && baseMapper.store != nil
}

func (baseMapper *BaseMapper) GetStore() store.KVStore {
	return baseMapper.store
}

func (baseMapper *BaseMapper) SetStore(store store.KVStore) {
	baseMapper.store = store
}

func (baseMapper *BaseMapper) GetStoreKey() store.StoreKey {
	return baseMapper.key
}

func (baseMapper *BaseMapper) Get(key []byte, ptr interface{}) (exsits bool) {

	if !baseMapper.isRegistered() {
		panic("mapper it's not prepared to work. you may forgot to register this mapper")
	}

	bz := baseMapper.store.Get(key)
	if bz == nil {
		exsits = false
		return
	}
	exsits = true
	baseMapper.DecodeObject(bz, ptr)
	return
}

func (baseMapper *BaseMapper) Set(key []byte, val interface{}) {

	if !baseMapper.isRegistered() {
		panic("mapper it's not prepared to work. you may forgot to register this mapper")
	}

	bz := baseMapper.EncodeObject(val)
	baseMapper.store.Set(key, bz)
}

func (baseMapper *BaseMapper) Del(key []byte) {

	if !baseMapper.isRegistered() {
		panic("mapper it's not prepared to work. you may forgot to register this mapper")
	}

	baseMapper.store.Delete(key)
}

func (baseMapper *BaseMapper) GetCodec() *go_amino.Codec {
	return baseMapper.cdc
}

func (baseMapper *BaseMapper) SetCodec(cdc *go_amino.Codec) {
	baseMapper.cdc = cdc
}

func (baseMapper *BaseMapper) EncodeObject(obj interface{}) []byte {
	bytes, err := baseMapper.cdc.MarshalBinaryBare(obj)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (baseMapper *BaseMapper) DecodeObject(bytes []byte, ptr interface{}) {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic("ptr must be a pointer")
	}

	err := baseMapper.cdc.UnmarshalBinaryBare(bytes, ptr)
	if err != nil {
		panic(err)
	}
}
