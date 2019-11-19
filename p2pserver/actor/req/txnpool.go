package req

import (
	"time"

	"github.com/sixexorg/magnetic-ring/p2pserver/temp"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/log"

	// "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	p2pcommon "github.com/sixexorg/magnetic-ring/p2pserver/common"
	tc "github.com/sixexorg/magnetic-ring/txpool/mainchain"
)

const txnPoolReqTimeout = p2pcommon.ACTOR_TIMEOUT * time.Second

var txnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID) {
	txnPoolPid = txnPid
}

//add txn to txnpool
func AddTransaction(msgTrn *p2pcommon.Trn) {
	if msgTrn.SyncType == p2pcommon.SYNC_DATA_ORG {
		addOrgTransaction(msgTrn)
		return
	}
	if txnPoolPid == nil {
		log.Error("[p2p]net_server AddTransaction(): txnpool pid is nil")
		return
	}
	txReq := tc.TxReq{
		Tx:       msgTrn.Txn,
		RespChan: nil,
	}
	txnPoolPid.Tell(txReq)
}

func addOrgTransaction(msgTrn *p2pcommon.Trn) {
	container, err := temp.GetContainerByOrgID(msgTrn.OrgID)
	if err != nil {
		log.Error("GetLedger err", "orgid", msgTrn.OrgID, "err", err)
		return
	}
	orgtxpool := container.TxPool()
	if orgtxpool == nil {
		log.Error("orgtxpool is null ", "orgid", msgTrn.OrgID, "err", err)
		return
	}
	txnSubPoolPid := orgtxpool.GetTxpoolPID()
	if txnSubPoolPid == nil {
		log.Error("[p2p]net_server addOrgTransaction(): txnSubPoolPid pid is nil")
		return
	}
	txReq := tc.TxReq{
		Tx:       msgTrn.Txn,
		RespChan: nil,
	}
	txnSubPoolPid.Tell(txReq)

}
