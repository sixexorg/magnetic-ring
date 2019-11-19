package extstates

import (
	"encoding/binary"
	"math/big"
	"sort"

	"github.com/sixexorg/magnetic-ring/common/sink"

	"io"

	"io/ioutil"

	"bytes"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

type LeagueAccountState struct {
	Address  common.Address
	LeagueId common.Address
	hash     common.Hash
	Height   uint64
	Data     *Account
}

type Account struct {
	Nonce       uint64
	Balance     *big.Int
	EnergyBalance *big.Int
	BonusHeight uint64
}

type LeagueAccountStates []*LeagueAccountState

func (s LeagueAccountStates) Len() int           { return len(s) }
func (s LeagueAccountStates) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s LeagueAccountStates) Less(i, j int) bool { return s[i].hash.String() < s[j].hash.String() }
func (s LeagueAccountStates) GetHashRoot() common.Hash {
	sort.Sort(s)
	hashes := make([]common.Hash, 0, s.Len())
	for _, v := range s {
		hashes = append(hashes, v.Hash())
	}
	return common.ComputeMerkleRoot(hashes)
}

func (this *LeagueAccountState) Hash() common.Hash {
	buff := new(bytes.Buffer)
	this.Serialization(buff)
	hash, _ := common.ParseHashFromBytes(common.Sha256(buff.Bytes()))
	this.hash = hash
	return hash
}
func (this *LeagueAccountState) Serialize(w io.Writer) error {
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
func (this *LeagueAccountState) Serialization(w io.Writer) error {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteAddress(this.Address)
	sk.WriteUint64(this.Height)
	sk.WriteUint64(this.Data.Nonce)
	c1, _ := sink.BigIntToComplex(this.Data.Balance)
	sk.WriteComplex(c1)
	c2, _ := sink.BigIntToComplex(this.Data.EnergyBalance)
	sk.WriteComplex(c2)
	sk.WriteUint64(this.Data.BonusHeight)
	w.Write(sk.Bytes())
	//fmt.Println("⛔️ hash ", this.Address.ToString(), this.Height, this.Data.Balance.Uint64(), this.Data.EnergyBalance.Uint64(), this.Data.BonusHeight)
	return nil
}

func (this *LeagueAccountState) Deserialize(r io.Reader) error {
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var eof bool
	source := sink.NewZeroCopySource(buff)
	_, eof = source.NextByte()
	this.LeagueId, eof = source.NextAddress()
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

func (this *LeagueAccountState) GetKey() []byte {
	buff := make([]byte, 1+2*common.AddrLength+8)
	buff[0] = byte(mcom.EXT_LEAGUE_ACCOUNT)
	copy(buff[1:common.AddrLength+1], this.LeagueId[:])
	copy(buff[common.AddrLength+1:common.AddrLength*2+1], this.Address[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength*2+1:], this.Height)
	return buff
}
func (this *LeagueAccountState) Account() common.Address {
	return this.Address
}
func (this *LeagueAccountState) League() common.Address {
	return this.LeagueId
}
func (this *LeagueAccountState) BalanceAdd(num *big.Int) {
	this.Data.Balance.Add(this.Data.Balance, num)
}
func (this *LeagueAccountState) BalanceSub(num *big.Int) {
	this.Data.Balance.Sub(this.Data.Balance, num)
}
func (this *LeagueAccountState) EnergyAdd(num *big.Int) {
	this.Data.EnergyBalance.Add(this.Data.EnergyBalance, num)
}
func (this *LeagueAccountState) EnergySub(num *big.Int) {
	this.Data.EnergyBalance.Sub(this.Data.EnergyBalance, num)
}
func (this *LeagueAccountState) NonceSet(num uint64) {
	this.Data.Nonce = num
}
func (this *LeagueAccountState) Balance() *big.Int {
	return this.Data.Balance
}
func (this *LeagueAccountState) Energy() *big.Int {
	return this.Data.EnergyBalance
}
func (this *LeagueAccountState) Nonce() uint64 {
	return this.Data.Nonce
}
func (this *LeagueAccountState) BonusHeight() uint64 {
	return this.Data.BonusHeight
}
func (this *LeagueAccountState) GetHeight() uint64 {
	return this.Height
}
func (this *LeagueAccountState) GetVal() interface{} {
	return this.Data.Balance
}
func (this *LeagueAccountState) BonusHSet(height uint64) {
	this.Data.BonusHeight = height
}
func (this *LeagueAccountState) HeightSet(height uint64) {
	this.Height = height
}
