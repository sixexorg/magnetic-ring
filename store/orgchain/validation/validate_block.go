package validation

import (
	"fmt"

	"sort"

	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

func ValidateBlock(block *types.Block, ledgerStore storelaw.Ledger4Validation) (blkInfo *storelaw.OrgBlockInfo, err error) {
	sv := NewStateValidate(ledgerStore, block.Header.LeagueId)
	txLen := block.Transactions.Len()
	if txLen > 0 {
		txsch := make(chan *types.Transaction, block.Transactions.Len())
		txs := types.TxByNonce(block.Transactions)
		sort.Sort(txs)
		go func(txs types.TxByNonce) {
			for k, _ := range txs {
				txsch <- txs[k]
			}
		}(txs)
		count := 0
		breakCount := block.Transactions.Len() * block.Transactions.Len()

		ii := 0
		for ch := range txsch {
			count++
			re, _ := sv.VerifyTx(ch)
			if re == 0 {
				txTmp := ch
				txsch <- txTmp
			} else if re == -1 {
				err = errors.ERR_BLOCK_ABNORMAL_TXS
				return
			} else if re == 1 {
				ii++
				if ii == block.Transactions.Len() {
					close(txsch)
				}
			}
			if count > breakCount {
				err = errors.ERR_BLOCK_ABNORMAL_TXS
				return
			}
		}
		//to clear
		if count > breakCount {
			err = errors.ERR_BLOCK_ABNORMAL_TXS
			return
		}
	}
	blkInfo = sv.ExecuteOplogs()
	/*	fmt.Printf("blockNew:\n txroot:%s\n receipt:%s\n account:%s\n league:%s\n",
		blockNew.Header.TxRoot.String(),
		blockNew.Header.ReceiptsRoot.String(),
		blockNew.Header.StateRoot.String(),
	)*/
	/*	for k, v := range receipts {
			fmt.Printf("num:%d, status:%t, txhash:%s, gasused:%d \n", k, v.Status, v.TxHash.String(), v.GasUsed)
		}
		fmt.Println("---------------------------root cmp-------------------------------")

	*/
	log.Info(fmt.Sprintf("root cmp blocknew: receipt:%s  state:%s\n", blkInfo.Block.Header.ReceiptsRoot.String(), blkInfo.Block.Header.StateRoot.String()))
	log.Info(fmt.Sprintf("root cmp blockold: receipt:%s  state:%s\n", block.Header.ReceiptsRoot.String(), block.Header.StateRoot.String()))

	if blkInfo.Block.Header.ReceiptsRoot != block.Header.ReceiptsRoot ||
		blkInfo.Block.Header.StateRoot != block.Header.StateRoot {
		err = errors.ERR_BLOCK_ABNORMAL_TXS
	}
	blkInfo.Block.Header.Difficulty = block.Header.Difficulty
	blkInfo.Block.Header.Timestamp = block.Header.Timestamp
	blkInfo.Block.Header.Coinbase = block.Header.Coinbase
	blkInfo.Block.Header.Extra = block.Header.Extra
	return
}
