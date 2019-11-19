package states

import (
	"encoding/binary"
	"math/big"
	"sort"

	"github.com/sixexorg/magnetic-ring/common/sink"

	"io"

	"io/ioutil"

	"bytes"

	"encoding/gob"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

func init() {
	gob.Register(&big.Int{})
}

type AccountState struct {
	Address     common.Address
	hash        common.Hash
	Height      uint64
	Data        *Account
	LeftBalance *big.Int //only for accountlevel
}

type Account struct {
	Nonce       uint64
	Balance     *big.Int
	EnergyBalance *big.Int
	BonusHeight uint64
}

func (this *AccountState) GetHeight() uint64 {
	return this.Height
}
func (this *AccountState) GetVal() interface{} {
	return this.Data.Balance
}

type AccountStates []*AccountState

func (s AccountStates) Len() int           { return len(s) }
func (s AccountStates) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s AccountStates) Less(i, j int) bool { return s[i].Hash().String() < s[j].Hash().String() }
func (s AccountStates) GetHashRoot() common.Hash {
	sort.Sort(s)
	hashes := make([]common.Hash, 0, s.Len())
	for _, v := range s {
		hashes = append(hashes, v.Hash())
	}
	return common.ComputeMerkleRoot(hashes)
}

func (this *AccountState) Hash() common.Hash {
	buff := new(bytes.Buffer)
	buff.Write(this.GetKey())
	this.Serialize(buff)
	hash, _ := common.ParseHashFromBytes(common.Sha256(buff.Bytes()))
	this.hash = hash
	return hash
}
func (this *AccountState) Serialize(w io.Writer) error {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteUint64(this.Data.Nonce)
	c1, _ := sink.BigIntToComplex(this.Data.Balance)
	sk.WriteComplex(c1)
	c2, _ := sink.BigIntToComplex(this.Data.EnergyBalance)
	sk.WriteComplex(c2)
	sk.WriteUint64(this.Data.BonusHeight)
	w.Write(sk.Bytes())
	return nil
}

func (this *AccountState) Deserialize(r io.Reader) error {
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var eof bool
	source := sink.NewZeroCopySource(buff)
	_, eof = source.NextByte()
	this.Address, eof = source.NextAddress()
	this.Height, eof = source.NextUint64()
	this.Data = &Account{}
	this.Data.Nonce, eof = source.NextUint64()
	bal, eof := source.NextComplex()
	this.Data.Balance, err = bal.ComplexToBigInt()
	if err != nil {
		return err
	}
	bbl, eof := source.NextComplex()
	this.Data.EnergyBalance, err = bbl.ComplexToBigInt()
	if err != nil {
		return err
	}
	this.Data.BonusHeight, eof = source.NextUint64()
	if eof {
		return errors.ERR_TXRAW_EOF
	}
	return nil
}

func (this *AccountState) GetKey() []byte {
	buff := make([]byte, 1+common.AddrLength+8)
	buff[0] = byte(mcom.ST_ACCOUNT)
	copy(buff[1:common.AddrLength+1], this.Address[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength+1:], this.Height)
	return buff
}
