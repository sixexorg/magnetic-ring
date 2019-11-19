package types

import (
	"fmt"
	"io"
	"math/big"
	"reflect"

	"sort"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/errors"
)

const (
	MSG        TransactionType = 0x01
	TransferUT TransactionType = 0x10

	ReplyVote      TransactionType = 0x20
	VoteIncreaseUT TransactionType = 0x21
	VoteApply      TransactionType = 0x22

	GetBonus TransactionType = 0x50

	//contact
	Deploy TransactionType = 0x70
	Invoke TransactionType = 0x71

	//transactions that point to the public chain
	Join         TransactionType = 0x80
	Leave        TransactionType = 0x81
	EnergyFromMain TransactionType = 0xa0
	EnergyToMain   TransactionType = 0xa1
	RaiseUT      TransactionType = 0xa2

	TxVersion = 0x01
)

//for vote comfirm
const (
	Against    VoteAttitude = 0
	Agree      VoteAttitude = 1
	Abstention VoteAttitude = 2
)

func init() {
	txMap = make(map[TransactionType]common.Arguments)
	initTxMap()
	sortTxMap()
}

var txMap map[TransactionType]common.Arguments

func sortTxMap() {
	for k, _ := range txMap {
		sort.Sort(txMap[k])
	}
}

//init transactionType required properties
func initTxMap() {
	commonTxMapInit()
	voteTxMapInit()
	systemTxMapInit()
	bonusTxMapInit()

}
func GetMethod(transactionType TransactionType) common.Arguments {
	return txMap[transactionType]
}

//todo AAttributes are currently incomplete
type TxData struct {
	//3
	From     common.Address `json:"from"`
	Account  common.Address `json:"account"`
	LeagueId common.Address `json:"league_id"`
	//2
	Froms *common.TxIns `json:"froms"`
	Tos   *common.TxOuts `json:"tos"`
	//4
	Msg    common.Hash `json:"msg"`
	TxHash common.Hash `json:"tx_hash"`
	Name   common.Hash `json:"name"`
	VoteId common.Hash `json:"vote_id"`
	Symbol common.Symbol `json:"symbol"`

	//11
	VoteResult bool `json:"vote_result"`
	VoteReply  uint8 `json:"vote_reply"`
	Rate       uint32 `json:"rate"`
	Nonce      uint64 `json:"nonce"`
	Amount     uint64 `json:"amount"`
	Start      uint64 `json:"start"`
	End        uint64 `json:"end"`

	MinBox     *big.Int `json:"min_box"`
	MetaBox    *big.Int `json:"meta_box"`
	UnlockEnergy *big.Int `json:"unlock_energy"`
	GasLimit   *big.Int `json:"gas_limit"`
	Fee        *big.Int `json:"fee"`
	Energy       *big.Int `json:"b_gas"`
}

func (data *TxData) Serialize(txp TransactionType, version byte) (raw []byte, err error) {
	if data == nil {
		return nil, errors.ERR_TX_TXDATA_NIL
	}
	sk := sink.NewZeroCopySink(nil)
	sk.WriteByte(version)
	sk.WriteByte(byte(txp))
	args := GetMethod(txp)
	txdataV := reflect.ValueOf(data)
	for _, v := range args {
		v2 := txdataV.Elem().FieldByName(v.Name)
		if v.Required {
			a := reflect.Zero(v2.Type())
			if reflect.DeepEqual(a.Interface(), v2.Interface()) {
				fmt.Printf("name=%s\n",v.Name)
				return nil, errors.ERR_TX_PARAM_REQUIRD
			}
		}
		err := sink.ZeroCopySinkAppend(sk, v2)
		if err != nil {
			return nil, err
		}
	}
	return sk.Bytes(), nil
}
func (tx *Transaction) Serialize(w io.Writer) error {
	if tx.Filled == false || len(tx.Raw) == 0 {
		raw, err := tx.TxData.Serialize(tx.TxType, tx.Version)
		if err != nil {
			return err
		}
		tx.Raw = raw
		tx.Filled = true
	}
	_, err := w.Write(tx.Raw)
	return err
}

/*func (tx *Transaction) Deserialize2(r io.Reader) error {
	byteArr, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	err = tx.Deserialization(sink.NewZeroCopySource(byteArr))
	return err
}*/
func (tx *Transaction) Deserialize(r io.Reader) error {
	var err error
	if tx.Version, err = serialization.ReadByte(r); err != nil {
		return err
	}

	var txType byte
	if txType, err = serialization.ReadByte(r); err != nil {
		return err
	}
	tx.TxType = TransactionType(txType)
	tx.TxData = &TxData{}
	args := GetMethod(tx.TxType)
	err = tx.setProperty(args, r)
	return err
}

func (tx *Transaction) setProperty(args common.Arguments, r io.Reader) error {
	//tx.TxData = &TxData{}
	for _, v := range args {
		out, err := serialization.ReadRealase(r, v.Type)
		if err != nil {
			return err
		}
		switch v.Name {
		case "From":
			tx.TxData.From = out.(common.Address)
		case "Froms":
			tx.TxData.Froms = out.(*common.TxIns)
		case "Tos":
			tx.TxData.Tos = out.(*common.TxOuts)
		case "Msg":
			tx.TxData.Msg = out.(common.Hash)
		case "TxHash":
			tx.TxData.TxHash = out.(common.Hash)
		case "VoteResult":
			tx.TxData.VoteResult = out.(bool)
		case "MinBox":
			tx.TxData.MinBox = out.(*big.Int)
		case "MetaBox":
			tx.TxData.MetaBox = out.(*big.Int)
		case "Rate":
			tx.TxData.Rate = out.(uint32)
		case "LeagueId":
			tx.TxData.LeagueId = out.(common.Address)
		case "Name":
			tx.TxData.Name = out.(common.Hash)
		case "UnlockEnergy":
			tx.TxData.UnlockEnergy = out.(*big.Int)
		case "GasLimit":
			tx.TxData.GasLimit = out.(*big.Int)
		case "Nonce":
			tx.TxData.Nonce = out.(uint64)
		case "Fee":
			tx.TxData.Fee = out.(*big.Int)
		case "VoteId":
			tx.TxData.VoteId = out.(common.Hash)
		case "VoteReply":
			tx.TxData.VoteReply = out.(uint8)
		case "Account":
			tx.TxData.Account = out.(common.Address)
		case "Amount":
			tx.TxData.Amount = out.(uint64)
		case "Energy":
			tx.TxData.Energy = out.(*big.Int)
		case "Start":
			tx.TxData.Start = out.(uint64)
		case "End":
			tx.TxData.End = out.(uint64)
		case "Symbol":
			tx.TxData.Symbol = out.(common.Symbol)

		}
	}
	return nil
}

func (tx *Transaction) ToRaw() error {
	raw, err := tx.TxData.Serialize(tx.TxType, tx.Version)
	if err != nil {
		return err
	}
	tx.Raw = raw
	return nil
}
func (tx *Transaction) Serialization(sk *sink.ZeroCopySink) error {
	if tx.Filled == false || len(tx.Raw) == 0 {
		return errors.ERR_SERIALIZATION_FAILED
	}
	sk.WriteBytes(tx.Raw)
	return nil
}

func (tx *Transaction) Deserialization(source *sink.ZeroCopySource) error {
	if tx == nil {
		tx = &Transaction{}
	}
	pstart := source.Pos()
	var eof bool
	tx.Version, eof = source.NextByte()
	var txType byte
	txType, eof = source.NextByte()
	tx.TxType = TransactionType(txType)
	args := GetMethod(tx.TxType)
	tx.TxData = &TxData{}
	for _, v := range args {
		out, _, eof := sink.ZeroCopySourceRelease(source, v.Type)
		if eof {
			return errors.ERR_TX_DATA_OVERFLOW
		}
		switch v.Name {
		case "From":
			tx.TxData.From = out.(common.Address)
		case "Froms":
			tx.TxData.Froms = out.(*common.TxIns)
		case "Tos":
			tx.TxData.Tos = out.(*common.TxOuts)
		case "Msg":
			tx.TxData.Msg = out.(common.Hash)
		case "TxHash":
			tx.TxData.TxHash = out.(common.Hash)
		case "VoteResult":
			tx.TxData.VoteResult = out.(bool)
		case "MinBox":
			tx.TxData.MinBox = out.(*big.Int)
		case "MetaBox":
			tx.TxData.MetaBox = out.(*big.Int)
		case "Rate":
			tx.TxData.Rate = out.(uint32)
		case "LeagueId":
			tx.TxData.LeagueId = out.(common.Address)
		case "Name":
			tx.TxData.Name = out.(common.Hash)
		case "UnlockEnergy":
			tx.TxData.UnlockEnergy = out.(*big.Int)
		case "GasLimit":
			tx.TxData.GasLimit = out.(*big.Int)
		case "Nonce":
			tx.TxData.Nonce = out.(uint64)
		case "Fee":
			tx.TxData.Fee = out.(*big.Int)
		case "VoteId":
			tx.TxData.VoteId = out.(common.Hash)
		case "VoteReply":
			tx.TxData.VoteReply = out.(uint8)
		case "Account":
			tx.TxData.Account = out.(common.Address)
		case "Amount":
			tx.TxData.Amount = out.(uint64)
		case "Energy":
			tx.TxData.Energy = out.(*big.Int)
		case "Start":
			tx.TxData.Start = out.(uint64)
		case "End":
			tx.TxData.End = out.(uint64)
		}
	}
	if eof {
		return errors.ERR_TXRAW_EOF
	}
	pos := source.Pos()
	lenUnsigned := pos - pstart
	source.BackUp(lenUnsigned)
	tx.Raw, _ = source.NextBytes(lenUnsigned)
	tx.Filled = true
	return nil
}
