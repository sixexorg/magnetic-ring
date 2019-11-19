// Copyright (C) 2017 go-crystal authors
//
// This file is part of the go-crystal library.
//
// the go-crystal library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-crystal library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-crystal library.  If not, see <http://www.gnu.org/licenses/>.
//

package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type account struct {

	// key address
	addr common.Address

	// key save path
	path string
}

func checkMake(dir string) error {
	_,err := os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir,os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// refreshAccounts sync key files to memory
func (m *AccountManagerImpl) refreshAccounts() error {
	err := checkMake(m.keydir)
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(m.keydir)
	if err != nil {
		return err
	}
	var accounts []*account


	for _, file := range files {

		acc, err := m.loadKeyFile(file)
		if err != nil {
			// errors have been recorded

			continue
		}
		accounts = append(accounts, acc)
	}

	m.accounts = accounts
	return nil
}

func (m *AccountManagerImpl) refreshMulTmplts() error {
	err := checkMake(m.multiAddrDir)
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(m.multiAddrDir)
	if err != nil {
		return err
	}


	for _, file := range files {

		if file.IsDir(){
			childPath := filepath.Join(m.multiAddrDir,file.Name())
			fs, err := ioutil.ReadDir(childPath)
			if err != nil {
				log.Info("read childPath error","manager_file",err)
				continue
			}
			for _,f := range fs {
				mpk,err := m.loadMultiTmpltFile(file.Name(),f)
				if err !=nil {
					log.Info("load multi tmplt file err","refreshMultmplts",err)
					continue
				}
				units,err := mpk.Tmplt.ToMultiAccUnits()
				if err !=nil {
					log.Info("mpk.Tmplt.ToMultiAccUnits err","refreshMultmplts",err)
					continue
				}

				mulac := new(MultipleAccountImpl)
				mulac.Maus = units
				tempaddress,err := units.Address()
				if err != nil {
					log.Info("units.Address err","refreshMultmplts",err)
					continue
				}
				mulac.Addr=&tempaddress



				if mulmap, ok := m.acctMulMap[file.Name()]; !ok {
					mulmap = make(MultiMap)

					mulmap[mulac.Addr.ToString()] = *mulac
					m.acctMulMap[file.Name()] = mulmap

				} else {
					mulmap[mpk.Mulstr] = *mulac
				}
			}
		}

	}
	return nil
}

func (m *AccountManagerImpl) loadKeyFile(file os.FileInfo) (*account, error) {
	var (
		keyJSON struct {
			Address string `json:"address"`
		}
	)

	path := filepath.Join(m.keydir, file.Name())

	if file.IsDir() || strings.HasPrefix(file.Name(), ".") || strings.HasSuffix(file.Name(), "~") {
		//logging.VLog().WithFields(logrus.Fields{
		//	"path": path,
		//}).Warn("Skipped this key file.")
		return nil, errors.New("file need skip")
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		//logging.VLog().WithFields(logrus.Fields{
		//	"err":  err,
		//	"path": path,
		//}).Error("Failed to read the key file.")
		return nil, errors.New("failed to read the key file")
	}

	keyJSON.Address = ""
	err = json.Unmarshal(raw, &keyJSON)
	if err != nil {
		//logging.VLog().WithFields(logrus.Fields{
		//	"err":  err,
		//	"path": path,
		//}).Error("Failed to parse the key file.")
		return nil, errors.New("failed to parse the key file")
	}

	addr, err := common.ToAddress(keyJSON.Address)
	if err != nil {
		//logging.VLog().WithFields(logrus.Fields{
		//	"err":     err,
		//	"address": keyJSON.Address,
		//}).Error("Failed to parse the address.")
		return nil, errors.New("failed to parse the address")
	}


	acc := &account{addr, path}
	return acc, nil
}

func (m *AccountManagerImpl) loadMultiTmpltFile(prepath string,file os.FileInfo) (*MulPack, error) {
	path := filepath.Join(m.multiAddrDir,prepath,file.Name())

	if file.IsDir() || strings.HasPrefix(file.Name(), ".") || strings.HasSuffix(file.Name(), "~") {
		//logging.VLog().WithFields(logrus.Fields{
		//	"path": path,
		//}).Warn("Skipped this key file.")
		return nil, errors.New("file need skip")
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		//logging.VLog().WithFields(logrus.Fields{
		//	"err":  err,
		//	"path": path,
		//}).Error("Failed to read the key file.")
		return nil, errors.New("failed to read the key file")
	}

	obj := new(MulPack)
	err = json.Unmarshal(raw,obj)
	if err != nil {
		fmt.Printf("decode err -->%v\n",err)
		return nil,err
	}


	return obj, nil
}

// loadFile import key to keystore in keydir
func (m *AccountManagerImpl) loadFile(addr common.Address, passphrase []byte) error {
	acc, err := m.getAccount(addr)
	if err != nil {
		return err
	}

	raw, err := ioutil.ReadFile(acc.path)
	if err != nil {
		return err
	}
	_, err = m.Load(raw, passphrase)
	return err
}

func (am *AccountManagerImpl) ImportNormalAccount(filePath, password string) (common.Address,error) {
	empty := common.Address{}
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return empty,err
	}
	passphrase := []byte(password)
	acct, err := am.Load(raw, passphrase)

	if err != nil {
		return empty,err
	}

	return acct,nil
}

func (m *AccountManagerImpl) ExportNormalAccount(na NormalAccount, filePath, password string,overwrite bool) error {
	passphrase := []byte(password)
	addr := na.Address()
	raw, err := m.Export(addr, passphrase)
	if err != nil {
		return  err
	}

	//acc, err := m.getAccount(&addr)
	// acc not found

	if err := common.FileWrite(filePath, raw, overwrite); err != nil {
		return err
	}
	return nil
}

func (m *AccountManagerImpl) exportFile(addr common.Address, passphrase []byte, overwrite bool) (path string, err error) {
	raw, err := m.Export(addr, passphrase)
	if err != nil {
		return "", err
	}

	acc, err := m.getAccount(addr)
	// acc not found
	if err != nil {
		path = filepath.Join(m.keydir, addr.ToString())
	} else {
		path = acc.path
	}
	if err := common.FileWrite(path, raw, overwrite); err != nil {
		return "", err
	}
	return path, nil
}

func (m *AccountManagerImpl) getAccount(addr common.Address) (*account, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, acc := range m.accounts {
		if acc.addr.Equals(addr) {
			return acc, nil
		}
	}
	return nil, errors.New("account not found")
}

func (m *AccountManagerImpl) updateAccount(addr common.Address, path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var target *account
	for _, acc := range m.accounts {
		if acc.addr.Equals(addr) {
			target = acc
			break
		}
	}
	if target != nil {
		target.path = path
	} else {
		target = &account{addr: addr, path: path}
		m.accounts = append(m.accounts, target)
	}
}
