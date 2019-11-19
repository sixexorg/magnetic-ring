package types

import (
	"fmt"
	"io"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"

	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/errors"
)

type Block struct {
	Header       *Header
	Transactions Transactions
	Sigs         *SigData
}

type SigData struct {
	TimeoutSigs common.SigBuf
	FailerSigs  common.SigBuf
	ProcSigs    common.SigBuf
}

func (b *SigData) Serialize(w io.Writer) error {
	if b == nil {
		b = &SigData{}
	}
	if b.TimeoutSigs == nil {
		b.TimeoutSigs = make([][]byte, 0)
	}
	if b.FailerSigs == nil {
		b.FailerSigs = make([][]byte, 0)
	}
	if b.ProcSigs == nil {
		b.ProcSigs = make([][]byte, 0)
	}
	sk := sink.NewZeroCopySink(nil)
	t, err := sink.SigBufToComplex(b.TimeoutSigs)
	if err != nil {
		return err
	}
	sk.WriteComplex(t)

	f, err := sink.SigBufToComplex(b.FailerSigs)
	if err != nil {
		return err
	}
	sk.WriteComplex(f)

	p, err := sink.SigBufToComplex(b.ProcSigs)
	if err != nil {
		return err
	}
	sk.WriteComplex(p)

	w.Write(sk.Bytes())
	return nil
}

/*func (this *SigData) Deserialization(source *sink.ZeroCopySource) error {
	var eof bool
	cmplx0, eof := source.NextComplex()
	t, err := cmplx0.ComplexToSigBuf()
	if err != nil {
		return err
	}
	this.TimeoutSigs = t
	cmplx1, eof := source.NextComplex()
	f, err := cmplx1.ComplexToSigBuf()
	if err != nil {
		return err
	}
	this.FailerSigs = f
	cmplx2, eof := source.NextComplex()
	p, err := cmplx2.ComplexToSigBuf()
	if err != nil {
		return err
	}
	this.ProcSigs = p
	if eof {
		return errors.ERR_TXRAW_EOF
	}
	return nil
}*/
func (this *SigData) Deserialize(r io.Reader) error {
	cmplx0, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	t, err := cmplx0.ComplexToSigBuf()
	if err != nil {
		return err
	}
	this.TimeoutSigs = t

	cmplx1, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	f, err := cmplx1.ComplexToSigBuf()
	if err != nil {
		return err
	}
	this.FailerSigs = f

	cmplx2, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	p, err := cmplx2.ComplexToSigBuf()
	if err != nil {
		return err
	}
	this.ProcSigs = p
	return nil
}
func (b *Block) Serialize(w io.Writer) error {
	err := b.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = serialization.WriteUint32(w, uint32(len(b.Transactions)))
	if err != nil {
		return fmt.Errorf("Block item Transactions length serialization failed: %s", err)
	}
	for _, transaction := range b.Transactions {
		err := transaction.Serialize(w)
		if err != nil {
			return err
		}
	}
	err = b.Sigs.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

func (this *Block) Deserialize(r io.Reader) error {
	var eof bool
	hder := new(Header)
	err := hder.Deserialize(r)
	if err != nil {
		return err
	}
	this.Header = hder

	tsize, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	txs := make(Transactions, 0)
	for i := 0; i < int(tsize); i++ {
		temptx := new(Transaction)
		err := temptx.Deserialize(r)
		if err != nil {
			return err
		}
		txs = append(txs, temptx)
	}
	this.Transactions = txs
	sigdata := new(SigData)

	err = sigdata.Deserialize(r)
	if err != nil {
		return err
	}

	this.Sigs = sigdata

	if eof {
		return errors.ERR_TXRAW_EOF
	}
	return nil
}

/*
func (b *Block) Serialization(sk *sink.ZeroCopySink) error {

	err := b.Header.Serialization(sk)
	if err != nil {
		return err
	}
	sk.WriteUint32(uint32(len(b.Transactions)))
	for _, transaction := range b.Transactions {
		err := transaction.Serialization(sk)
		if err != nil {
			return err
		}
	}
	return nil
}
*/
/*func (b *Block) Deserialization(source *sink.ZeroCopySource) error {
	if b.Header == nil {
		b.Header = new(Header)
	}
	err := b.Header.Deserialization(source)
	if err != nil {
		return err
	}
	length, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	var hashes []common.Hash
	for i := uint32(0); i < length; i++ {
		transaction := new(Transaction)
		// note currently all transaction in the block shared the same source
		err := transaction.Deserialization(source)
		if err != nil {
			return err
		}
		txhash := transaction.Hash()
		hashes = append(hashes, txhash)
		b.Transactions = append(b.Transactions, transaction)
	}

	b.Header.TxRoot = common.ComputeMerkleRoot(hashes)

	return nil
}
*/
func (b *Block) Hash() common.Hash {
	return b.Header.Hash()
}

func (b *Block) RebuildMerkleRoot() {
	txs := b.Transactions
	hashes := make([]common.Hash, 0, len(txs))
	for _, tx := range txs {
		hashes = append(hashes, tx.Hash())
	}
	hash := common.ComputeMerkleRoot(hashes)
	b.Header.TxRoot = hash
}
