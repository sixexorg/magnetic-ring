package node

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func Benchmark_PushStars(b *testing.B) {
	b.ReportAllocs()
	strs := []string{
		"04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
		"044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6",
		"0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f",
	}
	for i := 0; i < b.N; i++ { //use b.N for looping
		//st.pushStars(strs)
		st.popStars(strs)
	}
}

func Test_PushPop(t *testing.T) {
	strs := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"f",
		"g",
		"h",
		"i",
	}
	st.pushStars(strs)
	for _, v := range st.stars {
		fmt.Println(v)
	}
	strs = []string{
		"i",
		"a",
		"d",
	}
	fmt.Println("-------------------")
	st.popStars(strs)
	for _, v := range st.stars {
		fmt.Println(v)
	}
}

func TestStr2Pubkey(t *testing.T) {
	str := "04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f"
	pubuf, err := hex.DecodeString(str)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(hex.EncodeToString(pubuf))
}
