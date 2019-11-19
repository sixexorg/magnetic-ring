package httpactor

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/p2pserver/temp"
	txpool "github.com/sixexorg/magnetic-ring/txpool/mainchain"
	orgcom "github.com/sixexorg/magnetic-ring/txpool/orgchain"
)

var (
	txpoolPid *actor.PID
)

func SetTxPoolPid(pid *actor.PID) {
	txpoolPid = pid
}

func AppendTxToPool(txn *types.Transaction) (common.Hash, string) {
	//ch := make(chan *txpool.TxResp, 1)
	txReq := txpool.TxReq{txn, nil}
	txpoolPid.Tell(txReq)
	//if msg, ok := <-ch; ok {
	//
	//	fmt.Printf("what the hash like is-->%s\n", msg.Hash)
	//
	//	return msg.Hash, msg.Desc
	//}
	return txn.Hash(), ""
}

func AppendTxToSubPool(txn *orgtypes.Transaction, leagueId common.Address) (common.Hash, string) {
	//ch := make(chan *orgcom.TxResp, 1)
	container, err := temp.GetContainerByOrgID(leagueId)

	txReq := orgcom.TxReq{txn, nil}
	if err != nil {
		log.Info("get container by org id error", "error", err)
		return common.Hash{}, ""
	}
	container.TxPool().GetTxpoolPID().Tell(txReq)

	//if msg, ok := <-ch; ok {
	//
	//	fmt.Printf("what the hash like is-->%s\n", msg.Hash)
	//
	//	return msg.Hash, msg.Desc
	//}
	return txn.Hash(), ""
}
