## 客户端命令

命令类型:

1. Get

通过[GetCommands](https://github.com/QOSGroup/qbase/tree/master/client/types/flags.go#GetCommands)方法进行封装的命令具有以下通用参数:


| 参数 | 默认值 | 说明 |
| :--- | :---: | :--- |
|--chain-id| "" | Chain ID of tendermint node |
|--trust-node| false | Trust connected full node (don't verify proofs for responses) |
|--node| tcp://localhost:26657 | tcp://\<host\>:\<port\> to tendermint rpc interface for this chain |
|--height| 0 | block height to query, omit to get most recent provable block |
|--indent| false | add indent to json response |



2. Post

通过[PostCommands](https://github.com/QOSGroup/qbase/tree/master/client/types/flags.go#PostCommands)方法进行封装的命令具有以下通用参数:


| 参数 | 默认值 | 说明 |
| :--- | :---: | :--- |
|--nonce | 0 | account nonce to sign the tx |
|--max-gas| 0 | gas limit to set per tx |
|--chain-id| "" | Chain ID of tendermint node |
|--node| tcp://localhost:26657 | tcp://\<host\>:\<port\> to tendermint rpc interface for this chain |
|--async| false | broadcast transactions asynchronously |
|--trust-node| false | Trust connected full node |
|--qcp| false | enable qcp mode. send qcp tx |
|--qcp-signer| "" | qcp mode flag. qcp tx signer key name |
|--qcp-seq| 0 | qcp mode flag.  qcp in sequence |
|--qcp-from| "" | qcp mode flag. qcp tx source |
|--qcp-blockheight| 0 | qcp mode flag. original tx blockheight |
|--qcp-txindex| 0 | qcp mode flag. original tx index |
|--qcp-extends| "" | qcp mode flag. qcp tx extends info |
|--indent| false | add indent to json response |
|--nonce-node| "" | tcp://\<host\>:\<port\> to tendermint rpc interface for some chain to query account nonce |


`qbase`中内置了以下类型的的子命令集:

* keys: 密钥管理命令集
* tx: 发送具体transaction命令集
* query: 查询业务数据相关命令集
* qcp: 查询qcp相关数据命令集
* tendermint: 查询区块信息相关命令集

示例及用法参考[basecli](https://github.com/QOSGroup/qbase/tree/master/example/basecoin/cmd/basecli)

### Keys

Keys管理工具提供以下功能:
1. Key创建,生成助记符
2. 从助记符恢复Key
3. Key导入导出
4. Key密码修改及删除


Keys包含以下子命令:

|命令|说明|
|:---| :--- |
|list|List all keys|
|mnemonic|Compute the bip39 mnemonic|
|new|Interactive command to derive a new private key, encrypt it, and save to disk|
|add|Create a new key, or import from seed|
|delete|Delete the given key|
|update|Change the password used to protect private key|
|export|export key for the given name|
|import|Interactive command to import a new private key, encrypt it, and save to disk|

补充说明:

1. `add`命令可以使用`--recover`参数从助记符中恢复*key*
2. `import`命令可以使用`--file`参数从*ca私钥*文件中导入*key*


### Tx

Tx命令集包含发送具体的业务命令

1. 获取Tx命令:
```go
txCommand := bcli.TxCommand()
```

2. 添加具体的业务命令
```go
txCommand.AddCommand(ctypes.PostCommands(client.Commands(cdc)...)...)
```


### Query

Query(alias `q`)中包含以下命令:

|命令|说明|
|:---| :--- |
|account| Query account info by address or name |
|store| Query store data by low level |
|qcp| qcp subcommands|
`store`命令可以直接查询`abci app`中`store`存储的数据:

* --path=/store/STORENAME/key: 查询key值等于`data`的数据
* --path=/store/STORENAME/subspace: 查询所有前缀为`data`的数据

#### Qcp

Qcp中包含以下命令:

|命令|说明|
|:---| :--- |
|list| List all crossQcp chain's sequence info |
|out| Get max sequence to outChain |
|in| Get max sequence received from inChain |
|tx| Query qcp out tx info |

### Tendermint

Tendermint(alias `t`)中包含以下命令:

|命令|说明|
|:---| :--- |
|status|Query remote node for status|
|validators|Get validator set at given height|
|block|Get block info at given height|
|txs|Search for all transactions that match the given tags|
|tx|Query match hash tx in all commit block|

