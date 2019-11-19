package states_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

func TestLeagueState_Serialize(t *testing.T) {
	ls := new(states.LeagueState)
	ls.Address = mock.Mock_Address_1
	ls.Height = 1
	ls.Creator = mock.Mock_Address_2
	ls.MinBox = 1
	ls.Rate = 100

	ls.Data = new(states.League)

	ls.Data.Nonce = 3
	ls.Data.Name = common.Hash{}
	ls.Data.FrozenBox = big.NewInt(99)
	ls.Data.MemberRoot = common.Hash{}
	ls.Data.Private = false

	buf := new(bytes.Buffer)
	err := ls.Serialize(buf)

	if err != nil {
		t.Errorf("error occured -->%v\n", err)
		return
	}

	bufdata := buf.Bytes()
	kbuf := ls.GetKey()

	t.Logf("result-->%v\n", bufdata)

	reverse := new(states.LeagueState)

	dbuf := new(bytes.Buffer)

	dbuf.Write(kbuf)
	dbuf.Write(bufdata)

	err = reverse.Deserialize(dbuf)

	if err != nil {
		t.Errorf("error occured -->%v\n", err)
		return
	}

	t.Logf("height:%v\n", reverse.Height)
}
