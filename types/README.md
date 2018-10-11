# 基础类型定义
## Address
版本:
v0.1

日期:
2018年09月27日

### 简介
由公钥生成的唯一钱包地址
采用加密方法：ed25519

### 数据类型
byte数组

### 方法定义
- byte和string类型转换
- 判断空
- 判断两个地址是否相等
- Marshal返回byte数组/Unmashal设置地址值
- MarshalJson/UnmarshalJson
- 由公钥生成地址
- 已知公钥和地址，地址是否由该公钥生成

## Coin
版本:
v0.1

日期:
2018年09月27日

### 简介
链上的代币

### 数据类型
#### QOS
QOS公链提供的统一代币体系
```
type QOS struct {
	Name 	string 	`json:"coin_name"`
	Amount 	BigInt	`json:"amount"`
}
```
#### QSC
供联盟链自定义的代币统称
```
type QSC struct {
	Name	string	`json:"coin_name"`
	Amount	BigInt	`json:"amount"`
}
```

### 方法定义
- 成员变量的getters/setters