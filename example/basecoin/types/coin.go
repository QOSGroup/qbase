package types

import (
	"fmt"
	"github.com/QOSGroup/qbase/types"
	"strings"
)

type Coin struct {
	Name   string       `json:"name"`
	Amount types.BigInt `json:"amount"`
}

func NewCoin(name string, amount types.BigInt) Coin {
	return Coin{
		Name:   name,
		Amount: amount,
	}
}

func (coin *Coin) GetName() string {
	return coin.Name
}

func (coin *Coin) GetAmount() types.BigInt {
	return coin.Amount
}

func (coin *Coin) SetAmount(amount types.BigInt) {
	coin.SetAmount(amount)
}

// String provides a human-readable representation of a coin
func (coin Coin) String() string {
	return fmt.Sprintf("%v%v", coin.Amount, coin.Name)
}

// SameDenomAs returns true if the two coins are the same denom
func (coin Coin) SameNameAs(other Coin) bool {
	return (coin.Name == other.Name)
}

// IsZero returns if this represents no money
func (coin Coin) IsZero() bool {
	return coin.Amount.IsZero()
}

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value
func (coin Coin) IsGTE(other Coin) bool {
	return coin.SameNameAs(other) && (!coin.Amount.LT(other.Amount))
}

// IsLT returns true if they are the same type and the receiver is
// a smaller value
func (coin Coin) IsLT(other Coin) bool {
	return !coin.IsGTE(other)
}

// IsEqual returns true if the two sets of Coins have the same value
func (coin Coin) IsEqual(other Coin) bool {
	return coin.SameNameAs(other) && (coin.Amount.Equal(other.Amount))
}

// IsPositive returns true if coin amount is positive
func (coin Coin) IsPositive() bool {
	return (coin.Amount.Sign() == 1)
}

// IsNotNegative returns true if coin amount is not negative
func (coin Coin) IsNotNegative() bool {
	return (coin.Amount.Sign() != -1)
}

// Adds amounts of two coins with same denom
func (coin Coin) Plus(coinB Coin) Coin {
	if !coin.SameNameAs(coinB) {
		return coin
	}
	return Coin{coin.Name, coin.Amount.Add(coinB.Amount)}
}

// Subtracts amounts of two coins with same denom
func (coin Coin) Minus(coinB Coin) Coin {
	if !coin.SameNameAs(coinB) {
		return coin
	}
	return Coin{coin.Name, coin.Amount.Sub(coinB.Amount)}
}

//----------------------------------------
// Coins

// Coins is a set of Coin, one per currency
type Coins []Coin

func (coins Coins) String() string {
	if len(coins) == 0 {
		return ""
	}

	out := ""
	for _, coin := range coins {
		out += fmt.Sprintf("%v,", coin.String())
	}
	return out[:len(out)-1]
}

// IsValid asserts the Coins are sorted, and don't have 0 amounts
func (coins Coins) IsValid() bool {
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
			// we compare each coin against the last denom
			lowDenom = coin.Name
		}
		return true
	}
}

// Plus combines two sets of coins
// CONTRACT: Plus will never return Coins where one Coin has a 0 amount.
func (coins Coins) Plus(coinsB Coins) Coins {
	sum := ([]Coin)(nil)
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

func (coins Coins) PlusSingle(coin Coin) Coins {
	sum := ([]Coin)(nil)
	sum = append(sum, coin)
	return coins.Plus(sum)
}

// Negative returns a set of coins with all amount negative
func (coins Coins) Negative() Coins {
	res := make([]Coin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, Coin{
			Name:   coin.Name,
			Amount: coin.Amount.Neg(),
		})
	}
	return res
}

// Minus subtracts a set of coins from another (adds the inverse)
func (coins Coins) Minus(coinsB Coins) Coins {
	return coins.Plus(coinsB.Negative())
}

func (coins Coins) MinusSingle(coin Coin) Coins {
	sum := ([]Coin)(nil)
	sum = append(sum, coin)
	return coins.Minus(sum)
}

// IsGTE returns True iff coins is NonNegative(), and for every
// currency in coinsB, the currency is present at an equal or greater
// amount in coinsB
func (coins Coins) IsGTE(coinsB Coins) bool {
	diff := coins.Minus(coinsB)
	if len(diff) == 0 {
		return true
	}
	return diff.IsNotNegative()
}

// IsLT returns True iff every currency in coins, the currency is
// present at a smaller amount in coins
func (coins Coins) IsLT(coinsB Coins) bool {
	return !coins.IsGTE(coinsB)
}

// IsZero returns true if there are no coins
// or all coins are zero.
func (coins Coins) IsZero() bool {
	for _, coin := range coins {
		if !coin.IsZero() {
			return false
		}
	}
	return true
}

// IsEqual returns true if the two sets of Coins have the same value
func (coins Coins) IsEqual(coinsB Coins) bool {
	if len(coins) != len(coinsB) {
		return false
	}
	for i := 0; i < len(coins); i++ {
		if coins[i].Name != coinsB[i].Name || !coins[i].Amount.Equal(coinsB[i].Amount) {
			return false
		}
	}
	return true
}

// IsPositive returns true if there is at least one coin, and all
// currencies have a positive value
func (coins Coins) IsPositive() bool {
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

// IsNotNegative returns true if there is no currency with a negative value
// (even no coins is true here)
func (coins Coins) IsNotNegative() bool {
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

// Returns the amount of a denom from coins
func (coins Coins) AmountOf(denom string) types.BigInt {
	switch len(coins) {
	case 0:
		return types.ZeroInt()
	case 1:
		coin := coins[0]
		if coin.Name == denom {
			return coin.Amount
		}
		return types.ZeroInt()
	default:
		midIdx := len(coins) / 2 // 2:1, 3:1, 4:2
		coin := coins[midIdx]
		if denom < coin.Name {
			return coins[:midIdx].AmountOf(denom)
		} else if denom == coin.Name {
			return coin.Amount
		} else {
			return coins[midIdx+1:].AmountOf(denom)
		}
	}
}
