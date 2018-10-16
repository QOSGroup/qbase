package types

import "fmt"

// 币的通用接口
type Coin interface {
	// getters and setters
	GetName() string
	GetAmount() BigInt
	SetAmount(amount BigInt)

	// 判断是否同种币
	SameNameAs(Coin) bool
	String() string

	// 判断币的数量
	IsZero() bool
	IsPositive() bool
	IsNegative() bool
	IsGreaterThan(Coin) bool
	IsLessThan(Coin) bool
	IsEqual(Coin) bool

	// 币的数量运算
	Plus(coinB Coin) Coin
	Minus(coinB Coin) Coin
}

type BaseCoin struct {
	Name   string       `json:"coin_name"`
	Amount BigInt 		`json:"amount"`
}

func NewBaseCoin(name string, amount BigInt) *BaseCoin {
	return &BaseCoin{
		Name:   name,
		Amount: amount,
	}
}

func (coin *BaseCoin) GetName() string {
	return coin.Name
}

func (coin *BaseCoin) GetAmount() BigInt {
	return coin.Amount
}

func (coin *BaseCoin) SetAmount(amount BigInt) {
	coin.Amount = amount
}

// 将币的信息输出为可读字符串
func (coin *BaseCoin) String() string {
	return fmt.Sprintf("%v%v", coin.Amount, coin.Name)
}

// 判断是否与另一币同名
func (coin *BaseCoin) SameNameAs(another Coin) bool {
	return (coin.Name == another.GetName())
}

// 判断币的数量是否为零
func (coin *BaseCoin) IsZero() bool {
	return coin.Amount.IsZero()
}

// 判断币的数量是否为正值
func (coin *BaseCoin) IsPositive() bool {
	return (coin.Amount.Sign() == 1)
}

// 判断币的数量是否为负值
func (coin *BaseCoin) IsNegative() bool {
	return (coin.Amount.Sign() == -1)
}

// 同名币，判断数量是否相等
func (coin *BaseCoin) IsEqual(another Coin) bool {
	return coin.SameNameAs(another) && (coin.Amount.Equal(another.GetAmount()))
}

// 同名币，判断数量是否更大
func (coin *BaseCoin) IsGreaterThan(another Coin) bool {
	return coin.SameNameAs(another) && (coin.Amount.GT(another.GetAmount()))
}

// 同名币，判断数量是否更小
func (coin *BaseCoin) IsLessThan(another Coin) bool {
	return coin.SameNameAs(another) && coin.Amount.LT(another.GetAmount())
}

// 增加一定数量的币
func (coin *BaseCoin) PlusByAmount(amountplus BigInt){
	coin.SetAmount(coin.Amount.Add(amountplus))
}

// 对同名币的数量做加法运算；如果不同名则返回原值
func (coin *BaseCoin) Plus(coinB Coin) Coin {
	if !coin.SameNameAs(coinB) {
		return coin
	}
	return NewBaseCoin(coin.Name, coin.Amount.Add(coinB.GetAmount()))
}

// 减掉一定数量的币
func (coin *BaseCoin) MinusByAmount(amountminus BigInt){
	coin.SetAmount(coin.Amount.Sub(amountminus))
}

// 对同名币的数量做减法运算；如果不同名则返回原值
func (coin *BaseCoin) Minus(coinB Coin) Coin {
	if !coin.SameNameAs(coinB) {
		return coin
	}
	return NewBaseCoin(coin.Name, coin.Amount.Sub(coinB.GetAmount()))
}