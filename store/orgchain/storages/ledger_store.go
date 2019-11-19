package storages

import (
	"fmt"
	"sync"

	"math/big"

	"sort"

	"path"

	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/log"
	p2pcommon "github.com/sixexorg/magnetic-ring/p2pserver/common"
	scom "github.com/sixexorg/magnetic-ring/store/orgchain/common"
	"github.com/sixexorg/magnetic-ring/store/orgchain/states"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

const (
	SYSTEM_VERSION          = byte(1)      //Version of ledger store
	HEADER_INDEX_BATCH_SIZE = uint64(2000) //Bath size of saving header index
)

var (
	//Storage save path.
	DBDirBlock          = "block"
	DBDirState          = "states"
	DBDirAccountState   = "account"
	DBDirAccountRoot    = "aroot"
	DBDirVote           = "vote"
	DBDirUT             = "ut"
	MerkleTreeStorePath = "merkle_tree.db"
	p2pActor            *actor.PID
)

func initActor() error {
	var err error
	if p2pActor == nil {
		p2pActor, err = bactor.GetActorPid(bactor.P2PACTOR)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

//LedgerStoreImp is main store struct fo ledger
type LedgerStoreImp struct {
	blockStore *BlockStore //BlockStore for saving block & transaction data
	//stateStore       *StateStore //StateStore for saving state data, like balance, smart contract execution result, and so on.
	accountStore     *AccountStore
	accountRootStore *AccountRootStore
	voteStore        *VoteStore
	utStore          *UTStore

	storedIndexCount uint64      //record the count of have saved block index
	currBlockHeight  uint64      //Current block height
	currBlockHash    common.Hash //Current block hash
	currBlock        *types.Block
	totalDifficulty  *big.Int
	headerCache      map[common.Hash]*types.Header //BlockHash => Header
	headerIndex      map[uint64]common.Hash        //Header index, Mapping header height => block hash
	savingBlock      bool                          //is saving block now
	destroyedEnergy    *big.Int
	bonusCache       map[uint64]uint64
	lock             sync.RWMutex
}

//NewLedgerStore return LedgerStoreImp instance
func NewLedgerStore(dataDir string) (*LedgerStoreImp, error) {
	ledgerStore := &LedgerStoreImp{
		headerIndex:     make(map[uint64]common.Hash),
		headerCache:     make(map[common.Hash]*types.Header, 0),
		totalDifficulty: big.NewInt(0),
		destroyedEnergy:   big.NewInt(0),
		bonusCache:      make(map[uint64]uint64),
	}

	blockStore, err := NewBlockStore(path.Join(dataDir, DBDirBlock), true)
	if err != nil {
		return nil, fmt.Errorf("NewBlockStore error %s", err)
	}
	ledgerStore.blockStore = blockStore

	accountState, err := NewAccountStore(path.Join(dataDir, DBDirAccountState))
	if err != nil {
		return nil, fmt.Errorf("NewAccountStateStore error %s", err)
	}
	ledgerStore.accountStore = accountState
	accountRoot, err := NewAccountRootStore(path.Join(dataDir, DBDirAccountRoot))
	if err != nil {
		return nil, fmt.Errorf("NewAccountRootStore error %s", err)
	}
	ledgerStore.accountRootStore = accountRoot

	voteStore, err := NewVoteStore(path.Join(dataDir, DBDirVote), byte(scom.ST_VOTE_STATE), byte(scom.ST_VOTE_RECORD))
	if err != nil {
		return nil, fmt.Errorf("NewVoteStore error %s", err)
	}
	ledgerStore.voteStore = voteStore

	utStore, err := NewUTStore(path.Join(dataDir, DBDirUT))
	if err != nil {
		return nil, fmt.Errorf("NewUTStore error %s", err)
	}
	ledgerStore.utStore = utStore
	/*
		stateStore, err := NewStateStore(path.Join(dataDir, DBDirState), MerkleTreeStorePath)
		if err != nil {
			return nil, fmt.Errorf("NewStateStore error %s", err)
		}

		ledgerStore.stateStore = stateStore*/
	return ledgerStore, nil
}

func (this *LedgerStoreImp) ContainBlock(blockHash common.Hash) (bool, error) {
	return this.blockStore.ContainBlock(blockHash)

}
func (this *LedgerStoreImp) GetBlockByHash(blockHash common.Hash) (*types.Block, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	block, err := this.blockStore.GetBlock(blockHash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (this *LedgerStoreImp) verifyHeader(header *types.Header) error {
	if header.Height == 1 {
		return nil
	}
	prevHeader, err := this.GetHeaderByHash(header.PrevBlockHash)
	if err != nil {
		return err
	}
	if prevHeader == nil {
		return fmt.Errorf("cannot find pre header by blockHash %s", header.PrevBlockHash.String())
	}
	if prevHeader.Height+1 != header.Height {
		return fmt.Errorf("block height is incorrect")
	}
	/*	if prevHeader.Timestamp >= header.Timestamp {
		return fmt.Errorf("block timestamp is incorrect")
	}*/
	return nil
}

func (this *LedgerStoreImp) GetBlockHashByHeight(height uint64) (common.Hash, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	blockHash, err := this.blockStore.GetBlockHash(height)
	if err != nil {
		return common.Hash{}, err
	}
	return blockHash, err
}
func (this *LedgerStoreImp) GetBlockByHeight(height uint64) (*types.Block, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	blockHash, err := this.blockStore.GetBlockHash(height)
	if err != nil {
		return nil, err
	}
	return this.GetBlockByHash(blockHash)
}

func (this *LedgerStoreImp) setCurrentBlock(block *types.Block) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.currBlockHash = block.Hash()
	this.currBlockHeight = block.Header.Height
	this.currBlock = block
	if block.Header.Height%types.HWidth == 0 {
		this.bonusCache[block.Header.Height] = block.Header.Bonus
	}
}
func (this *LedgerStoreImp) setDestroyedEnergy(header *types.Header, feeSum *big.Int) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if header.Height%types.HWidth == 0 {
		this.destroyedEnergy.SetUint64(0)
	}
	rat := new(big.Rat).Set(types.BonusRate)
	rat.Mul(rat, new(big.Rat).SetInt(feeSum))
	rfTmp := big.NewFloat(0).SetRat(rat)
	integer, _ := rfTmp.Uint64()
	this.destroyedEnergy.Add(this.destroyedEnergy, big.NewInt(int64(integer)))
	fmt.Printf("⭕ leagueStore destoryEnergy:%d feeSu,:%d \n ", this.destroyedEnergy.Uint64(), feeSum.Uint64())
}
func (this *LedgerStoreImp) GetCurrentBlockInfo() *types.Block {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlock
}

func (this *LedgerStoreImp) GetCurrentBlock() (uint64, common.Hash) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight, this.currBlockHash
}

//GetCurrentBlockHash return the current block hash
func (this *LedgerStoreImp) GetCurrentBlockHash() common.Hash {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHash
}

//GetCurrentBlockHeight return the current block height
func (this *LedgerStoreImp) GetCurrentBlockHeight() uint64 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight
}

func (this *LedgerStoreImp) GetTransaction(txHash common.Hash, leagueId common.Address) (*types.Transaction, uint64, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.blockStore.GetTransaction(txHash)
}

func (this *LedgerStoreImp) Restore()  {
	
}

func (this *LedgerStoreImp) SaveAll(blockInfo *storelaw.OrgBlockInfo) error {
	if this.isSavingBlock() {
		//hash already saved or is saving
		return nil
	}
	defer this.resetSavingBlock()

	err := this.addStates(blockInfo.AccStates, blockInfo.Block.Header.Height)
	if err != nil {
		return err
	}
	err = this.addBlock(blockInfo.Block)
	if err != nil {
		return err
	}
	err = this.saveVotes(blockInfo.VoteStates, blockInfo.AccountVoteds, blockInfo.Block.Header.Height)
	if err != nil {
		return err
	}
	err = this.saveUT(blockInfo.UT, blockInfo.Block.Header.Height)
	if err != nil {
		return err
	}
	this.delHeaderCache(blockInfo.Block.Hash())
	if this.currBlockHeight > 0 {
		prevBlock := &types.Block{}
		common.DeepCopy(prevBlock, this.currBlock)

		/*	if actor.ConsensusPid != nil {
			actor.ConsensusPid.Tell(blockInfo.Block)
		}*/
		err := initActor()
		if err == nil {
			p2pActor.Tell(&p2pcommon.OrgPendingData{
				BANodeSrc: true,                            // true:ANode send staller false:staller send staller
				OrgId:     blockInfo.Block.Header.LeagueId, // orgid
				Block:     prevBlock,                       // tx
			})
			fmt.Println("!!!!!")
		}
		/*	fmt.Printf("⭕️ send to main radar the difficulty is %d the hash is %s \n", prevBlock.Header.Difficulty.Uint64(), prevBlock.Hash().String())
			fmt.Println("⭕️ send to  🚫 ",
				prevBlock.Header.Version,
				prevBlock.Header.PrevBlockHash.String(),
				prevBlock.Header.BlockRoot.String(),
				prevBlock.Header.LeagueId.ToString(),
				prevBlock.Header.TxRoot.String(),
				prevBlock.Header.StateRoot.String(),
				prevBlock.Header.ReceiptsRoot.String(),
				prevBlock.Header.Timestamp,
				prevBlock.Header.Height,
				prevBlock.Header.Difficulty.Uint64(),
				prevBlock.Header.Coinbase.ToString(),
				prevBlock.Header.Extra,
			)*/
	}
	this.setCurrentBlock(blockInfo.Block)
	this.setDestroyedEnergy(blockInfo.Block.Header, blockInfo.FeeSum)
	fmt.Printf("⭕ -----S---A---V---E----A---L---L---  Height:%d TxLen:%d blockHash:%s difficulty:%d bonus:%d bonusCache:%d \n",
		blockInfo.Block.Header.Height,
		blockInfo.Block.Transactions.Len(),
		blockInfo.Block.Hash().String(),
		blockInfo.Block.Header.Difficulty.Uint64(),
		blockInfo.Block.Header.Bonus,
		this.bonusCache[blockInfo.Height()],
	)
	return nil
}
func (this *LedgerStoreImp) saveUT(ut *big.Int, height uint64) error {
	amount := this.utStore.GetUTByHeight(height)
	if amount.Cmp(ut) == 0 {
		return nil
	}
	return this.utStore.Save(height, ut)
}
func (this *LedgerStoreImp) addStates(accountStates storelaw.AccountStaters, height uint64) error {
	if accountStates == nil {
		return nil
	}
	this.accountStore.NewBatch()
	err := this.accountStore.BatchSave(accountStates)
	if err != nil {
		return err
	}
	if err = this.accountStore.CommitTo(); err != nil {
		return err
	}
	sort.Sort(accountStates)
	hashes := []common.Address{}
	for _, v := range accountStates {
		hashes = append(hashes, v.Account())
	}
	asr := states.NewAccountStateRoot(height, hashes)
	err = this.accountRootStore.Save(asr)
	if err != nil {
		return err
	}
	return nil
}
func (this *LedgerStoreImp) addBlock(block *types.Block) error {
	currentHeight := this.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currentHeight {
		return nil
	}
	if blockHeight != currentHeight+1 {
		return fmt.Errorf("block height %d not equal next block height %d", blockHeight, currentHeight+1)
	}
	err := this.saveBlock(block)
	if err != nil {
		return err
	}
	return nil
}
func (this *LedgerStoreImp) saveBlock(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	this.blockStore.NewBatch()
	err := this.saveHeaderIndexList()
	if err != nil {
		return fmt.Errorf("saveHeaderIndexList error %s", err)
	}

	this.blockStore.SaveBlockHash(blockHeight, blockHash)
	err = this.blockStore.SaveBlock(block)
	if err != nil {
		return fmt.Errorf("SaveBlock height %d hash %s error %s", blockHeight, blockHash.String(), err)
	}
	err = this.blockStore.CommitTo()
	if err != nil {
		return err
	}
	this.setHeaderIndex(blockHeight, blockHash)
	//log.Info("this.totalDifficulty","difficulty",this.totalDifficulty,"block.Header.Difficulty",block.Header.Difficulty)
	this.totalDifficulty.Add(this.totalDifficulty, block.Header.Difficulty)
	return nil
}
func (this *LedgerStoreImp) saveVotes(voteStates []*storelaw.VoteState, AccountVoteds []*storelaw.AccountVoted, height uint64) error {
	this.voteStore.NewBatch()
	this.voteStore.SaveVotes(voteStates)
	this.voteStore.SaveAccountVoted(AccountVoteds, height)
	err := this.voteStore.CommitTo()
	return err
}
func (this *LedgerStoreImp) RollbackToHeight(height uint64) error {
	this.lock.RLock()
	defer this.lock.RUnlock()
	//accountState
	asts, err := this.accountRootStore.GetRange(height, this.currBlockHeight)
	if err != nil {
		return err
	}
	//accountRoot
	this.accountStore.NewBatch()
	heights := []uint64{}
	for k, v := range asts {
		this.accountStore.BatchRemove(v.Height, v.Accounts)
		heights = append(heights, asts[k].Height)
	}
	err = this.accountStore.CommitTo()
	if err != nil {
		return err
	}
	this.accountRootStore.NewBatch()
	this.accountRootStore.BatchRemove(heights)
	err = this.accountRootStore.CommitTo()
	if err != nil {
		return err
	}

	//block
	this.blockStore.NewBatch()
	for i := this.currBlockHeight; i > height; i-- {
		hash, err := this.blockStore.GetBlockHash(i)
		if err != nil {
			return err
		}
		block, err := this.blockStore.GetBlock(hash)
		if err != nil {
			return err
		}

		delete(this.headerCache, hash)
		delete(this.headerIndex, i)
		this.currBlockHash = block.Header.PrevBlockHash
		this.totalDifficulty.Sub(this.totalDifficulty, block.Header.Difficulty)

		this.blockStore.RemoveBlockHash(i)
		this.blockStore.RemoveHeader(hash)
		for _, v := range block.Transactions {
			this.blockStore.RemoveTransaction(v.Hash())
		}
	}
	err = this.blockStore.CommitTo()

	block, err := this.GetBlockByHeight(height)
	if err != nil {
		return err
	}
	this.setCurrentBlock(block)
	return nil
}
func (this *LedgerStoreImp) TotalDifficulty() *big.Int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.totalDifficulty
}
func (this *LedgerStoreImp) GetDifficultyByBlockHash(hash common.Hash) (*big.Int, error) {
	return this.blockStore.GetDifficultyByBlockHash(hash)
}
func (this *LedgerStoreImp) SaveAccount(state storelaw.AccountStater) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.accountStore.Save(state)
}
func (this *LedgerStoreImp) isSavingBlock() bool {
	this.lock.Lock()
	defer this.lock.Unlock()

	if !this.savingBlock {
		this.savingBlock = true
		return false
	}
	return true
}

func (this *LedgerStoreImp) resetSavingBlock() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.savingBlock = false
}

////////////////////Ledger4Validation///////////////////
func (this *LedgerStoreImp) GetCurrentBlockHeight4V(leagueId common.Address) uint64 {
	return this.GetCurrentBlockHeight()
}
func (this *LedgerStoreImp) GetCurrentHeaderHash4V(leagueId common.Address) common.Hash {
	return this.GetCurrentHeaderHash()
}
func (this *LedgerStoreImp) GetBlockHashByHeight4V(leagueId common.Address, height uint64) (common.Hash, error) {
	return this.GetBlockHashByHeight(height)
}

func (this *LedgerStoreImp) GetPrevAccount4V(height uint64, account, leagueId common.Address) (storelaw.AccountStater, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.accountStore.GetPrev(height, account, leagueId)
}
func (this *LedgerStoreImp) ContainTx4V(txHash common.Hash) bool {
	bool, err := this.blockStore.ContainTransaction(txHash)
	if err != nil {
		return false
	}
	return bool
}
func (this *LedgerStoreImp) GetVoteState(voteId common.Hash, height uint64) (*storelaw.VoteState, error) {
	return this.voteStore.GetVoteState(voteId, height)
}
func (this *LedgerStoreImp) AlreadyVoted(voteId common.Hash, account common.Address) bool {
	return this.voteStore.AlreadyVoted(voteId, account)
}
func (this *LedgerStoreImp) GetUTByHeight(height uint64, leagueId common.Address) *big.Int {
	return this.utStore.GetUTByHeight(height)
}
func (this *LedgerStoreImp) GetHeaderBonus(leagueId common.Address) *big.Int {
	return this.destroyedEnergy
}
func (this *LedgerStoreImp) GetAccountRange(start, end uint64, account, leagueId common.Address) (storelaw.AccountStaters, error) {
	return this.accountStore.GetAccountRange(start, end, account)
}
func (this *LedgerStoreImp) GetBonus(leagueId common.Address) map[uint64]uint64 {
	return this.bonusCache
}

////////////////////Ledger4Validation///////////////////
