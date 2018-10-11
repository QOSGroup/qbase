# 签名、加密方法使用
qos目前直接调用tendermint封装的ed25519加密方法

见 [tendermint/tendermint/crypto/ed25519/ed25519.go](https://github.com/tendermint/tendermint/blob/master/crypto/ed25519/ed25519.go)

[TOC]

## 私钥
使用其定义 `PrivKeyEd25519`  https://github.com/tendermint/tendermint/blob/0c9c3292c918617624f6f3fbcd95eceade18bcd5/crypto/ed25519/ed25519.go#L41

### 生成私钥
#### 随机生成
https://github.com/tendermint/tendermint/blob/0c9c3292c918617624f6f3fbcd95eceade18bcd5/crypto/ed25519/ed25519.go#L102
```
// GenPrivKey generates a new ed25519 private key.
// It uses OS randomness in conjunction with the current global random seed
// in tendermint/libs/common to generate the private key.
func GenPrivKey() PrivKeyEd25519 {
	return genPrivKey(crypto.CReader())
}
```
#### 使用密码生成
参数为密码
https://github.com/tendermint/tendermint/blob/0c9c3292c918617624f6f3fbcd95eceade18bcd5/crypto/ed25519/ed25519.go#L124
```
// GenPrivKeyFromSecret hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyFromSecret(secret []byte) PrivKeyEd25519 {
	privKey32 := crypto.Sha256(secret) // Not Ripemd160 because we want 32 bytes.
	privKey := new([64]byte)
	copy(privKey[:32], privKey32)
	// ed25519.MakePublicKey(privKey) alters the last 32 bytes of privKey.
	// It places the pubkey in the last 32 bytes of privKey, and returns the
	// public key.
	ed25519.MakePublicKey(privKey)
	return PrivKeyEd25519(*privKey)
}
```
### 使用私钥签名
参数为被签名的消息
https://github.com/tendermint/tendermint/blob/0c9c3292c918617624f6f3fbcd95eceade18bcd5/crypto/ed25519/ed25519.go#L49
```
func (privKey PrivKeyEd25519) Sign(msg []byte) ([]byte, error) {
	privKeyBytes := [64]byte(privKey)
	signatureBytes := ed25519.Sign(&privKeyBytes, msg)
	return signatureBytes[:], nil
}
```


## 公钥
使用其定义 `PubKeyEd25519`
https://github.com/tendermint/tendermint/blob/0c9c3292c918617624f6f3fbcd95eceade18bcd5/crypto/ed25519/ed25519.go#L137

### 生成公钥

从私钥生成公钥
https://github.com/tendermint/tendermint/blob/0c9c3292c918617624f6f3fbcd95eceade18bcd5/crypto/ed25519/ed25519.go#L55
```
// PubKey gets the corresponding public key from the private key.
func (privKey PrivKeyEd25519) PubKey() crypto.PubKey {
	privKeyBytes := [64]byte(privKey)
	initialized := false
	// If the latter 32 bytes of the privkey are all zero, compute the pubkey
	// otherwise privkey is initialized and we can use the cached value inside
	// of the private key.
	for _, v := range privKeyBytes[32:] {
		if v != 0 {
			initialized = true
			break
		}
	}
	if initialized {
		var pubkeyBytes [PubKeyEd25519Size]byte
		copy(pubkeyBytes[:], privKeyBytes[32:])
		return PubKeyEd25519(pubkeyBytes)
	}

	pubBytes := *ed25519.MakePublicKey(&privKeyBytes)
	return PubKeyEd25519(pubBytes)
}
```

### 生成地址
对公钥进行SHA256编码得到地址，即hexAddress
https://github.com/tendermint/tendermint/blob/0c9c3292c918617624f6f3fbcd95eceade18bcd5/crypto/ed25519/ed25519.go#L146

```
// Address is the SHA256-20 of the raw pubkey bytes.
func (pubKey PubKeyEd25519) Address() crypto.Address {
	return crypto.Address(tmhash.Sum(pubKey[:]))
}
```
需要bech32编码，可以进而通过address.go中提供的的方法进行转换

### 使用公钥对签名进行验证
传入被签名的消息，签名
https://github.com/tendermint/tendermint/blob/0c9c3292c918617624f6f3fbcd95eceade18bcd5/crypto/ed25519/ed25519.go#L159
```
func (pubKey PubKeyEd25519) VerifyBytes(msg []byte, sig_ []byte) bool {
	// make sure we use the same algorithm to sign
	if len(sig_) != SignatureSize {
		return false
	}
	sig := new([SignatureSize]byte)
	copy(sig[:], sig_)
	pubKeyBytes := [PubKeyEd25519Size]byte(pubKey)
	return ed25519.Verify(&pubKeyBytes, msg, sig)
}
```