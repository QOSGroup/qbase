# BaseCoin Example

basecoin example基于qbase实现了简单的单次单币种，单个发送/接收账户的转账功能

## 使用步骤

1. Install basecoind and basecli</br>
在qbase项目根目录下
```
cd example/basecoin/cmd/basecoind
go install
cd ../basecli
go install
```
2. 初始化
```
basecoind init
{
  "chain_id": "basecoin",
  "node_id": "bada889c78e1a3936863e6a89eb766c28b398032",
  "app_message": {
    "name": "Jia",
    "pass": "12345678",
    "secret": "problem dutch dilemma climb endorse client despair ostrich cannon path once suspect place base brisk deposit area spike veteran coin injury dove electric famous"
  }
}
```
创建配置文件，以及创世账户“Jia”

3. 创建账户"Liu"
```
basecli keys add Liu
Enter a passphrase for your key:
Repeat the passphrase:
NAME:	TYPE:	ADDRESS:						PUBKEY:
Liu	local	address1q55ay4hdv33uplvvxpq0j8r7lpxunx8ytsvgkn	PubKeyEd25519{8329335BCB2B26E6D6B26854B72C4722B8740AE70A4454A499468E08D16DD29C}
**Important** write this seed phrase in a safe place.
It is the only way to recover your account if you ever forget your password.
book distance cart design another view olympic orbit leopard indoor tumble dutch random feel glad brother obvious sweet unlock degree eyebrow south final rather
```
4. 启动basecoin app
```
basecoind start --with-tendermint=true
```
5. 账户查询状态
```
basecli account --name=Jia
{
  "type": "basecoin/AppAccount",
  "value": {
    "base_account": {
      "account_address": "address10ly5e3qz3v3xy84ha46dylnyuuq773exa8xcxz",
      "public_key": null,
      "nonce": "0"
    },
    "coins": [
      {
        "coin_name": "qstar",
        "amount": "100000000"
      }
    ]
  }
} <nil>
```
6. 链内交易
```
basecli send --from=Jia --to=Liu --coin-name=qstar --coin-amount=10
Password to sign with 'Jia':
{"check_tx":{},"deliver_tx":{},"hash":"0677BB2E156496064960ED759BFEDBE6D09A8282","height":"22"}
```
7. 账户查询状态
```
basecli account --name=Jia
{
  "type": "basecoin/AppAccount",
  "value": {
    "base_account": {
      "account_address": "address10ly5e3qz3v3xy84ha46dylnyuuq773exa8xcxz",
      "public_key": {
        "type": "tendermint/PubKeyEd25519",
        "value": "8i28U4DmFV+szuTNyzOpOurAXmN9dAuPzzsvTDNHx54="
      },
      "nonce": "1"
    },
    "coins": [
      {
        "coin_name": "qstar",
        "amount": "99999990"
      }
    ]
  }
} <nil>
basecli account --name=Liu
{
  "type": "basecoin/AppAccount",
  "value": {
    "base_account": {
      "account_address": "address10j7njmfrfmfe5myr2scv4e8qw62027jw0kgtfu",
      "public_key": null,
      "nonce": "0"
    },
    "coins": [
      {
        "coin_name": "qstar",
        "amount": "10"
      }
    ]
  }
} <nil>
```

8. 查询交易
```
basecli tx 0677BB2E156496064960ED759BFEDBE6D09A8282
{
  "hash": "Bne7LhVklgZJYO11m/7b5tCagoI=",
  "height": "22",
  "tx": {
    "type": "qbase/txs/stdtx",
    "value": {
      "itx": {
        "type": "basecoin/SendTx",
        "value": {
          "from": "address10ly5e3qz3v3xy84ha46dylnyuuq773exa8xcxz",
          "to": "address10j7njmfrfmfe5myr2scv4e8qw62027jw0kgtfu",
          "coin": {
            "coin_name": "qstar",
            "amount": "10"
          }
        }
      },
      "sigature": [
        {
          "pubkey": {
            "type": "tendermint/PubKeyEd25519",
            "value": "8i28U4DmFV+szuTNyzOpOurAXmN9dAuPzzsvTDNHx54="
          },
          "signature": "oRb4JAXgUexpBrDuu0Ez/K9cq63rvaQMA4reL/nbt2OhwUrdHT3KoIEt1bOR00G/oo+STI1QdoDs0z+NGevACA==",
          "nonce": "1"
        }
      ],
      "chainid": "test-chain-vHi9Q2",
      "maxgas": "0"
    }
  },
  "result": {}
} <nil>
```

9. QCP交易</br>
qstar PriKey:</br>
0xa3288910405746e29aeec7d5ed56fac138b215e651e3244e6d995f25cc8a74c40dd1ef8d2e8ac876faaa4fb281f17fb9bebb08bc14e016c3a88c6836602ca97595ae32300b
```
basecli send-qcp --from=Jia --to=Liu --coin-name=qstar --coin-amount=10 --qcp-chain=qstar
Password to sign with 'Jia':
PriKey to sign with qstar chain:
{"check_tx":{},"deliver_tx":{"tags":[{"key":"cWNwLmZyb20=","value":"dGVzdC1jaGFpbi12SGk5UTI="},{"key":"cWNwLnRv","value":"cXN0YXI="},{"key":"cWNwLnNlcXVlbmNl","value":"MQ=="},{"key":"cWNwLmhhc2g=","value":"hpXUKq6grBHBbJtGm0g6kkVbu7wUTlNCUrj0HudTYHo="}]},"hash":"CB1BA7356F59CD41C4538CD2B0757CF1A5D17062","height":"83"}
```

10. QCP sequence 查询
```
basecli qcp inseq --chain-id=qstar
1
basecli qcp outseq --chain-id=qstar
1
```
11. QCP 交易结果查询

```
basecli qcp outtx --chain-id=qstar --seq=1
{
  "type": "qbase/txs/qcptx",
  "value": {
    "txstd": {
      "itx": {
        "type": "qbase/txs/qcpresult",
        "value": {
          "result": {
            "Code": 0,
            "Data": null,
            "Log": "",
            "GasWanted": "0",
            "GasUsed": "0",
            "FeeAmount": "0",
            "FeeDenom": "",
            "Tags": [
              {
                "key": "cWNwLmZyb20=",
                "value": "dGVzdC1jaGFpbi12SGk5UTI="
              },
              {
                "key": "cWNwLnRv",
                "value": "cXN0YXI="
              }
            ]
          },
          "qcporiginalsequence": "1",
          "qcpextends": "",
          "info": ""
        }
      },
      "sigature": null,
      "chainid": "qstar",
      "maxgas": "0"
    },
    "from": "test-chain-vHi9Q2",
    "to": "qstar",
    "sequence": "1",
    "sig": {
      "pubkey": null,
      "signature": null,
      "nonce": "0"
    },
    "blockheight": "83",
    "txindex": "0",
    "isresult": true,
    "extends": ""
  }
} <nil>
```

更多命令，查阅
```
basecli --help
```