# BaseCoin Example

basecoin example基于qbase实现了简单的单次单币种，单个发送/接收账户的转账功能

## 使用步骤

1. Install basecoind and basecli</br>
在qbase项目根目录下
```
$ cd example/basecoin/cmd/basecoind
$ go install
$ cd ../basecli
$ go install
```
2. 初始化
```
$ basecoind init --chain-id basecoin --moniker basecoin-node
```

```
{
 "moniker": "basecoin-node",
 "chain_id": "basecoin",
 "node_id": "2227601e0dfb1561977bf5877a7f7de6357bc183",
 "gentxs_dir": "",
 "app_message": {
  "qcps": [
   {
    "name": "qstar",
    "chain_id": "qstar",
    "pub_key": {
     "type": "tendermint/PubKeyEd25519",
     "value": "ish2+qpPsoHxf7m+uwi8FOAWw6iMaDZgLKl1la4yMAs="
    }
   }
  ],
  "accounts": [
   {
    "address": "basecoin1pv085zl4scejgs3ns9xumk4eg5jdxwj7ydq34e",
    "coins": [
     {
      "coin_name": "qstar",
      "amount": "100000000"
     }
    ]
   }
  ]
 }
}

```

命令执行完成后,配置文件初始化完成并创建了创世账户“basecoin1pv085zl4scejgs3ns9xumk4eg5jdxwj7ydq34e”.

> 配置文件默认目录为$HOME/.basecoind/config

> 创世账户默认名称为`Jia`,密码为`123456`. 可以通过`basecli keys命令进行查看操作`



3. 创建本地账户"Liu"

```
$ basecli keys add Liu
```

```
Enter a passphrase for your key:
Repeat the passphrase:
NAME:   TYPE:   ADDRESS:                                                PUBKEY:
Liu     local   basecoin1sgp03a7l0jcuzeue2tzrucdfe8f0k9vtg269me basecoinpub1zcjduepqmcegq0pzuw6uaw7v3swpaxlt0f58hpht2d65tvax4quk5r73amvsme39q8
**Important** write this seed phrase in a safe place.
It is the only way to recover your account if you ever forget your password.

damage glimpse badge immense fix tent rebel excess news want base stone toddler decide warfare engine juice balance swing cup candy title fruit deposit
```
4. 启动basecoin app
```
$ basecoind start
```
5. 查询账户信息
```
$ basecli query account Jia
```

```
{"type":"basecoin/AppAccount","value":{"base_account":{"account_address":"basecoin1pv085zl4scejgs3ns9xumk4eg5jdxwj7ydq34e","public_key":null,"nonce":"0"},"coins":[{"coin_name":"qstar","amount"
:"100000000"}]}}
```

```
$ basecli query account Liu
```

```
ERROR: account not exists
```

> 本地账户`Liu`未在链上

6. 链内交易
```
$ basecli tx send --from=Jia --to=Liu --coin-name=qstar --coin-amount=10
```

```
Password to sign with 'Jia':
{"check_tx":{"gasWanted":"100000","gasUsed":"6958"},"deliver_tx":{"gasWanted":"100000","gasUsed":"17000","tags":[{"key":"YWN0aW9u","value":"c2VuZA=="},{"key":"c2VuZGVy","value":"YmFzZWNvaW4xcH
YwODV6bDRzY2VqZ3MzbnM5eHVtazRlZzVqZHh3ajd5ZHEzNGU="},{"key":"cmVjZWl2ZXI=","value":"YmFzZWNvaW4xc2dwMDNhN2wwamN1emV1ZTJ0enJ1Y2RmZThmMGs5dnRnMjY5bWU="}]},"hash":"B904250C76C38F573A601C0E3E0C31485732808B936DE27098F09E3BDD3A4F59","height":"13"}
```
7. 查询账户信息
```
$ basecli query account Jia
```

```
{"type":"basecoin/AppAccount","value":{"base_account":{"account_address":"basecoin1pv085zl4scejgs3ns9xumk4eg5jdxwj7ydq34e","public_key":{"type":"tendermint/PubKeyEd25519","value":"Nj2Kib7eYTaK
0cxUFTkNn8ZbopG13z4Q5Et5/3ncEjE="},"nonce":"1"},"coins":[{"coin_name":"qstar","amount":"99999973"}]}}
```

```
$ basecli query account Liu
```

```
{"type":"basecoin/AppAccount","value":{"base_account":{"account_address":"basecoin1sgp03a7l0jcuzeue2tzrucdfe8f0k9vtg269me","public_key":null,"nonce":"0"},"coins":[{"coin_name":"qstar","amount"
:"10"}]}}

```

8. 查询交易

```
$ basecli tendermint tx B904250C76C38F573A601C0E3E0C31485732808B936DE27098F09E3BDD3A4F59
```

```
{"hash":"b904250c76c38f573a601c0e3e0c31485732808b936de27098f09e3bdd3a4f59","height":"13","tx":{"type":"qbase/txs/stdtx","value":{"itx":[{"type":"basecoin/SendTx","value":{"from":"basecoin1pv08
5zl4scejgs3ns9xumk4eg5jdxwj7ydq34e","to":"basecoin1sgp03a7l0jcuzeue2tzrucdfe8f0k9vtg269me","coin":{"coin_name":"qstar","amount":"10"}}}],"sigature":[{"pubkey":{"type":"tendermint/PubKeyEd25519
","value":"Nj2Kib7eYTaK0cxUFTkNn8ZbopG13z4Q5Et5/3ncEjE="},"signature":"WIl4t4vOJVwRFfaVqg/2r/q1Y7lRr/T3LtcZcTMYCS3wv+H9w0rABxFiztvuNkyntBaKHz3R4NSZ2QkdXVI8DA==","nonce":"1"}],"chainid":"baseco
in","maxgas":"100000"}},"result":{"gas_wanted":"100000","gas_used":"17000","tags":[{"key":"YWN0aW9u","value":"c2VuZA=="},{"key":"c2VuZGVy","value":"YmFzZWNvaW4xcHYwODV6bDRzY2VqZ3MzbnM5eHVtazRl
ZzVqZHh3ajd5ZHEzNGU="},{"key":"cmVjZWl2ZXI=","value":"YmFzZWNvaW4xc2dwMDNhN2wwamN1emV1ZTJ0enJ1Y2RmZThmMGs5dnRnMjY5bWU="}]}}

```

9. QCP交易</br>
qstar PriKey:</br>
0xa3288910405746e29aeec7d5ed56fac138b215e651e3244e6d995f25cc8a74c40dd1ef8d2e8ac876faaa4fb281f17fb9bebb08bc14e016c3a88c6836602ca97595ae32300b

*ed25519格式:* V0bimu7H1e1W+sE4shXmUeMkTm2ZXyXMinTEDdHvjS6KyHb6qk+ygfF/ub67CLwU4BbDqIxoNmAsqXWVrjIwCw==

使用ed25519格式将签名私钥导入:

```
$ basecli keys import qcpsigner
```

```
> Enter ed25519 private key:
V0bimu7H1e1W+sE4shXmUeMkTm2ZXyXMinTEDdHvjS6KyHb6qk+ygfF/ub67CLwU4BbDqIxoNmAsqXWVrjIwCw==
> Enter a passphrase for your key:
> Repeat the passphrase:
```

```
$ basecli keys list
```

```
NAME:	TYPE:	ADDRESS:						PUBKEY:
Jia     local   basecoin1pv085zl4scejgs3ns9xumk4eg5jdxwj7ydq34e basecoinpub1zcjduepqxc7c4zd7mesndzk3e32p2wgdnlr9hg53kh0nuy8yfdul77wuzgcsqhgez8
Liu     local   basecoin1sgp03a7l0jcuzeue2tzrucdfe8f0k9vtg269me basecoinpub1zcjduepqmcegq0pzuw6uaw7v3swpaxlt0f58hpht2d65tvax4quk5r73amvsme39q8
qcpsigner       import  basecoin103eak408d4yp944wv58epp3neyah8z5d899xvx basecoinpub1zcjduepq3ty8d742f7egrutlhxltkz9uznspdsag335rvcpv496ett3jxq9stcdtnh

```


发送qcp跨链交易:

```
$ basecli tx send --from=Jia --to=Liu --coin-name=qstar --coin-amount=10 --qcp --qcp-from=qstar --qcp-signer=qcpsigner --qcp-blockheight=10
```

```
> step 1. build and sign TxStd
Password to sign with 'Jia':
> step 2. build and sign TxQcp
Password to sign with 'qcpsigner':
{"check_tx":{"gasWanted":"100000","gasUsed":"4141"},"deliver_tx":{"gasWanted":"100000","gasUsed":"21000","tags":[{"key":"YWN0aW9u","value":"c2VuZA=="},{"key":"c2VuZGVy","value":"cW9zYWNjMWowdjJzODl5NHpkamE3aDB5eWNjaHNlbmtmd3hhcjdkNmtxbTZk"},{"key":"cmVjZWl2ZXI=","value":"cW9zYWNjMTRhbWg1d213bDA2cHB2bXA3dHhkd2QzMGtsM3dndnI3OHloZzBr"},{"key":"cWNwLmZyb20=","value":"YmFzZWNvaW4="},{"key":"cWNwLnRv","value":"cXN0YXI="},{"key":"cWNwLnNlcXVlbmNl","value":"MQ=="},{"key":"cWNwLmhhc2g=","value":"y4aaWFIhIK6dhZOKUnnBovDaH0ZC5kcz0WIzLBs7l9k="}]},"hash":"4D0D07080DA9D1E78023DC6F47228A8106DAC5EA86D1938096259CC7C668F0B2","height":"44"}


```

10. QCP sequence 查询

```
$ basecli query qcp list
```

```
|Chain |Type |MaxSequence |
|----- |---- |----------- |
|qstar |in   |1           |
|qstar |out  |1           |

```

```
$ basecli query qcp in qstar
```

```
1
```

```
$ basecli query qcp out qstar
```

```
1
```
11. QCP 交易结果查询

```
$ basecli query qcp tx qstar  --seq 1
```

```
{"type":"qbase/txs/qcptx","value":{"txstd":{"itx":[{"type":"qbase/txs/qcpresult","value":{"result":{"Code":0,"Codespace":"","Data":null,"Log":"","GasWanted":"100000","GasUsed":"21000","FeeAmount":"0","FeeDenom":"","Tags":[{"key":"YWN0aW9u","value":"c2VuZA=="},{"key":"c2VuZGVy","value":"cW9zYWNjMWowdjJzODl5NHpkamE3aDB5eWNjaHNlbmtmd3hhcjdkNmtxbTZk"},{"key":"cmVjZWl2ZXI=","value":"cW9zYWNjMTRhbWg1d213bDA2cHB2bXA3dHhkd2QzMGtsM3dndnI3OHloZzBr"},{"key":"cWNwLmZyb20=","value":"YmFzZWNvaW4="},{"key":"cWNwLnRv","value":"cXN0YXI="}]},"qcporiginalsequence":"1","qcpextends":"","info":""}}],"sigature":null,"chainid":"qstar","maxgas":"0"},"from":"basecoin","to":"qstar","sequence":"1","sig":{"pubkey":null,"signature":null,"nonce":"0"},"blockheight":"44","txindex":"0","isresult":true,"extends":""}}

```

更多命令，查阅
```
$ basecli --help
```
