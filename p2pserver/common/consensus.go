package common

import (
	comm "github.com/sixexorg/magnetic-ring/common"
	sink "github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/crypto"
)

type P2PConsPld struct {
	Version         uint32
	PrevHash        comm.Hash
	Height          uint32
	BookkeeperIndex uint16
	Timestamp       uint32
	Data            []byte
	Owner           []byte
	Signature       []byte
	PeerId          uint64
	hash            comm.Hash
}

type Consensus struct {
	Cons ConsensusPayload
}

func ConsensusToP2PConsPld(cons *Consensus) *P2PConsPld{
	res := new(P2PConsPld)
	res.Version = cons.Cons.Version
	res.PrevHash = cons.Cons.PrevHash
	res.Height = cons.Cons.Height
	res.BookkeeperIndex = cons.Cons.BookkeeperIndex
	res.Timestamp = cons.Cons.Timestamp
	res.Data = cons.Cons.Data
	res.Owner = cons.Cons.Owner.Bytes()
	res.Signature = cons.Cons.Signature
	res.PeerId = cons.Cons.PeerId
	res.hash = cons.Cons.hash

	return res
}

func P2PConsPldToConsensus(cons *P2PConsPld) (*Consensus,error){
	own,err := crypto.UnmarshalPubkey(cons.Owner) 
	// keypair.DeserializePublicKey(cons.Owner)
	if err != nil {
		return nil,err
	}
	res := ConsensusPayload{
		Version: cons.Version,
		PrevHash: cons.PrevHash,
		Height: cons.Height,
		BookkeeperIndex: cons.BookkeeperIndex,
		Timestamp: cons.Timestamp,
		Data: cons.Data,
		Owner: own,
		Signature: cons.Signature,
		PeerId: cons.PeerId,
		hash: cons.hash,
	}

	return &Consensus{
		Cons:res,
	},nil
}

func (this *P2PConsPld) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *P2PConsPld) CmdType() string {
	return CONSENSUS_TYPE
}

//Deserialize message payload
func (this *P2PConsPld) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

//Serialize message payload
func (this *Consensus) Serialization(sink *sink.ZeroCopySink) error {
	return this.Cons.Serialization(sink)
}

func (this *Consensus) CmdType() string {
	return CONSENSUS_TYPE
}

//Deserialize message payload
func (this *Consensus) Deserialization(source *sink.ZeroCopySource) error {
	return this.Cons.Deserialization(source)
}

type PeerAnn struct {
	PeerId  uint64
	Height  uint64
}

type NotifyBlk struct {
	BlkHeight uint64
	BlkHash   comm.Hash
}