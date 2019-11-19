package common

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sixexorg/magnetic-ring/errors"

	"io"
	"sort"
)

//address type
type AddressType byte

const (
	AddrPrefix     = "bx"
	AddrLength     = 21
	HashLength     = 32
	PubkHashLength = 65
)

// address type enum
const (

	NormalAddress AddressType = iota

	MultiAddress

	OrganizationAddress

	ContractAddress

	NodeAddress

	NoneAddress
)

type Address [AddrLength]byte

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (addr *Address) SetBytes(hash []byte, addrType AddressType) {
	if len(hash) > AddrLength-1 {
		hash = hash[len(hash)-AddrLength-1:]
	}

	copy(addr[AddrLength-len(hash)-1:], hash)
	addr[AddrLength-1] = addrType.ToByte()
}

func (addr Address) Equals(ag Address) bool {
	return bytes.Compare(addr[:], ag[:]) == 0

}

func (addr Address) ToString() string {
	data := checksum(addr[:20], ToBase32(addr[:20]))
	return fmt.Sprintf("%s%d%s", AddrPrefix, addr[20], data)
}

func (h Address) MarshalText() ([]byte, error) {
	return []byte(h.ToString()), nil
}

func checksum(hash []byte, base32 string) string {
	digest := Sha256(hash)
	result := []byte(base32)
	for i := 0; i < len(result); i++ {
		hashByte := digest[i/8]

		if i%2 == 0 {
			hashByte &= 0xf
		} else {
			hashByte = hashByte >> 4
		}

		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}

	return string(result)
}

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte, addrType AddressType) Address {
	var a Address
	a.SetBytes(b, addrType)
	return a
}

func Bytes2Address(b []byte) Address {

	var a Address = Address{}
	copy(a[:], b[:])
	return a
}


func ToAddress(str string) (address Address, err error) {
	if len(str) <= 3 {
		return address, errors.ERR_ACCT_FMT1
	}

	if strings.ToLower(str[0:2]) != AddrPrefix {
		return address, errors.ERR_ACCT_FMT2
	}

	addrType := GetAddressType(str[2])
	if addrType == NoneAddress {
		return address, errors.ERR_ACCT_FMT3
	}

	data, err := FromBase32(strings.ToLower(str[3:]))
	if err != nil {
		return address, errors.ERR_ACCT_FMT4
	}

	address.SetBytes(data, addrType)
	return address, nil
}

func (addrType AddressType) ToByte() byte {
	return byte(addrType)
}

func GetAddressType(bit uint8) AddressType {
	switch bit {
	case 48:
		return NormalAddress
	case 49:
		return MultiAddress
	case 50:
		return OrganizationAddress
	case 51:
		return ContractAddress
	case 52:
		return NodeAddress
	default:
		return NoneAddress
	}
}

type Hash [HashLength]byte
type PubHash [PubkHashLength]byte

func StringToHash(hashStr string) (Hash, error) {
	buf, err := Hex2Bytes(hashStr)
	if err != nil {
		return Hash{}, err
	}
	ha := Hash{}
	copy(ha[:], buf)
	return ha, nil
}
func (hash Hash) ToBytes() []byte {
	return hash[:]
}
func (hash Hash) String() string {
	return Bytes2Hex(hash[:])
}
func (hash *Hash) Serialize(w io.Writer) error {
	_, err := w.Write(hash[:])
	return err
}
func (hash *Hash) Deserialize(r io.Reader) error {
	_, err := io.ReadFull(r, hash[:])
	return err
}

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

type SigBuf [][]byte

type HashArray []Hash

func (s HashArray) Len() int           { return len(s) }
func (s HashArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s HashArray) Less(i, j int) bool { return s[i].String() < s[j].String() }
func (s HashArray) GetHashRoot() Hash {
	sort.Sort(s)
	return ComputeMerkleRoot(s)
}

func CreateLeagueAddress(leagueBuf []byte) (Address, error) {

	hash := Sha256Ripemd160(leagueBuf)
	leagueAddr := BytesToAddress(hash, OrganizationAddress)

	return leagueAddr, nil
}

type SigMap []*SigMapItem
type SigMapItem struct {
	Key Address
	Val []byte
}

type SigTmplt struct {
	SigData []*Maus
}

func NewSigTmplt() *SigTmplt {
	tmplt := new(SigTmplt)
	tmplt.SigData = make([]*Maus, 0)
	return tmplt
}

type SigPack struct {
	SigData []*SigMap
}

func (s *SigPack) ToBytes() []byte {
	return nil
}

type Maus []Mau
type Mau struct {
	M     uint8
	Pubks []PubHash
}
