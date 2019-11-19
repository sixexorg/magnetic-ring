package storages

import (
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

type LightLedger struct {
	ledger *LedgerStoreImp
}

func GetLightLedger() *LightLedger {
	ledger := GetLedgerStore()
	return &LightLedger{ledger: ledger}
}

func (this *LightLedger) ContainBlock(blockHash common.Hash) (bool, error) {
	return this.ledger.blockStore.ContainBlock(blockHash)

}
func (this *LightLedger) GetBlockByHash(blockHash common.Hash) (*types.Block, error) {
	block, err := this.ledger.GetBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	return block, nil
}
func (this *LightLedger) GetBlockHashByHeight(height uint64) (common.Hash, error) {
	blockHash, err := this.ledger.GetBlockHashByHeight(height)
	if err != nil {
		return common.Hash{}, err
	}
	return blockHash, err
}
func (this *LightLedger) GetBlockByHeight(height uint64) (*types.Block, error) {
	return this.ledger.GetBlockByHeight(height)
}

func (this *LightLedger) GetCurrentHeaderHeight() uint64 {
	return this.ledger.GetCurrentHeaderHeight()
}

func (this *LightLedger) GetCurrentHeaderHash() common.Hash {
	return this.ledger.GetCurrentHeaderHash()
}

//GetCurrentBlock return the current block height, and block hash.
//Current block means the latest block in store.
func (this *LightLedger) GetCurrentBlock() (uint64, common.Hash) {
	return this.ledger.GetCurrentBlock()
}

//GetCurrentBlockHash return the current block hash
func (this *LightLedger) GetCurrentBlockHash() common.Hash {
	return this.ledger.GetCurrentBlockHash()
}

//GetCurrentBlockHeight return the current block height
func (this *LightLedger) GetCurrentBlockHeight() uint64 {
	return this.ledger.GetCurrentBlockHeight()
}

func (this *LightLedger) GetAccountByHeight(height uint64, account common.Address) (*states.AccountState, error) {
	return this.ledger.GetAccountByHeight(height, account)
}
func (this *LightLedger) GetLeagueByHeight(height uint64, leagueId common.Address) (*states.LeagueState, error) {
	return this.ledger.GetLeagueByHeight(height, leagueId)
}

/*func (this *LightLedger) GetMetaData(leagueId common.Address) (rate uint32, creator common.Address, minBox uint64, err error) {
	return this.ledger.GetMetaData(leagueId)
}*/
func (this *LightLedger) GetLeagueMemberByHeight(height uint64, leagueId, account common.Address) (*states.LeagueMember, error) {
	return this.ledger.GetLeagueMemberByHeight(height, leagueId, account)

}

func (this *LightLedger) GetTxByHash(txHash common.Hash) (*types.Transaction, uint64, error) {
	return this.ledger.GetTxByHash(txHash)
}

func (this *LightLedger) GetHeaderByHash(blockHash common.Hash) (*types.Header, error) {
	return this.ledger.GetHeaderByHash(blockHash)
}

func (this *LightLedger) GetTxByLeagueId(leagueId common.Address) (*types.Transaction, uint64, error) {
	return this.ledger.GetTxByLeagueId(leagueId)
}
