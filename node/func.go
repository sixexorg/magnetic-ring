package node

import (
	"fmt"

	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/meacount"
)

func GurStars() []string {
	return st.curStars()
	/*return []string{
		"04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
		"044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6",
		"0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f",

		//"045d54e7cebc80e03c52e53a9a3b7601041183514230ae5776996a131942f8f25425144adc67bbe266200fc7357781a9cc579b3814b37b4f46b3de67cf12177da8",
		//"04a3a9d49d883984c61e3f902f73c4fd0b4067bef3199b3f248a1973d02879e09a89b1b869ea03d40c1816a01fcc2d0a95288c98cfd4386883d9df3015e3f2d72d",
	}*/
}

func CurEarth() string {
	return "04d2db562f13d94fd31d5d500152cac0bfd1692b9fc1185f2fbea712dbd34f7e6c65ce05303ee3a4ce772e0513c75e95a3f3dcc97ea45e22cfebbe3a658de4a493"
}

func PushStars(pubStrs []string) {
	st.pushStars(pubStrs)
}
func PopStars(pubStrs []string) {
	st.popStars(pubStrs)
}
func IsStar() bool {
	own := meacount.GetOwner()
	ndactImpl := own.(*account.NormalAccountImpl)
	pubkstr := fmt.Sprintf("%x", ndactImpl.PublicKey().Bytes())
	for _, v := range GurStars() {
		if pubkstr == v {
			return true
		}
	}
	return false
}
