package tx

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/viper"

	"github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/keys"
	ctypes "github.com/QOSGroup/qbase/client/types"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
)

func BuildAndSignStdTx(ctx context.CLIContext, tx txs.ITx) (*txs.TxStd, error) {

	chainID := viper.GetString(ctypes.FlagChainID)
	accountNonce := viper.GetInt64(ctypes.FlagNonce)
	maxGas := viper.GetInt64(ctypes.FlagMaxGas)

	if chainID == "" {
		defaultChainID, err := getDefaultChainID(ctx)
		if err != nil || defaultChainID == "" {
			return nil, err
		}
		chainID = defaultChainID
	}

	if maxGas < 0 {
		return nil, errors.New("max-gas flag not correct")
	}

	signers, err := sortSigners(ctx, tx.GetSigner())
	if err != nil {
		return nil, err
	}
	txStd := txs.NewTxStd(tx, chainID, types.NewInt(maxGas))

	isUseFlagAccountNonce := accountNonce > 0
	for _, signerName := range signers {
		info, err := keys.GetKeyInfo(ctx, signerName)
		if err != nil {
			return nil, err
		}

		if isUseFlagAccountNonce {
			txStd, err = SignStdTx(ctx, signerName, accountNonce+1, txStd)
			if err != nil {
				return nil, fmt.Errorf("name %s signStdTx error: %s", signerName, err.Error())
			}
		} else {
			nonce, err := getDefaultAccountNonce(ctx, info.GetAddress().Bytes())
			if err != nil || nonce < 0 {
				return nil, err
			}
			txStd, err = SignStdTx(ctx, signerName, nonce+1, txStd)
			if err != nil {
				return nil, fmt.Errorf("name %s signStdTx error: %s", signerName, err.Error())
			}
		}

	}

	return txStd, nil
}

func getDefaultChainID(ctx context.CLIContext) (string, error) {
	client, err := ctx.GetNode()
	genesis, err := client.Genesis()
	if err != nil {
		return "", err
	}

	return genesis.Genesis.ChainID, nil
}

func getDefaultAccountNonce(ctx context.CLIContext, address []byte) (int64, error) {
	return account.GetAccountNonce(ctx, address)
}

func SignStdTx(ctx context.CLIContext, signerKeyName string, nonce int64, txStd *txs.TxStd) (*txs.TxStd, error) {

	keybase, err := keys.GetKeyBase(ctx)
	if err != nil {
		return nil, err
	}

	info, err := keybase.Get(signerKeyName)
	if err != nil {
		return nil, err
	}

	addr := info.GetAddress()
	ok := false

	for _, signer := range txStd.ITx.GetSigner() {
		if bytes.Equal(addr.Bytes(), signer.Bytes()) {
			ok = true
		}
	}

	if !ok {
		return nil, fmt.Errorf("Name %s is not signer", signerKeyName)
	}

	pass, err := keys.GetPassphrase(ctx, signerKeyName)
	if err != nil {
		return nil, fmt.Errorf("Get %s Passphrase error: %s", signerKeyName, err.Error())
	}

	sigdata := append(txStd.GetSignData(), types.Int2Byte(nonce)...)
	sig, pubkey, err := keybase.Sign(signerKeyName, pass, sigdata)

	if err != nil {
		return nil, fmt.Errorf("sign stdTx error: %s", err.Error())
	}

	txStd.Signature = append(txStd.Signature, txs.Signature{
		Pubkey:    pubkey,
		Signature: sig,
		Nonce:     nonce,
	})

	return txStd, nil
}

func sortSigners(ctx context.CLIContext, txSignerAddrs []types.Address) ([]string, error) {

	var sortNames []string

	for _, addr := range txSignerAddrs {

		keybase, err := keys.GetKeyBase(ctx)
		if err != nil {
			return nil, err
		}

		info, err := keybase.GetByAddress(addr)
		if err != nil {
			return nil, fmt.Errorf("signer addr: %s not in current keybase. err:%s", addr, err.Error())
		}

		sortNames = append(sortNames, info.GetName())
	}

	return sortNames, nil
}
