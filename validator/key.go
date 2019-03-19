package validator

const (
	ValidatorMapperName = "_base_validator_"
)

var (
	//EnableValidatorUpdatedKey 是否启用更新validator特性,默认关闭
	EnableValidatorUpdatedKey = []byte("_enable_validator_updated_")

	//ValidatorUpdateSetKey 保存本块中validator变化结果
	ValidatorUpdateSetKey = []byte("_validator_update_set_")

	//LastBlockProposerKey 上一块验证人
	LastBlockProposerKey = []byte("_last_block_proposer_")
)
