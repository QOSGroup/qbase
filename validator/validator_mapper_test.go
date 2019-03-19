package validator

import (
	"testing"

	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/store"
	"github.com/QOSGroup/qbase/types"
	"github.com/stretchr/testify/require"
	go_amino "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, types.Address) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := types.Address(pub.Address())
	return key, pub, addr
}

func defaultContext(key store.StoreKey, mapperMap map[string]mapper.IMapper) context.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, store.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := context.NewContext(cms, abci.Header{}, false, log.NewNopLogger(), mapperMap)
	return ctx
}

func getValidatorMapper() *ValidatorMapper {
	cdc := go_amino.NewCodec()

	seedMapper := NewValidatorMapper()
	seedMapper.SetCodec(cdc)

	mapperMap := make(map[string]mapper.IMapper)
	mapperMap[seedMapper.MapperName()] = seedMapper

	ctx := defaultContext(seedMapper.GetStoreKey(), mapperMap)
	return GetValidatorMapper(ctx)
}

func TestValidatorMapper(t *testing.T) {

	valMapper := getValidatorMapper()
	b := valMapper.IsEnableValidatorUpdated()
	require.Equal(t, false, b)

	valMapper.EnableValidatorUpdated()

	b = valMapper.IsEnableValidatorUpdated()
	require.Equal(t, true, b)

	valMapper.DisableValidatorUpdated()

	b = valMapper.IsEnableValidatorUpdated()
	require.Equal(t, false, b)

	s := valMapper.GetValidatorUpdateSet()
	require.Equal(t, 0, len(s))

	s = []abci.ValidatorUpdate{
		{}, {},
	}

	valMapper.SetValidatorUpdateSet(s)
	s = valMapper.GetValidatorUpdateSet()
	require.Equal(t, 2, len(s))

	valMapper.ClearValidatorUpdateSet()

	s = valMapper.GetValidatorUpdateSet()
	require.Equal(t, 0, len(s))

	addr, _ := valMapper.GetLastBlockProposer()
	require.Equal(t, true, addr.Empty())

	addr = types.Address{12, 20, 32}
	valMapper.SetLastBlockProposer(addr)

	addr, _ = valMapper.GetLastBlockProposer()
	require.Equal(t, false, addr.Empty())

	valMapper.ClearValidatorUpdateSet()
	_, pub, _ := keyPubAddr()
	_, pub1, _ := keyPubAddr()
	_, pub2, _ := keyPubAddr()
	_, pub3, _ := keyPubAddr()

	valMapper.AddValidatorUpdate(pub, uint64(0))

	r := valMapper.GetValidatorUpdateSet()
	require.Equal(t, int(1), len(r))

	valMapper.AddValidatorUpdate(pub1, uint64(1))

	valMapper.AddValidatorUpdate(pub2, uint64(2))

	valMapper.AddValidatorUpdate(pub3, uint64(3))

	r = valMapper.GetValidatorUpdateSet()
	require.Equal(t, int(4), len(r))

	valMapper.AddValidatorUpdate(pub, uint64(4))

	r = valMapper.GetValidatorUpdateSet()
	require.Equal(t, int(4), len(r))

}
