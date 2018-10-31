## Quick Start

### Basic

#### Tx

定义具体的事务内容，可以包含其他的信息。 自定义Tx需要实现`txs.ITx`接口:

> `qbase`使用[go-amino](https://github.com/tendermint/go-amino)进行编码解码，自定义Tx中的成员必须为可导出的（即成员首字母需为大写）。

```go

type ITx interface {
  //用于校验Tx数据是否合法。
  //可以在此方法中校验Tx是否包含完整的数据，用户账户是否有足够的金额进行转账等操作。
	ValidateData(ctx context.Context) bool

  //执行Tx事务
  //result: 事务执行结果
  //crossTxQcp: 若需要跨链处理，则返回跨链TxQcp。 crossTxQcp中仅需包含`To`及 `TxStd`
	Exec(ctx context.Context) (result types.Result, crossTxQcp *TxQcp)

  //获取签名者地址
	GetSigner() []types.Address
	CalcGas() types.BigInt
  GetGasPayer() types.Address

  // 获取Tx中待签名数据
	GetSignData() []byte
}

```


#### TxStd

区块链接收的标准Tx,包含以下成员:

```go

type TxStd struct {
  //ITx接口，将被具体Tx结构实例化
	ITx       ITx          `json:"itx"`

  //签名数组: 签名顺序要与ITx.GetSigner()保持一致
	Signature []Signature  `json:"sigature"`

  //执行TxStd的ChainID
	ChainID   string       `json:"chainid"`
  //Gas消耗的最大值
	MaxGas    types.BigInt `json:"maxgas"`
}

```


#### TxQcp

用于进行跨链的`TxStd`的封装:

```go

type TxQcp struct {
	TxStd       *TxStd    `json:"txstd"`       //包含的TxStd
	From        string    `json:"from"`        //跨链来源chainId
	To          string    `json:"to"`          //跨链处理的目标chainId
	Sequence    int64     `json:"sequence"`    //发送Sequence
	Sig         Signature `json:"sig"`         //签名
	BlockHeight int64     `json:"blockheight"` //Tx所在block高度
	TxIndx      int64     `json:"txindx"`      //Tx在block的位置
	IsResult    bool      `json:"isresult"`    //是否为Result
	Extends      string    `json:"extends"`      //扩展字段
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
	GetKVStoreName() string
	GetStoreKey() store.StoreKey

	SetStore(store store.KVStore)
	SetCodec(cdc *go_amino.Codec)
}

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

#### BaseApp

  `abci.Application`接口的实现。并提供以下注册方法:

* RegisterAccountProto(proto func() account.Account)： 注册自定义账户。调用此方法时，将注册account.AccountMapper

* RegisterMapper(mapper mapper.IMapper)： 注册自定义Mapper

* RegisterCustomQueryHandler(handler CustomQueryHandler): 注册自定义查询handler

* RegisterTxQcpSigner(signer crypto.PrivKey): 注册对跨链TxQcp进行签名的私钥

* RegisterTxQcpResultHandler(txQcpResultHandler TxQcpResultHandler): 注册对跨链交易结果处理handler


#### ABCI

  BaseApp中提供以下方法对*ABCI*中`initChain`,`beginBlock`,`endBlock`阶段进行自定义处理:

  * SetInitChainer(initChainer InitChainHandler)

  * SetBeginBlocker(beginBlocker BeginBlockHandler)

  * SetEndBlocker(endBlocker EndBlockHandler)




### Example


#### kvstore app

1. 自定义Tx

```go

type KvstoreTx struct {
	Key   []byte
	Value []byte
	Bytes []byte
}

```

2. 创建Mapper封装Tx操作

```go

const KvKVStoreName = "kvmapper"

func NewKvMapper() *KvMapper {
	var txMapper = KvMapper{}
	txMapper.BaseMapper = mapper.NewBaseMapper(nil, KvKVStoreName)
	return &txMapper
}

func (mapper *KvMapper) Copy() mapper.IMapper {
	cpyMapper := &KvMapper{}
	cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
	return cpyMapper
}

var _ mapper.IMapper = (*KvMapper)(nil)

func (mapper *KvMapper) SaveKV(key string, value string) {
	mapper.Set([]byte(key), value)
}

func (mapper *KvMapper) GetKey(key string) (v string) {
	mapper.Get([]byte(key), &v)
	return
}

```


3. 自定义Tx实现txs.ITx接口

```go

func (kv *KvstoreTx) ValidateData(ctx context.Context) bool {
	if len(kv.Key) < 0 {
		return false
	}
	return true
}

func (kv *KvstoreTx) Exec(ctx context.Context) (result types.Result, crossTxQcps *txs.TxQcp) {
	//获取注册的mapper：
	kvMapper := ctx.Mapper(KvMapperName).(*KvMapper)
	kvMapper.SaveKV(key, string(kv.Value))
	return
}

func (kv *KvstoreTx) GetSigner() []types.Address {
	return nil
}

func (kv *KvstoreTx) CalcGas() types.BigInt {
	return types.ZeroInt()
}

func (kv *KvstoreTx) GetGasPayer() types.Address {
	return types.Address{}
}

func (kv *KvstoreTx) GetSignData() []byte {
	return nil
}

```


3. 创建BaseApp,注册cdc编码，注册Mapper

```go

db, err := dbm.NewGoLevelDB("kvstore", filepath.Join("", "data"))

var baseapp = baseabci.NewBaseApp("kvstore", logger, db, func(cdc *go_amino.Codec){
  //将自定义的Tx注册到cdc编码中
  cdc.RegisterConcrete(&KvstoreTx{}, "kvstore/main/kvstoretx", nil)
})


//注册自定义的Mapper
	var kvMapper = kvstore.NewKvMapper(store.NewKVStoreKey("kv"))
	baseapp.RegisterMapper(kvMapper)

//加载kvstore
baseapp.LoadLatestVersion()

```




4. 启动app

```go

	// Start the ABCI server
  srv, err := server.NewServer("0.0.0.0:26658", "socket", baseapp)

  srv.Start()

```



### Advance

#### 自定义账户

1. 自定义账户类型

```go

type AppAccount struct {
	account.BaseAccount `json:"base_account"`
	Coins               Coins `json:"coins"`
}

```

2. 通过baseapp注册账户生成Proto,用于创建新账户。

> 调用`baseapp.RegisterAccountProto`后,`qbase`会自动注册 `account.AccountMapper`

> `account.AccountMapper`封装了对账户的基础操作

> 可以通过 `context.Mapper(account.AccountMapperName).(*account.AccountMapper)`获取accountMapper实例

```go

//将AppAccount注册到amino.Codec中
baseapp.GetCdc().RegisterConcrete(&AppAccount{}, "YourAppName/AppAccount", nil)

//注册accountProto方法，用于创建新账户:
baseapp.RegisterAccountProto(func() account.Account {
  return &AppAccount{
		BaseAccount: account.BaseAccount{},
		Coins:       []Coin{},
	}
})

```

3. 在自定义Tx中的Exec方法中，可以通过以下方法来获取注册的account.AccountMapper:

```go

func (tx *YourTx) Exec(ctx context.Context) (result types.Result, crossTxQcps *txs.TxQcp) {
	//获取注册的account mapper：
  accountMapper := baseabci.GetAccountMapper(ctx)

  //通过账户地址获取account
  account := accountMapper.GetAccount(accAddress).(*AppAccount)

  ...
  return
}

```

#### 实现QCP交易

联盟链--公链 QCP交易模型流程如下:

1. 联盟链中执行自定义Tx.Exec(...)方法返回需要跨链交易的crossTxQcp, crossTxQcp不为空，则表示需要跨链处理。
2. 使用在baseapp.RegisterTxQcpSigner(signer crypto.PrivKey)中注册的签名者对crossTxQcp进行签名并保存
3. *中继* 获取跨链`crossTxQcp`后，将其发送到对应的公链中
4. 公链执行`crossTxQcp`,并将执行结果保存为`crossTxQcp'`(crossTxQcp'.IsResult==true)
5. *中继* 从公链获取跨链`crossTxQcp'`，对`crossTxQcp'` 进行签名，将其发送到对应的联盟链中
6. 联盟链获取公链的执行结果`crossTxQcp'`后，根据返回的结果调用`TxQcpResultHandler`进行相应的业务处理

> `qbase`中提供了`qcpMapper`用于对跨链TxQcp的封装。通过`context.Mapper(qcp.QcpMapperName).(*qcp.QcpMapper)` 获取qcpMapper实例


##### 联盟链

###### 跨链处理操作步骤

1. 注册联盟链签名者

```go

//获取联盟链签名者私钥
// 该方法需要联盟链实现，根据私钥保存的地址，可以从文件，数据库等地方获取
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

###### 生成跨链TxQcp

1. 自定义ITx.Exec中返回crossTxQcp

```go

func (tx *YourTx) Exec(ctx context.Context) (result types.Result, crossTxQcp *txs.TxQcp) {
  ...

  //根据业务组装crossTxQcp

  ...
  return
}

```


##### 公链

###### 跨链处理操作步骤

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


#### 初始化创世块

1. 自定义`genesis.json`



2. 创世块中保存配置数据




#### 启动参数配置





#### 查询接口

`qbase`提供以下方式用于查询数据:

* `/store/{KVStoreKey}/key`

通过key查询index为{KVStoreKey}的数据。




* `/custom/{CustomPath1}/{CustomPath2}/...`

自定义查询路径。需要实现`baseabci.CustomQueryHandler`方法:

```go

//
var customQueryHandler = func(ctx ctx.Context, route []string, req abci.RequestQuery) (res []byte, err types.Error) {
	//检查route长度
	// if route[0] == {CustomPath1} && route[1] == {CustomPath2} ... {
	// do something...
	//}
}


//注册自定义查询Handler
baseapp.RegisterCustomQueryHandler(customQueryHandler)

```




















