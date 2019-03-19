package validator

import (
	"bytes"

	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/mapper"
	"github.com/QOSGroup/qbase/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

type ValidatorMapper struct {
	*mapper.BaseMapper
}

func (mapper *ValidatorMapper) ClearValidatorUpdateSet() {
	mapper.Del(ValidatorUpdateSetKey)
}

func (mapper *ValidatorMapper) GetValidatorUpdateSet() (allValidatorUpdate []abci.ValidatorUpdate) {
	mapper.Get(ValidatorUpdateSetKey, &allValidatorUpdate)
	return
}

func (mapper *ValidatorMapper) AddValidatorUpdate(pubkey crypto.PubKey, power uint64) error {

	updateInfo := abci.ValidatorUpdate{
		PubKey: tmtypes.TM2PB.PubKey(pubkey),
		Power:  int64(power),
	}

	allValidatorUpdate := mapper.GetValidatorUpdateSet()
	isReplace := false

	for i, vu := range allValidatorUpdate {
		if bytes.Equal(updateInfo.PubKey.Data, vu.PubKey.Data) {
			isReplace = true
			allValidatorUpdate[i] = updateInfo
			break
		}
	}

	if !isReplace {
		allValidatorUpdate = append(allValidatorUpdate, updateInfo)
	}

	mapper.SetValidatorUpdateSet(allValidatorUpdate)
	return nil
}

func (mapper *ValidatorMapper) SetValidatorUpdateSet(allValidatorUpdate []abci.ValidatorUpdate) {
	mapper.Set(ValidatorUpdateSetKey, allValidatorUpdate)
}

func (mapper *ValidatorMapper) SetLastBlockProposer(address types.Address) {
	mapper.Set(LastBlockProposerKey, address)
}

func (mapper *ValidatorMapper) GetLastBlockProposer() (address types.Address, exsits bool) {
	exsits = mapper.Get(LastBlockProposerKey, &address)
	return
}

func (mapper *ValidatorMapper) IsEnableValidatorUpdated() bool {
	if v, exsits := mapper.GetBool(EnableValidatorUpdatedKey); exsits {
		return v
	}
	return false
}

func (mapper *ValidatorMapper) EnableValidatorUpdated() {
	mapper.Set(EnableValidatorUpdatedKey, true)
}

func (mapper *ValidatorMapper) DisableValidatorUpdated() {
	mapper.Set(EnableValidatorUpdatedKey, false)
}

var _ mapper.IMapper = (*ValidatorMapper)(nil)

func NewValidatorMapper() *ValidatorMapper {
	var validatorMapper = ValidatorMapper{}
	validatorMapper.BaseMapper = mapper.NewBaseMapper(nil, ValidatorMapperName)
	return &validatorMapper
}

func GetValidatorMapper(ctx context.Context) *ValidatorMapper {
	return ctx.Mapper(ValidatorMapperName).(*ValidatorMapper)
}

func (mapper *ValidatorMapper) Copy() mapper.IMapper {
	validatorMapper := &ValidatorMapper{}
	validatorMapper.BaseMapper = mapper.BaseMapper.Copy()
	return validatorMapper
}
