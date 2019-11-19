package extstates

import (
	"encoding/binary"
	"io"
	"sort"

	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	common2 "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

type LeagueBlockSimple struct {
	*orgtypes.Header
	EnergyUsed *big.Int
	TxHashes common.HashArray
}

type LeaguegBlocks []*LeagueBlockSimple

func (s LeaguegBlocks) Len() int           { return len(s) }
func (s LeaguegBlocks) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s LeaguegBlocks) Less(i, j int) bool { return s[i].Hash().String() < s[j].Hash().String() }
func (s LeaguegBlocks) GetHashRoot() common.Hash {
	sort.Sort(s)
	hashes := make([]common.Hash, 0, s.Len())
	for _, v := range s {
		hashes = append(hashes, v.Hash())
	}
	return common.ComputeMerkleRoot(hashes)
}

func (this *LeagueBlockSimple) Serialize(w io.Writer) error {
	/*	err := this.Header.Serialize(w)
		if err != nil {
			return err
		}*/
	sk := sink.NewZeroCopySink(nil)
	err := this.Header.Serialization(sk)
	if err != nil {
		return err
	}
	gasUsed, err := sink.BigIntToComplex(this.EnergyUsed)
	if err != nil {
		return err
	}
	sk.WriteComplex(gasUsed)
	t, err := sink.HashArrayToComplex(this.TxHashes)
	if err != nil {
		return err
	}
	sk.WriteComplex(t)
	_, err = w.Write(sk.Bytes())
	return err
}

func (this *LeagueBlockSimple) Deserialize(r io.Reader) error {
	if this == nil {
		this = &LeagueBlockSimple{}
	}
	if this.Header == nil {
		this.Header = &orgtypes.Header{}
	}
	this.Header.Deserialize(r)
	cp1, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	this.EnergyUsed, err = cp1.ComplexToBigInt()
	if err != nil {
		return err
	}
	cp2, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	this.TxHashes, err = cp2.ComplexToHashArray()
	return err

	/*	buff, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		var eof bool
		source := sink.NewZeroCopySource(buff)
		_, eof = source.NextByte()
		this.Header = new(orgtypes.Header)
		this.Header.Version, eof = source.NextByte()
		prevBlockHash, eof := source.NextHash()

		this.Header.PrevBlockHash = prevBlockHash
		txRoot, eof := source.NextHash()
		this.Header.TxRoot = txRoot
		stateRoot, eof := source.NextHash()
		this.Header.StateRoot = stateRoot

		this.Header.ReceiptsRoot, eof = source.NextHash()
		this.Header.Timestamp, eof = source.NextInt64()
		this.Header.Height, eof = source.NextUint64()
		if eof {
			return errors.ERR_TXRAW_EOF
		}
		return nil*/
}

func (this *LeagueBlockSimple) GetKey() []byte {
	buff := make([]byte, 1+common.AddrLength+8)
	buff[0] = byte(common2.EXT_LEAGUE_DATA_BLOCK)
	copy(buff[1:common.AddrLength+1], this.LeagueId[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength+1:], this.Header.Height)
	return buff
}
