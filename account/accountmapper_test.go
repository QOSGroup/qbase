package account

import (
	"fmt"
	"testing"

	"github.com/QOSGroup/qbase/mapper"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/store"
	"github.com/QOSGroup/qbase/types"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

func defaultContext(key store.StoreKey, mapperMap map[string]mapper.IMapper) context.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, store.StoreTypeIAVL, db)
	// latestVersion也是int64经过amino编码后存储在相应键值下的
	cms.LoadLatestVersion()
	ctx := context.NewContext(cms, abci.Header{}, false, log.NewNopLogger(), mapperMap)
	return ctx
}

func TestAccountMapperGetSet(t *testing.T) {

	cdc := MakeCdc()

	seedMapper := NewAccountMapper(ProtoBaseAccount)
	seedMapper.SetCodec(cdc)

	mapperMap := make(map[string]mapper.IMapper)
	mapperMap[seedMapper.Name()] = seedMapper

	ctx := defaultContext(seedMapper.GetStoreKey(), mapperMap)

	mapper, _ := ctx.Mapper(AccountMapperName).(*AccountMapper)

	for i := 0; i < 100; i++ {
		pubkey := ed25519.GenPrivKey().PubKey()
		addr := types.Address(pubkey.Address())

		// 没有存过该addr，取出来应为nil
		acc := mapper.GetAccount(addr)
		require.Nil(t, acc)

		acc = mapper.NewAccountWithAddress(addr)
		require.NotNil(t, acc)
		require.Equal(t, addr, acc.GetAddress())
		require.EqualValues(t, nil, acc.GetPubicKey())
		require.EqualValues(t, 0, acc.GetNonce())

		// 新的account尚未存储，依然取出nil
		require.Nil(t, mapper.GetAccount(addr))

		nonce := int64(20)
		acc.SetNonce(nonce)
		acc.SetPublicKey(pubkey)
		// 存储account
		mapper.SetAccount(acc)

		// 将account以地址取出并验证
		acc = mapper.GetAccount(addr)
		require.NotNil(t, acc)
		require.Equal(t, nonce, acc.GetNonce())

	}
	// 批量处理特定前缀存储的账户
	mapper.IterateAccounts(func(acc Account) bool {
		fmt.Println(acc.GetAddress())
		bz := mapper.EncodeObject(acc)
		var acc1 Account
		mapper.DecodeObject(bz, &acc1)
		require.Equal(t, acc, acc1)
		return false
	})
}
