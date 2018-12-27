package types

import (
	"encoding/json"
	"math"
	"testing"

	"math/big"
	"math/rand"
)

func newIntegerFromString(s string) (*big.Int, bool) {
	return new(big.Int).SetString(s, 0)
}

func equal(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == 0 }

func gt(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == 1 }

func lt(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == -1 }

func add(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Add(i, i2) }

func sub(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Sub(i, i2) }

func mul(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mul(i, i2) }

func div(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Div(i, i2) }

func mod(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mod(i, i2) }

func neg(i *big.Int) *big.Int { return new(big.Int).Neg(i) }

func random(i *big.Int) *big.Int { return new(big.Int).Rand(rand.New(rand.NewSource(rand.Int63())), i) }

func min(i *big.Int, i2 *big.Int) *big.Int {
	if i.Cmp(i2) == 1 {
		return new(big.Int).Set(i2)
	}
	return new(big.Int).Set(i)
}

// MarshalAmino for custom encoding scheme
func marshalAmino(i *big.Int) (string, error) {
	bz, err := i.MarshalText()
	return string(bz), err
}

// UnmarshalAmino for custom decoding scheme
func unmarshalAmino(i *big.Int, text string) (err error) {
	return i.UnmarshalText([]byte(text))
}

// MarshalJSON for custom encoding scheme
// Must be encoded as a string for JSON precision
func marshalJSON(i *big.Int) ([]byte, error) {
	text, err := i.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(text))
}

// UnmarshalJSON for custom decoding scheme
// Must be encoded as a string for JSON precision
func unmarshalJSON(i *big.Int, bz []byte) error {
	var text string
	err := json.Unmarshal(bz, &text)
	if err != nil {
		return err
	}
	return i.UnmarshalText([]byte(text))
}

// BigInt wraps integer with 256 bit range bound
// Checks overflow, underflow and division by zero
// Exists in range from -(2^255-1) to 2^255-1
type BigInt struct {
	i *big.Int
}

// BigInt converts BigInt to big.BigInt
func (i BigInt) BigInt() *big.Int {
	return new(big.Int).Set(i.i)
}

// NewInt constructs BigInt from int64
func NewInt(n int64) BigInt {
	return BigInt{big.NewInt(n)}
}

// NewIntFromBigInt constructs BigInt from big.BigInt
func NewIntFromBigInt(i *big.Int) BigInt {
	if i.BitLen() > 255 {
		panic("NewIntFromBigInt() out of bound")
	}
	return BigInt{i}
}

// NewIntFromString constructs BigInt from string
func NewIntFromString(s string) (res BigInt, ok bool) {
	i, ok := newIntegerFromString(s)
	if !ok {
		return
	}
	// Check overflow
	if i.BitLen() > 255 {
		ok = false
		return
	}
	return BigInt{i}, true
}

// NewIntWithDecimal constructs BigInt with decimal
// Result value is n*10^dec
func NewIntWithDecimal(n int64, dec int) BigInt {
	if dec < 0 {
		panic("NewIntWithDecimal() decimal is negative")
	}
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
	i := new(big.Int)
	i.Mul(big.NewInt(n), exp)

	// Check overflow
	if i.BitLen() > 255 {
		panic("NewIntWithDecimal() out of bound")
	}
	return BigInt{i}
}

// ZeroInt returns BigInt value with zero
func ZeroInt() BigInt { return BigInt{big.NewInt(0)} }

// OneInt returns BigInt value with one
func OneInt() BigInt { return BigInt{big.NewInt(1)} }

// Int64 converts BigInt to int64
// Panics if the value is out of range
func (i BigInt) Int64() int64 {
	if !i.i.IsInt64() {
		panic("Int64() out of bound")
	}
	return i.i.Int64()
}

// IsInt64 returns true if Int64() not panics
func (i BigInt) IsInt64() bool {
	return i.i.IsInt64()
}

// 判断bigint中的实体数值是否为空
func (bi BigInt) IsNil() bool {
	return bi.i == nil
}

// BigInt nil值转换成0值
func (i BigInt) NilToZero() BigInt {
	if i.IsNil() {
		return ZeroInt()
	}
	return i
}

// IsZero returns true if BigInt is zero
func (i BigInt) IsZero() bool {
	return i.i.Sign() == 0
}

// Sign returns sign of BigInt
func (i BigInt) Sign() int {
	return i.i.Sign()
}

// Equal compares two Ints
func (i BigInt) Equal(i2 BigInt) bool {
	return equal(i.i, i2.i)
}

// GT returns true if first BigInt is greater than second
func (i BigInt) GT(i2 BigInt) bool {
	return gt(i.i, i2.i)
}

// LT returns true if first BigInt is lesser than second
func (i BigInt) LT(i2 BigInt) bool {
	return lt(i.i, i2.i)
}

// Add adds BigInt from another
func (i BigInt) Add(i2 BigInt) (res BigInt) {
	res = BigInt{add(i.i, i2.i)}
	// Check overflow
	if res.i.BitLen() > 255 {
		panic("BigInt overflow")
	}
	return
}

// AddRaw adds int64 to BigInt
func (i BigInt) AddRaw(i2 int64) BigInt {
	return i.Add(NewInt(i2))
}

// Sub subtracts BigInt from another
func (i BigInt) Sub(i2 BigInt) (res BigInt) {
	res = BigInt{sub(i.i, i2.i)}
	// Check overflow
	if res.i.BitLen() > 255 {
		panic("BigInt overflow")
	}
	return
}

// SubRaw subtracts int64 from BigInt
func (i BigInt) SubRaw(i2 int64) BigInt {
	return i.Sub(NewInt(i2))
}

// Mul multiples two Ints
func (i BigInt) Mul(i2 BigInt) (res BigInt) {
	// Check overflow
	if i.i.BitLen()+i2.i.BitLen()-1 > 255 {
		panic("BigInt overflow")
	}
	res = BigInt{mul(i.i, i2.i)}
	// Check overflow if sign of both are same
	if res.i.BitLen() > 255 {
		panic("BigInt overflow")
	}
	return
}

// MulRaw multipies BigInt and int64
func (i BigInt) MulRaw(i2 int64) BigInt {
	return i.Mul(NewInt(i2))
}

// Div divides BigInt with BigInt
func (i BigInt) Div(i2 BigInt) (res BigInt) {
	// Check division-by-zero
	if i2.i.Sign() == 0 {
		panic("Division by zero")
	}
	return BigInt{div(i.i, i2.i)}
}

// DivRaw divides BigInt with int64
func (i BigInt) DivRaw(i2 int64) BigInt {
	return i.Div(NewInt(i2))
}

// Mod returns remainder after dividing with BigInt
func (i BigInt) Mod(i2 BigInt) BigInt {
	if i2.Sign() == 0 {
		panic("division-by-zero")
	}
	return BigInt{mod(i.i, i2.i)}
}

// ModRaw returns remainder after dividing with int64
func (i BigInt) ModRaw(i2 int64) BigInt {
	return i.Mod(NewInt(i2))
}

// Neg negates BigInt
func (i BigInt) Neg() (res BigInt) {
	return BigInt{neg(i.i)}
}

// Return the minimum of the ints
func MinInt(i1, i2 BigInt) BigInt {
	return BigInt{min(i1.BigInt(), i2.BigInt())}
}

// Human readable string
func (i BigInt) String() string {
	return i.i.String()
}

// Testing purpose random BigInt generator
func randomInt(i BigInt) BigInt {
	return NewIntFromBigInt(random(i.BigInt()))
}

// MarshalAmino defines custom encoding scheme
func (i BigInt) MarshalAmino() (string, error) {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return marshalAmino(i.i)
}

// UnmarshalAmino defines custom decoding scheme
func (i *BigInt) UnmarshalAmino(text string) error {
	if i.i == nil { // Necessary since default BigInt initialization has i.i as nil
		i.i = new(big.Int)
	}
	return unmarshalAmino(i.i, text)
}

// MarshalJSON defines custom encoding scheme
func (i BigInt) MarshalJSON() ([]byte, error) {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return marshalJSON(i.i)
}

// UnmarshalJSON defines custom decoding scheme
func (i *BigInt) UnmarshalJSON(bz []byte) error {
	if i.i == nil { // Necessary since default BigInt initialization has i.i as nil
		i.i = new(big.Int)
	}
	return unmarshalJSON(i.i, bz)
}

// BigInt wraps integer with 256 bit range bound
// Checks overflow, underflow and division by zero
// Exists in range from 0 to 2^256-1
type Uint struct {
	i *big.Int
}

// BigInt converts Uint to big.Unt
func (i Uint) BigInt() *big.Int {
	return new(big.Int).Set(i.i)
}

// NewUint constructs Uint from int64
func NewUint(n uint64) Uint {
	i := new(big.Int)
	i.SetUint64(n)
	return Uint{i}
}

// NewUintFromBigUint constructs Uint from big.Uint
func NewUintFromBigInt(i *big.Int) Uint {
	// Check overflow
	if i.Sign() == -1 || i.Sign() == 1 && i.BitLen() > 256 {
		panic("Uint overflow")
	}
	return Uint{i}
}

// NewUintFromString constructs Uint from string
func NewUintFromString(s string) (res Uint, ok bool) {
	i, ok := newIntegerFromString(s)
	if !ok {
		return
	}
	// Check overflow
	if i.Sign() == -1 || i.Sign() == 1 && i.BitLen() > 256 {
		ok = false
		return
	}
	return Uint{i}, true
}

// NewUintWithDecimal constructs Uint with decimal
// Result value is n*10^dec
func NewUintWithDecimal(n uint64, dec int) Uint {
	if dec < 0 {
		panic("NewUintWithDecimal() decimal is negative")
	}
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
	i := new(big.Int)
	i.Mul(new(big.Int).SetUint64(n), exp)

	// Check overflow
	if i.Sign() == -1 || i.Sign() == 1 && i.BitLen() > 256 {
		panic("NewUintWithDecimal() out of bound")
	}
	return Uint{i}
}

// ZeroUint returns Uint value with zero
func ZeroUint() Uint { return Uint{big.NewInt(0)} }

// OneUint returns Uint value with one
func OneUint() Uint { return Uint{big.NewInt(1)} }

// Uint64 converts Uint to uint64
// Panics if the value is out of range
func (i Uint) Uint64() uint64 {
	if !i.i.IsUint64() {
		panic("Uint64() out of bound")
	}
	return i.i.Uint64()
}

// IsUint64 returns true if Uint64() not panics
func (i Uint) IsUint64() bool {
	return i.i.IsUint64()
}

func (i Uint) IsNil() bool {
	return i.i == nil
}

func (i Uint) NilToZero() Uint {
	if i.IsNil() {
		return ZeroUint()
	}
	return i
}

// IsZero returns true if Uint is zero
func (i Uint) IsZero() bool {
	return i.i.Sign() == 0
}

// Sign returns sign of Uint
func (i Uint) Sign() int {
	return i.i.Sign()
}

// Equal compares two Uints
func (i Uint) Equal(i2 Uint) bool {
	return equal(i.i, i2.i)
}

// GT returns true if first Uint is greater than second
func (i Uint) GT(i2 Uint) bool {
	return gt(i.i, i2.i)
}

// LT returns true if first Uint is lesser than second
func (i Uint) LT(i2 Uint) bool {
	return lt(i.i, i2.i)
}

// Add adds Uint from another
func (i Uint) Add(i2 Uint) (res Uint) {
	res = Uint{add(i.i, i2.i)}
	// Check overflow
	if res.Sign() == -1 || res.Sign() == 1 && res.i.BitLen() > 256 {
		panic("Uint overflow")
	}
	return
}

// AddRaw adds uint64 to Uint
func (i Uint) AddRaw(i2 uint64) Uint {
	return i.Add(NewUint(i2))
}

// Sub subtracts Uint from another
func (i Uint) Sub(i2 Uint) (res Uint) {
	res = Uint{sub(i.i, i2.i)}
	// Check overflow
	if res.Sign() == -1 || res.Sign() == 1 && res.i.BitLen() > 256 {
		panic("Uint overflow")
	}
	return
}

// SubRaw subtracts uint64 from Uint
func (i Uint) SubRaw(i2 uint64) Uint {
	return i.Sub(NewUint(i2))
}

// Mul multiples two Uints
func (i Uint) Mul(i2 Uint) (res Uint) {
	// Check overflow
	if i.i.BitLen()+i2.i.BitLen()-1 > 256 {
		panic("Uint overflow")
	}
	res = Uint{mul(i.i, i2.i)}
	// Check overflow
	if res.Sign() == -1 || res.Sign() == 1 && res.i.BitLen() > 256 {
		panic("Uint overflow")
	}
	return
}

// MulRaw multipies Uint and uint64
func (i Uint) MulRaw(i2 uint64) Uint {
	return i.Mul(NewUint(i2))
}

// Div divides Uint with Uint
func (i Uint) Div(i2 Uint) (res Uint) {
	// Check division-by-zero
	if i2.Sign() == 0 {
		panic("division-by-zero")
	}
	return Uint{div(i.i, i2.i)}
}

// Div divides Uint with uint64
func (i Uint) DivRaw(i2 uint64) Uint {
	return i.Div(NewUint(i2))
}

// Mod returns remainder after dividing with Uint
func (i Uint) Mod(i2 Uint) Uint {
	if i2.Sign() == 0 {
		panic("division-by-zero")
	}
	return Uint{mod(i.i, i2.i)}
}

// ModRaw returns remainder after dividing with uint64
func (i Uint) ModRaw(i2 uint64) Uint {
	return i.Mod(NewUint(i2))
}

// Return the minimum of the Uints
func MinUint(i1, i2 Uint) Uint {
	return Uint{min(i1.BigInt(), i2.BigInt())}
}

// Human readable string
func (i Uint) String() string {
	return i.i.String()
}

// Testing purpose random Uint generator
func randomUint(i Uint) Uint {
	return NewUintFromBigInt(random(i.BigInt()))
}

// MarshalAmino defines custom encoding scheme
func (i Uint) MarshalAmino() (string, error) {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return marshalAmino(i.i)
}

// UnmarshalAmino defines custom decoding scheme
func (i *Uint) UnmarshalAmino(text string) error {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return unmarshalAmino(i.i, text)
}

// MarshalJSON defines custom encoding scheme
func (i Uint) MarshalJSON() ([]byte, error) {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return marshalJSON(i.i)
}

// UnmarshalJSON defines custom decoding scheme
func (i *Uint) UnmarshalJSON(bz []byte) error {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return unmarshalJSON(i.i, bz)
}

//__________________________________________________________________________

// UintOverflow returns true if a given unsigned integer overflows and false
// otherwise.
func UintOverflow(x Uint) bool {
	return x.i.Sign() == -1 || x.i.Sign() == 1 && x.i.BitLen() > 256
}

// AddUint64Overflow performs the addition operation on two uint64 integers and
// returns a boolean on whether or not the result overflows.
func AddUint64Overflow(a, b uint64) (uint64, bool) {
	if math.MaxUint64-a < b {
		return 0, true
	}

	return a + b, false
}

// intended to be used with require/assert:  require.True(IntEq(...))
func IntEq(t *testing.T, exp, got BigInt) (*testing.T, bool, string, string, string) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
