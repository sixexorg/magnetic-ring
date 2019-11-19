package types

import (
	"fmt"
	"io"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
)

type Block struct {
	Header       *Header
	Transactions Transactions
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
	return nil
}

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

/*func (b *Block) Deserialize(source *sink.ZeroCopySource) error {
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
}*/
func (b *Block) Deserialize(r io.Reader) error {
	header := new(Header)
	err := header.Deserialize(r)
	if err != nil {
		return err
	}
	length, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	var hashes []common.Hash
	for i := uint32(0); i < length; i++ {
		transaction := new(Transaction)
		// note currently all transaction in the block shared the same source
		err := transaction.Deserialize(r)
		if err != nil {
			return err
		}
		txhash := transaction.Hash()
		hashes = append(hashes, txhash)
		b.Transactions = append(b.Transactions, transaction)
	}
	return nil
}

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
