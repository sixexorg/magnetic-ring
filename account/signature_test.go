package account_test

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/signer"
	"testing"
)

func TestSignTransaction(t *testing.T) {
	mgr, err := account.NewManager()
	if err != nil {
		fmt.Println("err0001-->", err)
		return
	}
	password := "jjjdddkkk"

	normalAccount, err := mgr.GenerateNormalAccount(password)
	address := normalAccount.Address()

	//muladdres,err := common.ToAddress("ct18x9yzad9Vx3uPyJpN65cStK813Y2S6UC")
	muladdres, err := mgr.CreateMultipleAccount(address, nil)

	tx := new(types.Transaction)
	tx.TxData = new(types.TxData)
	tx.TxData.From = muladdres.Address()
	tx.Raw = []byte("this is a test raw buf")

	tx.Templt = common.NewSigTmplt()

	maue := make(common.Maus, 0)
	mau := common.Mau{}
	mau.M = 1
	mau.Pubks = make([]common.PubHash, 0)

	pubkstr := "0445d75cecf5e1ecb74ca90e3e1cec63dd00984c34b81254357697b31a43fb5663eba021949436149128a7f14635b0c953b197b42e445a4407e02c099d3c801a68"
	puhash := common.PubHash{}

	pkbuf, _ := common.Hex2Bytes(pubkstr)

	t.Logf("pkbuflen=%d\n", len(pkbuf))

	copy(puhash[:], pkbuf[:])

	mau.Pubks = append(mau.Pubks, puhash)

	maue = append(maue, mau)

	tx.Templt.SigData = append(tx.Templt.SigData, &maue)

	tx.Sigs = new(common.SigPack)

	err = signer.SignTransaction(mgr, &address, tx, password)
	if err != nil {
		fmt.Println("oops,got an err-->", err)
	}

	rt, err := tx.VerifySignature()
	if err != nil {
		t.Errorf("verify err-->%v\n", err)
		return
	}

	t.Logf("verify result-->%v\n", rt)

	//address2,err := common.ToAddress("ct0x2x4jQyMfNrKn5h2ZqYb1m16gymxkmp7")

	//err = signer.SignTransaction(mgr,&address2,tx,password)
	//if err != nil {
	//	fmt.Println("oops,got an err-->",err)
	//}

	//ma,err := mgr.GetMultipleAccount(address,muladdres,password)
	//if err != nil {
	//	t.Error("err occured when GetMultipleAccount-->",err)
	//	return
	//}

	//maiml := ma.(*account.MultipleAccountImpl)

	tx.Raw = []byte("this is a test raw buf")

	//signresult,err := signer.VerifyTransaction(maiml.Maus,tx)

	//if err != nil {
	//	t.Error("sign error-->",signresult)
	//	return
	//}
	//
	//fmt.Printf("sign result-->%t\n",signresult)

}

func TestVerifySig(t *testing.T) {
	pubkstt := "04723F8F4DC833D7A9EC82BAE72F09B18650114A69FC87A589C4F64A2A1343A9109E0F1E5E1C4E7D113CCE5520BE34BCA41857966C68EF5E61AFDC7D3B881D3F2E"

	pubkbuf, _ := common.Hex2Bytes(pubkstt)

	t.Logf("pubkbuflen======>%d\n", len(pubkbuf))

	pubk, err := crypto.UnmarshalPubkey(pubkbuf)

	if err != nil {
		t.Errorf("err-->%v\n", err)
		return
	}

	rawbuf := []byte("hello world!")

	//buf := common.Sha256(rawbuf)
	//buf := sha256.Sum256(rawbuf)

	t.Logf("the buf hashed -->%X\n", rawbuf)

	t.Logf("rawbuf--->%x\n", rawbuf)

	sig := "2aaef1c6f7619bc0725a0d3b509241a8d30d44fdac90d98f5ae81a8d75ec26ad0a8fc00246e2eac6bd9d1ece035b962187622910bf0c9e3d762a3b913123163e"

	sigbuf, _ := common.Hex2Bytes(sig)

	result, err := pubk.Verify(rawbuf, sigbuf)
	if err != nil {
		t.Errorf("err-->%v\n", err)
		return
	}

	t.Logf("restult-->%v\n", result)

}
