package mainchain

import (
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/log"
)

type TxPoolActor struct {
	txnpool *TxPool
}

func (this *TxPoolActor) setPool(s *TxPool) {
	this.txnpool = s
}

func NewTxActor(s *TxPool) *TxPoolActor {
	actor := &TxPoolActor{}
	actor.setPool(s)

	return actor
}

func (tpa *TxPoolActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case TxReq:

		log.Debug("magnetic actor receive transaction", "tx", msg.Tx.Hash())

		err := tpa.txnpool.TxEnqueue(msg.Tx)
		resp := TxResp{}
		empt := common.Hash{}
		if err != nil {
			resp.Hash = empt.String()
			resp.Desc = err.Error()
			if msg.RespChan != nil {
				msg.RespChan <- &resp
			}
			return
		}

		log.Debug("txpool_actor keep msg", "txhash", msg.Tx.Hash())

		resp.Hash = empt.String()
		resp.Desc = ""

		if msg.RespChan != nil {
			msg.RespChan <- &resp
		}

	case bactor.HeightChange:
		log.Info("func txpool Receive", "oldTargetHeight", tpa.txnpool.stateValidator.TargetHeight, "txlen", tpa.txnpool.stateValidator.GetTxLen())
		tpa.txnpool.RefreshValidator()

	case MustPackTxsReq:
		tpa.txnpool.AppendMustPackTx(msg.Txs...)

	default:
		log.Info("txpool_actor.go txpool-tx actor: unknown msg", "msg", msg, "type", reflect.TypeOf(msg))
	}
}
