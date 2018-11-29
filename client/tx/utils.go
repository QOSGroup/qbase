package tx

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/viper"

	"github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/keys"
	"github.com/QOSGroup/qbase/client/qcp"

	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"

	cflags "github.com/QOSGroup/qbase/client/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type ITxBuilder func(ctx context.CLIContext) (txs.ITx, error)

func BroadcastTxAndPrintResult(cdc *amino.Codec, txBuilder ITxBuilder) error {
	result, err := BroadcastTx(cdc, txBuilder)
	cliCtx := context.NewCLIContext().WithCodec(cdc)
	cliCtx.PrintResult(result)
	return err
}

func BroadcastTx(cdc *amino.Codec, txBuilder ITxBuilder) (*ctypes.ResultBroadcastTxCommit, error) {
	cliCtx := context.NewCLIContext().WithCodec(cdc)
	signedTx, err := buildAndSignTx(cliCtx, txBuilder)
	if err != nil {
		return nil, err
	}

	return cliCtx.BroadcastTx(cdc.MustMarshalBinaryBare(signedTx))
}

func buildAndSignTx(ctx context.CLIContext, txBuilder ITxBuilder) (signedTx types.Tx, err error) {

	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("buildAndSignTx recovered: %v\n", r)
			signedTx = nil
			err = errors.New(log)
		}
	}()

	itx, err := txBuilder(ctx)
	if err != nil {
		return nil, err
	}
	qcpMode := viper.GetBool(cflags.FlagQcp)
	if qcpMode {
		return BuildAndSignQcpTx(ctx, itx)
	} else {
		return BuildAndSignStdTx(ctx, itx, getChainID(ctx))
	}
}

func BuildAndSignQcpTx(ctx context.CLIContext, tx txs.ITx) (*txs.TxQcp, error) {

	qcpSigner := viper.GetString(cflags.FlagQcpSigner)
	qcpFrom := viper.GetString(cflags.FlagQcpFrom)

	if qcpSigner == "" || qcpFrom == "" {
		return nil, errors.New("in qcp mode, --qcp-from and --qcp-signer flag must set")
	}
	qcpSignerInfo, err := keys.GetKeyInfo(ctx, qcpSigner)
	if err != nil {
		return nil, errors.New("query qcp singer info error.")
	}

	toChainID := getChainID(ctx)
	qcpSeq := getQcpInSequence(ctx, qcpFrom)

	fmt.Println("> step 1. build and sign TxStd")

	txStd, err := BuildAndSignStdTx(ctx, tx, qcpFrom)
	if err != nil {
		return nil, err
	}

	fmt.Println("> step 2. build and sign TxQcp")
	_, ok := tx.(*txs.QcpTxResult)

	txQcp := txs.NewTxQCP(txStd, qcpFrom,
		toChainID,
		qcpSeq+1,
		viper.GetInt64(cflags.FlagQcpBlockHeight),
		viper.GetInt64(cflags.FlagQcpTxIndex),
		ok,
		viper.GetString(cflags.FlagQcpExtends),
	)

	sig, pubkey := signData(ctx, qcpSignerInfo.GetName(), txQcp.BuildSignatureBytes())
	txQcp.Sig = txs.Signature{
		Pubkey:    pubkey,
		Signature: sig,
	}

	return txQcp, nil
}

func BuildAndSignStdTx(ctx context.CLIContext, tx txs.ITx, txStdFromChainID string) (*txs.TxStd, error) {

	accountNonce := viper.GetInt64(cflags.FlagNonce)
	maxGas := viper.GetInt64(cflags.FlagMaxGas)
	if maxGas < 0 {
		return nil, errors.New("max-gas flag not correct")
	}

	chainID := getChainID(ctx)
	signers := getSigners(ctx, tx.GetSigner())
	txStd := txs.NewTxStd(tx, chainID, types.NewInt(maxGas))

	isUseFlagAccountNonce := accountNonce > 0
	for _, signerName := range signers {
		info, err := keys.GetKeyInfo(ctx, signerName)
		if err != nil {
			return nil, err
		}

		var actualNonce int64
		if isUseFlagAccountNonce {
			actualNonce = accountNonce + 1
		} else {
			nonce, err := getDefaultAccountNonce(ctx, info.GetAddress().Bytes())
			if err != nil || nonce < 0 {
				return nil, err
			}
			actualNonce = nonce + 1
		}

		txStd, err = signStdTx(ctx, signerName, actualNonce, txStd, txStdFromChainID)
		if err != nil {
			return nil, fmt.Errorf("name %s signStdTx error: %s", signerName, err.Error())
		}
	}

	return txStd, nil
}

func signStdTx(ctx context.CLIContext, signerKeyName string, nonce int64, txStd *txs.TxStd, txStdFromChainID string) (*txs.TxStd, error) {

	info, err := keys.GetKeyInfo(ctx, signerKeyName)
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

	sigdata := txStd.BuildSignatureBytes(nonce, txStdFromChainID)
	sig, pubkey := signData(ctx, signerKeyName, sigdata)

	txStd.Signature = append(txStd.Signature, txs.Signature{
		Pubkey:    pubkey,
		Signature: sig,
		Nonce:     nonce,
	})

	return txStd, nil
}

func signData(ctx context.CLIContext, name string, data []byte) ([]byte, crypto.PubKey) {

	pass, err := keys.GetPassphrase(ctx, name)
	if err != nil {
		panic(fmt.Sprintf("Get %s Passphrase error: %s", name, err.Error()))
	}

	keybase, err := keys.GetKeyBase(ctx)
	if err != nil {
		panic(err.Error())
	}

	sig, pubkey, err := keybase.Sign(name, pass, data)
	if err != nil {
		panic(err.Error())
	}

	return sig, pubkey
}

func getSigners(ctx context.CLIContext, txSignerAddrs []types.Address) []string {

	var sortNames []string

	for _, addr := range txSignerAddrs {

		keybase, err := keys.GetKeyBase(ctx)
		if err != nil {
			panic(err.Error())
		}

		info, err := keybase.GetByAddress(addr)
		if err != nil {
			panic(fmt.Sprintf("signer addr: %s not in current keybase. err:%s", addr, err.Error()))
		}

		sortNames = append(sortNames, info.GetName())
	}

	return sortNames
}

func getQcpInSequence(ctx context.CLIContext, inChainID string) int64 {
	seq := viper.GetInt64(cflags.FlagQcpSequence)
	if seq > 0 {
		return seq
	}

	seq, _ = qcp.GetInChainSequence(ctx, inChainID)
	return seq
}

func getChainID(ctx context.CLIContext) string {
	chainID := viper.GetString(cflags.FlagChainID)
	if chainID != "" {
		return chainID
	}

	defaultChainID, err := getDefaultChainID(ctx)
	if err != nil || defaultChainID == "" {
		panic("get chain id error")
	}

	return defaultChainID
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

	if ctx.NonceNodeURI == "" {
		return account.GetAccountNonce(ctx, address)
	}

	//NonceNodeURI不为空,使用NonceNodeURI查询account nonce值
	rpc := rpcclient.NewHTTP(ctx.NonceNodeURI, "/websocket")
	newCtx := context.NewCLIContext().WithClient(rpc).WithCodec(ctx.Codec)

	return account.GetAccountNonce(newCtx, address)
}
