package common

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sixexorg/magnetic-ring/crypto"	
	sink "github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/common/serialization"	
	"github.com/sixexorg/magnetic-ring/log"
)

type ConsensusPayload struct {
	Version         uint32
	PrevHash        common.Hash
	Height          uint32
	BookkeeperIndex uint16
	Timestamp       uint32
	Data            []byte
	Owner           crypto.PublicKey
	Signature       []byte
	PeerId          uint64
	hash            common.Hash
}

//get the consensus payload hash
func (this *ConsensusPayload) Hash() common.Hash {
	return common.Hash{}
}

//Check whether header is correct
func (this *ConsensusPayload) Verify() error {
	buf := new(bytes.Buffer)
	err := this.SerializeUnsigned(buf)
	if err != nil {
		return err
	}
	_,err = this.Owner.Verify(common.Sha256(buf.Bytes()), this.Signature)
	// err = signature.Verify(this.Owner, buf.Bytes(), this.Signature)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetVerifyFail, fmt.Sprintf("signature verify error. buf:%v", buf))
	}
	return nil
}

//serialize the consensus payload
func (this *ConsensusPayload) ToArray() []byte {
	b := new(bytes.Buffer)
	err := this.Serialize(b)
	if err != nil {
		log.Error("consensus payload serialize error in ToArray()","payload", this)
		return nil
	}
	return b.Bytes()
}

//return inventory type
func (this *ConsensusPayload) InventoryType() common.InventoryType {
	return common.CONSENSUS
}

func (this *ConsensusPayload) GetMessage() []byte {
	//TODO: GetMessage
	//return sig.GetHashData(cp)
	return []byte{}
}

func (this *ConsensusPayload) Type() common.InventoryType {

	//TODO:Temporary add for Interface signature.SignableData use.
	return common.CONSENSUS
}

func (this *ConsensusPayload) Serialization(sink *sink.ZeroCopySink) error {
	this.serializationUnsigned(sink)
	// buf := keypair.SerializePublicKey(this.Owner)
	buf := this.Owner.Bytes()
	sink.WriteVarBytes(buf)
	sink.WriteVarBytes(this.Signature)

	return nil
}

//Serialize message payload
func (this *ConsensusPayload) Serialize(w io.Writer) error {
	err := this.SerializeUnsigned(w)
	if err != nil {
		return err
	}
	// buf := keypair.SerializePublicKey(this.Owner)
	buf := this.Owner.Bytes()
	err = serialization.WriteVarBytes(w, buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write publickey error. publickey buf:%v", buf))
	}

	err = serialization.WriteVarBytes(w, this.Signature)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write Signature error. Signature:%v", this.Signature))
	}

	return nil
}

//Deserialize message payload
func (this *ConsensusPayload) Deserialization(source *sink.ZeroCopySource) error {
	err := this.deserializationUnsigned(source)
	if err != nil {
		return err
	}
	buf, _, irregular, eof := source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return sink.ErrIrregularData
	}

	// this.Owner, err = keypair.DeserializePublicKey(buf)
	this.Owner, err = crypto.UnmarshalPubkey(buf) 
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "deserialize publickey error")
	}

	this.Signature, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return sink.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}

//Deserialize message payload
func (this *ConsensusPayload) Deserialize(r io.Reader) error {
	err := this.DeserializeUnsigned(r)
	if err != nil {
		return err
	}
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "read buf error")
	}
	// this.Owner, err = keypair.DeserializePublicKey(buf)
	this.Owner, err = crypto.UnmarshalPubkey(buf)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "deserialize publickey error")
	}

	this.Signature, err = serialization.ReadVarBytes(r)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "read Signature error")
	}

	return err
}

func (this *ConsensusPayload) serializationUnsigned(sink *sink.ZeroCopySink) {
	sink.WriteUint32(this.Version)
	sink.WriteHash(this.PrevHash)
	sink.WriteUint32(this.Height)
	sink.WriteUint16(this.BookkeeperIndex)
	sink.WriteUint32(this.Timestamp)
	sink.WriteVarBytes(this.Data)
}

//Serialize message payload
func (this *ConsensusPayload) SerializeUnsigned(w io.Writer) error {
	err := serialization.WriteUint32(w, this.Version)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. version:%v", this.Version))
	}
	err = this.PrevHash.Serialize(w)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("serialize error. PrevHash:%v", this.PrevHash))
	}
	err = serialization.WriteUint32(w, this.Height)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. Height:%v", this.Height))
	}
	err = serialization.WriteUint16(w, this.BookkeeperIndex)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. BookkeeperIndex:%v", this.BookkeeperIndex))
	}
	err = serialization.WriteUint32(w, this.Timestamp)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. Timestamp:%v", this.Timestamp))
	}
	err = serialization.WriteVarBytes(w, this.Data)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. Data:%v", this.Data))
	}
	return nil
}

func (this *ConsensusPayload) deserializationUnsigned(source *sink.ZeroCopySource) error {
	var irregular, eof bool
	this.Version, eof = source.NextUint32()
	this.PrevHash, eof = source.NextHash()
	this.Height, eof = source.NextUint32()
	this.BookkeeperIndex, eof = source.NextUint16()
	this.Timestamp, eof = source.NextUint32()
	this.Data, _, irregular, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return sink.ErrIrregularData
	}

	return nil
}

//Deserialize message payload
func (this *ConsensusPayload) DeserializeUnsigned(r io.Reader) error {
	var err error
	this.Version, err = serialization.ReadUint32(r)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "read version error")
	}

	preBlock := new(common.Hash)
	err = preBlock.Deserialize(r)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "read preBlock error")
	}
	this.PrevHash = *preBlock

	this.Height, err = serialization.ReadUint32(r)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "read Height error")
	}

	this.BookkeeperIndex, err = serialization.ReadUint16(r)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "read BookkeeperIndex error")
	}

	this.Timestamp, err = serialization.ReadUint32(r)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "read Timestamp error")
	}

	this.Data, err = serialization.ReadVarBytes(r)
	if err != nil {

		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "read Data error")
	}

	return nil
}
