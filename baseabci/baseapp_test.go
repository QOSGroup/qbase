package baseabci

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/context"
	"github.com/QOSGroup/qbase/store"
	"github.com/QOSGroup/qbase/txs"
	"github.com/QOSGroup/qbase/types"
	"github.com/stretchr/testify/require"

	go_amino "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

const cid = "test"

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "base/app")
}

func mockApp() *BaseApp {

	logger := defaultLogger()
	db := dbm.NewMemDB()

	app := NewBaseApp("test", logger, db, func(cdc *go_amino.Codec) {
		cdc.RegisterConcrete(&transferTx{}, "baseapp/test/transferTx", nil)
		cdc.RegisterConcrete(&testAccount{}, "baseapp/test/testAccount", nil)
	}, SetPruning(store.PruneSyncable))

	app.RegisterAccountProto(func() account.Account {
		baseAccount := &account.BaseAccount{}
		t := &testAccount{}
		t.BaseAccount = baseAccount
		return t
	})

	app.SetInitChainer(func(ctx context.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		for i := int64(0); i < 10; i++ {
			createAccount(i, int64(5500), ctx)
		}

		//测试： 使用签名者的公钥作为trustKey. 正式环境不能这么用
		if app.txQcpSigner != nil {
			qcpMapper := GetQcpMapper(ctx)
			qcpMapper.SetChainInTrustPubKey(cid, app.txQcpSigner.PubKey())
		}

		return abci.ResponseInitChain{}
	})

	return app
}

func TestLoadVersion(t *testing.T) {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	name := t.Name()
	app := NewBaseApp(name, logger, db, nil)
	app.cms.SetPruning(store.PruneSyncable)

	capKey := types.NewKVStoreKey("main")
	app.mountStoresIAVL(capKey)
	err := app.LoadLatestVersion()
	require.Nil(t, err)

	emptyCommitID := store.CommitID{}

	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID1 := store.CommitID{1, res.Data}

	header = abci.Header{Height: 2}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res = app.Commit()
	commitID2 := store.CommitID{2, res.Data}

	app = NewBaseApp(name, logger, db, nil)
	app.mountStoresIAVL(capKey)
	err = app.LoadLatestVersion()
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(2), commitID2)

	app = NewBaseApp(name, logger, db, nil)
	app.mountStoresIAVL(capKey)
	err = app.LoadVersion(1)
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(1), commitID1)
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()
	testLoadVersionHelper(t, app, int64(2), commitID2)
}

func testLoadVersionHelper(t *testing.T, app *BaseApp, expectedHeight int64, expectedID store.CommitID) {
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

func TestInitChainer(t *testing.T) {
	name := t.Name()
	// keep the db and logger ourselves so
	// we can reload the same  app later
	db := dbm.NewMemDB()
	logger := defaultLogger()
	app := NewBaseApp(name, logger, db, nil)
	capKey := types.NewKVStoreKey("main")
	capKey2 := types.NewKVStoreKey("key2")
	app.mountStoresIAVL(capKey, capKey2)

	// set a value in the store on init chain
	key, value := []byte("hello"), []byte("goodbye")
	var initChainer InitChainHandler = func(ctx context.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		store := ctx.KVStore(capKey)
		store.Set(key, value)
		return abci.ResponseInitChain{}
	}

	query := abci.RequestQuery{
		Path: "/store/main/key",
		Data: key,
	}

	// initChainer is nil - nothing happens
	app.InitChain(abci.RequestInitChain{ChainId: cid})
	res := app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// set initChainer and try again - should see the value
	app.SetInitChainer(initChainer)

	// stores are mounted and private members are set - sealing baseapp
	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)

	app.InitChain(abci.RequestInitChain{AppStateBytes: []byte("{}"), ChainId: "test-chain-id"}) // must have valid JSON genesis file, even if empty

	// assert that chainID is set correctly in InitChain
	chainID := app.deliverState.ctx.ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in deliverState not set correctly in InitChain")

	chainID = app.checkState.ctx.ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in checkState not set correctly in InitChain")

	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)

	// reload app
	app = NewBaseApp(name, logger, db, nil)
	app.SetInitChainer(initChainer)
	app.mountStoresIAVL(capKey, capKey2)
	err = app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)

	// ensure we can still query after reloading
	res = app.Query(query)
	require.Equal(t, value, res.Value)

	// commit and ensure we can still query
	app.BeginBlock(abci.RequestBeginBlock{})
	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

func TestTxQcpResult(t *testing.T) {

	app := mockApp()

	app.RegisterTxQcpResultHandler(func(ctx context.Context, txQcpResult interface{}) {
		qcpResult, _ := txQcpResult.(*txs.QcpTxResult)
		fmt.Println(qcpResult)
		return
	})

	signer := ed25519.GenPrivKey()
	app.RegisterTxQcpSigner(signer)

	app.LoadLatestVersion()

	//init chain
	app.InitChain(abci.RequestInitChain{ChainId: cid})
	app.Commit()

	var txQcpBytes [][]byte

	for i := uint64(1); i < 10; i++ {

		var code types.CodeType
		seed := rand.Int63n(10)

		if seed > int64(5) {
			code = types.CodeOK
		} else {
			code = types.CodeInternal
		}

		qcpResult := &txs.QcpTxResult{
			Result: types.Result{
				Code:    code,
				GasUsed: types.OneUint().Uint64(),
				Tags: types.Tags{
					types.MakeTag("key", "value"),
				},
			},
			QcpOriginalSequence: int64(i),
		}

		stdTx := txs.NewTxStd(qcpResult, cid, types.NewInt(10000))
		txQcp := txs.NewTxQCP(stdTx, cid, cid, int64(i), 2, 0, true, "")

		signature, _ := txQcp.SignTx(signer)
		txQcp.Sig.Pubkey = signer.PubKey()
		txQcp.Sig.Signature = signature

		txQcpBytes = append(txQcpBytes, app.GetCdc().MustMarshalBinaryBare(txQcp))
	}

	for _, txQcpByte := range txQcpBytes {
		res := app.CheckTx(txQcpByte)
		require.Equal(t, int64(0), int64(res.Code))
	}

	header := abci.Header{Height: 2, ChainID: cid}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	for _, txQcpByte := range txQcpBytes {
		app.DeliverTx(txQcpByte)
	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// -------------------------------------
	inKey := []byte(fmt.Sprintf("sequence/in/%s", cid))
	query := abci.RequestQuery{
		Path:   "/store/qcp/key",
		Data:   inKey,
		Height: 2,
	}

	res := app.Query(query)
	var seq int64
	app.GetCdc().UnmarshalBinaryBare(res.GetValue(), &seq)
	require.Equal(t, int64(9), seq)

}

func TestTxQcp(t *testing.T) {

	app := mockApp()

	signer := ed25519.GenPrivKey()
	app.RegisterTxQcpSigner(signer)

	app.LoadLatestVersion()

	//init chain
	app.InitChain(abci.RequestInitChain{ChainId: cid})
	app.Commit()

	//查询出两个账户进行转账操作
	newCtx := app.NewContext(true, abci.Header{})
	accMapper := GetAccountMapper(newCtx)

	pidAccount1 := getAccount(accMapper, int64(1))
	pidAccount2 := getAccount(accMapper, int64(2))

	var txQcpBytes [][]byte

	for i := int64(1); i < 10; i++ {

		acc := accMapper.GetAccount(pidAccount1.GetAddress())
		acc.SetNonce(i)
		txstd := createTransformTxWithNoQcpTx(acc, pidAccount2, 1000)
		txQcp := txs.NewTxQCP(txstd, cid, cid, int64(i), i, 0, false, "")

		signature, _ := txQcp.SignTx(signer)
		txQcp.Sig.Pubkey = signer.PubKey()
		txQcp.Sig.Signature = signature

		txQcpBytes = append(txQcpBytes, app.GetCdc().MustMarshalBinaryBare(txQcp))
	}

	for _, txQcpByte := range txQcpBytes {
		res := app.CheckTx(txQcpByte)
		require.Equal(t, int64(0), int64(res.Code))
	}

	header := abci.Header{Height: 2, ChainID: cid}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	for _, txQcpByte := range txQcpBytes {
		app.DeliverTx(txQcpByte)
	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// -------------------------------------
	inKey := []byte(fmt.Sprintf("sequence/in/%s", cid))
	query := abci.RequestQuery{
		Path:   "/store/qcp/key",
		Data:   inKey,
		Height: 2,
	}

	res := app.Query(query)
	var seq int64
	app.GetCdc().UnmarshalBinaryBare(res.GetValue(), &seq)
	require.Equal(t, int64(9), seq)

	// -------------------------------------

	outKey := []byte(fmt.Sprintf("sequence/out/%s", cid))
	query = abci.RequestQuery{
		Path:   "/store/qcp/key",
		Data:   outKey,
		Height: 2,
	}

	res = app.Query(query)
	app.GetCdc().UnmarshalBinaryBare(res.GetValue(), &seq)
	require.Equal(t, int64(9), seq)

	// -------------------------------------

	for i := int64(1); i <= seq; i++ {
		key := []byte(fmt.Sprintf("tx/out/%s/%d", cid, i))
		query := abci.RequestQuery{
			Path:   "/store/qcp/key",
			Data:   key,
			Height: 2,
		}

		res = app.Query(query)

		var tx types.Tx
		app.GetCdc().UnmarshalBinaryBare(res.GetValue(), &tx)

		outQcpTx := tx.(*txs.TxQcp)

		require.Equal(t, true, outQcpTx.IsResult)

		qcpResult, _ := outQcpTx.TxStd.ITxs[0].(*txs.QcpTxResult)

		if i > 5 {
			require.NotEqual(t, int64(0), int64(qcpResult.Result.Code))
		} else {
			require.Equal(t, int64(0), int64(qcpResult.Result.Code))
		}

	}

}

func TestCrossStdTx(t *testing.T) {

	app := mockApp()

	signer := ed25519.GenPrivKey()
	app.RegisterTxQcpSigner(signer)

	app.LoadLatestVersion()

	//init chain
	app.InitChain(abci.RequestInitChain{ChainId: cid})
	app.Commit()

	checkContext := app.checkState.ctx
	accMapper := GetAccountMapper(checkContext)

	pidAccount3 := getAccount(accMapper, int64(3))
	pidAccount4 := getAccount(accMapper, int64(4))

	var txss [][]byte
	//创建转账stdTx: pid3 转账 pid4
	for i := int64(1); i < 10; i++ {

		acc := accMapper.GetAccount(pidAccount3.GetAddress())
		require.Equal(t, i-1, acc.GetNonce())

		acc.SetNonce(i)

		stdTx := createTransformTxWithNoQcpTx(acc, pidAccount4, 1000)
		stdTxBz, _ := app.GetCdc().MarshalBinaryBare(stdTx)

		res := app.CheckTx(stdTxBz)
		require.Equal(t, uint32(0), res.Code)

		txss = append(txss, stdTxBz)

		acc = accMapper.GetAccount(pidAccount3.GetAddress())
		require.Equal(t, i, acc.GetNonce())

	}

	header := abci.Header{Height: 2, ChainID: cid}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	for i, stdTxBz := range txss {
		res := app.DeliverTx(stdTxBz)

		for _, p := range res.GetTags() {
			fmt.Println(string(p.GetKey()), string(p.GetValue()))
		}

		if i >= 5 {
			require.NotEqual(t, uint32(0), uint32(res.Code))
		} else {
			require.Equal(t, uint32(0), uint32(res.Code))
		}

	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	k1 := []byte(fmt.Sprintf("sequence/out/%s", cid))
	queryHeight1 := abci.RequestQuery{
		Path:   "/store/qcp/key",
		Data:   k1,
		Height: 2,
	}

	res := app.Query(queryHeight1)
	var seq int64
	app.GetCdc().UnmarshalBinaryBare(res.GetValue(), &seq)
	require.Equal(t, int64(5), seq)

	for i := int64(1); i <= seq; i++ {
		k2 := []byte(fmt.Sprintf("tx/out/%s/%d", cid, i))
		queryHeight2 := abci.RequestQuery{
			Path:   "/store/qcp/key",
			Data:   k2,
			Height: 2,
		}

		res = app.Query(queryHeight2)

		var tx types.Tx
		app.GetCdc().UnmarshalBinaryBare(res.GetValue(), &tx)

		outQcpTx := tx.(*txs.TxQcp)

		require.Equal(t, i, outQcpTx.Sequence)
		require.Equal(t, cid, outQcpTx.From)
		require.Equal(t, i-1, outQcpTx.TxIndex)
	}

}

func TestStdTx(t *testing.T) {

	app := mockApp()
	app.LoadLatestVersion()

	//init chain
	app.InitChain(abci.RequestInitChain{ChainId: cid})
	app.Commit()

	checkContext := app.checkState.ctx
	accMapper := GetAccountMapper(checkContext)

	pidAccount1 := getAccount(accMapper, int64(1))
	pidAccount2 := getAccount(accMapper, int64(2))

	var txs [][]byte
	//创建转账stdTx: pid1 转账 pid2
	for i := int64(1); i < 10; i++ {

		acc := accMapper.GetAccount(pidAccount1.GetAddress())
		require.Equal(t, i-1, acc.GetNonce())

		acc.SetNonce(i)
		stdTx := createTransformTxWithNoQcpTx(acc, pidAccount2, 1000)
		stdTxBz, _ := app.GetCdc().MarshalBinaryBare(stdTx)

		res := app.CheckTx(stdTxBz)
		require.Equal(t, uint32(0), res.Code)

		txs = append(txs, stdTxBz)

		acc = accMapper.GetAccount(pidAccount1.GetAddress())
		require.Equal(t, i, acc.GetNonce())

	}

	header := abci.Header{Height: 2, ChainID: cid}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	for i, stdTxBz := range txs {
		res := app.DeliverTx(stdTxBz)

		if i >= 5 {
			require.NotEqual(t, uint32(0), uint32(res.Code))
		} else {
			require.Equal(t, uint32(0), uint32(res.Code))
		}

	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	k1 := account.AddressStoreKey(pidAccount1.GetAddress())
	queryHeight1 := abci.RequestQuery{
		Path:   "/store/acc/key",
		Data:   k1,
		Height: 1,
	}

	res := app.Query(queryHeight1)
	var acc1 testAccount
	app.GetCdc().UnmarshalBinaryBare(res.GetValue(), &acc1)
	require.Equal(t, int64(5500), acc1.Money)

	queryHeight2 := abci.RequestQuery{
		Path:   "/store/acc/key",
		Data:   k1,
		Height: 2,
	}
	res = app.Query(queryHeight2)
	var acc2 testAccount
	app.GetCdc().UnmarshalBinaryBare(res.GetValue(), &acc2)
	require.Equal(t, int64(500), acc2.Money)

}

func createTransformTxWithNoQcpTx(from, to account.Account, amount int64) *txs.TxStd {
	tx := &transferTx{
		FromUsers: []types.Address{from.GetAddress()},
		ToUsers:   []types.Address{to.GetAddress()},
		Amount:    amount,
	}

	stdTx := txs.NewTxStd(tx, cid, types.NewInt(50000))

	signerAccount, _ := from.(*testAccount)

	signature, err := stdTx.SignTx(signerAccount.PrivKey, from.GetNonce(), "", cid)
	if err != nil {
		panic("signer error")
	}

	stdTx.Signature = []txs.Signature{{
		Pubkey:    signerAccount.PrivKey.PubKey(),
		Signature: signature,
		Nonce:     from.GetNonce(),
	}}

	return stdTx
}

func createAccount(id int64, money int64, ctx context.Context) account.Account {
	privkey := ed25519.GenPrivKey()
	accMapper := GetAccountMapper(ctx)

	pubkeyAddress := types.Address(privkey.PubKey().Address())

	account := accMapper.NewAccountWithAddress(pubkeyAddress)

	testAccount, _ := account.(*testAccount)

	testAccount.SetAddress(pubkeyAddress)
	testAccount.SetPublicKey(privkey.PubKey())
	testAccount.PrivKey = privkey
	testAccount.Id = id
	testAccount.Money = money

	accMapper.SetAccount(testAccount)

	return account
}

func getAccount(accMapper *account.AccountMapper, id int64) *testAccount {
	var iAcc account.Account
	accMapper.IterateAccounts(func(acc account.Account) bool {
		t, _ := acc.(*testAccount)
		if t.Id == id {
			iAcc = acc
			return true
		}
		return false
	})

	if iAcc == nil {
		return nil
	}

	t, ok := iAcc.(*testAccount)
	if !ok {
		return nil
	}

	return t
}

//------------------------------------------------------------------------------------------------------

type testAccount struct {
	*account.BaseAccount
	Id      int64
	Money   int64
	PrivKey crypto.PrivKey
}

// const mapperName = "transferTxMapper"
// const storeKey = "transferTxMapperStoreKey"

// type transferTxMapper struct {
// 	*mapper.BaseMapper
// }

// func newTransferMapper() *transferTxMapper {
// 	var transferTxMapper = transferTxMapper{}
// 	transferTxMapper.BaseMapper = mapper.NewBaseMapper(store.NewKVStoreKey(storeKey))
// 	return &transferTxMapper
// }

// func (mapper *transferTxMapper) Copy() mapper.IMapper {
// 	cpyMapper := &transferTxMapper{}
// 	cpyMapper.BaseMapper = mapper.BaseMapper.Copy()
// 	return cpyMapper
// }

// func (mapper *transferTxMapper) Name() string {
// 	return mapperName
// }

type transferTx struct {
	Amount    int64
	FromUsers []types.Address
	ToUsers   []types.Address
}

var _ txs.ITx = (*transferTx)(nil)

func (t *transferTx) ValidateData(ctx context.Context) error {
	if t.Amount < 0 || len(t.FromUsers) < 1 || len(t.ToUsers) < 1 {
		return errors.New("transferTx ValidateBasicData error")
	}
	return nil
}

//TODO: 简单实现，只实现单用户对单用户转账
func (t *transferTx) Exec(ctx context.Context) (result types.Result, crossTxQcps *txs.TxQcp) {

	accMapper := GetAccountMapper(ctx)

	from, _ := accMapper.GetAccount(t.FromUsers[0]).(*testAccount)
	to, _ := accMapper.GetAccount(t.ToUsers[0]).(*testAccount)

	if from.Money < t.Amount {
		result = types.ErrInternal(fmt.Sprintf("user %s not much money. expect: %d . actual: %d", from.GetAddress(), t.Amount, from.Money)).Result()
		return
	}

	from.Money = from.Money - t.Amount
	to.Money = to.Money + t.Amount

	accMapper.SetAccount(from)
	accMapper.SetAccount(to)

	//from.id == 3  && to.id == 4时，生成跨链结果
	if from.Id == int64(3) && to.Id == int64(4) {

		crossTxQcps = &txs.TxQcp{}

		//仅测试用
		crossTxQcps.To = cid

		ttx := &transferTx{
			Amount:    t.Amount,
			FromUsers: []types.Address{from.AccountAddress},
			ToUsers:   []types.Address{to.AccountAddress},
		}

		crossTxQcps.TxStd = txs.NewTxStd(ttx, crossTxQcps.To, types.OneInt())
		txStdSig, _ := crossTxQcps.TxStd.SignTx(from.PrivKey, int64(from.Nonce), cid, cid)

		crossTxQcps.TxStd.Signature = []txs.Signature{
			txs.Signature{
				Pubkey:    from.PrivKey.PubKey(),
				Signature: txStdSig,
				Nonce:     int64(from.Nonce),
			},
		}

	}

	return
}

func (t *transferTx) GetSigner() []types.Address {
	return t.FromUsers
}

func (t *transferTx) CalcGas() types.BigInt {
	return types.ZeroInt()
}

func (t *transferTx) GetGasPayer() types.Address {
	return t.FromUsers[0]
}

func (t *transferTx) GetSignData() []byte {
	signData := make([]byte, 100)
	signData = append(signData, types.Int2Byte(t.Amount)...)

	for _, addr := range t.FromUsers {
		signData = append(signData, []byte(addr)...)
	}

	for _, addr := range t.ToUsers {
		signData = append(signData, []byte(addr)...)
	}

	return signData
}

func TestInfo(t *testing.T) {
	app := mockApp()
	app.SetName(t.Name())
	reqInfo := abci.RequestInfo{}
	res := app.Info(reqInfo)

	require.Equal(t, "", res.Version)
	require.Equal(t, t.Name(), res.GetData())
	require.Equal(t, int64(0), res.LastBlockHeight)
	require.Equal(t, []uint8(nil), res.LastBlockAppHash)

}

func TestDeferFuncArgs(t *testing.T) {

	flag := false

	defer func(f bool) {

		if r := recover(); r != nil {
			fmt.Println(r)
		}

		fmt.Println(f)
		fmt.Println(flag)
	}(flag)

	// var b []byte
	// b[1] = byte(1)

	flag = true
	var b []byte
	b[1] = byte(1)

}
