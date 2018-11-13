# QBASE

## 概览

### 简介

`qbase`是QOS通用区块链应用框架，基于此框架开发QOS公链和联盟区块链应用,提供了通用的存储、交易和QCP跨链协议。

### 代码结构

`qbase`代码结构:

* `account`: 对账户的抽象及基础操作的封装
* `baseabci`: 基于`tendermint`实现的基础ABCI APP模板
* `client`: 与`qbase`交互的客户端命令集
* `context`: 上下文工具类
* `example`: 提供`basecoin` 和 `kvstore` 示例
* `keys`: 助记符生成及恢复
* `mapper`: `store`操作的基础封装
* `qcp`: 跨链协议操作封装
* `server`: 启动服务封装
* `store`: 基于`Merkle`树的数据库通用存储
* `txs`: transaction封装
* `types`: 工具类及通用结构

### Application

`qbase`项目包含以下几类Application:

1. ABCI App

`Tendermint`定义的接口规范

> The basic ABCI interface allowing Tendermint to drive the applications state machine with transaction blocks.


2. baseabci App

`ABCI App`接口实现的基础模板，实现了账户签名校验、账户`nonce`校验及跨链`qcp`业务等通用操作，并提供了扩展接口用于自定义业务扩展。

3. basecoin App

基于`baseabci app` 实现的示例，提供了转账、跨链转账功能。

4. your custom App

基于`baseabci app`开发的项目APP，可以在App中实现以下功能:
* 增强`account`功能
* 自定义`mapper`存储数据
* 自定义`tx`业务逻辑
* ...


## 基础组件

### Tx

定义具体的业务内容，可以包含其他的信息。 自定义Tx需要实现`txs.ITx`接口:

> `qbase`使用[go-amino](https://github.com/tendermint/go-amino)进行编码解码，自定义Tx中的成员必须为可导出的（即成员首字母需为大写）。

```go

type ITx interface {
	ValidateData(ctx context.Context) error // 校验业务数据
	Exec(ctx context.Context) (result types.Result, crossTxQcp *TxQcp)
	GetSigner() []types.Address //签名者
	CalcGas() types.BigInt      //计算gas
	GetGasPayer() types.Address //gas付费人
	GetSignData() []byte        //获取签名字段
}

```


### TxStd

`qbase`SDK处理的标准Tx,包含以下字段:

```go

type TxStd struct {
	ITx       ITx          `json:"itx"`      //ITx接口，将被具体Tx结构实例化
	Signature []Signature  `json:"sigature"` //签名数组
	ChainID   string       `json:"chainid"`  //ChainID: 执行ITx.exec方法的链ID
	MaxGas    types.BigInt `json:"maxgas"`   //Gas消耗的最大值
}

```

### TxQcp

基于跨链QCP协议的跨链业务数据封装:

```go

type TxQcp struct {
	TxStd       *TxStd    `json:"txstd"`       //TxStd结构
	From        string    `json:"from"`        //qscName
	To          string    `json:"to"`          //qosName
	Sequence    int64     `json:"sequence"`    //发送Sequence
	Sig         Signature `json:"sig"`         //签名
	BlockHeight int64     `json:"blockheight"` //Tx所在block高度
	TxIndex     int64     `json:"txindex"`     //Tx在block的位置
	IsResult    bool      `json:"isresult"`    //是否为Result
	Extends     string    `json:"extends"`     //扩展字段
}

```

#### KVStore

数据持久层，对底层数据的通用访问接口。具体方法参见`store.store.go`文件

#### Mapper

  基于`KVStore`的业务封装，使用[amino](https://github.com/tendermint/go-amino)对数据进行编码保存。

  自定义Mapper需要实现`mapper.IMapper`接口:

> 自定义Mapper通过调用baseapp.RegisterMapper方法后使用

 ```go

type IMapper interface {
	Copy() IMapper

	//BaseMapper implement below methods
	MapperName() string
	GetStoreKey() store.StoreKey

	SetStore(store store.KVStore)
	SetCodec(cdc *go_amino.Codec)
}

 ```

`qbase`提供`BaseMapper`用于对`IMapper`的快速实现,自定义Mapper仅需实现`Copy()`方法即可:

```go

type YourMapper struct {
  *mapper.BaseMapper
  //other fields
}

var _ mapper.IMapper = (*YourMapper)(nil)

func NewYourMapper(mapperName string) *YourMapper {
	var yourMapper = YourMapper{}
	yourMapper.BaseMapper = mapper.NewBaseMapper(nil, mapperName)
	return &yourMapper
}


func (mapper *YourMapper) Copy() mapper.IMapper {
	cpyMapper := &YourMapper{}
  cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
  //other fields copy
	return cpyMapper
}

```


### Account

`account`包中提供了对账户体系的基础封装，包含账户的地址、`nonce`、`pubkey`等。

可以通过以下方式对账户进行扩展:

> 扩展的账户类型需要通过 app.RegisterAccountProto进行注册

> 调用`baseapp.RegisterAccountProto`后,`qbase`会自动注册 `account.AccountMapper`。 `account.AccountMapper`提供了对账户的查询，保存操作。

>

```go
type YourAccount struct {
	account.BaseAccount `json:"base_account"`
	Coins               types.BaseCoins `json:"coins"`
  //other fields
}

```


## QCP跨链协议

参见: [QCP](../docs/qcp.md)

## 扩展

### BaseApp

  `abci.Application`接口的实现。并提供以下注册方法:

* RegisterAccountProto(proto func() account.Account)： 注册自定义账户。调用此方法时，将注册account.AccountMapper

* RegisterMapper(mapper mapper.IMapper)： 注册自定义Mapper

* RegisterCustomQueryHandler(handler CustomQueryHandler): 注册自定义查询handler

* RegisterTxQcpSigner(signer crypto.PrivKey): 注册对跨链TxQcp进行签名的私钥

* RegisterTxQcpResultHandler(txQcpResultHandler TxQcpResultHandler): 注册对跨链交易结果处理handler


### ABCI

  BaseApp中提供以下方法对*ABCI*中`initChain`,`beginBlock`,`endBlock`阶段进行自定义处理:

  * SetInitChainer(initChainer InitChainHandler)

  * SetBeginBlocker(beginBlocker BeginBlockHandler)

  * SetEndBlocker(endBlocker EndBlockHandler)


### 配置文件


### 查询接口

`qbase`提供以下方式用于查询数据:

* `/store/{KVStoreKey}/key`: 通过key查询index为{KVStoreKey}的数据。

* `/custom/{CustomPath1}/{CustomPath2}/...` : 自定义查询路径。需要实现`baseabci.CustomQueryHandler`方法:

```go

var customQueryHandler = func(ctx ctx.Context, route []string, req abci.RequestQuery) (res []byte, err types.Error) {
	//检查route长度
	// if route[0] == {CustomPath1} && route[1] == {CustomPath2} ... {
	// do something...
	//}
}


//注册自定义查询Handler
baseapp.RegisterCustomQueryHandler(customQueryHandler)

```

### amino codec推荐用法




## 示例

* [basecoin](../example/basecoin/README.md)

* [kvstore](../example/kvstore/README.md)

