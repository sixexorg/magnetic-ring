package extstates_test

import (
	"bytes"
	"testing"

	"math/big"

	"github.com/stretchr/testify/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
)

func TestSerializeAndReverse(t *testing.T) {
	hash1 := common.Hash{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	lbs := &extstates.LeagueBlockSimple{
		Header: &types.Header{
			Version:       0xA1,
			PrevBlockHash: hash1,
			LeagueId:      mock.LeagueAddress1,
			TxRoot:        hash1,
			StateRoot:     hash1,
			ReceiptsRoot:  hash1,
			Timestamp:     1540784752,
			Difficulty:    big.NewInt(1023123012),
			Height:        12688,
		},
		TxHashes: common.HashArray{},
	}

	blkhash := lbs.Header.Hash()
	t.Logf("blkhash-->%v\n", blkhash)

	buffer1 := new(bytes.Buffer)

	err := lbs.Serialize(buffer1)

	if err != nil {
		t.Errorf("err=%v\n", err)
		return
	}

	t.Logf("sbuf-->%v\n", buffer1.Bytes())

	lbs2 := new(extstates.LeagueBlockSimple)
	err = lbs2.Deserialize(buffer1)
	if err != nil {
		t.Errorf("err-->%v\n", err)
		return
	}

	t.Logf("v=%v,height=%d", lbs.Header.Version, lbs.Header.Height)
	lbs2.Hash()
	assert.Equal(t, lbs, lbs2)
}
