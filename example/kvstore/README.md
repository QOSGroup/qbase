
## App开发步骤


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


## 启动

`kvstore`示例仅实现`abci app`,没有集成`server`包。参考[basecoin](../basecoin/README.md)集成`server`

1. 启动tendermint

```sh

tendermint init
tendermint node
```

2. 启动kvstore服务端

```sh
cd example/kvstore/cmd/kvstored
go build
./kvstored
```

## 客户端执行查询及发送命令

```sh
cd example/kvstore/cmd/kvstorecli
go build
```

1. 查询key值:

```sh
./kvstorecli -k abc

```

2. 设置key值:

```sh

./kvstorecli -m set -k abc -v 111

```


