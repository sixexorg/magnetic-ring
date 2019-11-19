package types_test

import (
	"testing"

	"bytes"

	"fmt"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

func TestHeaderSerialize(t *testing.T) {
	//address, _ := common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdx")
	hash1 := common.Hash{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	hash2 := common.Hash{2, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	consensusPayload := []byte("00740f40257a13bf03b40f54a9fe398c79a664bb21cfa2870ab07888b21eeba8988899899")

	h := &types.Header{
		Version:          0xA1,
		PrevBlockHash:    hash1,
		TxRoot:           hash1,
		StateRoot:        hash2,
		ReceiptsRoot:     hash1,
		Timestamp:        1540784752,
		Height:           12688,
		ConsensusData:    98,
		ConsensusPayload: consensusPayload,
		Lv1:              50,
		Lv2:              60,
		Lv3:              70,
		Lv4:              80,
		Lv5:              90,
		Lv6:              100,
		Lv7:              111,
		Lv8:              123,
		Lv9:              322,
	}
	sk := sink.NewZeroCopySink(nil)
	err := h.Serialization(sk)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	header := &types.Header{}
	/*	source := sink.NewZeroCopySource(sk.Bytes())
		err = header.Deserialization(source)
		if err != nil {
			t.Error(err)
			t.Fail()
			return
		}*/
	buf1 := bytes.NewBuffer(sk.Bytes())
	err = header.Deserialize(buf1)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	buff := bytes.NewBuffer(nil)
	err = h.Serialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	header2 := &types.Header{}
	err = header2.Deserialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	fmt.Println(string(header.ConsensusPayload))
	assert.Equal(t, h, header)
}
