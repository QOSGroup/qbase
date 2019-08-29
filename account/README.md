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
	GetAddress() types.AccAddress
	SetAddress(addr types.AccAddress) error
	GetPublicKey() crypto.PubKey
	SetPublicKey(pubKey crypto.PubKey) error
	GetNonce() int64
	SetNonce(nonce int64) error
}
```
## 数据结构
### BaseAccount
```
type BaseAccount struct {
	AccountAddress types.AccAddress `json:"account_address"` // account address
	Publickey      crypto.PubKey    `json:"public_key"`      // public key
	Nonce          int64            `json:"nonce"`           // identifies tx_status of an account
}

```
## 方法规定
- 获得结构原型
- getter/setter
- 注册序列化的方法
