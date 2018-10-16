package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBool2Byte(t *testing.T) {
	bytrue := Bool2Byte(true)
	byfalse := Bool2Byte(false)
	require.NotNil(t, bytrue)
	require.NotNil(t, byfalse)
	require.NotEqual(t, bytrue, byfalse)
}

func TestInt2Byte(t *testing.T) {
	bysmall := Int2Byte(23)
	bylarge := Int2Byte(2345678910)
	require.NotNil(t, bysmall)
	require.NotNil(t, bylarge)
}

func TestCheckQsc(t *testing.T) {
	strOk := "qos34"
	strErr := "qsc$df3"
	strErr1 := "qsc34567891"

	bok := CheckQscName(strOk)
	berr := CheckQscName(strErr)
	berr1 := CheckQscName(strErr1)

	require.Equal(t, bok, true)
	require.Equal(t, berr, false)
	require.Equal(t, berr1, false)
}
