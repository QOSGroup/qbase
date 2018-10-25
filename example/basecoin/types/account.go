package types

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/types"
)

type AppAccount struct {
	account.BaseAccount `json:"base_account"`
	Coins               types.BaseCoins `json:"coins"`
}

func NewAppAccount() account.Account {
	return &AppAccount{
		BaseAccount: account.BaseAccount{},
		Coins:       types.BaseCoins{},
	}
}

func (acc *AppAccount) GetCoins() types.BaseCoins {
	return acc.Coins
}

func (acc *AppAccount) SetCoins(coins types.BaseCoins) error {
	acc.Coins = coins
	return nil
}
