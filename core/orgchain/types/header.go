package types

import (
	"bytes"
	"io"

	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
)

const (
	HWidth uint64 = 1
)

var (
	BonusRate = big.NewRat(45, 100)
)

type Header struct {
	Version       byte
	PrevBlockHash common.Hash
	BlockRoot     common.Hash //txroot aggregation merkle
	LeagueId      common.Address

	Bonus  uint64
	TxRoot common.Hash

	StateRoot    common.Hash
	ReceiptsRoot common.Hash

	Timestamp  uint64
	Height     uint64
	Difficulty *big.Int

	blockHash common.Hash
	//// by zp
	Coinbase common.Address
	Extra    []byte
}

func NewOrgHeader(version byte, leagueAddr common.Address, height uint64, blkroot, prvblkhash, blkhash common.Hash) *Header {
	hdr := new(Header)

	hdr.Height = height
	hdr.BlockRoot = blkroot
	hdr.PrevBlockHash = prvblkhash
	hdr.Version = version
	hdr.LeagueId = leagueAddr
	hdr.blockHash = blkhash
	return hdr
}

func (h *Header) SetDifficulty(d *big.Int) {
	h.Difficulty = big.NewInt(0).Set(d)
}

//Serialize the blockheader
func (h *Header) Serialize(w io.Writer) error {
	sk := sink.NewZeroCopySink(nil)
	err := h.Serialization(sk)
	if err != nil {
		return err

	}
	w.Write(sk.Bytes())
	return nil
}

func (h *Header) Serialization(sk *sink.ZeroCopySink) error {
	sk.WriteByte(h.Version)
	sk.WriteAddress(h.LeagueId)
	sk.WriteHash(h.PrevBlockHash)
	sk.WriteHash(h.TxRoot)
	sk.WriteHash(h.StateRoot)
	sk.WriteHash(h.ReceiptsRoot)
	sk.WriteUint64(h.Timestamp)
	cp, err := sink.BigIntToComplex(h.Difficulty)
	if err != nil {
		return err
	}
	sk.WriteComplex(cp)
	sk.WriteUint64(h.Height)
	sk.WriteAddress(h.Coinbase)
	if h.Height%HWidth == 0 {
		sk.WriteUint64(h.Bonus)
	}

	//sk.WriteBytes(h.Extra)
	return nil
}

func (h *Header) Deserialize(r io.Reader) error {
	var err error
	if h == nil {
		h = &Header{}
	}
	h.Version, err = serialization.ReadByte(r)
	if err != nil {
		return err
	}
	leagueId, err := serialization.ReadBytes(r, common.AddrLength)
	if err != nil {
		return err
	}
	copy(h.LeagueId[:], leagueId)
	prevBlockHash, err := serialization.ReadBytes(r, common.HashLength)
	if err != nil {
		return err
	}
	h.PrevBlockHash, _ = common.ParseHashFromBytes(prevBlockHash)
	txRoot, err := serialization.ReadBytes(r, common.HashLength)
	if err != nil {
		return err
	}
	h.TxRoot, _ = common.ParseHashFromBytes(txRoot)
	stateRoot, err := serialization.ReadBytes(r, common.HashLength)
	if err != nil {
		return err
	}
	h.StateRoot, _ = common.ParseHashFromBytes(stateRoot)

	receiptsRoot, err := serialization.ReadBytes(r, common.HashLength)
	if err != nil {
		return err
	}
	h.ReceiptsRoot, _ = common.ParseHashFromBytes(receiptsRoot)

	timp, err := serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	h.Timestamp = timp
	cp, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	diff, err := cp.ComplexToBigInt()
	if err != nil {
		return err
	}
	h.Difficulty = diff
	h.Height, err = serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	addrByte, err := serialization.ReadBytes(r, common.AddrLength)
	if err != nil {
		return err
	}
	addr := common.Bytes2Address(addrByte)
	h.Coinbase = addr
	if h.Height%HWidth == 0 {
		h.Bonus, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
	}
	return err
}

func (h *Header) Deserialization(source *sink.ZeroCopySource) error {
	var eof bool
	h.Version, eof = source.NextByte()
	h.LeagueId, eof = source.NextAddress()
	h.PrevBlockHash, eof = source.NextHash()
	h.TxRoot, eof = source.NextHash()
	h.StateRoot, eof = source.NextHash()
	h.ReceiptsRoot, eof = source.NextHash()
	h.Timestamp, eof = source.NextUint64()
	cp, eof := source.NextComplex()
	diff, err := cp.ComplexToBigInt()
	if err != nil {
		return err
	}
	h.Difficulty = diff
	h.Height, eof = source.NextUint64()
	h.Coinbase, eof = source.NextAddress()
	if h.Height%HWidth == 0 {
		h.Bonus, eof = source.NextUint64()
	}
	if eof {
		return nil
	}
	return nil
}

func (h *Header) Hash() common.Hash {
	buff := new(bytes.Buffer)
	h.Serialize(buff)
	hash, _ := common.ParseHashFromBytes(common.Sha256(buff.Bytes()))
	h.blockHash = hash
	return hash
}
