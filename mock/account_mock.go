package mock

import (
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/crypto"
)

var (
	AccountNormal_1 *account.NormalAccountImpl
	AccountNormal_2 *account.NormalAccountImpl
	Mgr             *account.AccountManagerImpl
)

func init() {

	//var err error

	/*	Mgr, err = account.NewManager()
		if err != nil {
			//fmt.Printf("create wallet error:%v\n", err)
			return
		}*/
	//str := "92f94dc2f24aa392b55e59a9867e3a4f6a7bfd15c04caad018b304a6f376560c"
	str := "70fe34923ddd9762b9ed9e0258721b406a29b46628680c75aeb18b6ededeab8a"
	priv, _ := crypto.HexToPrivateKey(str)
	//pubksrcbuf := priv.Public().Bytes()
	//pubkbuf := common.Sha256Ripemd160(pubksrcbuf)

	bufsss, _ := account.PubkToAddress(priv.Public())
	//fmt.Printf("bufsss=%s\n", bufsss.ToString())

	AccountNormal_1 = &account.NormalAccountImpl{
		PrivKey: priv,
		Addr:    bufsss,
	}

	str2 := "33dcf1912962fe9d8c7538476add705b1ab344e35af6d9f9f91cd79c27c5d7f9"
	priv2, _ := crypto.HexToPrivateKey(str2)
	pubksrcbuf2 := priv.Public().Bytes()
	pubkbuf2 := common.Sha256Ripemd160(pubksrcbuf2)

	bufsss2 := common.BytesToAddress(pubkbuf2, common.NodeAddress)

	/*	acct, err := Mgr.ImportKeyNormal("92f94dc2f24aa392b55e59a9867e3a4f6a7bfd15c04caad018b304a6f376560c", "jjjkkk")
		if err != nil {
			//fmt.Printf("error:%v\n", err)
			return
		}

		mult := account.MultiAccountUnit{1, []crypto.PublicKey{acct.PublicKey()}}

		muls := make(account.MultiAccountUnits, 1)
		muls[0] = mult*/

	/*mulacct, err := Mgr.CreateMultipleAccount(acct.Address(), muls)

	if err != nil {
		return
	}*/
	//pubksrcbuf5 := acct.PublicKey().Bytes()

	/*fmt.Printf("public key:%x\n", pubksrcbuf5)
	fmt.Printf("address:%s\n", acct.Address().ToString())
	fmt.Printf("default multiple address:%s\n", mulacct.Address().ToString())*/

	AccountNormal_2 = &account.NormalAccountImpl{
		PrivKey: priv2,
		Addr:    bufsss2,
	}

}
