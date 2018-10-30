package types

import "github.com/tendermint/tendermint/crypto"

// app_state in genesis.json
type GenesisState struct {
	QCPs []*QCPConfig `json:"qcps"`
}

// QCP配置
type QCPConfig struct {
	Name    string        `json:"name"`
	ChainId string        `json:"chain_id"`
	PubKey  crypto.PubKey `json:"pub_key"`
}
