package account

import (
	"bytes"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/crypto"
	errors2 "github.com/sixexorg/magnetic-ring/errors"
)

type Account interface {
	Sign(hash []byte) (sig []byte, err error)
	Verify(hash, sig []byte) (bool, error)
	Address() common.Address
}

type LeagueAccount interface {
	Address() common.Address
}

type NormalAccount interface {
	Account
	PublicKey() crypto.PublicKey
}

type NodeAccount interface {
	Account
	PublicKey() crypto.PublicKey
}

type MultipleAccount interface {
	Account
	ExistPubkey(pubKey crypto.PublicKey) bool
}

type AccountsManager interface {
	GenerateNormalAccount(password string) (NormalAccount, error)

	ImportNormalAccount(filePath, password string) (*common.Address, error)
	ExportNormalAccount(acc NormalAccount, filePath, password string, overide bool) error

	CreateMultipleAccount(addr common.Address, units MultiAccountUnits) (MultipleAccount, error)

	GetNormalAccount(addr common.Address, password string) (NormalAccount, error)
	GetMultipleAccount(addr, multiAddr common.Address, selfPass string) (MultipleAccount, error)
	DeleteNormalAccount(addr common.Address, password string) error
	DeleteMultipleAccount(addr, multiAddr common.Address, selfPass string) error
}

type MultiAccountUnit struct {
	Threshold uint8
	Pubkeys   []crypto.PublicKey
}

type MultiEle struct {
	Threshold uint8    `json:"m"`
	Pubkeys   []string `json:"pbks"`
}

type MultiAccObj []MultiEle

type MulPack struct {
	Tmplt  MultiAccObj
	Mulstr string
}

type MultiAccountUnits []MultiAccountUnit

func (maus MultiAccountUnits) ToObj() MultiAccObj {
	obj := make(MultiAccObj, len(maus))

	for i, unit := range maus {
		me := MultiEle{}
		me.Threshold = unit.Threshold
		me.Pubkeys = make([]string, len(unit.Pubkeys))
		for j, item := range unit.Pubkeys {
			str := common.Bytes2Hex(item.Bytes())
			me.Pubkeys[j] = str
		}
		obj[i] = me
	}
	return obj
}

func (maos MultiAccObj) ToMultiAccUnits() (MultiAccountUnits, error) {
	units := make(MultiAccountUnits, len(maos))

	for i, unit := range maos {
		me := MultiAccountUnit{}
		me.Threshold = unit.Threshold
		me.Pubkeys = make([]crypto.PublicKey, len(unit.Pubkeys))
		for j, item := range unit.Pubkeys {
			buf, err := common.Hex2Bytes(item)
			if err != nil {
				return nil, err
			}

			pubk, err := crypto.UnmarshalPubkey(buf)
			if err != nil {
				return nil, err
			}

			me.Pubkeys[j] = pubk
		}
		units[i] = me
	}
	return units, nil
}

func (units MultiAccountUnits) Address() (addr common.Address, err error) {
	buf, err := units.ToBytes()
	if err != nil {
		return addr, err
	}

	hash := common.Sha256Ripemd160(buf)
	addr.SetBytes(hash, common.MultiAddress)
	return addr, nil
}

func (units MultiAccountUnits) ToBytes() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	for _, unit := range units {
		if len(unit.Pubkeys) < int(unit.Threshold) {
			return nil, errors2.ERR_ACCT_CRT_MULADDR
		}

		_, err := buf.Write([]byte{unit.Threshold})
		if err != nil {
			return nil, err
		}

		for _, pubKey := range unit.Pubkeys {
			_, err = buf.Write(pubKey.Bytes())
		}
	}

	return buf.Bytes(), nil
}

func PubkToAddress(pubk crypto.PublicKey) (common.Address, error) {
	mult := MultiAccountUnit{1, []crypto.PublicKey{pubk}}

	muls := make(MultiAccountUnits, 1)
	muls[0] = mult

	return muls.Address()
}
