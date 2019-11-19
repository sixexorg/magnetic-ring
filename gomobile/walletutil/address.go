package walletutil

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/crypto"
)

func ToAddress(pubkbufstr string) string {

	pubkbuf,err := common.Hex2Bytes(pubkbufstr)
	if err != nil {
		return "hex 2 bytes error"
	}

	hash := common.Sha256Ripemd160(pubkbuf)

	addr := common.BytesToAddress(hash,common.NormalAddress)
	return addr.ToString()
}


func ToMagneticAddress(pubkbufstr string) string {

	buf,err := common.Hex2Bytes(pubkbufstr)
	if err != nil {
		return "hex 2 bytes error"
	}

	pubk,err := crypto.UnmarshalPubkey(buf)

	if err != nil {
		return newErrorResp(err.Error())
	}

	mult := account.MultiAccountUnit{1, []crypto.PublicKey{pubk}}

	muls := make(account.MultiAccountUnits, 1)
	muls[0] = mult

	muladdress,err := muls.Address()
	if err != nil {
		return newErrorResp(err.Error())
	}
	return newResp(fmt.Sprintf("%s", muladdress.ToString()))
}