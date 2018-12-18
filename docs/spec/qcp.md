## QCP跨链

### 业务流程

![qcp-arth](https://github.com/QOSGroup/static/blob/master/qcp_relay_qos.jpg?raw=true)

### 数据存储

> qcp包中封装了对qcp store的操作

qcp store 保存qcp交易sequence,tx列表

#### qcp

```
sequence/out/[chainId] //需要输出到"chainId"的qcp tx最大序号
tx/out/[chainId]/[sequence] //需要输出到"chainId"的每个qcp tx
sequence/in/[chainId] //已经接受到来自"chainId"的qcp tx最大序号
pubkey/in/[chainId] //接受来自"chainId"的合法公钥

```

### 交易模型示例

联盟链--公链 QCP交易模型流程如下:

1. 联盟链中执行自定义Tx.Exec(...)方法返回需要跨链交易的crossTxQcp, crossTxQcp不为空，则表示需要跨链处理。
2. 使用在baseapp.RegisterTxQcpSigner(signer crypto.PrivKey)中注册的签名者对crossTxQcp进行签名并保存
3. *中继* 获取跨链`crossTxQcp`后，将其发送到对应的公链中
4. 公链执行`crossTxQcp`,并将执行结果保存为`crossTxQcp'`(crossTxQcp'.IsResult==true)
5. *中继* 从公链获取跨链`crossTxQcp'`，对`crossTxQcp'` 进行签名，将其发送到对应的联盟链中
6. 联盟链获取公链的执行结果`crossTxQcp'`后，根据返回的结果调用`TxQcpResultHandler`进行相应的业务处理

> `qbase`中提供了`qcpMapper`用于对跨链TxQcp的封装。通过`context.Mapper(qcp.QcpMapperName).(*qcp.QcpMapper)` 获取qcpMapper实例


*前提*:

* 联盟链接受来自*中继*的数据，需要设置*中继*的公钥为可信公钥

* 联盟链需要对跨链Tx进行签名，需要在app中注册私钥进行

* 公链接受来自联盟链的数据，需要设置联盟链的公钥为可信公钥


#### 联盟链

##### 跨链处理操作步骤

1. 注册联盟链签名者

```go

//获取联盟链签名者私钥
//该方法需要联盟链实现，根据私钥保存的地址，可以从文件，数据库等地方获取
//***请妥善保管私钥，不能泄漏****
priKey := getSignerPriKey()

//注册签名者
baseapp.RegisterTxQcpSigner(priKey)

```

2. baseapp.InitChain中设置*中继*的公钥为可信公钥

```go

baseapp.SetInitChainer(func(ctx context.Context, req abci.RequestInitChain){
  ...

  qcpMapper := GetQcpMapper(ctx)

  // chainId , pubKey可以从配置文件或其他保存的地方获取
  // chainId := xxx 中继从chainId获取的跨链
  // pubKey := xxx  中继公钥
  qcpMapper.SetChainInTrustPubKey(chainId, pubKey)

  ...
})


```

3. 注册TxQcpResultHandler

实现TxQcpResultHandler方法，并调用baseapp.RegisterTxQcpResultHandler(txQcpResultHandler TxQcpResultHandler)方法进行注册


```go

txQcpResultHandler := func(ctx ctx.Context, txQcpResult interface{}) types.Result {

  qcpResult, ok := CovertTxQcpResult(txQcpResult)
  if !ok {
  return types.ErrInternal("wrong type").Result()
  }

  //根据结果进行业务处理

}


```

##### 生成跨链TxQcp

1. 自定义ITx.Exec中返回crossTxQcp

```go

func (tx *YourTx) Exec(ctx context.Context) (result types.Result, crossTxQcp *txs.TxQcp) {
  ...

  //根据业务组装crossTxQcp

  ...
  return
}

```


#### 公链

##### 跨链处理操作步骤

1. baseapp.InitChain中设置联盟链的公钥为可信公钥

```go

baseapp.SetInitChainer(func(ctx context.Context, req abci.RequestInitChain){
  ...

  qcpMapper := GetQcpMapper(ctx)

  // chainId , pubKey可以从配置文件或其他保存的地方获取
  // chainId := xxx： 联盟链chainId
  // pubKey := xxx : 联盟链公钥
  qcpMapper.SetChainInTrustPubKey(chainId, pubKey)

  ...
})


```
