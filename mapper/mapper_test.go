package mapper

import (
	"testing"

	"github.com/stretchr/testify/require"

	go_amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
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
	baseMapper = &BaseMapper{cdc: cdc}

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
