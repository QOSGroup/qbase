package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type BaseCoin struct {
	name   string `json:"coin_name"`
	amount BigInt `json:"amount"`
}

func NewBaseCoin(name string, amount BigInt) BaseCoin {
	return BaseCoin{
		name:   strings.ToUpper(name), //币种不区分大小写,默认大写
		amount: amount.NilToZero(),
	}
}

func NewInt64BaseCoin(name string, amount int64) BaseCoin {
	return NewBaseCoin(name, NewInt(amount))
}

func (coin BaseCoin) GetName() string {
	return strings.ToUpper(coin.name)
}

func (coin BaseCoin) GetAmount() BigInt {
	return coin.amount
}

// 将币的信息输出为可读字符串
func (coin BaseCoin) String() string {
	return fmt.Sprintf("%v%v", coin.GetAmount(), coin.GetName())
}

// 判断是否与另一币同名
func (coin BaseCoin) SameNameAs(another BaseCoin) bool {
	return coin.GetName() == another.GetName()
}

// 判断币的数量是否为零
func (coin BaseCoin) IsZero() bool {
	return coin.GetAmount().IsZero()
}

// 判断币的数量是否为正值
func (coin BaseCoin) IsPositive() bool {
	return (coin.GetAmount().Sign() == 1)
}

// 判断币的数量是否为非负值
func (coin BaseCoin) IsNotNegative() bool {
	return (coin.GetAmount().Sign() != -1)
}

// 判断币的数量是否为负值
func (coin BaseCoin) IsNegative() bool {
	return (coin.GetAmount().Sign() == -1)
}

// 同名币，判断数量是否相等
func (coin BaseCoin) IsEqual(another BaseCoin) bool {
	return coin.SameNameAs(another) && (coin.GetAmount().Equal(another.GetAmount()))
}

// 同名币，判断数量是否更大
func (coin BaseCoin) IsGreaterThan(another BaseCoin) bool {
	return coin.SameNameAs(another) && (coin.GetAmount().GT(another.GetAmount()))
}

// 同名币，判断数量是否更小
func (coin BaseCoin) IsLessThan(another BaseCoin) bool {
	return coin.SameNameAs(another) && coin.GetAmount().LT(another.GetAmount())
}

// 增加一定数量的币
func (coin BaseCoin) PlusByAmount(amountplus BigInt) BaseCoin {
	return NewBaseCoin(coin.GetName(), coin.GetAmount().Add(amountplus))
}

// 对同名币的数量做加法运算；如果不同名则返回原值
func (coin BaseCoin) Plus(coinB BaseCoin) BaseCoin {
	if !coin.SameNameAs(coinB) {
		return coin
	}
	return NewBaseCoin(coin.GetName(), coin.GetAmount().Add(coinB.GetAmount()))
}

// 减掉一定数量的币
func (coin BaseCoin) MinusByAmount(amountminus BigInt) BaseCoin {
	return BaseCoin{coin.GetName(), coin.GetAmount().Sub(amountminus)}
}

// 对同名币的数量做减法运算；如果不同名则返回原值
func (coin BaseCoin) Minus(coinB BaseCoin) BaseCoin {
	if !coin.SameNameAs(coinB) {
		return coin
	}
	return NewBaseCoin(coin.GetName(), coin.GetAmount().Sub(coinB.GetAmount()))
}

//----------------------------------------
// BaseCoins

// BaseCoin集合
type BaseCoins []BaseCoin

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
		lowDenom := coins[0].GetName()
		for _, coin := range coins[1:] {
			if coin.GetName() <= lowDenom {
				return false
			}
			if coin.IsZero() {
				return false
			}
			// we compare each coin against the last name
			lowDenom = coin.GetName()
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
	sum := ([]BaseCoin)(nil)
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
		switch strings.Compare(coinA.GetName(), coinB.GetName()) {
		case -1:
			sum = append(sum, coinA)
			indexA++
		case 0:
			if coinA.GetAmount().Add(coinB.GetAmount()).IsZero() {
				// ignore 0 sum coin type
			} else {
				sum = append(sum, coinA.Plus(coinB))
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
	res := make([]BaseCoin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, NewBaseCoin(coin.GetName(), coin.GetAmount().Neg()))
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
		if coins[i].GetName() != coinsB[i].GetName() || !coins[i].GetAmount().Equal(coinsB[i].GetAmount()) {
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
	coins = coins.Sort()
	upperName := strings.ToUpper(name)

	switch len(coins) {
	case 0:
		return ZeroInt()
	case 1:
		coin := coins[0]
		if coin.GetName() == upperName {
			return coin.GetAmount()
		}
		return ZeroInt()
	default:
		coins = coins.Sort()
		midIdx := len(coins) / 2
		coin := coins[midIdx]
		if upperName < coin.GetName() {
			return coins[:midIdx].AmountOf(upperName)
		} else if upperName == coin.GetName() {
			return coin.GetAmount()
		} else {
			return coins[midIdx+1:].AmountOf(upperName)
		}
	}
}

//----------------------------------------
// 排序

func (coins BaseCoins) Len() int           { return len(coins) }
func (coins BaseCoins) Less(i, j int) bool { return coins[i].GetName() < coins[j].GetName() }
func (coins BaseCoins) Swap(i, j int)      { coins[i], coins[j] = coins[j], coins[i] }

var _ sort.Interface = BaseCoins{}

func (coins BaseCoins) Sort() BaseCoins {
	sort.Sort(coins)
	return coins
}

func ParseCoins(str string) ([]BaseCoin, error) {
	if len(str) == 0 {
		return nil, nil
	}
	reDnm := `[[:alpha:]][[:alnum:]]{2,15}`
	reAmt := `[[:digit:]]+`
	reSpc := `[[:space:]]*`
	reCoin := regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reAmt, reSpc, reDnm))

	var coins []BaseCoin
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

		coins = append(coins, BaseCoin{
			coin[2],
			NewInt(amount),
		})
	}

	return coins, nil
}

//amino序列化支持
//amino格式: len(amount) + amount + name
func (coin BaseCoin) MarshalAmino() (string, error) {
	bz, err := marshalBaseCoinAmino(coin)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func (coin *BaseCoin) UnmarshalAmino(text string) error {
	return unmarshalBaseCoinAmino([]byte(text), coin)
}

func marshalBaseCoinAmino(coin BaseCoin) ([]byte, error) {
	if coin.amount.IsNil() {
		coin.amount = NewInt(0)
	}

	aminoBytes := make([]byte, 1, 37)

	a, err := coin.amount.MarshalAmino()
	if err != nil {
		return nil, err
	}

	bz := []byte(a)
	len := len(bz)

	aminoBytes[0] = (byte)(int8(len))
	aminoBytes = append(aminoBytes, bz...)
	aminoBytes = append(aminoBytes, []byte(coin.name)...)

	return aminoBytes, nil
}

func unmarshalBaseCoinAmino(bz []byte, coin *BaseCoin) error {
	amountBz := bz[1 : int8(bz[0])+1]
	err := coin.amount.UnmarshalAmino(string(amountBz))
	if err != nil {
		return err
	}
	coin.name = string(bz[int8(bz[0])+1:])
	return nil
}

//json序列化支持

type innerBaseCoin struct {
	Name   string `json:"coin_name"`
	Amount BigInt `json:"amount"`
}

func (coin BaseCoin) MarshalJSON() ([]byte, error) {
	return json.Marshal(innerBaseCoin{
		Name:   coin.name,
		Amount: coin.amount,
	})
}

func (coin *BaseCoin) UnmarshalJSON(bz []byte) error {
	var i innerBaseCoin

	err := json.Unmarshal(bz, &i)
	if err != nil {
		return err
	}

	coin.amount = i.Amount
	coin.name = i.Name

	return nil
}
