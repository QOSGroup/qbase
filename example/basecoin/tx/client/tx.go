package client

import (
	"encoding/hex"
	"fmt"
	"github.com/QOSGroup/qbase/client"
	cliacc "github.com/QOSGroup/qbase/client/account"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/keys"
	cliqcp "github.com/QOSGroup/qbase/client/qcp"
	btx "github.com/QOSGroup/qbase/client/tx"
	"github.com/QOSGroup/qbase/example/basecoin/tx"
	"github.com/QOSGroup/qbase/txs"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
)

const (
	flagFrom       = "from"
	flagTo         = "to"
	flagCoinName   = "coin-name"
	flagCoinAmount = "coin-amount"

	flagQCPChain = "qcp-chain"
)

// 链内交易
func SendTxCmd(cdc *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().
				WithCodec(cdc)

			fromName := viper.GetString(flagFrom)
			fromInfo, err := keys.GetKeyInfo(cliCtx, fromName)
			if err != nil {
				return err
			}
			from, err := cliacc.GetAccount(cliCtx, fromInfo.GetAddress())
			if err != nil {
				return err
			}

			toName := viper.GetString(flagTo)
			toInfo, err := keys.GetKeyInfo(cliCtx, toName)
			if err != nil {
				return err
			}

			name := viper.GetString(flagCoinName)
			amount, err := strconv.ParseInt(viper.GetString(flagCoinAmount), 10, 64)

			sendTx := tx.NewSendTx(from.GetAddress(), toInfo.GetAddress(), btypes.BaseCoin{name, btypes.NewInt(amount)})

			chainId, err := getChainId()
			if err != nil {
				return err
			}

			tx := txs.NewTxStd(&sendTx, chainId, btypes.NewInt(int64(0)))
			tx, err = btx.SignStdTx(cliCtx, fromName, from.GetNonce()+1, tx)
			if err != nil {
				return err
			}

			result, err := cliCtx.BroadcastTx(cdc.MustMarshalBinaryBare(tx))

			msg, _ := cdc.MarshalJSON(result)
			fmt.Println(string(msg))

			return err
		},
	}

	cmd.Flags().String(flagFrom, "", "Address to send coins")
	cmd.Flags().String(flagTo, "", "Address to receive coins")
	cmd.Flags().String(flagCoinName, "", "Name of coin to send")
	cmd.Flags().String(flagCoinAmount, "", "Amount of coin to send")

	return cmd
}

// 跨链交易
func SendQCPTxCmd(cdc *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-qcp",
		Short: "Create and sign a qcp send tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().
				WithCodec(cdc)

			fromName := viper.GetString(flagFrom)
			fromInfo, err := keys.GetKeyInfo(cliCtx, fromName)
			if err != nil {
				return err
			}
			from, err := cliacc.GetAccount(cliCtx, fromInfo.GetAddress())
			if err != nil {
				return err
			}

			toName := viper.GetString(flagTo)
			toInfo, err := keys.GetKeyInfo(cliCtx, toName)
			if err != nil {
				return err
			}

			name := viper.GetString(flagCoinName)
			amount, err := strconv.ParseInt(viper.GetString(flagCoinAmount), 10, 64)

			qcpChain := viper.GetString(flagQCPChain)
			seq, _ := cliqcp.GetInChainSequence(cliCtx, qcpChain)

			chainId, err := getChainId()
			if err != nil {
				return err
			}

			sendTx := tx.NewSendTx(from.GetAddress(), toInfo.GetAddress(), btypes.BaseCoin{name, btypes.NewInt(amount)})
			tx := txs.NewTxStd(&sendTx, chainId, btypes.NewInt(int64(0)))
			tx, err = btx.SignStdTx(cliCtx, fromName, from.GetNonce()+1, tx)
			if err != nil {
				return err
			}

			buf := client.BufferStdin()

			fmt.Print(fmt.Sprintf("PriKey to sign with %s chain:", qcpChain))
			hexPriKey, err := client.GetPassword("", buf)
			if err != nil {
				return err
			}
			qcpTx := txs.NewTxQCP(tx, qcpChain, chainId, seq+1, 0, 0, false, "")
			caHex, _ := hex.DecodeString(hexPriKey[2:])
			var caPriKey ed25519.PrivKeyEd25519
			cdc.MustUnmarshalBinaryBare(caHex, &caPriKey)
			sig, _ := qcpTx.SignTx(caPriKey)
			qcpTx.Sig.Nonce = seq
			qcpTx.Sig.Signature = sig
			qcpTx.Sig.Pubkey = caPriKey.PubKey()

			result, err := cliCtx.BroadcastTx(cdc.MustMarshalBinaryBare(qcpTx))

			msg, _ := cdc.MarshalJSON(result)
			fmt.Println(string(msg))

			return err
		},
	}

	cmd.Flags().String(flagFrom, "", "Address to send coins")
	cmd.Flags().String(flagTo, "", "Address to receive coins")
	cmd.Flags().String(flagCoinName, "", "Name of coin to send")
	cmd.Flags().String(flagCoinAmount, "", "Amount of coin to send")
	cmd.Flags().String(flagQCPChain, "", "QCP chain id")

	return cmd
}

func getChainId() (string, error) {
	chainId := viper.GetString(client.FlagChainID)
	if len(chainId) == 0 {
		cfg, err := tcmd.ParseConfig()
		if err != nil {
			return "", err
		}
		return btypes.GetChainID(cfg.BaseConfig.RootDir)
	}
	return chainId, nil
}
