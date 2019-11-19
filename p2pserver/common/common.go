package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/rlp"
)

type Msg struct {
	Code       uint64
	Type       string
	Size       uint32 // size of the paylod
	Payload    io.Reader
	ReceivedAt time.Time
}

// for test ......
func EncAndDec(msgcode uint64, data interface{}) error {
	size, r, err := rlp.EncodeToReader(data)
	if err != nil {
		return err
	}
	rawmsg := Msg{Code: msgcode, Size: uint32(size), Payload: r}

	codestr := TransCodeUToStr(rawmsg.Code)
	fmt.Println(" ******** codestr:", codestr)
	val, err := MakeEmptyMessage(codestr)
	if err != nil {
		return err
	}

	s := rlp.NewStream(rawmsg.Payload, uint64(rawmsg.Size))
	if err := s.Decode(val); err != nil {
		return newPeerError(errInvalidMsg, "(code %x) (size %d) %v", rawmsg.Code, rawmsg.Size, err)
		// return errors.New("(code %x) (size %d) %v")
	}
	return nil
}

//
func Send(w MsgWriter, msgcode uint64, data interface{}) error {
	size, r, err := rlp.EncodeToReader(data)
	if err != nil {
		return err
	}
	return w.WriteMsg(Msg{Code: msgcode, Size: uint32(size), Payload: r})
}

// Decode parses the RLP content of a message into
// the given value, which must be a pointer.
//
// For the decoding rules, please see package rlp.
func (msg Msg) Decode(val interface{}) error {
	s := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if err := s.Decode(val); err != nil {
		return newPeerError(errInvalidMsg, "(code %d) (size %d) %v", msg.Code, msg.Size, err)
		// return errors.New("(code %x) (size %d) %v")
	}
	return nil
}

func (msg Msg) String() string {
	return fmt.Sprintf("msg #%v (%v bytes)", msg.Code, msg.Size)
}

// Discard reads any remaining payload data into a black hole.
func (msg Msg) Discard() error {
	_, err := io.Copy(ioutil.Discard, msg.Payload)
	return err
}
func (msg *Msg) DistinctSk() (common.Hash, error) {
	buff, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return common.Hash{}, err
	}
	return common.ParseHashFromBytes(common.Sha256(buff))
}

type MsgReader interface {
	ReadMsg() (Msg, error)
}

type MsgWriter interface {
	// WriteMsg sends a message. It will block until the message's
	// Payload has been consumed by the other end.
	//
	// Note that messages can be sent only once because their
	// payload reader is drained.
	WriteMsg(Msg) error
}

// MsgReadWriter provides reading and writing of encoded messages.
// Implementations should ensure that ReadMsg and WriteMsg can be
// called simultaneously from multiple goroutines.
type MsgReadWriter interface {
	MsgReader
	MsgWriter
	Close(err ...error)
}
