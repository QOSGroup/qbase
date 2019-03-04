package mapper

import (
	"github.com/QOSGroup/qbase/types"
	"reflect"
	"strconv"
	"testing"

	"github.com/QOSGroup/qbase/store"
	"github.com/stretchr/testify/require"
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

func getMapper() *BaseMapper {
	var cdc = go_amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)
	cdc.RegisterConcrete(&mockStruct{}, "mock/struct", nil)
	cdc.RegisterInterface((*mockInterface)(nil), nil)

	storeKey := types.NewKVStoreKey("base")

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(storeKey, types.StoreTypeIAVL, db)
	cms.LoadLatestVersion()

	cms.GetStore(storeKey)

	return &BaseMapper{cdc: cdc, store: cms.GetStore(storeKey).(store.KVStore)}
}

func TestBaseMapper_EncodeObject(t *testing.T) {

	baseMapper := getMapper()

	strKey := []byte("stringkey")
	_, exsits := baseMapper.GetString(strKey)
	require.Equal(t, false, exsits)

	baseMapper.Set(strKey, "1111")
	s, exsits := baseMapper.GetString(strKey)
	require.Equal(t, true, exsits)
	require.Equal(t, "1111", s)

	require.Panics(t, func() {
		_, _ = baseMapper.GetInt64(strKey)
	})

	intKey := []byte("interkey")
	_, exsits = baseMapper.GetInt64(intKey)
	require.Equal(t, false, exsits)

	baseMapper.Set(intKey, int64(98))
	i, exsits := baseMapper.GetInt64(intKey)
	require.Equal(t, true, exsits)
	require.Equal(t, int64(98), i)

	require.Panics(t, func() {
		_, _ = baseMapper.GetString(intKey)
	})

	booKey := []byte("booKey")
	_, exsits = baseMapper.GetBool(booKey)
	require.Equal(t, false, exsits)

	baseMapper.Set(booKey, true)
	b, exsits := baseMapper.GetBool(booKey)
	require.Equal(t, true, exsits)
	require.Equal(t, true, b)

	require.Panics(t, func() {
		_, _ = baseMapper.GetString(booKey)
	})

	count := 0
	baseMapper.Iterator(nil, func(bz []byte) (stop bool) {
		count++
		return
	})
	require.Equal(t, 3, count)

	count = 0
	baseMapper.Iterator([]byte("interkey"), func(bz []byte) (stop bool) {
		count++
		return
	})
	require.Equal(t, 1, count)

	k1 := []byte("account")
	v1 := []byte("addressaaa")

	var v string
	exsits = baseMapper.Get(k1, &v)
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

func TestBaseMapper_IteratorWithType(t *testing.T) {

	//bool

	baseMapper := getMapper()

	bPrefix := "bool_"
	for i := 0; i <= 10; i++ {
		s := strconv.Itoa(i)
		key := append([]byte(bPrefix), []byte(s)...)
		if i < 5 {
			baseMapper.Set(key, true)
		} else {
			baseMapper.Set(key, false)
		}
	}

	baseMapper.IteratorWithType([]byte(bPrefix), reflect.TypeOf(true), func(key []byte, dataPtr interface{}) bool {
		bPtr := dataPtr.(*bool)
		b := *bPtr
		if b {
			require.Equal(t, true, b)
		}
		return false
	})

	//mockStruct
	sPrefix := "mocks_"
	for i := 0; i <= 10; i++ {
		s := strconv.Itoa(i)
		key := append([]byte(sPrefix), []byte(s)...)

		bz, _ := ed25519.GenPrivKey().Sign(key)
		baseMapper.Set(key, mockStruct{
			Nonce:  int64(i),
			Pubkey: ed25519.GenPrivKey().PubKey(),
			Name:   s,
			Sign:   bz,
		})
	}

	baseMapper.IteratorWithType([]byte(sPrefix), reflect.TypeOf(mockStruct{}), func(key []byte, dataPtr interface{}) bool {
		sPtr := dataPtr.(*mockStruct)
		mockStruct := *sPtr
		require.Equal(t, 64, len(mockStruct.Sign))
		return false
	})

}

func TestDiggggggggggggggggggggHole(t *testing.T) {
	baseMapper := getMapper()

	prefix := []byte("a")

	for i := 0; i < 10; i++ {
		s := strconv.Itoa(i)
		key := append(prefix, []byte(s)...)
		baseMapper.Set(key, i)
	}

	key := append(prefix, []byte("1")...)
	i, _ := baseMapper.GetInt64(key)

	//not 1
	require.Equal(t, int64(9), i)

	count := 0
	baseMapper.Iterator(prefix, func(_ []byte) bool {
		count++
		return false
	})

	//not 10
	require.Equal(t, 1, count)

}
