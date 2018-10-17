package types

import (
	"github.com/QOSGroup/qbase/account"
)

type AppAccount struct {
	account.BaseAccount `json:"base_account"`
	Coins               Coins `json:"coins"`
}

func NewAppAccount() account.Account {
	return &AppAccount{
		BaseAccount: account.BaseAccount{},
		Coins:       []Coin{},
	}
}

func (acc *AppAccount) GetCoins() Coins {
	return acc.Coins
}

func (acc *AppAccount) SetCoins(coins Coins) error {
	acc.Coins = coins
	return nil
}