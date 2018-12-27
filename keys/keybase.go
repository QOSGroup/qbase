package keys

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/QOSGroup/qbase/keys/hd"
	"github.com/QOSGroup/qbase/types"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/tyler-smith/go-bip39"

	go_amino "github.com/tendermint/go-amino"
	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	// used for deriving seed from mnemonic
	defaultBIP39Passphrase = ""

	// bits of entropy to draw when creating a mnemonic
	defaultEntropySize = 256

	addressSuffix = "address"
	infoSuffix    = "info"
)

type Keybase struct {
	db  dbm.DB
	cdc *go_amino.Codec
}

// New creates a new Keybase instance using the passed DB for reading and writing keys.
func New(db dbm.DB, cdc *go_amino.Codec) Keybase {
	return Keybase{
		db:  db,
		cdc: cdc,
	}
}

func (kb Keybase) IsNil() bool {
	return kb.db == nil
}

func (kb Keybase) CreateEnMnemonic(name, passwd string) (info Info, mnemonic string, err error) {

	entropy, err := bip39.NewEntropy(defaultEntropySize)
	if err != nil {
		return
	}

	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return
	}

	seed := bip39.NewSeed(mnemonic, defaultBIP39Passphrase)
	info, err = kb.persistDerivedKey(seed, passwd, name, hd.FullFundraiserPath)
	return
}

func (kb Keybase) Derive(name, mnemonic, passwd, hdPath string) (info Info, err error) {

	words := strings.Split(mnemonic, " ")
	if len(words) != 12 && len(words) != 24 {
		err = fmt.Errorf("recovering only works with 12 word (fundraiser) or 24 word mnemonics, got: %v words", len(words))
		return
	}

	_, err = hd.NewParamsFromPath(hdPath)
	if err != nil {
		return
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, defaultBIP39Passphrase)
	if err != nil {
		return
	}
	info, err = kb.persistDerivedKey(seed, passwd, name, hdPath)

	return
}

// CreateOffline creates a new reference to an offline keypair
// It returns the created key info
func (kb Keybase) CreateOffline(name string, pub crypto.PubKey) (Info, error) {
	return kb.writeOfflineKey(pub, name), nil
}

func (kb Keybase) CreateImportInfo(name, passwd string, priv crypto.PrivKey) (Info, error) {
	return kb.writeImportInfoKey(priv, name, passwd), nil
}

func (kb *Keybase) persistDerivedKey(seed []byte, passwd, name, fullHdPath string) (info Info, err error) {
	// create master key and derive first key:
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, fullHdPath)
	if err != nil {
		return
	}

	// if we have a password, use it to encrypt the private key and store it
	// else store the public key only
	priKey := ed25519.GenPrivKeyFromSecret(derivedPriv[:])

	if passwd != "" {
		info = kb.writeLocalKey(priKey, name, passwd)
	} else {
		info = kb.writeOfflineKey(priKey.PubKey(), name)
	}
	return
}

// List returns the keys from storage in alphabetical order.
func (kb Keybase) List() ([]Info, error) {
	var res []Info
	iter := kb.db.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		key := string(iter.Key())

		// need to include only keys in storage that have an info suffix
		if strings.HasSuffix(key, infoSuffix) {
			info, err := readInfo(kb.cdc, iter.Value())
			if err != nil {
				return nil, err
			}
			res = append(res, info)
		}
	}
	return res, nil
}

// Get returns the public information about one key.
func (kb Keybase) Get(name string) (Info, error) {
	bs := kb.db.Get(infoKey(name))
	if len(bs) == 0 {
		return nil, fmt.Errorf("Name: %s not found", name)
	}
	return readInfo(kb.cdc, bs)
}

func (kb Keybase) GetByAddress(address types.Address) (Info, error) {
	ik := kb.db.Get(addrKey(address))
	if len(ik) == 0 {
		return nil, fmt.Errorf("key with address %s not found", address)
	}
	bs := kb.db.Get(ik)
	return readInfo(kb.cdc, bs)
}

// Sign signs the msg with the named key.
// It returns an error if the key doesn't exist or the decryption fails.
func (kb Keybase) Sign(name, passphrase string, msg []byte) (sig []byte, pub crypto.PubKey, err error) {
	info, err := kb.Get(name)
	if err != nil {
		return
	}
	var priv crypto.PrivKey
	switch info.(type) {
	case *localInfo:
		linfo := info.(*localInfo)
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return
		}
		priv, err = UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, nil, err
		}
	case *importInfo:
		iinfo := info.(*importInfo)
		encBytes, err := base64.StdEncoding.DecodeString(iinfo.PrivKeyEncrypt)
		if err != nil {
			return nil, nil, err
		}
		priv, err = decryptPrivKey(iinfo.Salt, encBytes, passphrase)
		if err != nil {
			return nil, nil, err
		}
	case *offlineInfo:
		linfo := info.(*offlineInfo)
		_, err := fmt.Fprintf(os.Stderr, "Bytes to sign:\n%s", msg)
		if err != nil {
			return nil, nil, err
		}
		buf := bufio.NewReader(os.Stdin)
		_, err = fmt.Fprintf(os.Stderr, "\nEnter Amino-encoded signature:\n")
		if err != nil {
			return nil, nil, err
		}
		// Will block until user inputs the signature
		signed, err := buf.ReadString('\n')
		if err != nil {
			return nil, nil, err
		}
		kb.cdc.MustUnmarshalBinaryLengthPrefixed([]byte(signed), sig)
		return sig, linfo.GetPubKey(), nil
	}
	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}
	pub = priv.PubKey()
	return sig, pub, nil
}

func (kb Keybase) ExportPrivateKeyObject(name string, passphrase string) (crypto.PrivKey, error) {
	info, err := kb.Get(name)
	if err != nil {
		return nil, err
	}
	var priv crypto.PrivKey
	switch info.(type) {
	case *localInfo:
		linfo := info.(*localInfo)
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return nil, err
		}
		priv, err = UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, err
		}
	case *importInfo:
		iinfo := info.(*importInfo)
		encBytes, err := base64.StdEncoding.DecodeString(iinfo.PrivKeyEncrypt)
		if err != nil {
			return nil, err
		}
		priv, err = decryptPrivKey(iinfo.Salt, encBytes, passphrase)
		if err != nil {
			return nil, err
		}
	case offlineInfo:
		return nil, errors.New("Only works on local private keys")
	}
	return priv, nil
}

func (kb Keybase) Export(name string) (armor string, err error) {
	bz := kb.db.Get(infoKey(name))
	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}
	return ArmorInfoBytes(bz), nil
}

// ExportPubKey returns public keys in ASCII armored format.
// Retrieve a Info object by its name and return the public key in
// a portable format.
func (kb Keybase) ExportPubKey(name string) (armor string, err error) {
	bz := kb.db.Get(infoKey(name))
	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}
	info, err := readInfo(kb.cdc, bz)
	if err != nil {
		return
	}
	return ArmorPubKeyBytes(info.GetPubKey().Bytes()), nil
}

func (kb Keybase) Import(name string, armor string) (err error) {
	bz := kb.db.Get(infoKey(name))
	if len(bz) > 0 {
		return errors.New("Cannot overwrite data for name " + name)
	}
	infoBytes, err := UnarmorInfoBytes(armor)
	if err != nil {
		return
	}
	kb.db.Set(infoKey(name), infoBytes)
	return nil
}

// ImportPubKey imports ASCII-armored public keys.
// Store a new Info object holding a public key only, i.e. it will
// not be possible to sign with it as it lacks the secret key.
func (kb Keybase) ImportPubKey(name string, armor string) (err error) {
	bz := kb.db.Get(infoKey(name))
	if len(bz) > 0 {
		return errors.New("Cannot overwrite data for name " + name)
	}
	pubBytes, err := UnarmorPubKeyBytes(armor)
	if err != nil {
		return
	}

	var pubKey crypto.PubKey
	err = kb.cdc.UnmarshalBinaryBare(pubBytes, &pubKey)
	if err != nil {
		return
	}
	kb.writeOfflineKey(pubKey, name)
	return
}

// Delete removes key forever, but we must present the
// proper passphrase before deleting it (for security).
// A passphrase of 'yes' is used to delete stored
// references to offline and Ledger / HW wallet keys
func (kb Keybase) Delete(name, passphrase string) error {
	// verify we have the proper password before deleting
	info, err := kb.Get(name)
	if err != nil {
		return err
	}
	switch info.(type) {
	case *localInfo:
		linfo := info.(*localInfo)
		_, err = UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return err
		}
		kb.db.DeleteSync(addrKey(linfo.GetAddress()))
		kb.db.DeleteSync(infoKey(name))
		return nil
	case *importInfo:
		iinfo := info.(*importInfo)
		encBytes, err := base64.StdEncoding.DecodeString(iinfo.PrivKeyEncrypt)
		if err != nil {
			return nil
		}
		_, err = decryptPrivKey(iinfo.Salt, encBytes, passphrase)
		if err != nil {
			return nil
		}

		kb.db.DeleteSync(addrKey(iinfo.GetAddress()))
		kb.db.DeleteSync(infoKey(name))
		return nil

	case *offlineInfo:
		if passphrase != "yes" {
			return fmt.Errorf("enter 'yes' exactly to delete the key - this cannot be undone")
		}
		kb.db.DeleteSync(addrKey(info.GetAddress()))
		kb.db.DeleteSync(infoKey(name))
		return nil
	}

	return nil
}

// Update changes the passphrase with which an already stored key is
// encrypted.
//
// oldpass must be the current passphrase used for encryption,
// getNewpass is a function to get the passphrase to permanently replace
// the current passphrase
func (kb Keybase) Update(name, oldpass string, getNewpass func() (string, error)) error {
	info, err := kb.Get(name)
	if err != nil {
		return err
	}
	switch info.(type) {
	case *localInfo:
		linfo := info.(*localInfo)
		key, err := UnarmorDecryptPrivKey(linfo.PrivKeyArmor, oldpass)
		if err != nil {
			return err
		}
		newpass, err := getNewpass()
		if err != nil {
			return err
		}
		kb.writeLocalKey(key, name, newpass)
		return nil
	case *importInfo:
		iinfo := info.(*importInfo)
		encBytes, err := base64.StdEncoding.DecodeString(iinfo.PrivKeyEncrypt)
		if err != nil {
			return err
		}
		priv, err := decryptPrivKey(iinfo.Salt, encBytes, oldpass)
		if err != nil {
			return err
		}

		newpass, err := getNewpass()
		if err != nil {
			return err
		}
		kb.writeImportInfoKey(priv, name, newpass)
		return nil

	default:
		return fmt.Errorf("locally stored key required")
	}
}

// CloseDB releases the lock and closes the storage backend.
func (kb Keybase) CloseDB() {
	kb.db.Close()
}

func (kb Keybase) writeLocalKey(priv crypto.PrivKey, name, passphrase string) Info {
	// encrypt private key using passphrase
	privArmor := EncryptArmorPrivKey(priv, passphrase)
	// make Info
	pub := priv.PubKey()
	info := newLocalInfo(name, pub, privArmor)
	kb.writeInfo(info, name)
	return info
}

func (kb Keybase) writeImportInfoKey(priv crypto.PrivKey, name, passphrase string) Info {
	salt, privEncrypt := encryptPrivKey(priv, passphrase)
	pub := priv.PubKey()

	str := base64.StdEncoding.EncodeToString(privEncrypt)
	info := newImportInfo(name, pub, salt, str)
	kb.writeInfo(info, name)
	return info
}

func (kb Keybase) writeOfflineKey(pub crypto.PubKey, name string) Info {
	info := newOfflineInfo(name, pub)
	kb.writeInfo(info, name)
	return info
}

func (kb Keybase) writeInfo(info Info, name string) {
	// write the info by key
	key := infoKey(name)
	kb.db.SetSync(key, writeInfo(kb.cdc, info))
	// store a pointer to the infokey by address for fast lookup
	kb.db.SetSync(addrKey(info.GetAddress()), key)
}

func addrKey(address types.Address) []byte {
	return []byte(fmt.Sprintf("%s.%s", address.String(), addressSuffix))
}

func infoKey(name string) []byte {
	return []byte(fmt.Sprintf("%s.%s", name, infoSuffix))
}

// KeyType reflects a human-readable type for key listing.
type KeyType uint

// Info KeyTypes
const (
	TypeLocal   KeyType = 0 //从命令行生成或从助记符中恢复
	TypeOffline KeyType = 1
	TypeImport  KeyType = 2 //以私钥的方式导入

)

var keyTypes = map[KeyType]string{
	TypeLocal:   "local",
	TypeOffline: "offline",
	TypeImport:  "import",
}

// String implements the stringer interface for KeyType.
func (kt KeyType) String() string {
	return keyTypes[kt]
}

// Info is the publicly exposed information about a keypair
type Info interface {
	// Human-readable type for key listing
	GetType() KeyType
	// Name of the key
	GetName() string
	// Public key
	GetPubKey() crypto.PubKey
	// Address
	GetAddress() types.Address
}

var _ Info = &localInfo{}
var _ Info = &offlineInfo{}
var _ Info = &importInfo{}

type importInfo struct {
	Name           string        `json:"name"`
	PubKey         crypto.PubKey `json:"pubkey"`
	Salt           []byte        `json:"salt"`
	PrivKeyEncrypt string        `json:"encrypt"`
}

func newImportInfo(name string, pub crypto.PubKey, salt []byte, privKeyEncrypt string) Info {
	return &importInfo{
		Name:           name,
		PubKey:         pub,
		Salt:           salt,
		PrivKeyEncrypt: privKeyEncrypt,
	}
}

func (i importInfo) GetType() KeyType {
	return TypeImport
}

func (i importInfo) GetName() string {
	return i.Name
}

func (i importInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

func (i importInfo) GetAddress() types.Address {
	return i.PubKey.Address().Bytes()
}

// localInfo is the public information about a locally stored key
type localInfo struct {
	Name         string        `json:"name"`
	PubKey       crypto.PubKey `json:"pubkey"`
	PrivKeyArmor string        `json:"privkey.armor"`
}

func newLocalInfo(name string, pub crypto.PubKey, privArmor string) Info {
	return &localInfo{
		Name:         name,
		PubKey:       pub,
		PrivKeyArmor: privArmor,
	}
}

func (i localInfo) GetType() KeyType {
	return TypeLocal
}

func (i localInfo) GetName() string {
	return i.Name
}

func (i localInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

func (i localInfo) GetAddress() types.Address {
	return i.PubKey.Address().Bytes()
}

// offlineInfo is the public information about an offline key
type offlineInfo struct {
	Name   string        `json:"name"`
	PubKey crypto.PubKey `json:"pubkey"`
}

func newOfflineInfo(name string, pub crypto.PubKey) Info {
	return &offlineInfo{
		Name:   name,
		PubKey: pub,
	}
}

func (i offlineInfo) GetType() KeyType {
	return TypeOffline
}

func (i offlineInfo) GetName() string {
	return i.Name
}

func (i offlineInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

func (i offlineInfo) GetAddress() types.Address {
	return i.PubKey.Address().Bytes()
}

// encoding info
func writeInfo(cdc *go_amino.Codec, i Info) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(i)
}

// decoding info
func readInfo(cdc *go_amino.Codec, bz []byte) (info Info, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(bz, &info)
	return
}
