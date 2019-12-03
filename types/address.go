package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"strings"
	"sync"

	"github.com/tendermint/tendermint/crypto"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/libs/bech32"
)

const (
	AddrLen = 20

	DefaultAccountPrefix            = "qosacc"
	DefaultAccountPublicKeyPrefix   = "qosaccpub"
	DefaultValidatorPrefix          = "qosval"
	DefaultValidatorPublicKeyPrefix = "qosvalpub"
	DefaultConsensusPrefix          = "qoscons"
	DefaultConsensusPubKeyPrefix    = "qosconspub"
)

type AddressConfig struct {
	mtx                 sync.RWMutex
	sealed              bool
	bech32AddressPrefix map[string]string
	coinType            uint32
}

type Address interface {
	Equals(Address) bool
	Empty() bool
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
	Bytes() []byte
	String() string
	Format(s fmt.State, verb rune)
}

var _config = &AddressConfig{
	sealed: false,
	bech32AddressPrefix: map[string]string{
		"account_addr":   DefaultAccountPrefix,
		"validator_addr": DefaultValidatorPrefix,
		"consensus_addr": DefaultConsensusPrefix,
		"account_pub":    DefaultAccountPublicKeyPrefix,
		"validator_pub":  DefaultValidatorPublicKeyPrefix,
		"consensus_pub":  DefaultConsensusPubKeyPrefix,
	},
	coinType: uint32(389),
}

func GetAddressConfig() *AddressConfig {
	return _config
}

var _ Address = AccAddress{}
var _ Address = ValAddress{}
var _ Address = ConsAddress{}

// AccAddress: common user address
type AccAddress []byte

// validator address
type ValAddress []byte

// consensus address
type ConsAddress []byte

func AccAddressFromHex(address string) (addr AccAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("address is empty")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return AccAddress(bz), nil
}

func AccAddressFromBech32(address string) (addr AccAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return AccAddress{}, nil
	}

	prefix := GetAddressConfig().GetBech32AccountAddrPrefix()
	bz, err := GetFromBech32(address, prefix)
	if err != nil {
		return nil, err
	}

	err = verifyAddress(bz)
	if err != nil {
		return nil, err
	}

	return AccAddress(bz), nil
}

// Returns boolean for whether two AccAddresses are Equal
func (aa AccAddress) Equals(aa2 Address) bool {
	if aa.Empty() && aa2.Empty() {
		return true
	}

	return bytes.Equal(aa.Bytes(), aa2.Bytes())
}

// Returns boolean for whether an AccAddress is empty
func (aa AccAddress) Empty() bool {
	if aa == nil {
		return true
	}

	aa2 := AccAddress{}
	return bytes.Equal(aa.Bytes(), aa2.Bytes())
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (aa AccAddress) Marshal() ([]byte, error) {
	return aa, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (aa *AccAddress) Unmarshal(data []byte) error {
	*aa = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (aa AccAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(aa.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (aa *AccAddress) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	aa2, err := AccAddressFromBech32(s)
	if err != nil {
		return err
	}

	*aa = aa2
	return nil
}

// Bytes returns the raw address bytes.
func (aa AccAddress) Bytes() []byte {
	return aa
}

// String implements the Stringer interface.
func (aa AccAddress) String() string {
	if aa.Empty() {
		return ""
	}

	prefix := GetAddressConfig().GetBech32AccountAddrPrefix()
	bech32Addr, err := bech32.ConvertAndEncode(prefix, aa.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
func (aa AccAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(aa.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", aa)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(aa))))
	}
}

//---------------- validator -----------------------

func ValAddressFromHex(address string) (addr ValAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("validator address is empty.")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return ValAddress(bz), nil
}

// ValAddressFromBech32 creates a ValAddress from a Bech32 string.
func ValAddressFromBech32(address string) (addr ValAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return ValAddress{}, nil
	}

	prefix := GetAddressConfig().GetBech32ValidatorAddrPrefix()
	bz, err := GetFromBech32(address, prefix)
	if err != nil {
		return nil, err
	}

	err = verifyAddress(bz)
	if err != nil {
		return nil, err
	}

	return ValAddress(bz), nil
}

// Returns boolean for whether two ValAddresses are Equal
func (va ValAddress) Equals(va2 Address) bool {
	if va.Empty() && va2.Empty() {
		return true
	}

	return bytes.Equal(va.Bytes(), va2.Bytes())
}

// Returns boolean for whether an AccAddress is empty
func (va ValAddress) Empty() bool {
	if va == nil {
		return true
	}

	va2 := ValAddress{}
	return bytes.Equal(va.Bytes(), va2.Bytes())
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (va ValAddress) Marshal() ([]byte, error) {
	return va, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (va *ValAddress) Unmarshal(data []byte) error {
	*va = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (va ValAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(va.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (va *ValAddress) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	va2, err := ValAddressFromBech32(s)
	if err != nil {
		return err
	}

	*va = va2
	return nil
}

// Bytes returns the raw address bytes.
func (va ValAddress) Bytes() []byte {
	return va
}

// String implements the Stringer interface.
func (va ValAddress) String() string {
	if va.Empty() {
		return ""
	}

	prefix := GetAddressConfig().GetBech32ValidatorAddrPrefix()
	bech32Addr, err := bech32.ConvertAndEncode(prefix, va.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (va ValAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(va.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", va)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(va))))
	}
}

//------------------------consensus address-----------------------
func ConsAddressFromHex(address string) (addr ConsAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("consensus address is empty.")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return ConsAddress(bz), nil
}

func ConsAddressFromBech32(address string) (addr ConsAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return ConsAddress{}, nil
	}

	prefix := GetAddressConfig().GetBech32ConsensusAddrPrefix()
	bz, err := GetFromBech32(address, prefix)
	if err != nil {
		return nil, err
	}

	err = verifyAddress(bz)
	if err != nil {
		return nil, err
	}

	return ConsAddress(bz), nil
}

// get ConsAddress from pubkey
func GetConsAddress(pubkey crypto.PubKey) ConsAddress {
	return ConsAddress(pubkey.Address())
}

// Returns boolean for whether two ConsAddress are Equal
func (ca ConsAddress) Equals(ca2 Address) bool {
	if ca.Empty() && ca2.Empty() {
		return true
	}

	return bytes.Equal(ca.Bytes(), ca2.Bytes())
}

// Returns boolean for whether an ConsAddress is empty
func (ca ConsAddress) Empty() bool {
	if ca == nil {
		return true
	}

	ca2 := ConsAddress{}
	return bytes.Equal(ca.Bytes(), ca2.Bytes())
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (ca ConsAddress) Marshal() ([]byte, error) {
	return ca, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (ca *ConsAddress) Unmarshal(data []byte) error {
	*ca = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (ca ConsAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(ca.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (ca *ConsAddress) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	ca2, err := ConsAddressFromBech32(s)
	if err != nil {
		return err
	}

	*ca = ca2
	return nil
}

// Bytes returns the raw address bytes.
func (ca ConsAddress) Bytes() []byte {
	return ca
}

// String implements the Stringer interface.
func (ca ConsAddress) String() string {
	if ca.Empty() {
		return ""
	}

	prefix := GetAddressConfig().GetBech32ConsensusAddrPrefix()
	bech32Addr, err := bech32.ConvertAndEncode(prefix, ca.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (ca ConsAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(ca.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", ca)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(ca))))
	}
}

func stringifyPubKey(pubkey crypto.PubKey, prefix string) (string, error) {
	pk, ok := pubkey.(ed25519.PubKeyEd25519)
	if !ok {
		return "", errors.New("Not Support Pubkey Type")
	}
	return bech32.ConvertAndEncode(prefix, pk[:])
}

func AccPubKeyString(pubkey crypto.PubKey) (string, error) {
	prefix := GetAddressConfig().GetBech32AccountPubPrefix()
	return stringifyPubKey(pubkey, prefix)
}

func ValidatorPubKeyString(pubkey crypto.PubKey) (string, error) {
	prefix := GetAddressConfig().GetBech32ValidatorPubPrefix()
	return stringifyPubKey(pubkey, prefix)
}

func ConsensusPubKeyString(pubkey crypto.PubKey) (string, error) {
	prefix := GetAddressConfig().GetBech32ConsensusPubPrefix()
	return stringifyPubKey(pubkey, prefix)
}

func MustAccPubKeyString(pubkey crypto.PubKey) string {
	prefix := GetAddressConfig().GetBech32AccountPubPrefix()
	s, err := stringifyPubKey(pubkey, prefix)
	if err != nil {
		panic(err)
	}
	return s
}

func MustValidatorPubKeyString(pubkey crypto.PubKey) string {
	prefix := GetAddressConfig().GetBech32ValidatorPubPrefix()
	s, err := stringifyPubKey(pubkey, prefix)
	if err != nil {
		panic(err)
	}
	return s
}

func MustConsensusPubKeyString(pubkey crypto.PubKey) string {
	prefix := GetAddressConfig().GetBech32ConsensusPubPrefix()
	s, err := stringifyPubKey(pubkey, prefix)
	if err != nil {
		panic(err)
	}
	return s
}

func GetAccPubKeyBech32(pubkey string) (pk crypto.PubKey, err error) {
	prefix := GetAddressConfig().GetBech32AccountPubPrefix()
	return getPubKey(pubkey, prefix)
}

func GetValidatorPubKeyBech32(pubkey string) (pk crypto.PubKey, err error) {
	prefix := GetAddressConfig().GetBech32ValidatorPubPrefix()
	return getPubKey(pubkey, prefix)
}

func GetConsensusPubKeyBech32(pubkey string) (pk crypto.PubKey, err error) {
	prefix := GetAddressConfig().GetBech32ConsensusPubPrefix()
	return getPubKey(pubkey, prefix)
}

func getPubKey(pubkey, prefix string) (crypto.PubKey, error) {
	var pk crypto.PubKey
	var err error

	bz, err := GetFromBech32(pubkey, prefix)
	if err != nil {
		return nil, err
	}

	if len(bz) == ed25519.PubKeyEd25519Size {
		var b [ed25519.PubKeyEd25519Size]byte
		copy(b[:], bz)
		pk = ed25519.PubKeyEd25519(b)
	} else {
		pk, err = cryptoAmino.PubKeyFromBytes(bz)
	}

	if err != nil {
		return nil, err
	}

	return pk, nil
}

func verifyAddress(bz []byte) error {
	if len(bz) != AddrLen {
		return errors.New("Incorrect address length")
	}
	return nil
}

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func GetFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errors.New("decoding Bech32 address failed: must provide an address")
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid Bech32 prefix; expected %s, got %s", prefix, hrp)
	}

	return bz, nil
}

//config

func (config *AddressConfig) assertNotSealed() {
	config.mtx.Lock()
	defer config.mtx.Unlock()

	if config.sealed {
		panic("AddressConfig is sealed")
	}
}

func (config *AddressConfig) SetBech32PrefixForAccount(addressPrefix, pubKeyPrefix string) {
	config.assertNotSealed()
	config.bech32AddressPrefix["account_addr"] = addressPrefix
	config.bech32AddressPrefix["account_pub"] = pubKeyPrefix
}

func (config *AddressConfig) SetBech32PrefixForValidator(addressPrefix, pubKeyPrefix string) {
	config.assertNotSealed()
	config.bech32AddressPrefix["validator_addr"] = addressPrefix
	config.bech32AddressPrefix["validator_pub"] = pubKeyPrefix
}

func (config *AddressConfig) SetBech32PrefixForConsensusNode(addressPrefix, pubKeyPrefix string) {
	config.assertNotSealed()
	config.bech32AddressPrefix["consensus_addr"] = addressPrefix
	config.bech32AddressPrefix["consensus_pub"] = pubKeyPrefix
}

func (config *AddressConfig) SetCoinType(coinType uint32) {
	config.assertNotSealed()
	config.coinType = coinType
}

func (config *AddressConfig) Seal() *AddressConfig {
	config.mtx.Lock()
	defer config.mtx.Unlock()

	config.sealed = true
	return config
}

func (config *AddressConfig) GetBech32AccountAddrPrefix() string {
	return config.bech32AddressPrefix["account_addr"]
}

func (config *AddressConfig) GetBech32ValidatorAddrPrefix() string {
	return config.bech32AddressPrefix["validator_addr"]
}

func (config *AddressConfig) GetBech32ConsensusAddrPrefix() string {
	return config.bech32AddressPrefix["consensus_addr"]
}

func (config *AddressConfig) GetBech32AccountPubPrefix() string {
	return config.bech32AddressPrefix["account_pub"]
}

func (config *AddressConfig) GetBech32ValidatorPubPrefix() string {
	return config.bech32AddressPrefix["validator_pub"]
}

func (config *AddressConfig) GetBech32ConsensusPubPrefix() string {
	return config.bech32AddressPrefix["consensus_pub"]
}

func (config *AddressConfig) GetCoinType() uint32 {
	return config.coinType
}
