package main

import (
	"flag"
	"fmt"
	"github.com/sixexorg/magnetic-ring/crypto/cipher"
)

func main() {
	text := flag.String("text", "", "text you want encript")
	pasw := flag.String("pasw", "", "encript password")
	flag.Parse()
	if *text=="" || *pasw=="" {
		fmt.Printf("please give the text and pasw")
		return
	}

	fmt.Printf("listenAddr=%s\n",*text)
	cph := cipher.NewCipher()
	src := []byte(*text)
	password := []byte(*pasw)
	result,err := cph.Encrypt(src,password)

	if err != nil {
		fmt.Println("err=",err)
		return
	}


	fmt.Printf("encrypt success,result=%s\n",result)


	//decstr,err := cph.Decrypt(result,password)
	//if err != nil {
	//	fmt.Println("err=",err)
	//	return
	//}
	//
	//fmt.Printf("src=%s\n",decstr)

}
