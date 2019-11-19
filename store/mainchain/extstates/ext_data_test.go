package extstates

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/rlp"
)

func TestExtData(t *testing.T) {

	ed := &ExtData{
		Height: 1,
		LeagueBlock: &LeagueBlockSimple{
			Header: &types.Header{

				Difficulty: big.NewInt(1),
				Extra:      make([]byte, 0),
			},
			EnergyUsed: big.NewInt(1),
			TxHashes: make(common.HashArray, 0),
		},
		MainTxUsed:    make(common.HashArray, 0),
		AccountStates: make([]*EasyLeagueAccount, 0),
	}
	buff := bytes.NewBuffer(nil)
	err := rlp.Encode(buff, ed)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	var ed2 *ExtData
	err = rlp.Decode(buff, &ed2)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, ed, ed2)
}
