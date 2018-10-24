package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBigInt_IsNil(t *testing.T) {
	bi := BigInt{}
	require.True(t, bi.IsNil())
}
