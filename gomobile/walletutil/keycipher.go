package walletutil

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/crypto/cipher"
)

func EncriptyKey(privk,pasw string) string {
	cipher := cipher.NewCipher()

	prk,err := crypto.HexToPrivateKey(privk)
	if err != nil {
		return newErrorResp(err.Error())
	}

	pkbuf := prk.Bytes()

	enbuf,err := cipher.Encrypt(pkbuf,[]byte(pasw))
	if err!=nil {
		return newErrorResp(err.Error())
	}
	encjson := fmt.Sprintf("%s",enbuf)

	return newResp(fmt.Sprintf("%s", encjson))
}


func DecriptyKey(encbuf,pasw string) string {
	cipher := cipher.NewCipher()

	prkbuf := []byte(encbuf)


	debuf,err := cipher.Decrypt(prkbuf,[]byte(pasw))
	if err!=nil {
		return newErrorResp(err.Error())
	}
	encjson := fmt.Sprintf("%x",debuf)

	return newResp(fmt.Sprintf("%s", encjson))
}