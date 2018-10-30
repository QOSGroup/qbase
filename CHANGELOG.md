# Changelog

## v0.0.4

2018.10.29

修改一些bug，优化BaseMapper用法，统一account中nonce,qcp中sequence规则

### *BREAKING CHANGES:*

* [tx] nonce,sequence统一为int64，规则统一代表已接收tx序号，提交新交易需要加1填充
* [qcp] #67 deliverTx中check sequence错误不反馈QcpTxResult，且不增加qcp in sequence，等待中继补偿重发

### *FEATURES:*
    

### *IMPROVEMENTS:*
 
* [types] #57 BigInt增加是否为空方法
* [types] #60 BaseCoin默认值为0，增加BaseCoins类型
* [mapper] #63 #64 BaseMapper重构，mapper name和kvstore key合并
* [tx] #66 #71 account中nonce,qcp中sequence统一为int64，规则统一代表已接收tx序号
 
### *BUG FIXES:*
* [qcp] #67 deliverTxQcp中check sequence错误不反馈QcpTxResult且不增加qcp in sequence，等待中继补偿重发

## v0.0.3

2018.10.24

修改qcp event相关tag名称，修改qcp存储key命名规则

### *BREAKING CHANGES:*

* [tx] 修改qcp event相关tag名称。
* [tx] 修改qcp存储key命名规则。

### *FEATURES:*
    

### *IMPROVEMENTS:*


### *BUG FIXES:*


## v0.0.2

2018.10.23

本版本重点完成签名和qcp处理逻辑

### *BREAKING CHANGES:*

* [tx] ITx ValidateData方法增加context参数，支持验证时查询状态。
* [tx] TxQcp Payload修改为TxStd,并使用指针;TxIndx修改为TxIndex。

### *FEATURES:*
    
* [query] #11 增加customer query
* [baseabci] 增加qcp result回调方法

### *IMPROVEMENTS:*

* [baseabci] 完成签名验证逻辑
* [qcp] 完成qcp处理逻辑
* [basecoin] 增加签名验证和qcp处理案例
* [tx] 支持通用签名方法
* 修改一些参数和变量命名

### *BUG FIXES:*
* [sign] #27

