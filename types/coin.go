package types

// common interface
type Coin interface {
	GetName() string
	GetAmount() BigInt
	SetAmount(amount BigInt)
}
