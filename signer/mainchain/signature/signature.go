package signature

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/crypto"
	errors2 "github.com/sixexorg/magnetic-ring/errors"
)

/*
	check signature postion and state
*/
func CheckSignState(sp common.SigPack, units account.MultiAccountUnits) (apprlvlIndex int, enough bool, err error) {
	sigdata := sp.SigData

	if sigdata == nil {
		sp.SigData = make([]*common.SigMap, len(units))
	}
	sigdatalen := len(sigdata)

	if sigdatalen > len(units) {
		apprlvlIndex = -1
		err = errors2.SIG_ERR_SIZE
		return
	}

	if sigdatalen < 1 {
		return 0, false, nil
	}

	index := 0

	for index, sigmap := range sigdata {
		if len(*sigmap) < int(units[index].Threshold) {
			apprlvlIndex = index
			return
		}
	}
	apprlvlIndex = index
	enough = true
	return
}

func CheckAddressCanSign(units account.MultiAccountUnits, apprlvlIndex int, address common.Address) (bool, error) {
	if apprlvlIndex > (len(units)-1) || apprlvlIndex < 0 {
		return false, errors2.SIG_ERR_APPR_LVL
	}

	//unit := units[apprlvlIndex]

	muldddaddr, err := units.Address()
	if err != nil {
		return false, err
	}

	if muldddaddr.ToString() == address.ToString() {
		return true, nil
	}

	//for _, puk := range unit.Pubkeys {
	//
	//	pubytes := common.Sha256Ripemd160(puk.Bytes())
	//
	//	tempaddr := common.BytesToAddress(pubytes, common.NormalAddress)
	//	if address.Equals(tempaddr) {
	//		//apprlvl := *sp.SigData[apprlvlIndex]
	//		//if sigbytes,ok := apprlvl[address];ok {
	//		//	return sigbytes,nil
	//		//}
	//		return true, nil
	//	}
	//}

	return false, errors2.SIG_ERR_UNSIGN

}

func SignTransaction(m *account.AccountManagerImpl, address common.Address, tx *types.Transaction, pasw string) error {
	muladdr := tx.TxData.From
	mulacct, err := m.GetMultipleAccount(address, muladdr)
	if err != nil {
		return err
	}

	impl := mulacct.(*account.MultipleAccountImpl)

	addressstr := address.ToString()
	fmt.Printf("addressstr=%s\n", addressstr)

	key, err := m.GetNormalAccount(address, pasw)
	if err != nil {
		return err
	}

	b := impl.ExistPubkey(key.PublicKey())

	if !b {
		return errors2.SIG_ERR_WRONGOWNER
	}

	tx.Sigs = new(common.SigPack)
	err = tx.ToRaw()
	if err != nil {
		return err
	}

	sha256buf := common.Sha256(tx.Raw)
	sigbuf, err := key.Sign(sha256buf)
	if err != nil {
		return err
	}

	apprindex, _, err := CheckSignState(*tx.Sigs, impl.Maus)

	can, err := CheckAddressCanSign(impl.Maus, apprindex, address)

	if err != nil {
		return err
	}

	if !can {
		return errors2.SIG_ERR_CANOT
	}
	sd := tx.Sigs.SigData

	if sd == nil || len(sd) == 0 {
		sd = make([]*common.SigMap, 0)
		sm := make(common.SigMap, 0)
		itm := common.SigMapItem{address, sigbuf}
		//sm[*address] = sigbuf
		sm = append(sm, &itm)
		sd = append(sd, &sm)
		tx.Sigs.SigData = sd
	} else {
		fmt.Printf("sd=%v,apprindex=%d\n", sd, apprindex)
		apprlvl := sd[apprindex]
		if apprlvl == nil || len(*apprlvl) == 0 {
			temp := make(common.SigMap, 0)
			itm := common.SigMapItem{address, sigbuf}
			temp = append(temp, &itm)
			sd[apprindex] = &temp
		} else {
			itm := common.SigMapItem{address, sigbuf}
			*apprlvl = append(*apprlvl, &itm)
		}
	}

	return nil
}

func VerifyTransaction(unis account.MultiAccountUnits, tx *types.Transaction) (bool, error) {

	if unis == nil || len(unis) == 0 {
		return false, errors2.SIG_ERR_NULL_UNITS
	}

	if tx.Sigs.SigData == nil || len(tx.Sigs.SigData) < 1 {
		return false, errors2.SIG_ERR_NOSIG
	}

	if len(tx.Sigs.SigData) != len(unis) {
		return false, errors2.SIG_ERR_SIGS_NOTENOUGH
	}

	for index, unit := range unis {
		sigmaparr := tx.Sigs.SigData

		sigmp := sigmaparr[index]

		if len(*sigmp) < int(unit.Threshold) {
			return false, errors2.SIG_ERR_APPROVE_NOT_ENOUGH
		}

		for _, sigitm := range *sigmp {

			for _, pubk := range unit.Pubkeys {

				addrhash := common.Sha256Ripemd160(pubk.Bytes())

				tmpaddress := common.BytesToAddress(addrhash, common.NormalAddress)

				if tmpaddress.Equals(sigitm.Key) {
					bsign, err := pubk.Verify(tx.Raw, sigitm.Val)
					if err != nil {
						return false, err
					}

					if !bsign {
						return false, errors2.SIG_ERR_INVALID_SIG
					}
				}
			}
		}

	}

	return true, nil
}

func SignTransactionPubk(m *account.AccountManagerImpl, pubkstr string, tx *types.Transaction, pasw string) error {
	muladdr := tx.TxData.From

	buf, err := common.Hex2Bytes(pubkstr)
	if err != nil {
		return err
	}

	pubk, err := crypto.UnmarshalPubkey(buf)
	if err != nil {
		return err
	}
	bytes := pubk.Bytes()
	hash := common.Sha256Ripemd160(bytes)
	address := common.BytesToAddress(hash, common.NormalAddress)

	mulacct, err := m.GetMultipleAccount(address, muladdr)
	if err != nil {
		return err
	}

	impl := mulacct.(*account.MultipleAccountImpl)

	key, err := m.GetNormalAccount(address, pasw)
	if err != nil {
		return err
	}

	b := impl.ExistPubkey(key.PublicKey())

	if !b {
		return errors2.SIG_ERR_WRONGOWNER
	}

	tx.Sigs = new(common.SigPack)
	err = tx.ToRaw()
	if err != nil {
		return err
	}

	sha256buf := common.Sha256(tx.Raw)
	sigbuf, err := key.Sign(sha256buf)
	if err != nil {
		return err
	}

	apprindex, _, err := CheckSignState(*tx.Sigs, impl.Maus)

	can, err := CheckAddressCanSign(impl.Maus, apprindex, address)

	if err != nil {
		return err
	}

	if !can {
		return errors2.SIG_ERR_CANOT
	}
	sd := tx.Sigs.SigData

	if sd == nil || len(sd) == 0 {
		sd = make([]*common.SigMap, 0)
		sm := make(common.SigMap, 0)
		itm := common.SigMapItem{address, sigbuf}
		//sm[*address] = sigbuf
		sm = append(sm, &itm)
		sd = append(sd, &sm)
		tx.Sigs.SigData = sd
	} else {
		fmt.Printf("sd=%v,apprindex=%d\n", sd, apprindex)
		apprlvl := sd[apprindex]
		if apprlvl == nil || len(*apprlvl) == 0 {
			temp := make(common.SigMap, 0)
			itm := common.SigMapItem{address, sigbuf}
			temp = append(temp, &itm)
			sd[apprindex] = &temp
		} else {
			itm := common.SigMapItem{address, sigbuf}
			*apprlvl = append(*apprlvl, &itm)
		}
	}

	return nil
}

func SignTransactionWithPrivk(address common.Address, tx *types.Transaction, privk string) error {

	tplt := tx.Templt

	impl := new(account.MultipleAccountImpl)

	mulen := len(tplt.SigData)
	muls := make(account.MultiAccountUnits, mulen)
	for i, layer := range tplt.SigData {
		var mult account.MultiAccountUnit
		for _, pubdata := range *layer {
			pbks := make([]crypto.PublicKey, 0)

			for _, pbk := range pubdata.Pubks {
				pubkey, err := crypto.UnmarshalPubkey(pbk[:])
				if err != nil {
					return err
				}
				pbks = append(pbks, pubkey)
			}
			mult = account.MultiAccountUnit{pubdata.M, pbks}
		}
		muls[i] = mult

	}
	impl.Maus = muls

	addressstr := address.ToString()
	fmt.Printf("addressstr=%s\n", addressstr)

	key, err := crypto.HexToPrivateKey(privk)

	//key, err := m.GetNormalAccount(address, pasw)
	if err != nil {
		return err
	}

	b := impl.ExistPubkey(key.Public())

	if !b {
		return errors2.SIG_ERR_WRONGOWNER
	}

	tx.Sigs = new(common.SigPack)
	err = tx.ToRaw()
	if err != nil {
		return err
	}

	sha256buf := common.Sha256(tx.Raw)
	sigbuf, err := key.Sign(sha256buf)
	if err != nil {
		return err
	}

	apprindex, _, err := CheckSignState(*tx.Sigs, impl.Maus)

	can, err := CheckAddressCanSign(impl.Maus, apprindex, address)

	if err != nil {
		return err
	}

	if !can {
		return errors2.SIG_ERR_CANOT
	}
	sd := tx.Sigs.SigData

	if sd == nil || len(sd) == 0 {
		sd = make([]*common.SigMap, 0)
		sm := make(common.SigMap, 0)
		itm := common.SigMapItem{address, sigbuf}
		//sm[*address] = sigbuf
		sm = append(sm, &itm)
		sd = append(sd, &sm)
		tx.Sigs.SigData = sd
	} else {
		fmt.Printf("sd=%v,apprindex=%d\n", sd, apprindex)
		apprlvl := sd[apprindex]
		if apprlvl == nil || len(*apprlvl) == 0 {
			temp := make(common.SigMap, 0)
			itm := common.SigMapItem{address, sigbuf}
			temp = append(temp, &itm)
			sd[apprindex] = &temp
		} else {
			itm := common.SigMapItem{address, sigbuf}
			*apprlvl = append(*apprlvl, &itm)
		}
	}

	return nil
}
