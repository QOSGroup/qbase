package types

import (
	"encoding/binary"
	"encoding/json"
	"path/filepath"
	"regexp"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"
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
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(in))
	return bz
}

// 函数：bool 转化为 []byte
func Bool2Byte(in bool) []byte {
	if in {
		return []byte{1}
	}
	return []byte{0}
}

// 功能：检查 QscName 的合法性
// 备注：合法（3-10个字符，数字-字母-下划线）
func CheckQscName(qscName string) bool {
	ret := len(qscName) > 10 || len(qscName) < 3
	reg := regexp.MustCompile(`[^(a-z 0-9 A-Z _)]`)
	ret = !ret && !reg.Match([]byte(qscName))

	return ret
}

// Slight modification of the RFC3339Nano but it right pads all zeros and drops the time zone info
const SortableTimeFormat = "2006-01-02T15:04:05.000000000"

// Formats a time.Time into a []byte that can be sorted
func FormatTimeBytes(t time.Time) []byte {
	return []byte(t.UTC().Round(0).Format(SortableTimeFormat))
}

// Parses a []byte encoded using FormatTimeKey back into a time.Time
func ParseTimeBytes(bz []byte) (time.Time, error) {
	str := string(bz)
	t, err := time.Parse(SortableTimeFormat, str)
	if err != nil {
		return t, err
	}
	return t.UTC().Round(0), nil
}

func GetGenesisDoc(rootDir string) (*tmtypes.GenesisDoc, error) {
	path := filepath.Join(rootDir, "config", "genesis.json")
	doc, err := tmtypes.GenesisDocFromFile(path)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func GetChainID(rootDir string) (string, error) {
	doc, err := GetGenesisDoc(rootDir)
	if err != nil {
		return "", err
	}
	return doc.ChainID, nil
}
