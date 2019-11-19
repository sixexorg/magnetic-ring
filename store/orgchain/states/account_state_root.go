package states

import (
	"encoding/binary"
	"io"
	"io/ioutil"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/errors"
	mcom "github.com/sixexorg/magnetic-ring/store/orgchain/common"
)

type AccountStateRoot struct {
	Height   uint64
	Accounts []common.Address
}

func NewAccountStateRoot(height uint64, accounts []common.Address) *AccountStateRoot {
	return &AccountStateRoot{
		Height:   height,
		Accounts: accounts,
	}
}
func (this *AccountStateRoot) Serialize(w io.Writer) error {
	sk := sink.NewZeroCopySink(nil)
	cp, err := sink.AddressesToComplex(this.Accounts)
	if err != nil {
		return err
	}
	sk.WriteComplex(cp)
	w.Write(sk.Bytes())
	return nil
}

func (this *AccountStateRoot) Deserialize(r io.Reader) error {
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var eof bool
	source := sink.NewZeroCopySource(buff)
	_, eof = source.NextByte()
	this.Height, eof = source.NextUint64()
	cp, eof := source.NextComplex()
	this.Accounts, err = cp.ComplexToTxAddresses()
	if err != nil {
		return err
	}
	if eof {
		return errors.ERR_TXRAW_EOF
	}
	return nil
}

func (this *AccountStateRoot) GetKey() []byte {
	return GetAccountStateRootKey(this.Height)
}

func GetAccountStateRootKey(height uint64) []byte {
	buff := make([]byte, 1+8)
	buff[0] = byte(mcom.ST_ACCOUNT_ROOT)
	binary.LittleEndian.PutUint64(buff[1:], height)
	return buff
}
