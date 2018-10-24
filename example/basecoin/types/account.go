package types

import (
	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/types"
)

type AppAccount struct {
	account.BaseAccount `json:"base_account"`
	Coins               []types.BaseCoin `json:"coins"`
}

func NewAppAccount() account.Account {
	return &AppAccount{
		BaseAccount: account.BaseAccount{},
		Coins:       []types.BaseCoin{},
	}
}

func (acc *AppAccount) GetCoins() []types.BaseCoin {
	return acc.Coins
}

func (acc *AppAccount) SetCoins(coins []types.BaseCoin) error {
	acc.Coins = coins
	return nil
}
