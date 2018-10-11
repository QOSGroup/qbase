package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/server"
	"github.com/QOSGroup/qbase/server/config"
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

type AppAccount struct {
	account.BaseAccount `json:"base_account"`
	Coins               Coins `json:"coins"`
}

func NewAppAccount() account.Account {
	return &AppAccount{
		BaseAccount: account.BaseAccount{},
		Coins:       []Coin{},
	}
}

func (acc *AppAccount) GetCoins() Coins {
	return acc.Coins
}

func (acc *AppAccount) SetCoins(coins Coins) error {
	acc.Coins = coins
	return nil
}

// QOS初始状态
type GenesisState struct {
	CAPubKey crypto.PubKey     `json:"ca_pub_key"`
	Accounts []*GenesisAccount `json:"accounts"`
}

// 初始账户
type GenesisAccount struct {
	Address types.Address `json:"address"`
	Coins   Coins         `json:"coins"`
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

func InitBaseCoinInit() server.AppInit {
	return server.AppInit{
		AppGenTx:    InitBaseCoinGenTx,
		AppGenState: InitBaseCoinGenStateJSON,
	}
}

type BaseCoinGenTx struct {
	Addr types.Address `json:"addr"`
}

// Generate a genesis transaction
func InitBaseCoinGenTx(cdc *amino.Codec, pk crypto.PubKey, genTxConfig config.GenTx) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	var addr types.Address
	var secret string
	addr, secret, err = GenerateCoinKey()
	if err != nil {
		return
	}

	var bz []byte
	simpleGenTx := BaseCoinGenTx{addr}
	bz, err = cdc.MarshalJSON(simpleGenTx)
	if err != nil {
		return
	}
	appGenTx = json.RawMessage(bz)

	mm := map[string]string{"secret": secret}
	bz, err = cdc.MarshalJSON(mm)
	if err != nil {
		return
	}
	cliPrint = json.RawMessage(bz)

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  10,
	}
	return
}

func InitBaseCoinGenStateJSON(cdc *amino.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

	if len(appGenTxs) != 1 {
		err = errors.New("must provide a single genesis transaction")
		return
	}

	var genTx BaseCoinGenTx
	err = cdc.UnmarshalJSON(appGenTxs[0], &genTx)
	if err != nil {
		return
	}

	appState = json.RawMessage(fmt.Sprintf(`{
  "accounts": [{
    "address": "%s",
    "coins": [
      {
        "name":"qstar",
        "amount":"100000000"
      }
    ]
  }]
}`, genTx.Addr))
	return
}

func GenerateCoinKey() (addr types.Address, secret string, err error) {
	//ed25519
	addr, _ = types.GetAddrFromBech32("address1k0m8ucnqug974maa6g36zw7g2wvfd4sug6uxay")
	secret = "0xa328891040ae9b773bcd30005235f99a8d62df03a89e4f690f9fa03abb1bf22715fc9ca05613f2d8061492e9f8149510b5b67d340d199ff24f34c85dbbbd7e0df780e9a6cc"
	return
}
