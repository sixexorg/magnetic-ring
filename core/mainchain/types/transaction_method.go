package types

import (
	"fmt"
	"io"
	"math/big"

	"reflect"

	"github.com/sixexorg/magnetic-ring/common/sink"

	"sort"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/errors"
)

const (
	MSG TransactionType = 0x01

	TransferBox  TransactionType = 0x10
	TransferEnergy TransactionType = 0x11
	ReceiveEnergy  TransactionType = 0x12

	StartVote   TransactionType = 0x20
	ConfirmVote TransactionType = 0x21

	//league
	CreateLeague  TransactionType = 0x50
	RaiseUT       TransactionType = 0x51
	ApplyPass     TransactionType = 0x52
	JoinLeague    TransactionType = 0x53
	RecycleMinbox TransactionType = 0x54
	ExitLeague    TransactionType = 0x55
	EnergyToLeague  TransactionType = 0x56
	EnergyToMain    TransactionType = 0x57

	GetBonus TransactionType = 0x70
	//Authorized to miner
	AuthX TransactionType = 0xA0

	LockLeague   TransactionType = 0xB0
	UnlockLeague TransactionType = 0xB1

	PlaintextClear TransactionType = 0xC1

	ConsensusLeague TransactionType = 0xD1

	TxVersion = 0x01
)

var txMap map[TransactionType]common.Arguments

func init() {
	txMap = make(map[TransactionType]common.Arguments)

	initTxMap()
	sortTxMap()
}
func sortTxMap() {
	for k, _ := range txMap {
		sort.Sort(txMap[k])
	}
}

//init transactionType required properties
func initTxMap() {
	commonTxMapInit()
	systemTxMapInit()
	bonusTxMapInit()
}
func GetMethod(transactionType TransactionType) common.Arguments {
	return txMap[transactionType]
}

type TxData struct {
	//3
	From     common.Address
	LeagueId common.Address
	NodeId   common.Address
	Account  common.Address
	//4
	Symbol    common.Symbol
	BlockRoot common.Hash
	Msg       common.Hash
	TxHash    common.Hash

	//2
	Froms *common.TxIns
	Tos   *common.TxOuts

	//2
	Private    bool
	VoteResult bool

	//6
	Rate        uint32
	LeagueNonce uint64
	Nonce       uint64
	StartHeight uint64
	EndHeight   uint64
	MinBox      uint64

	//5
	UnlockEnergy *big.Int
	GasLimit   *big.Int
	MetaBox    *big.Int//
	Fee        *big.Int
	Energy       *big.Int

	NodePub []byte
	//3+2+ 2+2+2+ 6+5+1 = 23
}

func (data *TxData) Serialize(txp TransactionType, isHash bool, version byte, sigpack *common.SigPack, tmplt *common.SigTmplt) (raw []byte, err error) {
	if data == nil {
		return nil, errors.ERR_TX_TXDATA_NIL
	}
	sk := sink.NewZeroCopySink(nil)
	sk.WriteByte(version)
	sk.WriteByte(byte(txp))

	sigplex, err := sink.SigPackToComplex(sigpack)
	if err != nil {
		return nil, err
	}
	sk.WriteComplex(sigplex)

	tmpltplex, err := sink.SigTmpltToComplex(tmplt)
	if err != nil {
		return nil, err
	}
	sk.WriteComplex(tmpltplex)

	//sk.WriteComplex()
	args := GetMethod(txp)
	txdataV := reflect.ValueOf(data)
	for _, v := range args {
		if isHash && v.HashOut {
			continue
		}
		v2 := txdataV.Elem().FieldByName(v.Name)
		if v.Required {
			a := reflect.Zero(v2.Type())
			if reflect.DeepEqual(a.Interface(), v2.Interface()) {
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
	raw, err := tx.TxData.Serialize(tx.TxType, false, tx.Version, tx.Sigs, tx.Templt)
	if err != nil {
		return err
	}
	tx.Raw = raw
	tx.Filled = true
	_, err = w.Write(tx.Raw)
	return err
}

func (tx *Transaction) ToRaw() error {
	raw, err := tx.TxData.Serialize(tx.TxType, false, tx.Version, tx.Sigs, tx.Templt)
	if err != nil {
		return err
	}
	tx.Raw = raw
	return nil
}

func (tx *Transaction) getHashRaw() ([]byte, error) {
	raw, err := tx.TxData.Serialize(tx.TxType, true, tx.Version, tx.Sigs, tx.Templt)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (tx *Transaction) Deserialize(r io.Reader) error {
	if tx.TxData == nil {
		tx.TxData = &TxData{}
	}
	var err error
	if tx.Version, err = serialization.ReadByte(r); err != nil {
		return err
	}

	var txType byte
	if txType, err = serialization.ReadByte(r); err != nil {
		return err
	}
	cp1, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	sp, err := cp1.ComplexToSigpack()
	if err != nil {
		return err
	}
	tx.Sigs = sp
	cp2, err := serialization.ReadComplex(r)
	if err != nil {
		return err
	}
	spa, err := cp2.ComplexToSigTmplt()
	if err != nil {
		fmt.Printf("possible 002\n")
		return err
	}
	tx.Templt = spa
	tx.TxType = TransactionType(txType)
	args := GetMethod(tx.TxType)
	tx.TxData = &TxData{}

	for _, v := range args {
		out, err := serialization.ReadRealase(r, v.Type)
		if err != nil {
			return err
		}
		switch v.Name {
		case "From": //1
			tx.TxData.From = out.(common.Address)
		case "LeagueId": //2
			tx.TxData.LeagueId = out.(common.Address)
		case "NodeId": //3
			tx.TxData.NodeId = out.(common.Address)

		case "Symbol": //4
			tx.TxData.Symbol = out.(common.Symbol)
		case "Msg": //5
			tx.TxData.Msg = out.(common.Hash)
		case "TxHash": //6
			tx.TxData.TxHash = out.(common.Hash)
		case "BlockRoot": //7
			tx.TxData.BlockRoot = out.(common.Hash)

		case "Froms": //8
			tx.TxData.Froms = out.(*common.TxIns)
		case "Tos": //9
			tx.TxData.Tos = out.(*common.TxOuts)

		case "VoteResult": //10
			tx.TxData.VoteResult = out.(bool)
		case "Private": //11
			tx.TxData.Private = out.(bool)

		case "Rate": //12
			tx.TxData.Rate = out.(uint32)
		case "MinBox": //13
			tx.TxData.MinBox = out.(uint64)
		case "Nonce": //14
			tx.TxData.Nonce = out.(uint64)
		case "LeagueNonce": //15
			tx.TxData.LeagueNonce = out.(uint64)
		case "StartHeight": //16
			tx.TxData.StartHeight = out.(uint64)
		case "EndHeight": //17
			tx.TxData.EndHeight = out.(uint64)

		case "MetaBox": //18
			tx.TxData.MetaBox = out.(*big.Int)
		case "UnlockEnergy": //19
			tx.TxData.UnlockEnergy = out.(*big.Int)
		case "GasLimit": //20
			tx.TxData.GasLimit = out.(*big.Int)
		case "Fee": //21
			tx.TxData.Fee = out.(*big.Int)
		case "Energy": //22
			tx.TxData.Energy = out.(*big.Int)
		case "NodePub":
			tx.TxData.NodePub = out.([]byte)
		case "Account":
			tx.TxData.Account = out.(common.Address)
		}
	}
	return nil
}

/*func (tx *Transaction) Serialization(sk *sink.ZeroCopySink) error {
	if tx.Filled == false || len(tx.Raw) == 0 {
		return errors.ERR_SERIALIZATION_FAILED
	}
	sk.WriteBytes(tx.Raw)
	return nil
}*/

func (tx *Transaction) Deserialization(source *sink.ZeroCopySource) error {
	pstart := source.Pos()

	var eof bool
	tx.Version, eof = source.NextByte()
	var txType byte
	txType, eof = source.NextByte()
	tx.TxType = TransactionType(txType)

	dt, eof := source.NextComplex()
	sp, err := dt.ComplexToSigpack()
	if err != nil {
		fmt.Printf("possible 001\n")
		return err
	}
	tx.Sigs = sp

	dta, eof := source.NextComplex()
	spa, err := dta.ComplexToSigTmplt()
	if err != nil {
		fmt.Printf("possible 002\n")
		return err
	}
	tx.Templt = spa

	args := GetMethod(tx.TxType)
	tx.TxData = &TxData{}
	for _, v := range args {
		out, _, eof := sink.ZeroCopySourceRelease(source, v.Type)
		if eof {
			return errors.ERR_TX_DATA_OVERFLOW
		}
		switch v.Name {
		case "From": //1
			tx.TxData.From = out.(common.Address)
		case "LeagueId": //2
			tx.TxData.LeagueId = out.(common.Address)
		case "NodeId": //3
			tx.TxData.NodeId = out.(common.Address)

		case "Symbol": //4
			tx.TxData.Symbol = out.(common.Symbol)
		case "Msg": //5
			tx.TxData.Msg = out.(common.Hash)
		case "TxHash": //6
			tx.TxData.TxHash = out.(common.Hash)
		case "BlockRoot": //7
			tx.TxData.BlockRoot = out.(common.Hash)

		case "Froms": //8
			tx.TxData.Froms = out.(*common.TxIns)
		case "Tos": //9
			tx.TxData.Tos = out.(*common.TxOuts)

		case "VoteResult": //10
			tx.TxData.VoteResult = out.(bool)
		case "Private": //11
			tx.TxData.Private = out.(bool)

		case "Rate": //12
			tx.TxData.Rate = out.(uint32)
		case "MinBox": //13
			tx.TxData.MinBox = out.(uint64)
		case "Nonce": //14
			tx.TxData.Nonce = out.(uint64)
		case "LeagueNonce": //15
			tx.TxData.LeagueNonce = out.(uint64)
		case "StartHeight": //16
			tx.TxData.StartHeight = out.(uint64)
		case "EndHeight": //17
			tx.TxData.EndHeight = out.(uint64)

		case "MetaBox": //18
			tx.TxData.MetaBox = out.(*big.Int)
		case "UnlockEnergy": //19
			tx.TxData.UnlockEnergy = out.(*big.Int)
		case "GasLimit": //20
			tx.TxData.GasLimit = out.(*big.Int)
		case "Fee": //21
			tx.TxData.Fee = out.(*big.Int)
		case "Energy": //22
			tx.TxData.Energy = out.(*big.Int)
		case "NodePub":
			tx.TxData.NodePub = out.([]byte)
		case "Account": //1
			tx.TxData.Account = out.(common.Address)
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
func ToLeagueAddress(tx *Transaction) common.Address {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteAddress(tx.TxData.From)
	sk.WriteAddress(tx.TxData.NodeId)
	cp, _ := sink.BigIntToComplex(tx.TxData.MetaBox)
	sk.WriteComplex(cp)
	sk.WriteUint64(tx.TxData.MinBox)
	sk.WriteUint32(tx.TxData.Rate)
	sk.WriteBool(tx.TxData.Private)
	sk.WriteSymbol(tx.TxData.Symbol)
	league, _ := common.CreateLeagueAddress(sk.Bytes())
	return league
}
