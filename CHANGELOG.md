Changelog

v0.0.2
2018.10.23

本版本重点完成签名和qcp处理逻辑

BREAKING CHANGES:

* [tx] ITx ValidateData方法增加context参数，支持验证时查询状态。
* [tx] TxQcp Payload修改为TxStd,并使用指针;TxIndx修改为TxIndex。

FEATURES:
    
* [query] #11 增加customer query
* [baseabci] 增加qcp result回调方法

IMPROVEMENTS:   

* [baseabci] 完成签名验证逻辑
* [qcp] 完成qcp处理逻辑
* [basecoin] 增加签名验证和qcp处理案例
* [tx] 支持通用签名方法
* 修改一些参数和变量命名

BUG FIXES:
* [sign] #27

