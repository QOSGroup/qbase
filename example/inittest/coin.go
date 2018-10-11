package inittest

import "github.com/QOSGroup/qbase/types"

// coins for the specific regional chain
type QSC struct {
	Name   string       `json:"coin_name"`
	Amount types.BigInt `json:"amount"`
}

func NewQSC(name string, amount types.BigInt) *QSC {
	return &QSC{
		name,
		amount,
	}
}

// getter of qsc name
func (qsc *QSC) GetName() string {
	return qsc.Name
}

// getter of qsc name
func (qsc *QSC) GetAmount() types.BigInt {
	return qsc.Amount
}

// setter of qsc amount
func (qsc *QSC) SetAmount(amount types.BigInt) {
	qsc.Amount = amount
}
