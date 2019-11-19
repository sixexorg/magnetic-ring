package types_test

import (
	"testing"

	"math/big"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
)

func TestHeaderSerialize(t *testing.T) {
	hash1 := common.Hash{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	h := &types.Header{
		Version:       0xA1,
		PrevBlockHash: hash1,
		LeagueId:      mock.LeagueAddress1,
		TxRoot:        hash1,
		StateRoot:     hash1,
		ReceiptsRoot:  hash1,
		Timestamp:     1540784752,
		Difficulty:    big.NewInt(1023123012),
		Height:        12688,
		Coinbase:      common.Address{1, 2, 3, 4, 5, 6},
	}
	sk := sink.NewZeroCopySink(nil)
	err := h.Serialization(sk)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	header := &types.Header{}
	source := sink.NewZeroCopySource(sk.Bytes())
	err = header.Deserialization(source)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, h, header)

}
