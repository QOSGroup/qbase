package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp"
)

// SortedJSON takes any JSON and returns it sorted by keys. Also, all white-spaces
// are removed.
// This method can be used to canonicalize JSON to be returned by GetSignBytes,
// e.g. for the ledger integration.
// If the passed JSON isn't valid it will return an error.
func SortJSON(toSortJSON []byte) ([]byte, error) {
	var c interface{}
	err := json.Unmarshal(toSortJSON, &c)
	if err != nil {
		return nil, err
	}
	js, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return js, nil
}

// MustSortJSON is like SortJSON but panic if an error occurs, e.g., if
// the passed JSON isn't valid.
func MustSortJSON(toSortJSON []byte) []byte {
	js, err := SortJSON(toSortJSON)
	if err != nil {
		panic(err)
	}
	return js
}

// 函数：int64 转化为 []byte
func Int2Byte(in int64) []byte {
	var ret = bytes.NewBuffer([]byte{})
	err := binary.Write(ret, binary.BigEndian, in)
	if err != nil {
		fmt.Printf("Int2Byte error:%s", err.Error())
		return nil
	}

	return ret.Bytes()
}

// 函数：bool 转化为 []byte
func Bool2Byte(in bool) []byte {
	var ret = bytes.NewBuffer([]byte{})
	err := binary.Write(ret, binary.BigEndian, in)
	if err != nil {
		fmt.Printf("Bool2Byte error:%s", err.Error())
		return nil
	}

	return ret.Bytes()
}

// 功能：检查 QscName 的合法性
// 备注：合法（3-10个字符，数字-字母-下划线）
func CheckQscName(qscName string) bool {
	ret := len(qscName) > 10 || len(qscName) < 3
	reg := regexp.MustCompile(`[^(a-z 0-9 A-Z _)]`)
	ret = !ret && !reg.Match([]byte(qscName))

	return ret
}
