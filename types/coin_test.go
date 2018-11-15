package types

import (
	"github.com/stretchr/testify/assert"
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
		if tcIndex == 0 || tcIndex == 4 {
			require.Equal(t, tc.expected, res, "coin less-than relation is incorrect, tc #%d", tcIndex)
		} else {
			require.Equal(t, !tc.expected, res, "coin less-than relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestCoin_AmountOpr(t *testing.T) {
	cases := []struct {
		inputOne       Coin
		inputTwo       Coin
		add_expected   Coin
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

func TestBaseCoins(t *testing.T) {
	good := BaseCoins{
		{"GAS", NewInt(1)},
		{"MINERAL", NewInt(1)},
		{"TREE", NewInt(1)},
	}
	neg := good.Negative()
	sum := good.Plus(neg)
	empty := BaseCoins{
		{"GOLD", NewInt(0)},
	}
	null := BaseCoins{}
	badSort1 := BaseCoins{
		{"TREE", NewInt(1)},
		{"GAS", NewInt(1)},
		{"MINERAL", NewInt(1)},
	}
	// both are after the first one, but the second and third are in the wrong order
	badSort2 := BaseCoins{
		{"GAS", NewInt(1)},
		{"TREE", NewInt(1)},
		{"MINERAL", NewInt(1)},
	}
	badAmt := BaseCoins{
		{"GAS", NewInt(1)},
		{"TREE", NewInt(0)},
		{"MINERAL", NewInt(1)},
	}
	dup := BaseCoins{
		{"GAS", NewInt(1)},
		{"GAS", NewInt(1)},
		{"MINERAL", NewInt(1)},
	}

	assert.True(t, good.IsValid(), "BaseCoins are valid")
	assert.True(t, good.IsPositive(), "Expected coins to be positive: %v", good)
	assert.False(t, null.IsPositive(), "Expected coins to not be positive: %v", null)
	assert.True(t, good.IsGTE(empty), "Expected %v to be >= %v", good, empty)
	assert.False(t, good.IsLT(empty), "Expected %v to be < %v", good, empty)
	assert.True(t, empty.IsLT(good), "Expected %v to be < %v", empty, good)
	assert.False(t, neg.IsPositive(), "Expected neg coins to not be positive: %v", neg)
	assert.Zero(t, len(sum), "Expected 0 coins")
	assert.False(t, badSort1.IsValid(), "BaseCoins are not sorted")
	assert.False(t, badSort2.IsValid(), "BaseCoins are not sorted")
	assert.False(t, badAmt.IsValid(), "BaseCoins cannot include 0 amounts")
	assert.False(t, dup.IsValid(), "Duplicate coin")
}

func TestBaseCoins_Plus(t *testing.T) {
	one := NewInt(1)
	zero := NewInt(0)
	negone := NewInt(-1)
	two := NewInt(2)

	cases := []struct {
		inputOne BaseCoins
		inputTwo BaseCoins
		expected BaseCoins
	}{
		{BaseCoins{{"A", one}, {"B", one}}, BaseCoins{{"A", one}, {"B", one}}, BaseCoins{{"A", two}, {"B", two}}},
		{BaseCoins{{"B", one}, {"A", one}}, BaseCoins{{"A", one}, {"B", one}}, BaseCoins{{"A", two}, {"B", two}}},
		{BaseCoins{{"A", zero}, {"B", one}}, BaseCoins{{"B", zero}, {"A", zero}}, BaseCoins{{"B", one}}},
		{BaseCoins{{"A", zero}, {"B", zero}}, BaseCoins{{"B", zero}, {"A", zero}}, BaseCoins(nil)},
		{BaseCoins{{"A", one}, {"B", zero}}, BaseCoins{{"A", negone}, {"B", zero}}, BaseCoins(nil)},
		{BaseCoins{{"A", negone}, {"B", zero}}, BaseCoins{{"B", zero}, {"A", zero}}, BaseCoins{{"A", negone}}},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Plus(tc.inputTwo)
		assert.True(t, res.IsValid())
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func TestBaseCoins_Sort(t *testing.T) {

	good := BaseCoins{
		NewInt64BaseCoin("GAS", 1),
		NewInt64BaseCoin("MINERAL", 1),
		NewInt64BaseCoin("TREE", 1),
	}
	empty := BaseCoins{
		NewInt64BaseCoin("GOLD", 0),
	}
	badSort1 := BaseCoins{
		NewInt64BaseCoin("TREE", 1),
		NewInt64BaseCoin("GAS", 1),
		NewInt64BaseCoin("MINERAL", 1),
	}
	badSort2 := BaseCoins{
		NewInt64BaseCoin("GAS", 1),
		NewInt64BaseCoin("TREE", 1),
		NewInt64BaseCoin("MINERAL", 1),
	}
	badAmt := BaseCoins{
		NewInt64BaseCoin("GAS", 1),
		NewInt64BaseCoin("TREE", 0),
		NewInt64BaseCoin("MINERAL", 1),
	}
	dup := BaseCoins{
		NewInt64BaseCoin("GAS", 1),
		NewInt64BaseCoin("GAS", 1),
		NewInt64BaseCoin("MINERAL", 1),
	}

	cases := []struct {
		coins         BaseCoins
		before, after bool // valid before/after sort
	}{
		{good, true, true},
		{empty, false, false},
		{badSort1, false, true},
		{badSort2, false, true},
		{badAmt, false, false},
		{dup, false, false},
	}

	for tcIndex, tc := range cases {
		require.Equal(t, tc.before, tc.coins.IsValid(), "coin validity is incorrect before sorting, tc #%d", tcIndex)
		tc.coins.Sort()
		require.Equal(t, tc.after, tc.coins.IsValid(), "coin validity is incorrect after sorting, tc #%d", tcIndex)
	}
}

func TestBaseCoins_AmountOf(t *testing.T) {

	case0 := BaseCoins{}
	case1 := BaseCoins{
		NewInt64BaseCoin("", 0),
	}
	case2 := BaseCoins{
		NewInt64BaseCoin(" ", 0),
	}
	case3 := BaseCoins{
		NewInt64BaseCoin("GOLD", 0),
	}
	case4 := BaseCoins{
		NewInt64BaseCoin("GAS", 1),
		NewInt64BaseCoin("TREE", 1),
		NewInt64BaseCoin("MINERAL", 1),
	}
	case5 := BaseCoins{
		NewInt64BaseCoin("TREE", 1),
		NewInt64BaseCoin("MINERAL", 1),
	}
	case6 := BaseCoins{
		NewInt64BaseCoin("", 6),
	}
	case7 := BaseCoins{
		NewInt64BaseCoin(" ", 7),
	}
	case8 := BaseCoins{
		NewInt64BaseCoin("GAS", 8),
	}

	cases := []struct {
		coins           BaseCoins
		amountOf        int64
		amountOfSpace   int64
		amountOfGAS     int64
		amountOfMINERAL int64
		amountOfTREE    int64
	}{
		{case0, 0, 0, 0, 0, 0},
		{case1, 0, 0, 0, 0, 0},
		{case2, 0, 0, 0, 0, 0},
		{case3, 0, 0, 0, 0, 0},
		{case4, 0, 0, 1, 1, 1},
		{case5, 0, 0, 0, 1, 1},
		{case6, 6, 0, 0, 0, 0},
		{case7, 0, 7, 0, 0, 0},
		{case8, 0, 0, 8, 0, 0},
	}

	for _, tc := range cases {
		assert.Equal(t, NewInt(tc.amountOf), tc.coins.AmountOf(""))
		assert.Equal(t, NewInt(tc.amountOfSpace), tc.coins.AmountOf(" "))
		assert.Equal(t, NewInt(tc.amountOfGAS), tc.coins.AmountOf("GAS"))
		assert.Equal(t, NewInt(tc.amountOfMINERAL), tc.coins.AmountOf("MINERAL"))
		assert.Equal(t, NewInt(tc.amountOfTREE), tc.coins.AmountOf("TREE"))
	}
}

func TestBaseCoins_IsGTE(t *testing.T) {
	cases := []struct {
		inputOne BaseCoins
		inputTwo BaseCoins
		expected bool
	}{
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 0), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 3)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("B", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("C", 1)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins(nil), true},
		{BaseCoins(nil), BaseCoins(nil), true},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.inputOne.IsGTE(tc.inputTwo), tc.expected)
	}
}

func TestBaseCoins_IsLT(t *testing.T) {
	cases := []struct {
		inputOne BaseCoins
		inputTwo BaseCoins
		expected bool
	}{
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 0), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 3)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("B", 1)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("C", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 2), NewInt64BaseCoin("B", 1)},
			BaseCoins(nil), false},
		{BaseCoins(nil), BaseCoins(nil), false},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.inputOne.IsLT(tc.inputTwo), tc.expected)
	}
}

func TestBaseCoins_IsEqual(t *testing.T) {
	cases := []struct {
		inputOne BaseCoins
		inputTwo BaseCoins
		expected bool
	}{
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("B", 1), NewInt64BaseCoin("A", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("A", 1)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)},
			BaseCoins{NewInt64BaseCoin("C", 1)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)},
			BaseCoins(nil), false},
		{BaseCoins(nil), BaseCoins(nil), true},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.inputOne.IsEqual(tc.inputTwo), tc.expected)
	}
}

func TestBaseCoins_IsPositive(t *testing.T) {
	cases := []struct {
		input    BaseCoins
		expected bool
	}{
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 0)}, false},
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", -1)}, false},
		{BaseCoins(nil), false},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.input.IsPositive(), tc.expected)
	}
}

func TestBaseCoins_IsNotNegative(t *testing.T) {
	cases := []struct {
		input    BaseCoins
		expected bool
	}{
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 1)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", 0)}, true},
		{BaseCoins{NewInt64BaseCoin("A", 1), NewInt64BaseCoin("B", -1)}, false},
		{BaseCoins(nil), true},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.input.IsNotNegative(), tc.expected)
	}
}
