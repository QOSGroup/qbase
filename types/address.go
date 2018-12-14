package types

import (
	"bytes"
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bech32"
)

const (
	PREF_ADD = "address" // 地址前缀
)

// 地址，类型byte数组
type Address crypto.Address

// 地址转换为bytes
func (add Address) Bytes() []byte {
	return add[:]
}

// 判断地址是否为空
func (add Address) Empty() bool {
	if len(add[:]) == 0 {
		return true
	}
	return false
}

// 判断两地址是否相同
func (add Address) EqualsTo(anotherAdd Address) bool {
	if add.Empty() && anotherAdd.Empty() {
		return true
	}
	return bytes.Compare(add.Bytes(), anotherAdd.Bytes()) == 0
}

// 将base64的地址转换成bech32编码的字符串
func (add Address) String() string {
	bech32Addr, err := bech32.ConvertAndEncode(PREF_ADD, add.Bytes())
	if err != nil {
		panic(err)
	}
	return bech32Addr
}

// 由Bech32编码的地址解码为byte数组
// prefix_type 验证类型是否相符
func GetAddrFromBech32(bech32Addr string) (address []byte, err error) {
	prefix, bz, err := bech32.DecodeAndConvert(bech32Addr)
	address = bz
	if prefix != PREF_ADD {
		return nil, ErrInvalidAddress("Valid Address string should begin with %s" + PREF_ADD)
	}
	return
}

// photobuf: marshal得到地址原始byte数组
func (add Address) Marshal() ([]byte, error) {
	return add, nil
}

// photobuf：Unmarshal设置地址的byte数组值
func (add *Address) Unmarshal(data []byte) error {
	*add = data
	return nil
}

// 用Bech32编码将地址marshal为Json
func (add Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(add.String())
}

// 将Bech32编码的地址Json进行UnMarshal
func (add *Address) UnmarshalJSON(bech32Addr []byte) error {
	var s string
	err := json.Unmarshal(bech32Addr, &s)
	if err != nil {
		return err
	}
	add2, err := GetAddrFromBech32(s)
	if err != nil {
		return err
	}
	*add = add2
	return nil
}
