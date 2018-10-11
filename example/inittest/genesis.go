package inittest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/server"
	serverconfig "github.com/QOSGroup/qbase/server/config"
	"github.com/QOSGroup/qbase/types"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

// 创世配置app_state
// QOS初始状态
type GenesisState struct {
	CAPubKey crypto.PubKey     `json:"ca_pub_key"`
	Accounts []*GenesisAccount `json:"accounts"`
}

// 初始账户
type GenesisAccount struct {
	Address types.Address `json:"address"`
	Qos     types.BigInt  `json:"qos"`
	QscList []*QSC   `json:"qsc"`
}

// 给定 QOSAccount 创建 GenesisAccount
func NewGenesisAccount(aa *QOSAccount) *GenesisAccount {
	return &GenesisAccount{
		Address: aa.BaseAccount.GetAddress(),
		Qos:     aa.Qos,
		QscList: aa.QscList,
	}
}

// 给定 GenesisAccount 创建 QOSAccount
func (ga *GenesisAccount) ToQosAccount() (acc *QOSAccount, err error) {
	return &QOSAccount{
		BaseAccount: account.BaseAccount{
			AccountAddress: ga.Address,
		},
		Qos:     ga.Qos,
		QscList: ga.QscList,
	}, nil
}

func InitTestAppInit() server.AppInit {
	return server.AppInit{
		AppGenTx:    InitTestAppGenTx,
		AppGenState: InitTestAppGenStateJSON,
	}
}

type BaseCoinGenTx struct {
	Addr types.Address `json:"addr"`
}

// Generate a genesis transaction
func InitTestAppGenTx(cdc *amino.Codec, pk crypto.PubKey, genTxConfig serverconfig.GenTx) (
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

func InitTestAppGenStateJSON(cdc *amino.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

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
  "ca_pub_key": {
    "type": "tendermint/PubKeyEd25519",
    "value": "MpVfPwWAwh/d53kj5ZfNCxdQ69yUFuz19J5ygCByGCc="
  },
  "accounts": [{
    "address": "%s",
	"qos": "100000000",
    "qsc": [
      {
        "coin_name":"qstar",
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
