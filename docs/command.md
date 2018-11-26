## 客户端命令

`qbase`中内置了以下类型的的子命令集:

* keys: 密钥管理命令集
* tx: 发送具体transaction命令集
* query: 查询业务数据相关命令集
* qcp: 查询qcp相关数据命令集
* tendermint: 查询区块信息相关命令集

示例及用法参考[basecli](https://github.com/QOSGroup/qbase/tree/master/example/basecoin/cmd/basecli)

### Keys

Keys管理工具提供以下功能:
1. 账户创建,生成助记符
2. 从助记符恢复账户
3. 私钥导入导出
4. 账户密码修改及删除

Keys包含以下子命令:

* list:
* mnemonic:
* new
* add
* delete
* update
* export
* import

Available Commands:
  mnemonic    Compute the bip39 mnemonic
  new         Interactive command to derive a new private key, encrypt it, and save to disk
  add         Create a new key, or import from seed
  list        List all keys

  delete      Delete the given key
  update      Change the password used to protect private key

  export      export key for the given name
  import      Interactive command to import a new private key, encrypt it, and save to disk







### Tx


### Query


### Qcp



### Tendermint





