package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/QOSGroup/qbase/account"
	clikeys "github.com/QOSGroup/qbase/client/keys"
	"github.com/QOSGroup/qbase/keys"
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"

	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	DefaultAccountName = "Jia"
	DefaultAccountPass = "12345678"
)

var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.basecli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.basecoind")
)

// QOS初始状态
type GenesisState struct {
	CAPubKey crypto.PubKey     `json:"pub_key"`
	Accounts []*GenesisAccount `json:"accounts"`
}

// 初始账户
type GenesisAccount struct {
	Address types.Address   `json:"address"`
	Coins   types.BaseCoins `json:"coins"`
}

// 给定 AppAccpunt 创建 GenesisAccount
func NewGenesisAccount(aa *AppAccount) *GenesisAccount {
	return &GenesisAccount{
		Address: aa.BaseAccount.GetAddress(),
		Coins:   aa.Coins,
	}
}

// 给定 GenesisAccount 创建 AppAccpunt
func (ga *GenesisAccount) ToAppAccount() (acc *AppAccount, err error) {
	return &AppAccount{
		BaseAccount: account.BaseAccount{
			AccountAddress: ga.Address,
		},
		Coins: ga.Coins,
	}, nil
}

type BaseCoinGenTx struct {
	Addr types.Address `json:"addr"`
}

func BaseCoinAppGenState(cdc *amino.Codec, appGenTxs BaseCoinGenTx) (appState json.RawMessage, err error) {

	appState = json.RawMessage(fmt.Sprintf(`{
		"qcps":[{
			"name": "qstar",
			"chain_id": "qstar",
			"pub_key":{
        		"type": "tendermint/PubKeyEd25519",
        		"value": "ish2+qpPsoHxf7m+uwi8FOAWw6iMaDZgLKl1la4yMAs="
			}
		}],
  		"accounts": [{
    		"address": "%s",
    		"coins": [
      			{
        			"coin_name":"qstar",
        			"amount":"100000000"
      			}
			]
  		}]
	}`, appGenTxs.Addr))
	return
}

func GenerateCoinKey(cdc *amino.Codec, clientRoot string) (addr types.Address, mnemonic string, err error) {

	db, err := dbm.NewGoLevelDB(clikeys.KeyDBName, filepath.Join(clientRoot, "keys"))
	if err != nil {
		return types.Address([]byte{}), "", err
	}
	keybase := keys.New(db, cdc)

	info, secret, err := keybase.CreateEnMnemonic(DefaultAccountName, DefaultAccountPass)
	if err != nil {
		return types.Address([]byte{}), "", err
	}

	addr = types.Address(info.GetPubKey().Address())
	return addr, secret, nil
}
