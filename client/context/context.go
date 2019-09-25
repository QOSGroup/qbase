package context

import (
	"encoding/json"
	"errors"
	"fmt"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/viper"

	"github.com/QOSGroup/qbase/client/types"
	goAmino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	cTypes "github.com/tendermint/tendermint/rpc/core/types"
)

var nodeRe = regexp.MustCompile(`(?i:^tcp://\S+(:\d+)?$)`)

type BroadcastMode int

const (
	BroadcastBlock BroadcastMode = iota
	BroadcastSync
	BroadcastAsync
)

// CLIContext implements a typical CLI context created in SDK modules for
// transaction handling and queries.
type CLIContext struct {
	Codec        *goAmino.Codec
	Client       rpcClient.Client
	Height       int64
	NodeURI      string
	Mode         BroadcastMode
	TrustNode    bool
	NonceNodeURI string
	JSONIndent   bool
	ChainID      string
}

// NewCLIContext returns a new initialized CLIContext with parameters from the
// command line using Viper.
func NewCLIContext() CLIContext {
	//优先从$config-home/config.toml文件中加载选项
	loadCliConfiguration()

	var rpc rpcClient.Client
	nodeURI := viper.GetString(types.FlagNode)
	if nodeURI != "" {
		rpc = rpcClient.NewHTTP(nodeURI, "/websocket")
	}

	var nonceNodeURI string
	nonceNodeValue := viper.GetString(types.FlagNonceNode)
	if nonceNodeValue != "" && nodeRe.MatchString(nonceNodeValue) {
		nonceNodeURI = nonceNodeValue
	}

	return CLIContext{
		Client:       rpc,
		NodeURI:      nodeURI,
		ChainID:      viper.GetString(types.FlagChainID),
		Height:       viper.GetInt64(types.FlagHeight),
		Mode:         parseMode(viper.GetString(types.FlagBroadcastMode)),
		TrustNode:    viper.GetBool(types.FlagTrustNode),
		JSONIndent:   viper.GetBool(types.FlagJSONIndet),
		NonceNodeURI: nonceNodeURI,
	}
}

func parseMode(mode string) BroadcastMode {
	var broadcastMode BroadcastMode
	mode = strings.TrimSpace(mode)
	if strings.ToLower(mode) == "sync" {
		broadcastMode = BroadcastSync
	} else if strings.ToLower(mode) == "async" {
		broadcastMode = BroadcastAsync
	} else {
		broadcastMode = BroadcastBlock
	}
	return broadcastMode
}

// WithCodec returns a copy of the context with an updated codec.
func (ctx CLIContext) WithCodec(cdc *goAmino.Codec) CLIContext {
	ctx.Codec = cdc
	return ctx
}

func (ctx CLIContext) WithChainID(chainID string) CLIContext {
	ctx.ChainID = chainID
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
	ctx.Client = rpcClient.NewHTTP(nodeURI, "/websocket")
	return ctx
}

// WithClient returns a copy of the context with an updated RPC client
// instance.
func (ctx CLIContext) WithClient(client rpcClient.Client) CLIContext {
	ctx.Client = client
	return ctx
}

func (ctx CLIContext) WithHeight(height int64) CLIContext {
	ctx.Height = height
	return ctx
}

func (ctx CLIContext) WithBroadcastMode(mode string) CLIContext {
	ctx.Mode = parseMode(mode)
	return ctx
}

func (ctx CLIContext) GetCodec() (*goAmino.Codec, error) {
	if ctx.Codec == nil {
		return nil, errors.New("no Codec defined")
	}
	return ctx.Codec, nil
}

func (ctx CLIContext) GetNode() (rpcClient.Client, error) {
	if ctx.Client == nil {
		return nil, errors.New("no RPC client defined")
	}
	return ctx.Client, nil
}

func (ctx CLIContext) GetHeight() int64 {
	return ctx.Height
}

func (ctx CLIContext) GetMode() int64 {
	return int64(ctx.Mode)
}

func (ctx CLIContext) GetNonceNodeURI() string {
	return ctx.NonceNodeURI
}

func (ctx CLIContext) IsTrustNode() bool {
	return ctx.TrustNode
}

func (ctx CLIContext) IsJSONIndent() bool {
	return ctx.JSONIndent
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

	opts := rpcClient.ABCIQueryOptions{
		Height: ctx.Height,
		Prove:  ctx.TrustNode,
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

func (ctx CLIContext) BroadcastTx(txBytes []byte) (*cTypes.ResultBroadcastTxCommit, error) {

	if ctx.Mode == BroadcastAsync {
		res, err := ctx.BroadcastTxAsync(txBytes)
		if err != nil {
			return nil, err
		}

		resCommit := &cTypes.ResultBroadcastTxCommit{
			Hash: res.Hash,
		}
		return resCommit, nil
	}

	if ctx.Mode == BroadcastSync {
		res, err := ctx.BroadcastTxSync(txBytes)
		if err != nil {
			return nil, err
		}

		return &cTypes.ResultBroadcastTxCommit{
			CheckTx: abciTypes.ResponseCheckTx{
				Code: res.Code,
				Data: res.Data,
				Log:  res.Log,
			},
			Hash: res.Hash,
		}, nil
	}

	return ctx.BroadcastTxAndAwaitCommit(txBytes)
}

// BroadcastTxAndAwaitCommit broadcasts transaction bytes to a Tendermint node
// and waits for a commit.
func (ctx CLIContext) BroadcastTxAndAwaitCommit(tx []byte) (*cTypes.ResultBroadcastTxCommit, error) {
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
func (ctx CLIContext) BroadcastTxSync(tx []byte) (*cTypes.ResultBroadcastTx, error) {
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
func (ctx CLIContext) BroadcastTxAsync(tx []byte) (*cTypes.ResultBroadcastTx, error) {
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

func (ctx CLIContext) JSONResult(obj interface{}) ([]byte, error) {

	var bz []byte
	var err error

	if ctx.JSONIndent {
		bz, err = ctx.Codec.MarshalJSONIndent(obj, "", "  ")
	} else {
		bz, err = ctx.Codec.MarshalJSON(obj)
	}

	if err != nil {
		if ctx.JSONIndent {
			bz, err = json.MarshalIndent(obj, "", "  ")
		} else {
			bz, err = json.Marshal(obj)
		}
	}

	return bz, err
}

func (ctx CLIContext) PrintResult(obj interface{}) error {

	bz, err := ctx.JSONResult(obj)
	if err != nil {
		return err
	}

	resultFile := viper.GetString(types.FlagResultOutPut)
	appendMode := viper.GetBool(types.FlagResultOutPutAppend)

	if len(resultFile) == 0 {
		fmt.Println(string(bz))
		return nil
	}

	var file *os.File
	if appendMode {
		file, err = os.OpenFile(resultFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	} else {
		file, err = os.OpenFile(resultFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	}

	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(append(bz, []byte("\n")...))

	return err
}

func (ctx CLIContext) WithIndent(indent bool) CLIContext {
	ctx.JSONIndent = indent
	return ctx
}

func loadCliConfiguration() error {

	homeDir := viper.GetString(cli.HomeFlag)
	cfgFile := path.Join(homeDir, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		//override value
		// if tree, err := toml.LoadFile(cfgFile); err == nil {
		// 	for _, k := range tree.Keys() {
		// 		viper.Set(k, tree.Get(k))
		// 	}
		// }
	}

	return nil
}
