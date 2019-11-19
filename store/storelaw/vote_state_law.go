package storelaw

import (
	"math/big"

	"io"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
)

//voteId is tx
type VoteState struct {
	VoteId     common.Hash
	Height     uint64
	Agree      *big.Int
	Against    *big.Int
	Abstention *big.Int
}
type AccountVoted struct {
	VoteId   common.Hash
	Height   uint64
	Accounts []common.Address
}

func (this *VoteState) AddAgree(num *big.Int) {
	this.Agree.Add(this.Agree, num)
}
func (this *VoteState) AddAgainst(num *big.Int) {
	this.Against.Add(this.Against, num)
}
func (this *VoteState) AddAbstention(num *big.Int) {
	this.Abstention.Add(this.Abstention, num)
}
func (this *VoteState) GetVal() []byte {
	sk := sink.NewZeroCopySink(nil)
	c1, _ := sink.BigIntToComplex(this.Agree)
	c2, _ := sink.BigIntToComplex(this.Against)
	c3, _ := sink.BigIntToComplex(this.Abstention)
	sk.WriteComplex(c1)
	sk.WriteComplex(c2)
	sk.WriteComplex(c3)
	return sk.Bytes()
}

func (this *VoteState) Deserialize(r io.Reader) error {
	byteArr, err := serialization.ReadBytes(r, common.HashLength)
	if err != nil {
		return err
	}
	this.VoteId, err = common.ParseHashFromBytes(byteArr)
	if err != nil {
		return err
	}
	this.Height, err = serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	c1, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	this.Agree, err = c1.ComplexToBigInt()
	if err != nil {
		return err
	}
	c2, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	this.Against, err = c2.ComplexToBigInt()
	if err != nil {
		return err
	}

	c3, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	this.Abstention, err = c3.ComplexToBigInt()
	if err != nil {
		return err
	}
	return nil
}

//porportion:(0,100)
//1:pass,0:unfinished
func (this *VoteState) GetResult(total *big.Int, proportion int64) bool {
	if proportion < 0 {
		proportion = 0
	} else if proportion > 100 {
		proportion = 100
	}
	no := big.NewInt(0)
	no = no.Add(this.Against, this.Abstention)
	yes := big.NewInt(0).Set(this.Agree)
	if yes.Cmp(no) == 1 {
		//voteFor bigger
		totalTmp := big.NewInt(0).Set(total)
		totalTmp.Mul(totalTmp, big.NewInt(proportion))
		if totalTmp.Cmp(yes.Mul(yes, big.NewInt(100))) == -1 {
			return true
		}
	}
	return false
}
