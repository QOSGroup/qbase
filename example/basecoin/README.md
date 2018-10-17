# BaseCoin Example

basecoin example基于qbase实现了简单的单次单币种，单个发送/接收账户的转账功能

## 使用步骤

1. 编译basecoind
```
cd cmd/basecoind
go build
```
2. 初始化
```
./basecoind init --chain-id=basecoin-chain
```
3. 启动basecoin app
```
./basecoind start --with-tendermint=true
```
4. 编译basecli
```
cd cmd/basecli
go build
```
5. 链内交易
```
./basecli -m=stdtransfer -from=address1k0m8ucnqug974maa6g36zw7g2wvfd4sug6uxay -to=address1srrhd4quypqn0vu5sgrmutgudtnmgm2t2juwya -coin=qstar,10 -prikey=0xa328891040ae9b773bcd30005235f99a8d62df03a89e4f690f9fa03abb1bf22715fc9ca05613f2d8061492e9f8149510b5b67d340d199ff24f34c85dbbbd7e0df780e9a6cc -nonce=0
```
6. 账户查询状态
```
./basecli -m=accquery -addr=address1k0m8ucnqug974maa6g36zw7g2wvfd4sug6uxay
```
7. QCP交易
```
./basecli -m=qcptransfer -from=address1k0m8ucnqug974maa6g36zw7g2wvfd4sug6uxay -to=address1srrhd4quypqn0vu5sgrmutgudtnmgm2t2juwya -coin=qstar,10 -prikey=0xa328891040ae9b773bcd30005235f99a8d62df03a89e4f690f9fa03abb1bf22715fc9ca05613f2d8061492e9f8149510b5b67d340d199ff24f34c85dbbbd7e0df780e9a6cc -nonce=1 -chainid=qstar -qcpprikey=0xa3288910405746e29aeec7d5ed56fac138b215e651e3244e6d995f25cc8a74c40dd1ef8d2e8ac876faaa4fb281f17fb9bebb08bc14e016c3a88c6836602ca97595ae32300b -qcpseq=1
```
8. QCP sequence 查询
```
./basecli -m=qcpseq -chainid=qstar
```
9. QCP 查询
```
./basecli -m=qcpquery -chainid=qstar -qcpseq=1
```
10. QCP TxResult
```
./basecli -m=qcptxresult -chainid=qstar -qcpprikey=0xa3288910405746e29aeec7d5ed56fac138b215e651e3244e6d995f25cc8a74c40dd1ef8d2e8ac876faaa4fb281f17fb9bebb08bc14e016c3a88c6836602ca97595ae32300b -originseq=1 -qcpseq=2
```

参数说明：<br/>
-m            账户查询：accquery，QCP sequence 查询：qcpseq，QCP查询：qcpquery，链内交易：stdtransfer，QCP交易：qcptransfer，QCP TxResult：qcptxresult<br/>
-from         发送地址，bech32格式<br/>
-to           接收地址，bech32格式<br/>
-coin         币种,币值 半角逗号分隔<br/>
-addr         账户地址，bebech32格式<br/>
-prikey       发送账号私钥 hex<br/>
-chainid      QCP chainId<br/>
-qcpprikey    QCP私钥 hex<br/>
-originseq    此结果对应的TxQcp.Sequence<br/>
-qcpseq       QCP发送序号<br/>
