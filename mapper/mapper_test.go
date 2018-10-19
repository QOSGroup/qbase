package mapper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/QOSGroup/qbase/store"
	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
	dbm "github.com/tendermint/tendermint/libs/db"
)

type mockInterface interface {
	NName() string
}

type mockStruct struct {
	Name   string
	Pubkey crypto.PubKey
	Nonce  int64
	Sign   []byte
}

func (mockStruct *mockStruct) NName() string {
	return mockStruct.Name
}

func TestBaseMapper_EncodeObject(t *testing.T) {

	var cdc = go_amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)
	cdc.RegisterConcrete(&mockStruct{}, "mock/struct", nil)
	cdc.RegisterInterface((*mockInterface)(nil), nil)

	var baseMapper *BaseMapper
	storeKey := store.NewKVStoreKey("base")

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(storeKey, store.StoreTypeIAVL, db)
	cms.LoadLatestVersion()

	cms.GetStore(storeKey)

	baseMapper = &BaseMapper{cdc: cdc, store: cms.GetStore(storeKey).(store.KVStore)}

	k1 := []byte("account")
	v1 := []byte("addressaaa")

	var v string
	exsits := baseMapper.Get(k1, &v)
	require.Equal(t, false, exsits)

	baseMapper.Set(k1, v1)

	exsits = baseMapper.Get(k1, &v)
	require.Equal(t, true, exsits)

	baseMapper.Del(k1)

	exsits = baseMapper.Get(k1, &v)
	require.Equal(t, false, exsits)

	key := ed25519.GenPrivKey()
	pub := key.PubKey()

	bytes := []byte("hello world")
	nonce := int64(12)

	sig := mockStruct{
		Name:   "test",
		Pubkey: pub,
		Sign:   bytes,
		Nonce:  nonce,
	}

	encodeBytes := baseMapper.EncodeObject(sig)

	var mockInterface mockInterface
	baseMapper.DecodeObject(encodeBytes, &mockInterface)

	decodeSig, _ := mockInterface.(*mockStruct)

	require.Equal(t, nonce, decodeSig.Nonce)
	require.Equal(t, bytes, decodeSig.Sign)
	require.Equal(t, pub, decodeSig.Pubkey)

	nonceEncodeBytes := baseMapper.EncodeObject(nonce)
	var nonceInt int64
	baseMapper.DecodeObject(nonceEncodeBytes, &nonceInt)

	require.Equal(t, nonce, nonceInt)

}
