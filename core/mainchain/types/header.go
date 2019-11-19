package types

import (
	"io"

	"github.com/sixexorg/magnetic-ring/common/sink"

	"bytes"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
)

const (
	HWidth uint64 = 1
)

type Header struct {
	Version          byte        //1
	PrevBlockHash    common.Hash //2
	BlockRoot        common.Hash //3 txroot aggregation merkle
	TxRoot           common.Hash //4
	LeagueRoot       common.Hash //5
	StateRoot        common.Hash //6
	ReceiptsRoot     common.Hash //7
	Timestamp        uint64      //8
	Height           uint64      //9
	ConsensusData    uint64      //10
	ConsensusPayload []byte      //11
	NextBookkeeper common.Address //12
	//Bookkeeper     pubkey
	blockHash common.Hash

/*	Lv1 uint64
	Lv2 uint64
	Lv3 uint64
	Lv4 uint64
	Lv5 uint64
	Lv6 uint64
	Lv7 uint64
	Lv8 uint64
	Lv9 uint64*/
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
	sk.WriteByte(h.Version)              //1
	sk.WriteHash(h.PrevBlockHash)        //2
	sk.WriteHash(h.BlockRoot)            //3
	sk.WriteHash(h.TxRoot)               //4
	sk.WriteHash(h.LeagueRoot)           //5
	sk.WriteHash(h.StateRoot)            //6
	sk.WriteHash(h.ReceiptsRoot)         //7
	sk.WriteUint64(h.Timestamp)          //8
	sk.WriteUint64(h.Height)             //9
	sk.WriteUint64(h.ConsensusData)      //10
	sk.WriteVarBytes(h.ConsensusPayload) //11
	/*if h.Height%HWidth == 0 {
		sk.WriteUint64(h.Lv1)
		sk.WriteUint64(h.Lv2)
		sk.WriteUint64(h.Lv3)
		sk.WriteUint64(h.Lv4)
		sk.WriteUint64(h.Lv5)
		sk.WriteUint64(h.Lv6)
		sk.WriteUint64(h.Lv7)
		sk.WriteUint64(h.Lv8)
		sk.WriteUint64(h.Lv9)
	}*/
	//sk.WriteAddress(h.NextBookkeeper)    //12
	return nil
}

func (h *Header) Deserialize(r io.Reader) error {
	var err error
	h.Version, err = serialization.ReadByte(r) //1
	if err != nil {
		return err
	}
	prevBlockHash, err := serialization.ReadBytes(r, common.HashLength) //2
	if err != nil {
		return err
	}
	h.PrevBlockHash, _ = common.ParseHashFromBytes(prevBlockHash)

	blockRoot, err := serialization.ReadBytes(r, common.HashLength) //3
	if err != nil {
		return err
	}
	h.BlockRoot, _ = common.ParseHashFromBytes(blockRoot)

	txRoot, err := serialization.ReadBytes(r, common.HashLength) //4
	if err != nil {
		return err
	}
	h.TxRoot, _ = common.ParseHashFromBytes(txRoot)

	leagueRoot, err := serialization.ReadBytes(r, common.HashLength) //5
	if err != nil {
		return err
	}
	h.LeagueRoot, _ = common.ParseHashFromBytes(leagueRoot)

	stateRoot, err := serialization.ReadBytes(r, common.HashLength) //6
	if err != nil {
		return err
	}
	h.StateRoot, _ = common.ParseHashFromBytes(stateRoot)

	receiptsRoot, err := serialization.ReadBytes(r, common.HashLength) //7
	if err != nil {
		return err
	}
	h.ReceiptsRoot, _ = common.ParseHashFromBytes(receiptsRoot)

	timp, err := serialization.ReadUint64(r) //8
	if err != nil {
		return err
	}
	h.Timestamp = timp
	h.Height, err = serialization.ReadUint64(r) //9
	if err != nil {
		return err
	}
	h.ConsensusData, err = serialization.ReadUint64(r) //10
	if err != nil {
		return err
	}
	h.ConsensusPayload, err = serialization.ReadVarBytes(r) //11
	if err != nil {
		return err
	}
	/*if h.Height%HWidth == 0 {
		h.Lv1, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
		h.Lv2, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
		h.Lv3, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
		h.Lv4, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
		h.Lv5, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
		h.Lv6, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
		h.Lv7, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
		h.Lv8, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
		h.Lv9, err = serialization.ReadUint64(r)
		if err != nil {
			return err
		}
	}*/
	return err
}

/*func (h *Header) Deserialization(source *sink.ZeroCopySource) error {
	var eof bool
	h.Version, eof = source.NextByte()                    //1
	h.PrevBlockHash, eof = source.NextHash()              //2
	h.BlockRoot, eof = source.NextHash()                  //3
	h.TxRoot, eof = source.NextHash()                     //4
	h.LeagueRoot, eof = source.NextHash()                 //5
	h.StateRoot, eof = source.NextHash()                  //6
	h.ReceiptsRoot, eof = source.NextHash()               //7
	h.Timestamp, eof = source.NextInt64()                 //8
	h.Height, eof = source.NextUint64()                   //9
	h.ConsensusData, eof = source.NextUint64()            //10
	h.ConsensusPayload, _, _, eof = source.NextVarBytes() //11
	//h.NextBookkeeper, eof = source.NextAddress() //12
	if eof {
		return nil
	}
	return nil
}*/

func (h *Header) Hash() common.Hash {
	buff := new(bytes.Buffer)
	h.Serialize(buff)
	hash, _ := common.ParseHashFromBytes(common.Sha256(buff.Bytes()))
	h.blockHash = hash
	return hash
}
