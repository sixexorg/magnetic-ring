package validation

import (
	"sort"

	"fmt"

	"bytes"

	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/store/mainchain/genesis"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

func ValidateBlock(block *types.Block, ledgerStore *storages.LedgerStoreImp) (*storages.BlockInfo, error) {
	/*	for k, v := range block.Transactions {
			fmt.Println("num:", k, ",hash:", v.Hash().String(), ",from:", v.TxData.From.ToString(), ",nonce:", v.TxData.Froms.Tis[0].Nonce)
		}
	*/
	log.Info("func validation ValidateBlock start", "blockHeight", block.Header.Height, "txlen", block.Transactions.Len())
	blockHeight := block.Header.Height
	if ledgerStore.GetCurrentBlockHeight()+1 < blockHeight {
		return nil, errors.ERR_BLOCK_ABNORMAL_HEIGHT
	}
	if blockHeight > 1 {
		prevBlock, err := ledgerStore.GetBlockByHeight(blockHeight - 1)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(prevBlock.Hash().ToBytes(), block.Header.PrevBlockHash.ToBytes()) {
			return nil, errors.ERR_BLOCK_PREVHASH_DIFF
		}
	}

	/*	add, _ := common.ToAddress("ct0fHcR19s2nDaHxR3CVdUyYxHgsjzp8wrt")
		acc, err := ledgerStore.GetAccountByHeight(blockHeight, add)
		if err == nil {
			fmt.Println("accountNonce", acc.Data.Nonce)
		}*/
	sv := NewStateValidate(ledgerStore)
	txLen := block.Transactions.Len()
	if txLen > 0 {
		txsch := make(chan *types.Transaction, block.Transactions.Len())
		txs := types.TxByNonce(block.Transactions)
		sort.Sort(txs)
		go func(txs types.TxByNonce) {
			for _, v := range txs {
				txTmp := v
				txsch <- txTmp
			}
		}(txs)
		count := 0
		breakCount := block.Transactions.Len() * block.Transactions.Len()

		ii := 0
		for txTmp := range txsch {
			count++
			//txTmp := ch
			re, err := sv.VerifyTx(txTmp)
			//fmt.Println("validateBlock verifyTx err:", err, ",hash:", txTmp.Hash().String(), "from:", txTmp.TxData.From.ToString(), "nonce:", txTmp.TxData.Froms.Tis[0].Nonce)
			if re == 0 {
				txsch <- txTmp
			} else if re == -1 {
				return nil, fmt.Errorf("ValidateBlock err:%s,type:%s,result:-1\n", errors.ERR_BLOCK_ABNORMAL_TXS.Error(), err.Error())
			} else if re == 1 {
				ii++
				//fmt.Println("num:", ii, "hash:", ch.Hash().String())
				if ii == block.Transactions.Len() {
					close(txsch)
				}
			}
			if count > breakCount {
				return nil, fmt.Errorf("ValidateBlock err:%s,type:%s\n", errors.ERR_BLOCK_ABNORMAL_TXS.Error(), err.Error())
			}
		}
		//to clear
		if count > breakCount {
			return nil, fmt.Errorf("ValidateBlock err:%s\n", errors.ERR_BLOCK_ABNORMAL_TXS.Error())
		}
	}
	blockInfo := sv.ExecuteOplogs()
	if blockHeight == 1 {
		ass, txs := genesis.GetGenesisBlockState()
		blockInfo.AccountStates = ass
		blockInfo.Block.Header.TxRoot = txs.GetHashRoot()
		blockInfo.Block.Header.StateRoot = blockInfo.AccountStates.GetHashRoot()
		blockInfo.Block.Transactions = txs
	}

	/*fmt.Printf("blockNew:\n txroot:%s\n receipt:%s\n account:%s\n league:%s\n",
		blockInfo.Block.Header.TxRoot.String(),
		blockInfo.Block.Header.ReceiptsRoot.String(),
		blockInfo.Block.Header.StateRoot.String(),
		blockInfo.Block.Header.LeagueRoot.String(),
	)*/
	/*for k, v := range blockInfo.Receipts {
		fmt.Printf("Receipts2: num:%d, status:%t, txhash:%s, gasused:%d \n", k, v.Status, v.TxHash.String(), v.GasUsed)
	}*/
	/*
		for k, v := range blockInfo.AccountStates {
			fmt.Printf("States2: num:%d, Hash:%s, nonce:%d, balance:%d ,gasBalance:%d \n", k, v.Hash().String(), v.Data.Nonce, v.Data.Balance, v.Data.EnergyBalance)
		}*/
	/*	for k, v := range blockInfo.LeagueStates {
		fmt.Printf("league2: num:%d, address:%s, nonce:%d, MemberRoot:%d, FrozenBox:%d, Private:%t, minBox:%d, rate:%d, creator:%s  ", k, v.Address.ToString(), v.Data.Nonce, v.Data.MemberRoot, v.Data.FrozenBox.Uint64(), v.Data.Private, v.MinBox, v.Rate, v.Creator.ToString())
	}*/

	log.Info(fmt.Sprintf("func validation ValidateBlock root cmp blocknew : receipt:%s  league:%s state:%s", blockInfo.Block.Header.ReceiptsRoot.String(), blockInfo.Block.Header.LeagueRoot.String(), blockInfo.Block.Header.StateRoot.String()))
	log.Info(fmt.Sprintf("func validation ValidateBlock root cmp blockold : receipt:%s  league:%s state:%s", block.Header.ReceiptsRoot.String(), block.Header.LeagueRoot.String(), block.Header.StateRoot.String()))
	if blockInfo.Block.Header.ReceiptsRoot != block.Header.ReceiptsRoot ||
		blockInfo.Block.Header.LeagueRoot != block.Header.LeagueRoot ||
		blockInfo.Block.Header.StateRoot != block.Header.StateRoot {
		log.Info("func validation ValidateBlock failed", "blockHeight", blockInfo.Block.Header.Height, "txlen", blockInfo.Block.Transactions.Len())
		return nil, errors.ERR_BLOCK_ABNORMAL_TXS
	}
	blockInfo.Block.Header.ConsensusPayload = block.Header.ConsensusPayload
	blockInfo.Block.Header.Timestamp = block.Header.Timestamp
	blockInfo.Block.Sigs = block.Sigs
	blockInfo.Block.Header.ConsensusData = block.Header.ConsensusData
	blockInfo.Block.Header.NextBookkeeper = block.Header.NextBookkeeper
	log.Info("func validation ValidateBlock end", "blockHeight", blockInfo.Block.Header.Height, "txlen", blockInfo.Block.Transactions.Len())
	return blockInfo, nil
}
