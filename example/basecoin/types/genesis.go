package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/QOSGroup/qbase/account"
	clikeys "github.com/QOSGroup/qbase/client/keys"
	"github.com/QOSGroup/qbase/keys"
	"github.com/QOSGroup/qbase/server"
	"github.com/QOSGroup/qbase/server/config"
	"github.com/QOSGroup/qbase/types"
	"github.com/spf13/pflag"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
	"os"
	"path/filepath"

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
	Address types.Address `json:"address"`
	Coins   types.Coins   `json:"coins"`
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

func BaseCoinInit() server.AppInit {
	fsAppGenState := pflag.NewFlagSet("", pflag.ContinueOnError)

	fsAppGenTx := pflag.NewFlagSet("", pflag.ContinueOnError)
	fsAppGenTx.String(server.FlagName, "", "validator moniker, required")
	fsAppGenTx.String(server.FlagClientHome, DefaultCLIHome,
		"home directory for the client, used for key generation")
	fsAppGenTx.Bool(server.FlagOWK, false, "overwrite the accounts created")

	return server.AppInit{
		FlagsAppGenState: fsAppGenState,
		FlagsAppGenTx:    fsAppGenTx,
		AppGenTx:         BaseCoinAppGenTx,
		AppGenState:      BaseCoinAppGenState,
	}
}

type BaseCoinGenTx struct {
	Addr types.Address `json:"addr"`
}

// Generate a genesis transaction
func BaseCoinAppGenTx(cdc *amino.Codec, pk crypto.PubKey, genTxConfig config.GenTx) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	var addr types.Address
	var secret string
	addr, secret, err = GenerateCoinKey(cdc, genTxConfig.CliRoot)
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

	mm := map[string]string{"name": DefaultAccountName, "pass": DefaultAccountPass, "secret": secret}
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

func BaseCoinAppGenState(cdc *amino.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

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
	}`, genTx.Addr))
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
