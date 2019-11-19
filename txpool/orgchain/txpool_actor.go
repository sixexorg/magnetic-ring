package orgchain

import (
	"fmt"
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/log"
)

type TxPoolActor struct {
	txnpool *SubPool
}

func (this *TxPoolActor) setPool(s *SubPool) {
	this.txnpool = s
}

func NewTxActor(s *SubPool) *TxPoolActor {
	actor := &TxPoolActor{}
	actor.setPool(s)

	return actor
}

func (tpa *TxPoolActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case TxReq:
		fmt.Println("📩 txpool actor receive")
		log.Info("magnetic actor receive transaction", "tx", msg.Tx.Hash())

		err := tpa.txnpool.TxEnqueue(msg.Tx)
		resp := TxResp{}
		if err != nil {
			resp.Hash = common.Hash{}
			resp.Desc = err.Error()
			if msg.RespChan != nil {
				msg.RespChan <- &resp
			}
			return
		}

		log.Info("txpool_actor keep msg", "txhash", msg.Tx.Hash())

		resp.Hash = msg.Tx.Hash()
		resp.Desc = ""
		if msg.RespChan != nil {
			msg.RespChan <- &resp
		}

	case MustPackTxsReq:
		tpa.txnpool.AppendMustPackTx(msg.Txs...)

	default:
		log.Info("orgtxpoolactor.go txpool-tx actor: unknown msg", "msg", msg, "type", reflect.TypeOf(msg))
	}
}
