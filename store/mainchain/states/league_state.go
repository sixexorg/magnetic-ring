package states

import (
	"io"
	"math/big"

	"github.com/sixexorg/magnetic-ring/common/sink"

	"encoding/binary"

	"io/ioutil"

	"bytes"

	"sort"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

type LeagueState struct {
	Address common.Address
	hash    common.Hash
	Symbol  common.Symbol
	Height  uint64
	MinBox  uint64         // not stored in the db
	Creator common.Address // not stored in the db
	Rate    uint32         // not stored in the db
	Private bool
	Data    *League
}

type League struct {
	Nonce      uint64
	Name       common.Hash
	FrozenBox  *big.Int
	MemberRoot common.Hash
}

func (this *LeagueState) IsPrivate() bool {
	return this.Private
}

type LeagueStates []*LeagueState

func (s LeagueStates) Len() int           { return len(s) }
func (s LeagueStates) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s LeagueStates) Less(i, j int) bool { return s[i].Hash().String() < s[j].Hash().String() }
func (s LeagueStates) GetHashRoot() common.Hash {
	sort.Sort(s)
	hashes := make([]common.Hash, 0, s.Len())
	for _, v := range s {
		hashes = append(hashes, v.Hash())
	}
	return common.ComputeMerkleRoot(hashes)
}

type LeagueAccountStatus byte

const (
	LAS_Apply  LeagueAccountStatus = 0x01
	LAS_Normal LeagueAccountStatus = 0x02
	LAS_Exit   LeagueAccountStatus = 0x03
)

func (this *LeagueState) Hash() common.Hash {
	buff := new(bytes.Buffer)
	buff.Write(this.GetKey())
	this.Serialize(buff)
	hash, _ := common.ParseHashFromBytes(common.Sha256(buff.Bytes()))
	this.hash = hash
	return hash
}

func (this *LeagueState) Serialize(w io.Writer) error {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteUint64(this.Data.Nonce)
	sk.WriteHash(this.Data.Name)
	c1, _ := sink.BigIntToComplex(this.Data.FrozenBox)
	sk.WriteComplex(c1)
	sk.WriteHash(this.Data.MemberRoot)
	w.Write(sk.Bytes())
	return nil
}

func (this *LeagueState) Deserialize(r io.Reader) error {
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var eof bool
	source := sink.NewZeroCopySource(buff)
	_, eof = source.NextByte()
	this.Address, eof = source.NextAddress()
	this.Height, eof = source.NextUint64()
	this.Data = &League{}
	this.Data.Nonce, eof = source.NextUint64()
	this.Data.Name, eof = source.NextHash()
	c1, eof := source.NextComplex()
	this.Data.FrozenBox, err = c1.ComplexToBigInt()
	if err != nil {
		return err
	}
	this.Data.MemberRoot, eof = source.NextHash()
	if eof {
		return errors.ERR_TXRAW_EOF
	}
	return nil
}
func (this *LeagueState) GetKey() []byte {
	buff := make([]byte, 1+common.AddrLength+8)
	buff[0] = byte(mcom.ST_LEAGUE)
	copy(buff[1:common.AddrLength+1], this.Address[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength+1:], this.Height)
	return buff
}

func (this *LeagueState) AddMetaBox(num *big.Int) {
	fmt.Println("☀️ AddMetaBox ", this.Data.FrozenBox.Uint64(), num.Uint64())
	this.Data.FrozenBox.Add(this.Data.FrozenBox, num)
}
