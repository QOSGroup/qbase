package block

import (
	"errors"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/types"
	"github.com/QOSGroup/qbase/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	go_amino "github.com/tendermint/go-amino"
)

const (
	flagPath = "path"
	flagData = "data"
)

func storeCommand(cdc *go_amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "store",
		Short: "Query store data by low level",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(types.FlagTrustNode, true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			node, err := cliCtx.GetNode()
			if err != nil {
				return err
			}

			result, err := node.ABCIQuery(viper.GetString(flagPath), []byte(viper.GetString(flagData)))
			if err != nil {
				return err
			}

			valueBz := result.Response.GetValue()
			if len(valueBz) == 0 {
				return errors.New("response empty value")
			}

			val, err := tryDecodeValue(cliCtx.Codec, valueBz, true)
			if err != nil {
				return err
			}

			return cliCtx.PrintResult(val)
		},
	}

	cmd.Flags().String(flagPath, "", "store query path")
	cmd.Flags().String(flagData, "", "store query data")
	cmd.Flags().StringP(types.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().Bool(types.FlagJSONIndet, false, "print indent result json")
	viper.BindPFlag(types.FlagNode, cmd.Flags().Lookup(types.FlagNode))

	cmd.MarkFlagRequired(flagPath)
	cmd.MarkFlagRequired(flagData)

	return cmd
}

func noPaincRegisterInterface(cdc *go_amino.Codec) {
	defer func() {
		if r := recover(); r != nil {
			//nothing
		}
	}()
	cdc.RegisterInterface((*interface{})(nil), nil)
}

func noPaincRegisterKVPairResult(cdc *go_amino.Codec) {
	defer func() {
		if r := recover(); r != nil {
			//nothing
		}
	}()
	cdc.RegisterConcrete([]kvPairResult{}, "client/store/kvPairResults", nil)
}

type kvPairResult struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func tryDecodeValue(cdc *go_amino.Codec, bz []byte, useKVPairFlag bool) (interface{}, error) {
	noPaincRegisterInterface(cdc)

	// if len(bz) != 1 {
	var vInterface interface{}
	err := cdc.UnmarshalBinaryBare(bz, &vInterface)
	if err == nil {
		return vInterface, nil
	}

	if useKVPairFlag {
		var vKVPair []store.KVPair
		err = cdc.UnmarshalBinaryLengthPrefixed(bz, &vKVPair)
		if err == nil {
			var pairResults []kvPairResult
			for _, pair := range vKVPair {
				val, _ := tryDecodeValue(cdc, pair.Value, false)
				pairResults = append(pairResults, kvPairResult{
					Key:   string(pair.Key),
					Value: val,
				})
			}
			// noPaincRegisterKVPairResult(cdc)
			return pairResults, nil
		}
	}

	// }

	var vBool bool
	err = cdc.UnmarshalBinaryBare(bz, &vBool)
	if err == nil {
		return vBool, nil
	}

	var vInt int64
	err = cdc.UnmarshalBinaryBare(bz, &vInt)
	if err == nil {
		return vInt, nil
	}

	var vString string
	err = cdc.UnmarshalBinaryBare(bz, &vString)
	if err == nil {
		return vString, nil
	}

	return bz, errors.New("can't decode value")
}
