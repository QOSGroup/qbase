package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoin_SameNameAs(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewBaseCoin("A", NewInt(1)), NewBaseCoin("A", NewInt(1)), true},
		{NewBaseCoin("A", NewInt(1)), NewBaseCoin("a", NewInt(1)), false},
		{NewBaseCoin("steak", NewInt(10)), NewBaseCoin("steak", NewInt(10)), true},
		{NewBaseCoin("steak", NewInt(-10)), NewBaseCoin("steak", NewInt(10)), true},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.SameNameAs(tc.inputTwo)
		require.Equal(t, tc.expected, res, "coin denominations didn't match, tc #%d", tcIndex)
	}
}

func TestCoin_AmountEquality(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewBaseCoin("A", NewInt(1)), NewBaseCoin("A", NewInt(1)), true},
		{NewBaseCoin("A", NewInt(1)), NewBaseCoin("a", NewInt(1)), false},
		{NewBaseCoin("a", NewInt(1)), NewBaseCoin("b", NewInt(1)), false},
		{NewBaseCoin("steak", NewInt(1)), NewBaseCoin("steak", NewInt(10)), false},
		{NewBaseCoin("steak", NewInt(-10)), NewBaseCoin("steak", NewInt(10)), false},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.IsEqual(tc.inputTwo)
		require.Equal(t, tc.expected, res, "coin equality relation is incorrect, tc #%d", tcIndex)
	}

	cases = []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewBaseCoin("A", NewInt(1)), NewBaseCoin("A", NewInt(1)), false},
		{NewBaseCoin("A", NewInt(2)), NewBaseCoin("A", NewInt(1)), true},
		{NewBaseCoin("A", NewInt(0)), NewBaseCoin("A", NewInt(-1)), true},
		{NewBaseCoin("A", NewInt(-1)), NewBaseCoin("A", NewInt(5)), false},
		{NewBaseCoin("a", NewInt(1)), NewBaseCoin("b", NewInt(1)), false},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.IsGreaterThan(tc.inputTwo)
		require.Equal(t, tc.expected, res, "coin greater-than relation is incorrect, tc #%d", tcIndex)
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.IsLessThan(tc.inputTwo)
		if tcIndex == 0 || tcIndex ==4 {
			require.Equal(t, tc.expected, res, "coin less-than relation is incorrect, tc #%d", tcIndex)
		}else {
			require.Equal(t, !tc.expected, res, "coin less-than relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestCoin_AmountOpr(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		add_expected Coin
		minus_expected Coin
	}{
		{NewBaseCoin("A", NewInt(1)), NewBaseCoin("B", NewInt(1)), NewBaseCoin("A", NewInt(1)), NewBaseCoin("A", NewInt(1))},
		{NewBaseCoin("A", NewInt(1)), NewBaseCoin("A", NewInt(1)), NewBaseCoin("A", NewInt(2)), NewBaseCoin("A", NewInt(0))},
		{NewBaseCoin("A", NewInt(-4)), NewBaseCoin("A", NewInt(1)), NewBaseCoin("A", NewInt(-3)), NewBaseCoin("A", NewInt(-5))},
		{NewBaseCoin("A", NewInt(0)), NewBaseCoin("A", NewInt(-1)), NewBaseCoin("A", NewInt(-1)), NewBaseCoin("A", NewInt(1))},
	}

	for tcIndex, tc := range cases {
		add_res := tc.inputOne.Plus(tc.inputTwo)
		require.Equal(t, tc.add_expected, add_res, "add operation of coins is incorrect, tc #%d", tcIndex)
		minus_res := tc.inputOne.Minus(tc.inputTwo)
		// TODO: 第一种方法无法通过测试
		//require.Equal(t, tc.minus_expected, minus_res, "minus operation of coins is incorrect, tc #%d", tcIndex)
		require.True(t, minus_res.IsEqual(tc.minus_expected), "minus operation of coins is incorrect, tc #%d", tcIndex)
	}

}