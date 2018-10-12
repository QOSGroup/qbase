# qbase

[![version](https://img.shields.io/github/tag/QOSGroup/qbase.svg)](https://github.com/QOSGroup/qbase/releases/latest)
[![Build Status](https://travis-ci.org/QOSGroup/qbase.svg?branch=master)](https://travis-ci.org/QOSGroup/qbase)
[![codecov](https://codecov.io/gh/QOSGroup/qbase/branch/master/graph/badge.svg)](https://codecov.io/gh/QOSGroup/qbase)
[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/QOSGroup/qbase)
[![Go version](https://img.shields.io/badge/go-1.11.0-blue.svg)](https://github.com/moovweb/gvm)
[![license](https://img.shields.io/github/license/QOSGroup/qbase.svg)](https://github.com/QOSGroup/qbase/blob/master/LICENSE)
[![](https://tokei.rs/b1/github/QOSGroup/qbase?category=lines)](https://github.com/QOSGroup/qbase)

qbase是QOS通用区块链应用框架，基于此框架开发QOS公链和联盟区块链应用,提供了通用的存储、交易和QCP跨链协议。

感谢tendermint团队，此框架基于[tendermint](https://github.com/tendermint/tendermint)研发，实现了ABCI应用的通用封装。

感谢cosmos团队，本框架实现参考了[cosmos-sdk](https://github.com/cosmos/cosmos-sdk)代码。

当前非正式版本，我们会持续完善。

## Examples
* [kvstore](https://github.com/QOSGroup/qbase/blob/master/example/kvstore)
* [basecoin](https://github.com/QOSGroup/qbase/tree/master/example/basecoin)

## Quick Start
* 创建一个新abci
* 交易实现
* qcp交易

## 已实现区块链应用
* [qos](https://github.com/QOSGroup/qos)，QOS公链，支持双层代币和跨链交易
* [qstars](https://github.com/QOSGroup/qstars)，星云联盟链
