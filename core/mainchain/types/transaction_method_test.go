package types_test

import (
	"testing"

	"bytes"

	"fmt"

	"reflect"

	"math/big"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
)

func TestTransactionLife(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	err := mock.Mock_Tx_CreateLeague.Serialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	tx := &types.Transaction{}
	err = tx.Deserialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, mock.Mock_Tx_CreateLeague.TxData, tx.TxData)
}

func TestGenesisBlock(t *testing.T) {

}

func TestTransactionEnergyToLeague(t *testing.T) {
	nodeId, _ := common.ToAddress("ct3qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	txCL := &types.Transaction{
		Version: 0x01,
		TxType:  types.CreateLeague,
		TxData: &types.TxData{
			From:    mock.AccountNormal_1.Address(),
			Nonce:   1,
			Fee:     big.NewInt(20),
			Rate:    1000,
			MinBox:  50,
			MetaBox: big.NewInt(5000),
			NodeId:  nodeId,
			Private: true,
		},
	}
	txEnergyToLeague := &types.Transaction{
		Version: 0x01,
		TxType:  types.EnergyToLeague,
		TxData: &types.TxData{
			From:     mock.AccountNormal_1.Address(),
			Nonce:    2,
			Fee:      big.NewInt(20),
			Energy:     big.NewInt(5000),
			LeagueId: types.ToLeagueAddress(txCL),
		},
	}

	buff := bytes.NewBuffer(nil)
	err := txEnergyToLeague.Serialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	tx := &types.Transaction{TxData: &types.TxData{}}

	err = tx.Deserialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	fmt.Println(txEnergyToLeague.TxData.From.ToString(),
		txEnergyToLeague.TxData.Fee.Uint64(),
		txEnergyToLeague.TxData.Energy.Uint64(),
		txEnergyToLeague.TxData.LeagueId.ToString(),
		txEnergyToLeague.TxData.Nonce)
	fmt.Println(tx.TxData.From.ToString(),
		tx.TxData.Fee.Uint64(),
		tx.TxData.Energy.Uint64(),
		tx.TxData.LeagueId.ToString(),
		tx.TxData.Nonce)
	assert.Equal(t, txEnergyToLeague.TxData, tx.TxData)
}

func BenchmarkTransactionLife(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		/*buff := bytes.NewBuffer(nil)
		err := mock.Mock_Tx_CreateLeague.Serialize(buff)
		if err != nil {
			return
		}
		buf1 := bytes.NewBuffer(buff.Bytes())
		tx := &types.Transaction{}
		err = tx.Deserialize(buf1)
		if err != nil {
			return
		}*/
		buff := bytes.NewBuffer(nil)
		err := mock.Mock_Tx_CreateLeague.Serialize(buff)
		if err != nil {
			return
		}
		tx := &types.Transaction{}
		sk := sink.NewZeroCopySource(buff.Bytes())
		err = tx.Deserialization(sk)
		if err != nil {
			return
		}
	}
}

func TestToLeagueAddress(t *testing.T) {
	nodeId, _ := common.ToAddress("ct2qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	tx := &types.Transaction{
		Version: 0x01,
		TxType:  types.CreateLeague,
		TxData: &types.TxData{
			From:    mock.AccountNormal_1.Addr,
			Nonce:   1,
			Fee:     big.NewInt(20),
			Rate:    1000,
			MinBox:  10,
			MetaBox: big.NewInt(5000),
			NodeId:  nodeId,
			Private: false,
		},
	}
	fmt.Println(types.ToLeagueAddress(tx).ToString())

	fmt.Println(mock.AccountNormal_1.Addr.ToString())
}
func TestGetMethod(t *testing.T) {
	a := reflect.TypeOf(uint64(64))
	if a == common.BType_Uint64 {
		fmt.Println("OK")
	}
	if a == common.BType_Bool {
		fmt.Println("NO")
	}

	args := types.GetMethod(types.CreateLeague)
	for k, v := range args {
		fmt.Println(k, v.Name, v.Type.String(), v.Required)
		fmt.Println(v.Type.Kind())
		switch v.Type {
		case common.BType_Bool:
			fmt.Println("bool")
		case common.BType_Uint32:
			fmt.Println("uin32")
		case common.BType_Uint64:
			fmt.Println("uint64")
		case common.BType_BigInt:
			fmt.Println("bigint")
		case common.BType_Hash:
			fmt.Println("hash")
		case common.BType_Address:
			fmt.Println("address")
		case common.BType_TxIns:
			fmt.Println("txins")
		case common.BType_TxOuts:
			fmt.Println("txouts")
		case common.BType_TxHashArray:
			fmt.Println("txhasharray")
		case common.BType_SigBuf:
			fmt.Println("buf")
		}
	}
}
