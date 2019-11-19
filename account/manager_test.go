package account

import (
	"testing"
	"fmt"
	"path/filepath"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/crypto"
	"io/ioutil"
	"encoding/json"
	"strings"
)

func TestAccountManagerImpl_GenerateNormalAccount(t *testing.T) {
	mgr,err := NewManager()
	if err != nil {
		fmt.Println("err0001-->",err)
		return
	}

	password := "jjjdddkkk"

	acct,err := mgr.GenerateNormalAccount(password)
	if err != nil {
		fmt.Println("generate normal account error-->",err)
		return
	}

	pubksrcbuf := acct.PublicKey().Bytes()

	t.Logf("pubklen=%d,pubk:%x\n",len(pubksrcbuf),pubksrcbuf)

	pubkbuf := common.Sha256Ripemd160(pubksrcbuf)

	t.Logf("pubk:%x\n",pubkbuf)

	bufsss := acct.Address()

	t.Logf("len=%d,address=%s\n",len(acct.Address().ToString()),bufsss[:])


	str := acct.Address().ToString()

	address,err := common.ToAddress(str)

	t.Logf("address byte:%T \t %v %v",address,address,bufsss[:])


}



func TestAccountManagerImpl_ExportNormalAccount(t *testing.T) {
	mgr,err := NewManager()
	if err != nil {
		fmt.Println("err0001-->",err)
		return
	}

	customPath := "customPath"

	password := "jjjdddkkk"

	acct,err := mgr.GenerateNormalAccount(password)
	if err != nil {
		fmt.Println("generate normal account error-->",err)
		return
	}

	fullpath := filepath.Join(mgr.keydir, customPath)


	err = mgr.ExportNormalAccount(acct,fullpath,password,false)

	if err != nil {
		fmt.Println("export normal account err-->",err)
		return
	}

	fmt.Printf("address %s export success\n",acct.Address().ToString())

}

func TestAccountManagerImpl_ImportNormalAccount(t *testing.T) {

	mgr,err := NewManager()
	if err != nil {
		fmt.Println("err0001-->",err)
		return
	}

	customPath := "customPath"

	password := "jjjdddkkk"

	fullpath := filepath.Join(mgr.keydir, customPath)


	acct,err := mgr.ImportNormalAccount(fullpath,password)

	if err != nil {
		fmt.Println("export normal account err-->",err)
		return
	}

	fmt.Printf("address %s import success\n",acct.ToString())

}

func TestAccountManagerImpl_CreateMultipleAccount(t *testing.T) {
	mgr,err := NewManager()
	if err != nil {
		fmt.Println("err0001-->",err)
		return
	}
	password := "jjjdddkkk"

	normalacct,err := mgr.GenerateNormalAccount(password)
	if err != nil {
		t.Error(err)
	}

	//memberacct,err := mgr.GenerateNormalAccount(password)
	//if err != nil {
	//	t.Error(err)
	//}

	normalacct.PublicKey().Bytes()

	mult := MultiAccountUnit{1,[]crypto.PublicKey{normalacct.PublicKey()}}

	muls := make(MultiAccountUnits,1)
	muls[0]=mult

	fmt.Printf("lenlenelelne-->%d\n",len(muls))

	mulacct,err := mgr.CreateMultipleAccount(normalacct.Address(),muls)
	if err != nil {
		t.Error(err)
		return
	}
	if a, ok :=mulacct.(*MultipleAccountImpl);ok {
		fmt.Printf("mulacct address create success-->%s  -- %s\n", a.Addr.ToString(), a.Address().ToString())
	}
}

func TestDeseri(t *testing.T)  {
	obj := new(MulPack)

	norstr := "ct07c9guy3vq7tZ54mX62Bg92BjMzBm8e2v"
	mulstr := "ct1RZR58HT8sn3uqrsjkWwRgJ738bj2z321"

	mulpath := filepath.Join("keydir/multiaddrdir",norstr, mulstr)
	raw, err := ioutil.ReadFile(mulpath)
	if err != nil {
		fmt.Printf("read file err -->%v\n",err)
		return
	}

	err = json.Unmarshal(raw,obj)
	if err != nil {
		fmt.Printf("decode err -->%v\n",err)
		return
	}

	fmt.Printf("obj=%v\n",obj)

	units,err := obj.Tmplt.ToMultiAccUnits()

	if err!=nil {
		t.Error(err)
		return
	}

	decaddr,err := units.Address()

	if err!=nil {
		t.Error(err)
		return
	}

	fmt.Printf("deseriaddress-->%s,if equals src=%v\n",decaddr.ToString(),strings.EqualFold(decaddr.ToString(),mulstr))

}

func TestAccountManagerImpl_Sign(t *testing.T) {


	mgr,err := NewManager()
	if err != nil {
		fmt.Println("err0001-->",err)
		return
	}

	addrstr := "ct0x2x4jQyMfNrKn5h2ZqYb1m16gymxkmp7"

	addr,err := common.ToAddress(addrstr)

	if err != nil {
		t.Error(err)
		return
	}



	password := "jjjdddkkk"

	na,err := mgr.GetNormalAccount(addr,password)
	if err != nil {
		t.Error(err)
		return
	}

	raw := []byte("functionissrcoffunc")

	sig,err := na.Sign(raw)

	if err != nil {
		t.Error(err)
		return
	}

	b,err := na.Verify(raw,sig)

	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("verify result-->%t\n",b)

}
