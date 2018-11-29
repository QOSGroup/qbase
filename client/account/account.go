package account

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QOSGroup/qbase/account"
	"github.com/QOSGroup/qbase/client/context"
	"github.com/QOSGroup/qbase/client/keys"
	"github.com/QOSGroup/qbase/types"

	"github.com/spf13/viper"
)

var (
	ErrAccountNotExsits = errors.New("account not exists")
)

func queryAccount(ctx context.CLIContext, addr []byte) (account.Account, error) {
	path := account.BuildAccountStoreQueryPath()
	res, err := ctx.Query(string(path), account.AddressStoreKey(addr))
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, ErrAccountNotExsits
	}

	var acc account.Account
	err = ctx.Codec.UnmarshalBinaryBare(res, &acc)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func GetAccount(ctx context.CLIContext, address []byte) (account.Account, error) {
	return queryAccount(ctx, address)
}

func GetAccountFromBech32Addr(ctx context.CLIContext, bech32Addr string) (account.Account, error) {

	addrBytes, err := types.GetAddrFromBech32(bech32Addr)

	if err != nil {
		return nil, fmt.Errorf("%s is not a valid bech32Addr", bech32Addr)
	}

	return queryAccount(ctx, addrBytes)
}

func GetAccountNonce(ctx context.CLIContext, address []byte) (int64, error) {
	account, err := queryAccount(ctx, address)

	if err == ErrAccountNotExsits {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return account.GetNonce(), nil
}

func IsAccountExists(ctx context.CLIContext, address []byte) bool {
	_, err := queryAccount(ctx, address)

	if err != nil {
		return false
	}

	return true
}

func GetAddrFromFlag(ctx context.CLIContext, flag string) (types.Address, error) {
	value := viper.GetString(flag)
	return GetAddrFromValue(ctx, value)
}

func GetAddrFromValue(ctx context.CLIContext, value string) (types.Address, error) {
	if strings.HasPrefix(value, types.PREF_ADD) {
		addr, err := types.GetAddrFromBech32(value)
		if err == nil {
			return addr, nil
		}
	}

	info, err := keys.GetKeyInfo(ctx, value)
	if err != nil {
		return nil, err
	}

	return info.GetAddress(), nil
}
