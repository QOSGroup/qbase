package tx

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/keys"
	kb "github.com/QOSGroup/qbase/keys"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
)

func SignStdTx(ctx context.CLIContext, name string, nonce int64, txStd *txs.TxStd) (*txs.TxStd, error) {

	keybase, err := keys.GetKeyBase(ctx)
	if err != nil {
		return nil, err
	}

	info, err := keybase.Get(name)
	if err != nil {
		return nil, err
	}

	if info.GetType() == kb.TypeOffline {
		return nil, errors.New("offline keytype not support")
	}

	addr := info.GetAddress()
	ok := false

	for _, singer := range txStd.ITx.GetSigner() {
		if bytes.Equal(addr.Bytes(), singer.Bytes()) {
			ok = true
		}
	}

	if !ok {
		return nil, fmt.Errorf("Name %s is not singer", name)
	}

	pass, err := keys.GetPassphrase(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("Get %s Passphrase error: %s", name, err.Error())
	}

	sigdata := append(txStd.GetSignData(), types.Int2Byte(nonce)...)
	sig, pubkey, err := keybase.Sign(name, pass, sigdata)

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
