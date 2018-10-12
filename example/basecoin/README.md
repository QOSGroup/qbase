# BaseCoin Example

basecoin example基于qbase实现了简单的单次单币种，单个发送/接收账户的转账功能

## 使用步骤

1. 编译basecoind程序
```
cd cmd/basecoind
go build
```
2. 初始化
```
./basecoind init
```
3. 启动basecoin app
```
./basecoind start --with-tendermint=true
```
4. 编译basecoincli
```
cd cmd/basecoincli
go build
```
4. 发送交易
```
./basecoincli -m=transfer -from=address1k0m8ucnqug974maa6g36zw7g2wvfd4sug6uxay -to=address1srrhd4quypqn0vu5sgrmutgudtnmgm2t2juwya -coin=qstar,18
```
5. 查询状态
```
./basecoincli -m=query -addr=address1srrhd4quypqn0vu5sgrmutgudtnmgm2t2juwya
```

  参数说明：
-m 转账：transfer，查询：query
-from 发送地址，bech32格式
-to 接收地址，bech32格式
-coin 币种,币值 半角逗号分隔
-addr 账户地址，bebech32格式