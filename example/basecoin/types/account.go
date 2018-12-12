package types

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/types"
)

type AppAccount struct {
	account.BaseAccount `json:"base_account"`
	Coins               types.Coins `json:"coins"`
}

func NewAppAccount() account.Account {
	return &AppAccount{
		BaseAccount: account.BaseAccount{},
		Coins:       types.Coins{},
	}
}

func (acc *AppAccount) GetCoins() types.Coins {
	return acc.Coins
}

func (acc *AppAccount) SetCoins(coins types.Coins) error {
	acc.Coins = coins
	return nil
}
