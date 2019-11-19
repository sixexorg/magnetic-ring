package walletutil

import (
	"fmt"
	"testing"
)

func TestEncKey(t *testing.T) {

	EncriptyKey("e4454016c146cf9e4a75614dcd1d8453d314f9ab87cbea2234e8e95c81c97e3d","laogui")
}


func TestDecKey(t *testing.T) {

	prk:= DecriptyKey(`{"cipher":"aes-128-ctr","ciphertext":"364fef70891e6abc5356d13ae6de7a70831153306eac074b143a0b6dd6692742","cipherparams":{"iv":"9e56b0aa3ce1fddcd992ffc94265bbf6"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":4096,"p":1,"r":8,"salt":"e178e27dbbd19fb312c6975fab14a2606b5a4b72ffb2232867a145133aed89e8"},"mac":"221a0bbbcfca85027b1f2b21f93175f72c0958798767447c5594792a91a08678","machash":"sha3256"}`,"laogui")
	fmt.Printf("prk:%s\n",prk)
}