package account

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	acct "github.com/sixexorg/magnetic-ring/account/keystore"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/crypto/cipher"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/log"
)

type NormalAccountImpl struct {
	PrivKey crypto.PrivateKey
	Addr    common.Address
}

func (acct *NormalAccountImpl) Sign(hash []byte) (sig []byte, err error) {
	priv := acct.PrivKey
	return priv.Sign(hash[:])
}

func (acct *NormalAccountImpl) Verify(hash, sig []byte) (bool, error) {
	pub := acct.PublicKey()
	return pub.Verify(hash, sig)
}

func (acct *NormalAccountImpl) Address() common.Address {
	return acct.Addr
}

func (acct *NormalAccountImpl) PublicKey() crypto.PublicKey {
	return acct.PrivKey.Public()
}

type MultipleAccountImpl struct {
	Addr *common.Address
	Maus MultiAccountUnits
}

func (mai *MultipleAccountImpl) PublicKey() crypto.PublicKey {
	return nil
}

func (mai *MultipleAccountImpl) ExistPubkey(pubKey crypto.PublicKey) bool {

	for _, arr := range mai.Maus {
		for _, ele := range arr.Pubkeys {
			if bytes.Compare(pubKey.Bytes(), ele.Bytes()) == 0 {
				return true
			}
		}
	}

	return false
}

func (mai *MultipleAccountImpl) OnlyPubkey() crypto.PublicKey {
	if len(mai.Maus) != 1 {
		return nil
	}
	if len(mai.Maus[0].Pubkeys) != 1 {
		return nil
	}
	return mai.Maus[0].Pubkeys[0]
}

func (acct *MultipleAccountImpl) Address() common.Address {
	return *acct.Addr
}

func (acct *MultipleAccountImpl) Sign(hash []byte) (sig []byte, err error) {
	return nil, errors.SIG_ERR_NOT_SUPPORT
}

func (acct *MultipleAccountImpl) Verify(hash, sig []byte) (bool, error) {

	return false, errors.SIG_ERR_NOT_SUPPORT
}

type AccountManagerImpl struct {
	// keystore
	ks *acct.Keystore

	// key save path
	keydir string

	//multi addr save path
	multiAddrDir string

	// account slice
	accounts []*account

	mutex sync.Mutex

	acctMulMap  map[string]MultiMap
	accMulMutex sync.Mutex
}

type MultiMap map[string]MultipleAccountImpl

func NewManager() (*AccountManagerImpl, error) {
	m := new(AccountManagerImpl)
	m.ks = acct.DefaultKS
	tmpKeyDir, err := filepath.Abs(config.DefaultKeyDir)
	if err != nil {
		return nil, err
	}
	m.keydir = tmpKeyDir

	mulTmpDir, err := filepath.Abs(config.DefaultMultiAddrDir)
	if err != nil {
		return nil, err
	}
	m.multiAddrDir = mulTmpDir
	m.acctMulMap = make(map[string]MultiMap)
	if err := m.refreshAccounts(); err != nil {
		return nil, err
	}
	if err := m.refreshMulTmplts(); err != nil {
		return nil, err
	}
	return m, nil
}

func (am *AccountManagerImpl) GenerateNormalAccount(password string) (NormalAccount, error) {

	privk, err := crypto.GenerateKey()

	if err != nil {
		log.Error("generate key err", err)
		return nil, err
	}
	na := new(NormalAccountImpl)
	na.PrivKey = privk

	bytes := na.PublicKey().Bytes()
	hash := common.Sha256Ripemd160(bytes)
	na.Addr = common.BytesToAddress(hash, common.NormalAddress)

	passphrase := []byte(password)
	addr, err := am.setKeyStore(privk, passphrase, common.NormalAddress)
	if err != nil {
		return nil, err
	}

	path, err := am.exportFile(addr, passphrase, false)
	if err != nil {
		return nil, err
	}

	am.updateAccount(addr, path)

	return na, nil
}

func (am *AccountManagerImpl) CreateTraAccount(password string) (NormalAccount, error) {

	privk, err := crypto.GenerateKey()

	if err != nil {
		log.Error("generate key err", err)
		return nil, err
	}
	na := new(NormalAccountImpl)
	na.PrivKey = privk

	bytes := na.PublicKey().Bytes()
	hash := common.Sha256Ripemd160(bytes)
	na.Addr = common.BytesToAddress(hash, common.NormalAddress)

	passphrase := []byte(password)
	addr, err := am.setKeyStore(privk, passphrase, common.NormalAddress)
	if err != nil {
		return nil, err
	}

	path, err := am.exportFile(addr, passphrase, false)
	if err != nil {
		return nil, err
	}

	am.updateAccount(addr, path)

	return na, nil
}

func (am *AccountManagerImpl) OneKeyToAddress(pubk crypto.PublicKey) (common.Address, error) {
	mult := MultiAccountUnit{1, []crypto.PublicKey{pubk}}

	muls := make(MultiAccountUnits, 1)
	muls[0] = mult

	return muls.Address()
}

func (am *AccountManagerImpl) ImportKeyNormal(prvk, password string) (NormalAccount, error) {
	privk, err := crypto.HexToPrivateKey(prvk)

	if err != nil {
		log.Error("Hex To Private Key err", "error", err)
		return nil, err
	}
	na := new(NormalAccountImpl)
	na.PrivKey = privk


	mult := MultiAccountUnit{1, []crypto.PublicKey{privk.Public()}}

	muls := make(MultiAccountUnits, 1)
	muls[0] = mult

	muladdr, err := muls.Address()

	if err != nil {
		log.Error("create multiple account err", "units.Address", err)
		return nil, err
	}
	muladdrstr := muladdr.ToString()
	fmt.Printf("muladdrstr=%s\n", muladdrstr)

	paswphrase := []byte(password)
	err = am.setOneKeyStore(privk, paswphrase, muladdr)
	if err != nil {
		return nil, err
	}

	path, err := am.exportFile(muladdr, paswphrase, false)
	if err != nil {
		return nil, err
	}

	am.updateAccount(muladdr, path)

	mulpath := filepath.Join(am.multiAddrDir, muladdr.ToString(), muladdr.ToString())

	mulac := new(MultipleAccountImpl)
	mulac.Addr = &muladdr
	mulac.Maus = muls

	am.accMulMutex.Lock()

	if mulmap, ok := am.acctMulMap[muladdr.ToString()]; !ok {
		mulmap = make(MultiMap)
	} else {
		mulmap[muladdr.ToString()] = *mulac
	}

	am.accMulMutex.Unlock()

	//rawbuf := new(bytes.Buffer)

	mpk := new(MulPack)
	mpk.Mulstr = muladdr.ToString()
	mpk.Tmplt = muls.ToObj()

	rawbuf, err := json.Marshal(mpk)
	//err = rlp.Encode(rawbuf,mpk)

	if err != nil {
		return nil, err
	}

	if err := common.FileWrite(mulpath, rawbuf, false); err != nil {
		log.Error("create multiple account err", "common.FileWrite", err)
		return nil, err
	}

	return na, nil
}

func (am *AccountManagerImpl) ImportKeyNode(prvk, password string) (NormalAccount, error) {
	privk, err := crypto.HexToPrivateKey(prvk)

	if err != nil {
		log.Error("Hex To Private Key err", "error", err)
		return nil, err
	}
	na := new(NodeAccountImpl)
	na.PrivKey = privk

	bytes := na.PublicKey().Bytes()
	hash := common.Sha256Ripemd160(bytes)
	na.Addr = common.BytesToAddress(hash, common.NodeAddress)

	passphrase := []byte(password)
	addr, err := am.setKeyStore(privk, passphrase, common.NodeAddress)
	if err != nil {
		return nil, err
	}

	path, err := am.exportFile(addr, passphrase, false)
	if err != nil {
		return nil, err
	}

	am.updateAccount(addr, path)

	return na, nil
}

func (am *AccountManagerImpl) GenerateNodeAccount(password string) (NodeAccount, error) {

	privk, err := crypto.GenerateKey()

	if err != nil {
		log.Error("generate key err", err)
		return nil, err
	}
	na := new(NodeAccountImpl)
	na.PrivKey = privk

	bytes := na.PublicKey().Bytes()
	hash := common.Sha256Ripemd160(bytes)
	na.Addr = common.BytesToAddress(hash, common.NodeAddress)

	passphrase := []byte(password)
	addr, err := am.setKeyStore(privk, passphrase, common.NodeAddress)
	if err != nil {
		return nil, err
	}

	path, err := am.exportFile(addr, passphrase, false)
	if err != nil {
		return nil, err
	}

	am.updateAccount(addr, path)

	return na, nil
}

func (am *AccountManagerImpl) CreateMultipleAccount(addr common.Address, units MultiAccountUnits) (MultipleAccount, error) {

	muladdr, err := units.Address()

	if err != nil {
		log.Error("create multiple account err", "units.Address", err)
		return nil, err
	}

	mulpath := filepath.Join(am.multiAddrDir, addr.ToString(), muladdr.ToString())

	mulac := new(MultipleAccountImpl)
	mulac.Addr = &muladdr
	mulac.Maus = units

	am.accMulMutex.Lock()

	if mulmap, ok := am.acctMulMap[addr.ToString()]; !ok {
		mulmap = make(MultiMap)
	} else {
		mulmap[muladdr.ToString()] = *mulac
	}

	am.accMulMutex.Unlock()

	//rawbuf := new(bytes.Buffer)

	mpk := new(MulPack)
	mpk.Mulstr = muladdr.ToString()
	mpk.Tmplt = units.ToObj()

	rawbuf, err := json.Marshal(mpk)
	//err = rlp.Encode(rawbuf,mpk)

	if err != nil {
		return nil, err
	}

	if err := common.FileWrite(mulpath, rawbuf, false); err != nil {
		log.Error("create multiple account err", "common.FileWrite", err)
		return nil, err
	}

	return mulac, err
}

func (am *AccountManagerImpl) GetNormalAccount(addr common.Address, password string) (NormalAccount, error) {
	passphrase := []byte(password)

	addrstr := addr.ToString()
	res, err := am.ks.ContainsAlias(addrstr)
	if err != nil || res == false {
		err = am.loadFile(addr, passphrase)
		if err != nil {
			fmt.Printf("👠 fu ck\n")
			return nil, err
		}
	}

	privk, err := am.ks.GetKey(addr.ToString(), passphrase)
	if err != nil {
		fmt.Printf("👠 fu2ck,address=%s\n",addr.ToString())
		return nil, err
	}

	na := new(NormalAccountImpl)
	na.PrivKey = privk
	bytes := na.PublicKey().Bytes()
	hash := common.Sha256Ripemd160(bytes)
	na.Addr = common.BytesToAddress(hash, common.NormalAddress)
	return na, err
}

//func (am *AccountManagerImpl) GetMultipleAccount(addr, multiAddr common.Address, selfPass string) (MultipleAccount, error) {
//
//	//fmt.Printf("watch addr=%s\n", addr.ToString())
//	//fmt.Printf("watch mult=%s\n", multiAddr.ToString())
//	passphrase := []byte(selfPass)
//
//	res, err := am.ks.ContainsAlias(addr.ToString())
//	if err != nil || res == false {
//		err = am.loadFile(addr, passphrase)
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	_, err = am.ks.GetKey(addr.ToString(), passphrase)
//
//	if err != nil {
//		log.Info("GetMultipleAccount err", "AccountManagerImpl.ks.GetKey", err)
//		return nil, err
//	}
//
//	addrstr := addr.ToString()
//	multstr := multiAddr.ToString()
//
//	am.accMulMutex.Lock()
//	defer am.accMulMutex.Unlock()
//
//	rt, ok := am.acctMulMap[addrstr]
//
//	if !ok {
//		return nil, errors.ERR_ACCT_MULADDR_NOT_FOUND
//	}
//
//	fnrt, ok := rt[multstr]
//
//	if !ok {
//		return nil, errors.ERR_ACCT_MULADDR_NOT_FOUND
//
//	}
//	return &fnrt, nil
//}

func (am *AccountManagerImpl) GetMultipleAccount(addr, multiAddr common.Address) (MultipleAccount, error) {
	addrstr := addr.ToString()
	multstr := multiAddr.ToString()

	am.accMulMutex.Lock()
	defer am.accMulMutex.Unlock()

	rt, ok := am.acctMulMap[addrstr]

	if !ok {
		return nil, errors.ERR_ACCT_MULADDR_NOT_FOUND
	}

	fnrt, ok := rt[multstr]

	if !ok {
		return nil, errors.ERR_ACCT_MULADDR_NOT_FOUND

	}
	return &fnrt, nil
}

func (am *AccountManagerImpl) DeleteNormalAccount(addr common.Address, password string) error {
	passphrase := []byte(password)
	err := am.ks.Delete(addr.ToString(), passphrase)
	return err
}

func (am *AccountManagerImpl) DeleteMultipleAccount(addr, multiAddr common.Address, selfPass string) error {
	_, err := am.ks.GetKey(addr.ToString(), []byte(selfPass))
	if err != nil {
		return err
	}

	am.accMulMutex.Lock()

	delete(am.acctMulMap[addr.ToString()], multiAddr.ToString())

	am.accMulMutex.Unlock()

	return nil
}

func (m *AccountManagerImpl) Load(keyjson, passphrase []byte) (common.Address, error) {
	//var account account
	empty := common.Address{}
	cipher := cipher.NewCipher()
	data, err := cipher.DecryptKey(keyjson, passphrase)
	if err != nil {
		return empty, err
	}
	defer common.ZeroBytes(data)

	priv, err := crypto.ToPrivateKey(data)
	if err != nil {
		return empty, err
	}
	//defer priv.Clear()

	addr, err := m.setKeyStore(priv, passphrase, common.NormalAddress)
	if err != nil {
		return empty, err
	}

	if _, err := m.getAccount(addr); err != nil {
		m.mutex.Lock()
		acc := &account{addr: addr}
		m.accounts = append(m.accounts, acc)
		m.mutex.Unlock()
	}
	return addr, nil
}

func (m *AccountManagerImpl) setKeyStore(priv crypto.PrivateKey, passphrase []byte, addrType common.AddressType) (common.Address, error) {
	empty := common.Address{}
	//pub := priv.Public().Bytes()

	//hash := common.Sha256Ripemd160(pub)
	//
	//addr := common.BytesToAddress(hash, addrType)

	addr,err := m.OneKeyToAddress(priv.Public())
	if err != nil {
		return addr,err
	}

	// set key to keystore
	err = m.ks.SetKey(addr.ToString(), priv, passphrase)
	if err != nil {
		return empty, err
	}

	return addr, nil
}

func (m *AccountManagerImpl) setOneKeyStore(priv crypto.PrivateKey, passphrase []byte, addr common.Address) error {
	//empty := common.Address{}
	//pub := priv.Public().Bytes()

	//hash := common.Sha256Ripemd160(pub)

	//addr := common.BytesToAddress(hash, addrType)

	// set key to keystore
	err := m.ks.SetKey(addr.ToString(), priv, passphrase)
	if err != nil {
		return err
	}

	return nil
}

func (m *AccountManagerImpl) Export(addr common.Address, passphrase []byte) ([]byte, error) {
	key, err := m.ks.GetKey(addr.ToString(), passphrase)
	if err != nil {
		return nil, err
	}
	//defer key.Clear()

	data := key.Bytes()
	defer common.ZeroBytes(data)

	cipher := cipher.NewCipher()
	if err != nil {
		return nil, err
	}

	out, err := cipher.EncryptKey(addr.ToString(), data, passphrase)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type LeagueAccountImpl struct {
	Addr common.Address
}

func (la *LeagueAccountImpl) Address() common.Address {
	return la.Addr
}

func (am *AccountManagerImpl) CreateOneKeyAccount(password []byte) (MultipleAccount, error) {
	privk, err := crypto.GenerateKey()

	if err != nil {
		log.Error("generate key err", err)
		return nil, err
	}

	mult := MultiAccountUnit{1, []crypto.PublicKey{privk.Public()}}

	muls := make(MultiAccountUnits, 1)
	muls[0] = mult

	muladdr, err := muls.Address()

	if err != nil {
		log.Error("create multiple account err", "units.Address", err)
		return nil, err
	}

	err = am.setOneKeyStore(privk, password, muladdr)
	if err != nil {
		return nil, err
	}

	path, err := am.exportFile(muladdr, password, false)
	if err != nil {
		return nil, err
	}

	am.updateAccount(muladdr, path)

	mulpath := filepath.Join(am.multiAddrDir, muladdr.ToString(), muladdr.ToString())

	mulac := new(MultipleAccountImpl)
	mulac.Addr = &muladdr
	mulac.Maus = muls

	am.accMulMutex.Lock()

	if mulmap, ok := am.acctMulMap[muladdr.ToString()]; !ok {
		mulmap = make(MultiMap)
	} else {
		mulmap[muladdr.ToString()] = *mulac
	}

	am.accMulMutex.Unlock()

	//rawbuf := new(bytes.Buffer)

	mpk := new(MulPack)
	mpk.Mulstr = muladdr.ToString()
	mpk.Tmplt = muls.ToObj()

	rawbuf, err := json.Marshal(mpk)
	//err = rlp.Encode(rawbuf,mpk)

	if err != nil {
		return nil, err
	}

	if err := common.FileWrite(mulpath, rawbuf, false); err != nil {
		log.Error("create multiple account err", "common.FileWrite", err)
		return nil, err
	}

	return mulac, err
}


func (am *AccountManagerImpl) CreateBoxAccount(password []byte,privKeyHex string) (MultipleAccount, error) {
	privk, err := crypto.HexToPrivateKey(privKeyHex)

	if err != nil {
		log.Error("generate key err", err)
		return nil, err
	}

	fmt.Printf("private key for test:%s\n", privk.Hex())

	mult := MultiAccountUnit{1, []crypto.PublicKey{privk.Public()}}

	muls := make(MultiAccountUnits, 1)
	muls[0] = mult

	muladdr, err := muls.Address()

	if err != nil {
		log.Error("create multiple account err", "units.Address", err)
		return nil, err
	}

	err = am.setOneKeyStore(privk, password, muladdr)
	if err != nil {
		return nil, err
	}

	path, err := am.exportFile(muladdr, password, false)
	if err != nil {
		return nil, err
	}

	am.updateAccount(muladdr, path)

	mulpath := filepath.Join(am.multiAddrDir, muladdr.ToString(), muladdr.ToString())

	mulac := new(MultipleAccountImpl)
	mulac.Addr = &muladdr
	mulac.Maus = muls

	am.accMulMutex.Lock()

	if mulmap, ok := am.acctMulMap[muladdr.ToString()]; !ok {
		mulmap = make(MultiMap)
	} else {
		mulmap[muladdr.ToString()] = *mulac
	}

	am.accMulMutex.Unlock()

	//rawbuf := new(bytes.Buffer)

	mpk := new(MulPack)
	mpk.Mulstr = muladdr.ToString()
	mpk.Tmplt = muls.ToObj()

	rawbuf, err := json.Marshal(mpk)
	//err = rlp.Encode(rawbuf,mpk)

	if err != nil {
		return nil, err
	}

	if err := common.FileWrite(mulpath, rawbuf, false); err != nil {
		log.Error("create multiple account err", "common.FileWrite", err)
		return nil, err
	}

	return mulac, err
}