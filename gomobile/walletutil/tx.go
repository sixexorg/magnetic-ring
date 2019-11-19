package walletutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	mainTypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgTypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/signer/mainchain/signature"
	"math/big"
)

type Resp struct {
	ErrorMsg string `json:"error_msg"`
	Data     string `json:"data"`
}

func newErrorResp(errmsg string) string {
	rsp := Resp{errmsg, ""}
	buf, err := json.Marshal(rsp)
	if err != nil {
	}
	return fmt.Sprintf("%s", buf)
}

func newResp(obj string) string {
	rsp := Resp{"", obj}
	buf, err := json.Marshal(rsp)
	if err != nil {
		return fmt.Sprintf("{\"error_msg\":\"%s\",\"data\":\"\"}", err.Error())
	}
	return fmt.Sprintf("%s", buf)
}


func buildLeagueRaw(tx *mainTypes.Transaction) []byte {
	var (
		orgTp orgTypes.TransactionType
		orgtx = &orgTypes.Transaction{
			Version: orgTypes.TxVersion,
			TxType:  orgTp,
			TxData: &orgTypes.TxData{
				From:   tx.TxData.From,
				TxHash: tx.Hash(),
			},
		}
	)
	switch tx.TxType {
	case mainTypes.EnergyToLeague:
		orgTp = orgTypes.EnergyFromMain
		orgtx.TxData.Energy = tx.TxData.Energy
		break
	case mainTypes.JoinLeague:
		orgTp = orgTypes.Join
		break
	}
	orgtx.TxType = orgTp
	err := orgtx.ToRaw()
	if err != nil {
		return nil
	}
	return orgtx.Raw
}


func TransferBox(pubk, to, msg string, amount, fee, nonce int64) string {

	fromoutAddr,err := OnePubkAddress(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	toAddr, _ := common.ToAddress(to)
	amt := amount
	feeamt := fee
	txouts := &common.TxIns{}
	txout := common.TxIn{
		Address: fromoutAddr,
		Amount:  big.NewInt(amt),
		Nonce:   uint64(nonce),
	}
	txouts.Tis = append(txouts.Tis, &txout)

	totxins := &common.TxOuts{}
	totxin := common.TxOut{
		Address: toAddr,
		Amount:  big.NewInt(amt),
	}
	totxins.Tos = append(totxins.Tos, &totxin)

	msghash, err := common.StringToHash(msg)
	if err != nil {
		msghash = common.Hash{}
	}

	txdata := mainTypes.TxData{
		From:  fromoutAddr,
		Froms: txouts,
		Tos:   totxins,
		Fee:   big.NewInt(feeamt),
		Msg:   msghash,
	}
	tx, err := mainTypes.NewTransaction(mainTypes.TransferBox, mainTypes.TxVersion, &txdata)
	if err != nil {
		return newErrorResp(err.Error())
	}


	err,tmplttmp := OnePubkTmplt(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	tx.Templt = tmplttmp

	err = tx.ToRaw()
	if err != nil {
		return newErrorResp(err.Error())
	}

	outstr := fmt.Sprintf("%x", tx.Raw)
	result := newResp(outstr)
	return result
}


func TransferEnergy(pubk, to, msg string, amount, fee, nonce int64) string {
	fromoutAddr,err := OnePubkAddress(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}
	toAddr, _ := common.ToAddress(to)
	amt := amount
	feeamt := fee
	txouts := &common.TxIns{}
	txout := common.TxIn{
		Address: fromoutAddr,
		Amount:  big.NewInt(amt),
		Nonce:   uint64(nonce),
	}
	txouts.Tis = append(txouts.Tis, &txout)

	totxins := &common.TxOuts{}
	totxin := common.TxOut{
		Address: toAddr,
		Amount:  big.NewInt(amt),
	}
	totxins.Tos = append(totxins.Tos, &totxin)

	msghash, err := common.StringToHash(msg)
	if err != nil {
		msghash = common.Hash{}
	}

	txdata := mainTypes.TxData{
		From:  fromoutAddr,
		Froms: txouts,
		Tos:   totxins,
		Fee:   big.NewInt(feeamt),
		Msg:   msghash,
	}
	tx, err := mainTypes.NewTransaction(mainTypes.TransferEnergy, mainTypes.TxVersion, &txdata)
	if err != nil {
		return newErrorResp(err.Error())
	}

	err,tmplttmp := OnePubkTmplt(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	tx.Templt = tmplttmp

	err = tx.ToRaw()
	if err != nil {
		return newErrorResp(err.Error())
	}

	return newResp(fmt.Sprintf("%x", tx.Raw))
}

func CreateLeague(nodeIdstr, pubk, msg string, minbox, metabox, nonce, fee int64, rate int32,len5s string) string {
	nodeId, err := common.ToAddress(nodeIdstr)
	if err != nil {
		return newErrorResp(err.Error())
	}

	fromoutAddr,err := OnePubkAddress(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}
	msghash, err := common.StringToHash(msg)
	if err != nil {
		msghash = common.Hash{}
	}

	sbl,b := common.String2Symbol(len5s)
	fmt.Printf("string2symbol result=%v\n",b)

	tx := &mainTypes.Transaction{
		Version: mainTypes.MainTxVersion,
		TxType:  mainTypes.CreateLeague,
		TxData: &mainTypes.TxData{
			From:    fromoutAddr,
			Nonce:   uint64(nonce),
			Fee:     big.NewInt(int64(fee)),
			Rate:    uint32(rate),
			MinBox:  uint64(minbox),
			MetaBox: big.NewInt(int64(metabox)),
			NodeId:  nodeId,
			Private: false,
			Msg:     msghash,
			Symbol: sbl,
		},
	}


	err,tmplttmp := OnePubkTmplt(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	tx.Templt = tmplttmp

	err = tx.ToRaw()
	if err != nil {
		return newErrorResp(err.Error())
	}

	return newResp(fmt.Sprintf("%x", tx.Raw))
}



func EnergyToLeague(pubk, leagueId, msg string, quntity, fee, nonce int64) string {
	leageuAddress, err := common.ToAddress(leagueId)
	if err != nil {
		return newErrorResp(err.Error())
	}

	fromAddr,err := OnePubkAddress(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}
	msghash, err := common.StringToHash(msg)
	if err != nil {
		msghash = common.Hash{}
	}
	tx := &mainTypes.Transaction{
		Version: mainTypes.TxVersion,
		TxType:  mainTypes.EnergyToLeague,
		TxData: &mainTypes.TxData{
			From:     fromAddr,
			Nonce:    uint64(nonce),
			Fee:      big.NewInt(int64(fee)),
			LeagueId: leageuAddress,
			Energy:     big.NewInt(int64(quntity)),
			Msg:      msghash,
		},
	}


	err,tmplttmp := OnePubkTmplt(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	tx.Templt = tmplttmp

	err = tx.ToRaw()
	if err != nil {
		return newErrorResp(err.Error())
	}

	return newResp(fmt.Sprintf("%x", tx.Raw))
}

func JoinLeague(pubk, leagueId, msg string, quntity, fee, nonce, leagueNonce int64) string {
	leageuAddress, err := common.ToAddress(leagueId)
	if err != nil {
		return newErrorResp(err.Error())
	}

	fromAddr,err := OnePubkAddress(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}
	msghash, err := common.StringToHash(msg)
	if err != nil {
		msghash = common.Hash{}
	}
	tx := &mainTypes.Transaction{
		Version: mainTypes.TxVersion,
		TxType:  mainTypes.JoinLeague,
		TxData: &mainTypes.TxData{
			From:        fromAddr,
			Nonce:       uint64(nonce),
			LeagueNonce: uint64(leagueNonce),
			Fee:         big.NewInt(fee),
			LeagueId:    leageuAddress,
			MinBox:      uint64(quntity),
			Msg:         msghash,
		},
	}
	tx.TxData.From = fromAddr
	tx.TxData.Nonce = uint64(nonce)


	err,tmplttmp := OnePubkTmplt(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	tx.Templt = tmplttmp

	err = tx.ToRaw()
	if err != nil {
		return newErrorResp(err.Error())
	}

	return newResp(fmt.Sprintf("%x", tx.Raw))
}

func TransferUt(pubk, to, msg string, quntity, fee, nonce int64) string {
	txins := &common.TxIns{}
	fromAddr,err := OnePubkAddress(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	toAddr, err := common.ToAddress(to)
	if err != nil {
		return newErrorResp(err.Error())
	}

	txin := common.TxIn{
		Address: fromAddr,
		Amount:  big.NewInt(quntity),
		Nonce:   uint64(nonce),
	}
	txins.Tis = append(txins.Tis, &txin)

	totxins := &common.TxOuts{}
	totxin := common.TxOut{
		Address: toAddr,
		Amount:  big.NewInt(quntity),
	}
	totxins.Tos = append(totxins.Tos, &totxin)

	msghash, err := common.StringToHash(msg)
	if err != nil {
		msghash = common.Hash{}
	}
	txdata := &orgTypes.TxData{
		From:  fromAddr,
		Froms: txins,
		Tos:   totxins,
		Fee:   big.NewInt(fee),
		Msg:   msghash,
	}
	tx := &orgTypes.Transaction{
		Version: orgTypes.TxVersion,
		TxType:  orgTypes.TransferUT,
		TxData:  txdata,
	}


	err,tmplttmp := OnePubkTmplt(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	tx.Templt = tmplttmp

	err = tx.ToRaw()
	if err != nil {
		return newErrorResp(err.Error())
	}

	return newResp(fmt.Sprintf("%x", tx.Raw))
}

func SignMainTransaction(txhex, accountAddress, privk string) string {

	comaddress, err := common.ToAddress(accountAddress)
	if err != nil {
		return newErrorResp(err.Error())
	}

	txbuf, err := common.Hex2Bytes(txhex)
	if err != nil {
		fmt.Printf("can not deserialize data to transaction,err=%v\n", err)
		return newErrorResp(err.Error())
	}
	buff := bytes.NewBuffer(txbuf)
	tx := &mainTypes.Transaction{}
	sk := sink.NewZeroCopySource(buff.Bytes())
	err = tx.Deserialization(sk)
	if err != nil {
		fmt.Printf("can not deserialize data to transaction,err=%v\n", err)
		return newErrorResp(err.Error())
	}

	err = signature.SignTransactionWithPrivk(comaddress, tx, privk)
	if err != nil {
		fmt.Printf("sign transaction error=%v\n", err)
		return newErrorResp(err.Error())
	}

	b, e := tx.VerifySignature()
	if e != nil {
		fmt.Printf("error=%v\n", e)
		return newErrorResp(err.Error())
	}

	fmt.Printf("b=%t\n", b)

	serbufer := new(bytes.Buffer)
	err = tx.Serialize(serbufer)

	if err != nil {
		return newErrorResp(err.Error())
	}

	outbuf := serbufer.Bytes()

	returnstr := common.Bytes2Hex(outbuf)

	reverse, err := common.Hex2Bytes(returnstr)
	if err != nil {
		return newErrorResp(err.Error())
	}

	txrev := new(mainTypes.Transaction)

	source := sink.NewZeroCopySource(reverse)
	err = txrev.Deserialization(source)
	if err != nil {
		return newErrorResp(err.Error())
	}

	err = tx.ToRaw()
	if err != nil {
		return newErrorResp(err.Error())
	}

	return newResp(fmt.Sprintf("%x", tx.Raw))

}

func OnePubkTmplt(pubk string) (error,*common.SigTmplt) {
	pubkhash,err := common.HexToBytes(pubk)
	if err != nil {
		return err,nil
	}

	pbhash := common.PubHash{}
	copy(pbhash[:],pubkhash[:])

	return nil,&common.SigTmplt{
		[]*common.Maus{
			&common.Maus{
				common.Mau{
					1,
					[]common.PubHash{pbhash},
				},
			},
		},
	}
}

func OnePubkAddress(pubk string) (common.Address,error) {
	buf,err := common.Hex2Bytes(pubk)
	if err != nil {
		return common.Address{},err
	}

	publickKey,err := crypto.UnmarshalPubkey(buf)

	mult := account.MultiAccountUnit{1, []crypto.PublicKey{publickKey}}

	muls := make(account.MultiAccountUnits, 1)
	muls[0] = mult

	fromoutAddr, err := muls.Address()

	if err != nil {
		return common.Address{},err
	}
	return fromoutAddr,nil
}

func BuildLeagueId(txhash string) string {
	ra,err := common.Hex2Bytes(txhash)
	if err != nil {
		return newErrorResp(err.Error())
	}

	tx := &mainTypes.Transaction{}
	source := sink.NewZeroCopySource(ra)
	err = tx.Deserialization(source)
	if err != nil {
		return newErrorResp(err.Error())
	}

	leagueId := mainTypes.ToLeagueAddress(tx)
	return newResp(leagueId.ToString())

}

func PushMsg(pubk, to, msg string, fee, nonce int64) string {
	txins := &common.TxIns{}
	fromAddr,err := OnePubkAddress(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	toAddr, err := common.ToAddress(to)
	if err != nil {
		return newErrorResp(err.Error())
	}

	//txin := common.TxIn{
	//	Address: fromAddr,
	//	Amount:  big.NewInt(0),
	//	Nonce:   uint64(nonce),
	//}
	//txins.Tis = append(txins.Tis, &txin)

	totxins := &common.TxOuts{}
	totxin := common.TxOut{
		Address: toAddr,
		Amount:  big.NewInt(0),
	}
	totxins.Tos = append(totxins.Tos, &totxin)

	msghash, err := common.StringToHash(msg)
	if err != nil {
		msghash = common.Hash{}
	}
	txdata := &orgTypes.TxData{
		From:  fromAddr,
		Froms: txins,
		Tos:   totxins,
		Fee:   big.NewInt(fee),
		Msg:   msghash,
		Nonce:uint64(nonce),
	}
	tx := &orgTypes.Transaction{
		Version: orgTypes.TxVersion,
		TxType:  orgTypes.MSG,
		TxData:  txdata,
	}


	err,tmplttmp := OnePubkTmplt(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	tx.Templt = tmplttmp

	err = tx.ToRaw()
	if err != nil {
		return newErrorResp(err.Error())
	}

	return newResp(fmt.Sprintf("%x", tx.Raw))
}


func CreateMsgTx(pubk, to, msg string, fee, nonce int64) string {
	txins := &common.TxIns{}
	fromAddr,err := OnePubkAddress(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	toAddr, err := common.ToAddress(to)
	if err != nil {
		return newErrorResp(err.Error())
	}

	totxins := &common.TxOuts{}
	totxin := common.TxOut{
		Address: toAddr,
		Amount:  big.NewInt(0),
	}
	totxins.Tos = append(totxins.Tos, &totxin)

	msghash, err := common.StringToHash(msg)
	if err != nil {
		msghash = common.Hash{}
	}
	txdata := &orgTypes.TxData{
		From:  fromAddr,
		Froms: txins,
		Tos:   totxins,
		Fee:   big.NewInt(fee),
		Msg:   msghash,
		Nonce:uint64(nonce),
	}
	tx := &orgTypes.Transaction{
		Version: orgTypes.TxVersion,
		TxType:  orgTypes.MSG,
		TxData:  txdata,
	}


	err,tmplttmp := OnePubkTmplt(pubk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	tx.Templt = tmplttmp

	err = tx.ToRaw()
	if err != nil {
		return newErrorResp(err.Error())
	}

	return fmt.Sprintf("%x", tx.Raw)
}