package context

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"github.com/QOSGroup/qbase/client"

	go_amino "github.com/tendermint/go-amino"
	cmn "github.com/tendermint/tendermint/libs/common"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// CLIContext implements a typical CLI context created in SDK modules for
// transaction handling and queries.
type CLIContext struct {
	Codec     *go_amino.Codec
	Client    rpcclient.Client
	Height    int64
	NodeURI   string
	Async     bool
	TrustNode bool
}

// NewCLIContext returns a new initialized CLIContext with parameters from the
// command line using Viper.
func NewCLIContext() CLIContext {
	var rpc rpcclient.Client
	nodeURI := viper.GetString(client.FlagNode)
	if nodeURI != "" {
		rpc = rpcclient.NewHTTP(nodeURI, "/websocket")
	}

	return CLIContext{
		Client:    rpc,
		NodeURI:   nodeURI,
		Height:    viper.GetInt64(client.FlagHeight),
		Async:     viper.GetBool(client.FlagAsync),
		TrustNode: viper.GetBool(client.FlagTrustNode),
	}
}

// WithCodec returns a copy of the context with an updated codec.
func (ctx CLIContext) WithCodec(cdc *go_amino.Codec) CLIContext {
	ctx.Codec = cdc
	return ctx
}

func (ctx CLIContext) WithNodeIP(nodeIP string) CLIContext {
	return ctx.WithNodeIPAndPort(nodeIP, 0)
}

func (ctx CLIContext) WithNodeIPAndPort(nodeIP string, nodeRPCPort int) CLIContext {
	if nodeRPCPort == 0 {
		nodeRPCPort = 26657
	}

	nodeURI := fmt.Sprintf("tcp://%s:%d", nodeIP, nodeRPCPort)
	ctx.NodeURI = nodeURI
	ctx.Client = rpcclient.NewHTTP(nodeURI, "/websocket")
	return ctx
}

// WithClient returns a copy of the context with an updated RPC client
// instance.
func (ctx CLIContext) WithClient(client rpcclient.Client) CLIContext {
	ctx.Client = client
	return ctx
}

func (ctx CLIContext) GetNode() (rpcclient.Client, error) {
	if ctx.Client == nil {
		return nil, errors.New("no RPC client defined")
	}
	return ctx.Client, nil
}

func (ctx CLIContext) Query(path string, data []byte) (res []byte, err error) {
	return ctx.query(path, cmn.HexBytes(data))
}

// query performs a query from a Tendermint node with the provided store name
// and path.
func (ctx CLIContext) query(path string, key cmn.HexBytes) (res []byte, err error) {
	node, err := ctx.GetNode()
	if err != nil {
		return res, err
	}

	opts := rpcclient.ABCIQueryOptions{
		Height:  ctx.Height,
		Trusted: ctx.TrustNode,
	}

	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return res, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return res, fmt.Errorf(resp.Log)
	}

	return resp.Value, nil
}

func (ctx CLIContext) BroadcastTx(txBytes []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	if ctx.Async {
		res, err := ctx.BroadcastTxAsync(txBytes)
		if err != nil {
			return nil, err
		}

		resCommit := &ctypes.ResultBroadcastTxCommit{
			Hash: res.Hash,
		}
		return resCommit, err
	}

	return ctx.BroadcastTxAndAwaitCommit(txBytes)
}

// BroadcastTxAndAwaitCommit broadcasts transaction bytes to a Tendermint node
// and waits for a commit.
func (ctx CLIContext) BroadcastTxAndAwaitCommit(tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return res, err
	}

	if !res.CheckTx.IsOK() {
		return res, fmt.Errorf(res.CheckTx.Log)
	}

	if !res.DeliverTx.IsOK() {
		return res, fmt.Errorf(res.DeliverTx.Log)
	}

	return res, err
}

// BroadcastTxSync broadcasts transaction bytes to a Tendermint node
// synchronously.
func (ctx CLIContext) BroadcastTxSync(tx []byte) (*ctypes.ResultBroadcastTx, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxSync(tx)
	if err != nil {
		return res, err
	}

	return res, err
}

// BroadcastTxAsync broadcasts transaction bytes to a Tendermint node
// asynchronously.
func (ctx CLIContext) BroadcastTxAsync(tx []byte) (*ctypes.ResultBroadcastTx, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxAsync(tx)
	if err != nil {
		return res, err
	}

	return res, err
}

func (ctx CLIContext) ToJSONIndentStr(obj interface{}) (string, error) {

	bz, err := ctx.Codec.MarshalJSONIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(bz), nil
}
