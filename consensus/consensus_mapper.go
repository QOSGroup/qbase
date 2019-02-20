package consensus

import (
	"fmt"

	"github.com/QOSGroup/qbase/mapper"
	go_amino "github.com/tendermint/go-amino"
)

const (
	ConsensusMapperName = "consensus"
	consensusKey        = "cons_params"
)

//存储共识参数mapper
type ConsensusMapper struct {
	*mapper.BaseMapper
}

var _ mapper.IMapper = (*ConsensusMapper)(nil)

func BuildConsStoreQueryPath() []byte {
	return []byte(fmt.Sprintf("/store/%s/key", ConsensusMapperName))
}

func BuildConsKey() []byte {
	return []byte(consensusKey)
}

func NewConsensusMapper(cdc *go_amino.Codec) *ConsensusMapper {
	baseMapper := mapper.NewBaseMapper(cdc, ConsensusMapperName)
	return &ConsensusMapper{BaseMapper: baseMapper}
}

func (cons *ConsensusMapper) Copy() mapper.IMapper {
	copyBaseMapper := cons.BaseMapper.Copy()
	return &ConsensusMapper{BaseMapper: copyBaseMapper}
}
