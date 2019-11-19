package cmd

import (
	"bytes"
	"fmt"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/signer/mainchain/signature"
	"testing"
)

func TestNt(t *testing.T) {

	password := "jjjkkk"

	mgr, err := account.NewManager()
	if err != nil {
		fmt.Printf("create wallet error:%v\n", err)
		return
	}

	acctaddressstr := "ct1aajp7n9hVDSWHTPHukf934azZwNh6gU2"

	comaddress, err := common.ToAddress(acctaddressstr)
	if err != nil {
		return
	}

	passwordstr := string(password)

	rawTx := "0110020008c1c04e0009f84cf84af848f84601f843b84104f992c8418749fc5d4bbcaffc4ff120b20b5550f0b5744cff2dd55ad1bfa7f7db94891af2315443a35330a6527436037b7b3d6de5c41e16171fd461e73b5434500100036443d34df4118065f93c4bee56b5d61d4a3b3d6918011d0001dcdbda9543d34df4118065f93c4bee56b5d61d4a3b3d6918018203e80200000000000000000000000000000000000000000000000000000000000000001c0002dbdad9957417702301a3130edc4bdb35ef760fc47f53f2f9008203e8"
	txbuf,_ := common.Hex2Bytes(rawTx)
	buff := bytes.NewBuffer(txbuf)
	tx := &types.Transaction{}
	sk := sink.NewZeroCopySource(buff.Bytes())
	err = tx.Deserialization(sk)
	if err != nil {
		fmt.Printf("can not deserialize data to transaction,err=%v\n", err)
		return
	}

	err = signature.SignTransaction(mgr, comaddress, tx, passwordstr)
	if err != nil {
		fmt.Printf("sign transaction error=%v\n", err)
		return
	}

	b, e := tx.VerifySignature()
	if e != nil {
		fmt.Printf("error=%v\n", e)
		return
	}

	fmt.Printf("b=%t\n", b)

	serbufer := new(bytes.Buffer)

	err = tx.Serialize(serbufer)

	if err != nil {
		fmt.Printf("Serialize transaction error=%v\n", err)
		return
	}

	outbuf := serbufer.Bytes()

	returnstr := common.Bytes2Hex(outbuf)

	reverse,_ := common.Hex2Bytes(returnstr)

	txrev := &types.Transaction{}

	source := sink.NewZeroCopySource(reverse)
	err = txrev.Deserialization(source)
	if err != nil {
		fmt.Printf("transaction Deserialization error=", err)
		return
	}

	b2, e2 := txrev.VerifySignature()
	if e2 != nil {
		fmt.Printf("e2=%v\n", e2)
		return
	}

	fmt.Printf("b2=%v\n", b2)

	//fmt.Printf("%s\n", returnstr)
}
