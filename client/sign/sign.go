package sign

import (
	"errors"
	"fmt"
	"github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/client/context"
	clientTx "github.com/QOSGroup/qbase/client/tx"
	"github.com/QOSGroup/qbase/client/types"
	"github.com/QOSGroup/qbase/txs"
	qtypes "github.com/QOSGroup/qbase/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"io/ioutil"
)

func SignCommand(cdc *amino.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:    "sign [file]",
		Short:  "Sign transactions generated offline",
		PreRun: checkSignCmd,
		RunE:   makeSignCmd(cdc),
		Args:   cobra.ExactArgs(1),
	}

	cmd = types.PostCommands(cmd)[0]

	cmd.Flags().Bool(types.FlagSigOnly, false, "Print only the generated signature, then exit")
	cmd.Flags().Bool(types.FlagOffline, false, "Offline mode; Do not query a full node. --nonce must be set if offline")
	cmd.Flags().String(types.FlagSigner, "", "Signer address or keybase name")
	cmd.MarkFlagRequired(types.FlagSigner)

	return cmd
}

func makeSignCmd(cdc *amino.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.NewCLIContext().WithCodec(cdc)
		isOffline := viper.GetBool(types.FlagOffline)
		nonce := viper.GetInt64(types.FlagNonce)
		isSignOnly := viper.GetBool(types.FlagSigOnly)

		txFile := args[0]

		bz, err := ioutil.ReadFile(txFile)
		if err != nil {
			return fmt.Errorf("read sign file: %s error. err: %s", txFile, err.Error())
		}

		signerAddress, err := account.GetAddrFromFlag(ctx, types.FlagSigner)
		if err != nil {
			return err
		}

		var tx qtypes.Tx
		if err := cdc.UnmarshalJSON(bz, &tx); err != nil {
			return errors.New("unmarshalJSON sign file error")
		}

		implTx, ok := tx.(*txs.TxStd)
		if !ok {
			return errors.New("not support tx type.")
		}

		allSigners := implTx.GetSigners()
		index, err := searchAddress(allSigners, signerAddress)
		if err != nil {
			return err
		}

		if !isOffline {
			if nonce, err = account.GetAccountNonce(ctx, signerAddress); err != nil {
				return err
			}
		}

		actualNonce := nonce + 1
		data := implTx.BuildSignatureBytes(actualNonce, implTx.ChainID)
		sig, pubKey := clientTx.SignDataFromAddress(ctx, signerAddress, data)

		signature := txs.Signature{
			Pubkey:    pubKey,
			Signature: sig,
			Nonce:     actualNonce,
		}

		if isSignOnly {
			ctx.PrintResult(signature)
			return nil
		}

		siges := implTx.Signature
		if len(siges) == 0 {
			siges = make([]txs.Signature, len(allSigners))
		}
		siges[index] = signature

		implTx.Signature = siges
		ctx.PrintResult(implTx)

		return nil
	}
}

func checkSignCmd(cmd *cobra.Command, _ []string) {
	if viper.GetBool(types.FlagOffline) {
		cmd.MarkFlagRequired(types.FlagNonce)
	}
}

func searchAddress(addresses []qtypes.AccAddress, searchAddress qtypes.AccAddress) (index int, err error) {
	if len(addresses) == 0 {
		return -1, errors.New("search addresses list is empty.")
	}
	if searchAddress.Empty() {
		return -1, errors.New("search address is empty.")
	}

	index = -1
	err = errors.New("singer address not found. it's may not a signer in tx")

	for i, addr := range addresses {
		if addr.Equals(searchAddress) {
			index = i
			err = nil
			return
		}
	}
	return
}
