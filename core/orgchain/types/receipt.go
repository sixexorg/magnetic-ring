package types

import (
	"io"

	"github.com/sixexorg/magnetic-ring/common"

	"bytes"

	"sort"

	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
)

type Receipt struct {
	Status  bool
	GasUsed uint64
	TxHash  common.Hash
}

type Receipts []*Receipt

func (this Receipts) Len() int      { return len(this) }
func (this Receipts) Swap(i, j int) { this[i], this[j] = this[j], this[i] }
func (this Receipts) Less(i, j int) bool {
	return this[i].Hash().String() < this[j].Hash().String()
}
func (this Receipts) GetHashRoot() common.Hash {
	sort.Sort(this)
	hashes := make([]common.Hash, 0, this.Len())
	for _, v := range this {
		hashes = append(hashes, v.Hash())
	}
	return common.ComputeMerkleRoot(hashes)
}

//Serialize the blockheader
func (this *Receipt) Serialize(w io.Writer) error {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteHash(this.TxHash)
	sk.WriteBool(this.Status)
	sk.WriteUint64(this.GasUsed)
	_, err := w.Write(sk.Bytes())
	return err
}

func (this *Receipt) Deserialize(r io.Reader) error {
	if this == nil {
		this = &Receipt{}
	}
	var err error
	txHash, err := serialization.ReadBytes(r, common.HashLength)
	if err != nil {
		return err
	}
	this.TxHash, _ = common.ParseHashFromBytes(txHash)
	this.Status, err = serialization.ReadBool(r)
	if err != nil {
		return err
	}
	this.GasUsed, err = serialization.ReadUint64(r)
	return err
}

func (this *Receipt) Hash() common.Hash {
	buff := new(bytes.Buffer)
	this.Serialize(buff)
	hash, _ := common.ParseHashFromBytes(common.Sha256(buff.Bytes()))
	return hash
}
