# Changelog

## v0.0.10
2019.01.08

**BREAKING CHANGES**
* [baseapp] tendermint依赖版本升级到v0.27.3
* [server]  server init command修改

**FEATURES**
* [client] client命令行增加从配置文件读取参数
* [client] client tx签名工具类优化

## v0.0.9
2018.12.24

**IMPROVEMENTS**
* [client] 客户端部分命令修改

## v0.0.8
2018.12.21

**BREAKING CHANGES**
* [baseapp] 修改TxQcpResultHandler调用逻辑

**IMPROVEMENTS**
* [tx] TxStd屏蔽重复签名
* [consuses] 增加共识参数查询
* [baseapp] 删除多余依赖库
* [account] 修改Address类型


## v0.0.7
2018.11.29

**BREAKING CHANGES**
* [tx] TxStd签名时增加源ChainID字段

**FEATURES**
* [client] 增加client command相关命令
* [client] client command文档完善

**IMPROVEMENTS**
* [basecoin] basecoin client cmd 示例基于client command重构

**BUG FIXES**
* [client] keybase中类型为import的key无法删除
* [tx] BaseCoins排序问题

## v0.0.6
2018.11.5

**BREAKING CHANGES**
* [qcp] QcpTxResult返回字段直接包含为DeliverResult

**FEATURES**

**IMPROVEMENTS**
* [qcp] #82 TxQcp中增加Extends，QcpTxResult中增加QcpOriginalExtends，方便联盟链关联TxQcpQcpTxResult
* [qcp] DeliverResult.Tags中qcp.sequence保存为字符串数字，方便订阅比较
* [tx] ITx.validateData返回类型修改为error，方便在Result中保存这些error，客户端查看

**BUG FIXES**
* [qcp] QcpTxResult.TXSTD.ChainID设置为qcp来源chain

## v0.0.5

2018.10.30

initChain增加qcp pubkey初始化功能

**BREAKING CHANGES**

**FEATURES**

**IMPROVEMENTS**

* [qcp] #78 initChain增加从genesis.json文件初始化qcp pubkey功能

**BUG FIXES**

## v0.0.4

2018.10.29

修改一些bug，优化BaseMapper用法，统一account中nonce,qcp中sequence规则

**BREAKING CHANGES**

* [tx] nonce,sequence统一为int64，规则统一代表已接收tx序号，提交新交易需要加1填充
* [qcp] #67 deliverTx中check sequence错误不反馈QcpTxResult，且不增加qcp in sequence，等待中继补偿重发

**FEATURES**


**IMPROVEMENTS**

* [types] #57 BigInt增加是否为空方法
* [types] #60 BaseCoin默认值为0，增加BaseCoins类型
* [mapper] #63 #64 BaseMapper重构，mapper name和kvstore key合并
* [tx] #66 #71 account中nonce,qcp中sequence统一为int64，规则统一代表已接收tx序号

**BUG FIXES**
* [qcp] #67 deliverTxQcp中check sequence错误不反馈QcpTxResult且不增加qcp in sequence，等待中继补偿重发

## v0.0.3

2018.10.24

修改qcp event相关tag名称，修改qcp存储key命名规则

**BREAKING CHANGES**

* [tx] 修改qcp event相关tag名称。
* [tx] 修改qcp存储key命名规则。

**FEATURES**


**IMPROVEMENTS**


**BUG FIXES**


## v0.0.2

2018.10.23

本版本重点完成签名和qcp处理逻辑

**BREAKING CHANGES**

* [tx] ITx ValidateData方法增加context参数，支持验证时查询状态。
* [tx] TxQcp Payload修改为TxStd,并使用指针;TxIndx修改为TxIndex。

**FEATURES**

* [query] #11 增加customer query
* [baseabci] 增加qcp result回调方法

**IMPROVEMENTS**

* [baseabci] 完成签名验证逻辑
* [qcp] 完成qcp处理逻辑
* [basecoin] 增加签名验证和qcp处理案例
* [tx] 支持通用签名方法
* 修改一些参数和变量命名

**BUG FIXES**
* [sign] #27

