package account

import (
	"fmt"

	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/types"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
)

const (
	AccountMapperName = "acc"      // 用户获取账户存储的store的键名
	accountStoreKey   = "account:" // 便于获取全部账户的通用存储键名，继承BaseAccount时，可根据不同业务设置存储前缀
)

func BuildAccountStoreQueryPath() []byte {
	return []byte(fmt.Sprintf("/store/%s/key", AccountMapperName))
}

// 对BaseAccount存储操作进行包装的结构，可进行序列化
type AccountMapper struct {
	*mapper.BaseMapper
	proto func() Account
}

var _ mapper.IMapper = (*AccountMapper)(nil)

// 用给定编码和原型生成mapper
func NewAccountMapper(cdc *go_amino.Codec, proto func() Account) *AccountMapper {
	var accountMapper = AccountMapper{}
	accountMapper.BaseMapper = mapper.NewBaseMapper(cdc, AccountMapperName)
	accountMapper.proto = proto
	return &accountMapper
}

func (mapper *AccountMapper) Copy() mapper.IMapper {
	cpyMapper := &AccountMapper{}
	cpyMapper.proto = mapper.proto
	cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
	return cpyMapper
}

// 用指定地址生成账户返回
func (mapper *AccountMapper) NewAccountWithAddress(add types.Address) Account {
	acc := mapper.proto()
	err := acc.SetAddress(add)
	if err != nil {
		panic(err)
	}
	return acc
}

// 将地址转换成存储通用的key
func AddressStoreKey(addr types.Address) []byte {
	return append([]byte(accountStoreKey), addr.Bytes()...)
}

// 从存储中获得账户
func (mapper *AccountMapper) GetAccount(addr types.Address) (acc Account) {
	mapper.Get(AddressStoreKey(addr), &acc)
	return acc
}

// 存储账户
func (mapper *AccountMapper) SetAccount(acc Account) {
	mapper.Set(AddressStoreKey(acc.GetAddress()), acc)
}

// 遍历并用闭包批量处理所存储的全部账户
func (mapper *AccountMapper) IterateAccounts(process func(Account) (stop bool)) {
	iter := types.KVStorePrefixIterator(mapper.GetStore(), []byte(accountStoreKey))
	for {
		if !iter.Valid() {
			return
		}
		val := iter.Value()
		var acc Account
		mapper.DecodeObject(val, &acc)
		if process(acc) {
			return
		}
		iter.Next()
	}
}

// 获取地址代表账户的公钥
func (mapper *AccountMapper) GetPubKey(addr types.Address) (crypto.PubKey, types.Error) {
	acc := mapper.GetAccount(addr)
	if acc == nil {
		return nil, types.ErrUnknownAddress(addr.String())
	}
	return acc.GetPubicKey(), nil
}

// 获取地址代表账户的nonce
func (mapper *AccountMapper) GetNonce(addr types.Address) (int64, types.Error) {
	acc := mapper.GetAccount(addr)
	if acc == nil {
		return 0, types.ErrUnknownAddress(addr.String())
	}
	return acc.GetNonce(), nil
}

// 为特定账户设置新的nonce值
func (mapper *AccountMapper) SetNonce(addr types.Address, nonce int64) types.Error {
	acc := mapper.GetAccount(addr)
	if acc == nil {
		return types.ErrUnknownAddress(addr.String())
	}
	err := acc.SetNonce(nonce)
	if err != nil {
		panic(err)
	}
	mapper.SetAccount(acc)
	return nil
}
