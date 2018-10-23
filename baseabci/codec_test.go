package baseabci

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

type I interface {
	Name() string
}

type X struct {
	N string
}

type Y struct {
	N string
}

func (y *Y) Name() string {
	return y.N
}

type Z struct {
	Xfield X
}

type M struct {
	IField I
}

func TestMakeQBaseCodec(t *testing.T) {

	cdc := MakeQBaseCodec()
	x := X{N: "1"}
	bz, _ := cdc.MarshalBinaryBare(x)

	var xPtr X
	cdc.UnmarshalBinaryBare(bz, &xPtr)

	require.Equal(t, x.N, xPtr.N)

	z := Z{Xfield: x}

	bz, _ = cdc.MarshalBinaryBare(z)

	var zPtr Z
	cdc.UnmarshalBinaryBare(bz, &zPtr)

	fmt.Println(zPtr)
	require.Equal(t, zPtr.Xfield.N, x.N)

	require.Panics(t, func() {
		y := &Y{N: "n"}
		m := &M{IField: y}
		cdc.MustMarshalBinaryBare(m)
	})

	cdc.RegisterInterface((*I)(nil), nil)
	cdc.RegisterConcrete(&Y{}, "test/codec/y", nil)

	require.NotPanics(t, func() {
		y := &Y{N: "n"}
		m := &M{IField: y}
		bz := cdc.MustMarshalBinaryBare(m)

		fmt.Println(bz)
	})

}
