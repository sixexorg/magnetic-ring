package states

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
	mcom "github.com/sixexorg/magnetic-ring/store/orgchain/common"
)

type AccountState struct {
	Address common.Address
	hash    common.Hash
	Height  uint64
	Data    *Account
}

type Account struct {
	Nonce       uint64
	Balance     *big.Int
	EnergyBalance *big.Int
	BonusHeight uint64
}

type AccountStates []*AccountState

func (s AccountStates) Len() int           { return len(s) }
func (s AccountStates) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s AccountStates) Less(i, j int) bool { return s[i].hash.String() < s[j].hash.String() }
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
	this.Serialization(buff)
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
func (this *AccountState) Serialization(w io.Writer) error {
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
	//fmt.Println("⭕️ hash ", this.Address.ToString(), this.Height, this.Data.Balance.Uint64(), this.Data.EnergyBalance.Uint64(), this.Data.BonusHeight)
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
	return GetAccountStateKey(this.Address, this.Height)
}
func (this *AccountState) Account() common.Address {
	return this.Address
}
func (this *AccountState) League() common.Address {
	return common.Address{}
}
func (this *AccountState) BalanceAdd(num *big.Int) {
	this.Data.Balance.Add(this.Data.Balance, num)
}
func (this *AccountState) BalanceSub(num *big.Int) {
	this.Data.Balance.Sub(this.Data.Balance, num)
}
func (this *AccountState) EnergyAdd(num *big.Int) {
	this.Data.EnergyBalance.Add(this.Data.EnergyBalance, num)
}
func (this *AccountState) EnergySub(num *big.Int) {
	this.Data.EnergyBalance.Sub(this.Data.EnergyBalance, num)
}
func (this *AccountState) NonceSet(num uint64) {
	this.Data.Nonce = num
}
func (this *AccountState) Balance() *big.Int {
	return this.Data.Balance
}
func (this *AccountState) Energy() *big.Int {
	return this.Data.EnergyBalance
}
func (this *AccountState) Nonce() uint64 {
	return this.Data.Nonce
}
func (this *AccountState) BonusHeight() uint64 {
	return this.Data.BonusHeight
}

func (this *AccountState) GetHeight() uint64 {
	return this.Height
}
func (this *AccountState) GetVal() interface{} {
	return this.Data.Balance
}
func (this *AccountState) BonusHSet(num uint64) {
	this.Data.BonusHeight = num
}
func (this *AccountState) HeightSet(height uint64) {
	this.Height = height
}
func GetAccountStateKey(address common.Address, height uint64) []byte {
	buff := make([]byte, 1+common.AddrLength+8)
	buff[0] = byte(mcom.ST_ACCOUNT)
	copy(buff[1:common.AddrLength+1], address[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength+1:], height)
	return buff
}
func GetAccountStatePrifex(address common.Address) []byte {
	buff := make([]byte, 1+common.AddrLength)
	buff[0] = byte(mcom.ST_ACCOUNT)
	copy(buff[1:common.AddrLength+1], address[:])
	return buff
}
