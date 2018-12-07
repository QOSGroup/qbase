package types

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

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
	IsNotNegative() bool
	IsNegative() bool
	IsGreaterThan(Coin) bool
	IsLessThan(Coin) bool
	IsEqual(Coin) bool

	// 币的数量运算
	Plus(coinB Coin) Coin
	Minus(coinB Coin) Coin
}

type BaseCoin struct {
	Name   string `json:"coin_name"`
	Amount BigInt `json:"amount"`
}

func NewBaseCoin(name string, amount BigInt) *BaseCoin {
	return &BaseCoin{
		Name:   name,
		Amount: amount.NilToZero(),
	}
}

func NewInt64BaseCoin(name string, amount int64) *BaseCoin {
	return NewBaseCoin(name, NewInt(amount))
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

// 判断币的数量是否为非负值
func (coin *BaseCoin) IsNotNegative() bool {
	return (coin.Amount.Sign() != -1)
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
func (coin *BaseCoin) PlusByAmount(amountplus BigInt) {
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
func (coin *BaseCoin) MinusByAmount(amountminus BigInt) {
	coin.SetAmount(coin.Amount.Sub(amountminus))
}

// 对同名币的数量做减法运算；如果不同名则返回原值
func (coin *BaseCoin) Minus(coinB Coin) Coin {
	if !coin.SameNameAs(coinB) {
		return coin
	}
	return NewBaseCoin(coin.Name, coin.Amount.Sub(coinB.GetAmount()))
}

//----------------------------------------
// BaseCoins

// BaseCoin集合
type BaseCoins []*BaseCoin

func (coins BaseCoins) String() string {
	if len(coins) == 0 {
		return ""
	}

	out := ""
	for _, coin := range coins {
		out += fmt.Sprintf("%v,", coin.String())
	}
	return out[:len(out)-1]
}

// 校验
// 1.排序好
// 2.amount没有0值
func (coins BaseCoins) IsValid() bool {
	switch len(coins) {
	case 0:
		return true
	case 1:
		return !coins[0].IsZero()
	default:
		lowDenom := coins[0].Name
		for _, coin := range coins[1:] {
			if coin.Name <= lowDenom {
				return false
			}
			if coin.IsZero() {
				return false
			}
			// we compare each coin against the last name
			lowDenom = coin.Name
		}
		return true
	}
}

// BaseCoins相加
// 注意：任何一个BaseCoin如果amount有0值将不会返回正确值
func (coins BaseCoins) Plus(coinsB BaseCoins) BaseCoins {
	if len(coins) > 1 {
		coins = coins.Sort()
	}
	if len(coinsB) > 1 {
		coinsB = coinsB.Sort()
	}
	sum := ([]*BaseCoin)(nil)
	indexA, indexB := 0, 0
	lenA, lenB := len(coins), len(coinsB)
	for {
		if indexA == lenA {
			if indexB == lenB {
				return sum
			}
			return append(sum, coinsB[indexB:]...)
		} else if indexB == lenB {
			return append(sum, coins[indexA:]...)
		}
		coinA, coinB := coins[indexA], coinsB[indexB]
		switch strings.Compare(coinA.Name, coinB.Name) {
		case -1:
			sum = append(sum, coinA)
			indexA++
		case 0:
			if coinA.Amount.Add(coinB.Amount).IsZero() {
				// ignore 0 sum coin type
			} else {
				sum = append(sum, coinA.Plus(coinB).(*BaseCoin))
			}
			indexA++
			indexB++
		case 1:
			sum = append(sum, coinB)
			indexB++
		}
	}
}

// 返回相反值
func (coins BaseCoins) Negative() BaseCoins {
	res := make([]*BaseCoin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, &BaseCoin{
			Name:   coin.Name,
			Amount: coin.Amount.Neg(),
		})
	}
	return res
}

// 相减
func (coins BaseCoins) Minus(coinsB BaseCoins) BaseCoins {
	return coins.Plus(coinsB.Negative())
}

// 返回coins内币种币值是否均大于等于coinsB对应值
func (coins BaseCoins) IsGTE(coinsB BaseCoins) bool {
	if coins == nil && coinsB == nil {
		return true
	}
	diff := coins.Minus(coinsB)
	if len(diff) == 0 {
		return true
	}
	return diff.IsNotNegative()
}

// 返回coins内币种币值是否均小于coinsB对应值
func (coins BaseCoins) IsLT(coinsB BaseCoins) bool {
	return !coins.IsGTE(coinsB)
}

// 返回coins内币种币值是否均等于0
func (coins BaseCoins) IsZero() bool {
	for _, coin := range coins {
		if !coin.IsZero() {
			return false
		}
	}
	return true
}

// 返回coins是否和coinsB一致
func (coins BaseCoins) IsEqual(coinsB BaseCoins) bool {
	if len(coins) != len(coinsB) {
		return false
	}
	if len(coins) > 1 {
		coins = coins.Sort()
	}
	if len(coinsB) > 1 {
		coinsB = coinsB.Sort()
	}
	for i := 0; i < len(coins); i++ {
		if coins[i].Name != coinsB[i].Name || !coins[i].Amount.Equal(coinsB[i].Amount) {
			return false
		}
	}
	return true
}

// 返回coins内币种币值是否均大于0
func (coins BaseCoins) IsPositive() bool {
	if len(coins) == 0 {
		return false
	}
	for _, coin := range coins {
		if !coin.IsPositive() {
			return false
		}
	}
	return true
}

// 返回coins内币种币值是否均大于等于0
func (coins BaseCoins) IsNotNegative() bool {
	if len(coins) == 0 {
		return true
	}
	for _, coin := range coins {
		if !coin.IsNotNegative() {
			return false
		}
	}
	return true
}

// 返回coins内给定币种币值
func (coins BaseCoins) AmountOf(name string) BigInt {
	switch len(coins) {
	case 0:
		return ZeroInt()
	case 1:
		coin := coins[0]
		if coin.Name == name {
			return coin.Amount
		}
		return ZeroInt()
	default:
		coins = coins.Sort()
		midIdx := len(coins) / 2
		coin := coins[midIdx]
		if name < coin.Name {
			return coins[:midIdx].AmountOf(name)
		} else if name == coin.Name {
			return coin.Amount
		} else {
			return coins[midIdx+1:].AmountOf(name)
		}
	}
}

//----------------------------------------
// 排序

func (coins BaseCoins) Len() int           { return len(coins) }
func (coins BaseCoins) Less(i, j int) bool { return coins[i].Name < coins[j].Name }
func (coins BaseCoins) Swap(i, j int)      { coins[i], coins[j] = coins[j], coins[i] }

var _ sort.Interface = BaseCoins{}

func (coins BaseCoins) Sort() BaseCoins {
	sort.Sort(coins)
	return coins
}

func ParseCoins(str string) ([]*BaseCoin, error) {
	if len(str) == 0 {
		return nil, nil
	}
	reDnm := `[[:alpha:]][[:alnum:]]{2,15}`
	reAmt := `[[:digit:]]+`
	reSpc := `[[:space:]]*`
	reCoin := regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reAmt, reSpc, reDnm))

	var coins []*BaseCoin
	arr := strings.Split(str, ",")
	for _, q := range arr {
		coin := reCoin.FindStringSubmatch(q)
		if len(coin) != 3 {
			return coins, fmt.Errorf("coins str: %s parse faild", q)
		}
		coin[2] = strings.TrimSpace(coin[2])
		amount, err := strconv.ParseInt(strings.TrimSpace(coin[1]), 10, 64)
		if err != nil {
			return coins, err
		}

		coins = append(coins, &BaseCoin{
			coin[2],
			NewInt(amount),
		})
	}

	return coins, nil
}
