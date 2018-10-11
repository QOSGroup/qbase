# 账户定义
版本:
v0.1

日期:
2018年09月26日

## 简介：

账户提供基本的账户信息及相关方法
### 账户接口
```
type Account interface {
	GetAddress() types.Address
	SetAddress(addr types.Address) error
	GetPubicKey() crypto.PubKey
	SetPublicKey(pubKey crypto.PubKey) error
	GetNonce() uint64
	SetNonce(nonce uint64) error
}
```
## 数据结构
### BaseAccount
```
type BaseAccount struct{
	AccountAddress common.Address `json:"account_address"` // account address
	Publickey      crypto.PubKey  `json:"public_key"`		// public key
	Nonce          uint64         `json:"nonce"`			// identifies tx_status of an account
}
```
## 方法规定
- 获得结构原型
- getter/setter
- 注册序列化的方法
