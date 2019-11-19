package extstorages

import (
	"fmt"
	"os"
	"sync"

	"path"

	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	scom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	orgStorages "github.com/sixexorg/magnetic-ring/store/orgchain/storages"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

var (
	//Storage save path.
	DBDirBlock          = "extblock"
	DBDirState          = "extstates"
	DBDirMainTx         = "extmain"
	DBDirAccIdx         = "extaccidx"
	DBDirVote           = "extvote"
	DBDirFullTx         = "extfull"
	DBDirUT             = "extut"
	MerkleTreeStorePath = "extmerkle_tree.db"
	//radarActor          *actor.PID
)

/*func initActor() {
	if radarActor == nil {
		actor, err := bactor.GetActorPid(bactor.MAINRADARACTOR)
		if err != nil {
			panic(fmt.Errorf("extstorages init radarActor failed,error: %s", err.Error()))
		}
		radarActor = actor
	}
}
*/
//LedgerStoreImp is main store struct fo ledger
type LedgerStoreImp struct {
	blockStore      *ExternalLeagueBlock
	accountStore    *ExternalLeague
	mainTxUsedStore *MainTxUsed
	acctIdxStore    *ExtAccIndexStore
	voteStore       *orgStorages.VoteStore
	fullTxStore     *ExtFullTXStore
	utStore         *ExtUTStore

	LightLedger     *storages.LightLedger
	currBlockHeader map[common.Address]*orgtypes.Header //BlockHash => Header
	savingBlock     bool                                //is saving block now

	bonusCache    map[common.Address]map[uint64]uint64
	destroyedEnergy map[common.Address]*big.Int
	lock          sync.RWMutex
}

var ledgerStore *LedgerStoreImp

func (this *LedgerStoreImp) NewBlockBatch() {
	this.blockStore.NewBatch()
}

func (this *LedgerStoreImp) BlockBatchCommit() error {
	err := this.blockStore.CommitTo()
	if err != nil {
		return err
	}
	return nil
}
func GetLedgerStoreInstance() *LedgerStoreImp {
	return ledgerStore
}

//NewLedgerStore return LedgerStoreImp instance
func NewLedgerStore(dataDir string) (*LedgerStoreImp, error) {
	if ledgerStore != nil {
		return ledgerStore, nil
	}
	ledgerStore = &LedgerStoreImp{
		currBlockHeader: make(map[common.Address]*orgtypes.Header),
		bonusCache:      make(map[common.Address]map[uint64]uint64),
		destroyedEnergy:   make(map[common.Address]*big.Int),
	}

	accountStore, err := NewExternalLeague(path.Join(dataDir, DBDirState), false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewExtAccoutStore error %s\n", err)
		return nil, err
	}

	ledgerStore.accountStore = accountStore

	blockStore, err := NewExternalLeagueBlock(path.Join(dataDir, DBDirBlock), true)
	if err != nil {
		return nil, fmt.Errorf("NewExtBlockStore error %s", err)
	}
	ledgerStore.blockStore = blockStore

	mainTxUsed, err := NewMainTxUsed(path.Join(dataDir, DBDirMainTx))
	if err != nil {
		return nil, fmt.Errorf("NewExtMainTxUsedStore error %s", err)
	}
	ledgerStore.mainTxUsedStore = mainTxUsed

	accIdxStore, err := NewExtAccIndexStore(path.Join(dataDir, DBDirAccIdx), false)
	if err != nil {
		return nil, fmt.Errorf("NewExtAccIndexStore error %s", err)
	}
	ledgerStore.acctIdxStore = accIdxStore

	voteStore, err := orgStorages.NewVoteStore(path.Join(dataDir, DBDirVote), byte(scom.EXT_VOTE_STATE), byte(scom.EXT_VOTE_RECORD))
	if err != nil {
		return nil, fmt.Errorf("NewVoteStore error %s", err)
	}
	ledgerStore.voteStore = voteStore

	utStore, err := NewExtUTStore(path.Join(dataDir, DBDirUT))
	if err != nil {
		return nil, fmt.Errorf("NewExtUTStore error %s", err)
	}
	ledgerStore.utStore = utStore

	fullTxStore, err := NewExtFullTX(path.Join(dataDir, DBDirFullTx))
	if err != nil {
		return nil, fmt.Errorf("NewExtFullTX error %s", err)
	}
	ledgerStore.fullTxStore = fullTxStore

	ledgerStore.LightLedger = storages.GetLightLedger()
	return ledgerStore, nil
}
func (this *LedgerStoreImp) getCurrentBlockHeader(leagueId common.Address) *orgtypes.Header {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeader[leagueId]
}

func (this *LedgerStoreImp) MainTxUsedExist(txHash common.Hash) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.mainTxUsedStore.Exist(txHash)
}
func (this *LedgerStoreImp) GetBlockHashSpan(leagueId common.Address, start, end uint64) (hashArr common.HashArray, gasUsedSum *big.Int, err error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.blockStore.GetBlockHashSpan(leagueId, start, end)

}
func (this *LedgerStoreImp) GetExtDataByHeight(leagueId common.Address, height uint64) (*extstates.ExtData, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	fmt.Printf("🚫 💿 GetExtDataByHeight part0 leagueId:%s,height:%d\n", leagueId.ToString(), height)
	block, err := this.blockStore.GetBlock(leagueId, height)
	if err != nil {
		fmt.Printf("🚫 💿 GetExtDataByHeight part1 err:%s,leagueId:%s,height:%d\n", err, leagueId.ToString(), height)
		return nil, err
	}
	addrs, hashes, err := this.acctIdxStore.Get(height, leagueId)
	if err != nil {
		fmt.Printf("🚫 💿 GetExtDataByHeight part2 err:%s,leagueId:%s,height:%d\n", err, leagueId.ToString(), height)
		return nil, err
	}
	accountStates := make([]*extstates.EasyLeagueAccount, len(addrs))
	if len(addrs) > 0 {
		for _, acc := range addrs {
			extAcc, err := this.accountStore.GetAccountByHeight(acc, leagueId, height)
			if err != nil {
				fmt.Printf("🚫 💿 GetExtDataByHeight part3 err:%s,leagueId:%s,height:%d,account:%s\n", err, leagueId.ToString(), height, acc.ToString())
				return nil, err
			}
			accountStates = append(accountStates, &extstates.EasyLeagueAccount{acc, extAcc})
		} //3
	}
	ed := new(extstates.ExtData)
	ed.Height = height
	ed.LeagueBlock = block
	ed.MainTxUsed = hashes //2
	ed.AccountStates = accountStates
	return ed, nil
}

/*func (this *LedgerStoreImp) GetExtDataMapByHeight(leagueHeight map[common.Address]uint64) []*extstates.ExtData {
	this.lock.RLock()
	defer this.lock.RUnlock()
	wg := new(sync.WaitGroup)
	wg.Add(len(leagueHeight))
	etdCh := make(chan *extstates.ExtData, 1)
	etds := make([]*extstates.ExtData, 0, len(leagueHeight))
	for k, v := range leagueHeight {
		go func(league common.Address, height uint64) {
			ed := new(extstates.ExtData)
			ed.Height = height
			block, err := this.blockStore.GetBlock(league, height)
			if err != nil {
				ed.Err = err
				etdCh <- ed
				return
			}
			ed.LeagueBlock = block //1
			addrs, hashes, err := this.acctIdxStore.Get(height, league)
			if err != nil {
				etdCh <- ed
				return
			}
			ed.MainTxUsed = hashes //2
			ed.AccountStates = make(map[common.Address]*extstates.Account, len(addrs))
			for _, acc := range addrs {
				extAcc, err := this.accountStore.GetAccountByHeight(acc, league, height)
				if err != nil {
					ed.Err = err
				}
				ed.AccountStates[acc] = extAcc
			} //3
			etdCh <- ed
			return
		}(k, v)
	}
	go func() {
		for ch := range etdCh {
			etds = append(etds, ch)
			wg.Done()
		}
	}()
	close(etdCh)
	return etds
}*/
func (this *LedgerStoreImp) SaveAll(blkInfo *storelaw.OrgBlockInfo, mainTxUsed common.HashArray) error {
	if this.isSavingBlock() {
		return nil
	}
	defer this.resetSavingBlock()
	if blkInfo.Block == nil {
		return errors.ERR_COMMON_REFERENCE_EMPTY
	}
	currentHeight := this.GetCurrentBlockHeight4V(blkInfo.Block.Header.LeagueId)
	blockHeight := blkInfo.Block.Header.Height
	if blockHeight <= currentHeight {
		return nil
	}
	nextBlockHeight := currentHeight + 1
	if blockHeight != nextBlockHeight {
		return fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
	}

	leagueBlock := extstates.Block2LeagueBlock(blkInfo.Block)
	this.blockStore.NewBatch()
	err := this.blockStore.Save(leagueBlock)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}
	err = this.blockStore.CommitTo()
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}
	err = this.saveFullTxs(blkInfo.Block.Header.LeagueId, blkInfo.Block.Header.Height, blkInfo.Block.Transactions)
	if err != nil {
		return fmt.Errorf("saveFullTxs error %s", err)
	}
	err = this.acctIdxStore.Save(blockHeight, blkInfo.Block.Header.LeagueId, blkInfo.AccStates.GetAccounts(), mainTxUsed)
	if err != nil {
		return fmt.Errorf("saveAccIdx error %s", err)
	}
	this.accountStore.NewBatch()
	if err = this.accountStore.BatchSave(blkInfo.AccStates); err != nil {
		return fmt.Errorf("saveAccountState error %s", err)
	}
	if err = this.accountStore.CommitTo(); err != nil {
		return fmt.Errorf("saveAccountState error %s", err)
	}
	this.mainTxUsedStore.NewBatch()
	if err = this.mainTxUsedStore.BatchSave(mainTxUsed, blkInfo.Block.Header.Height); err != nil {
		return err
	}
	if err = this.mainTxUsedStore.CommitTo(); err != nil {
		return err
	}
	err = this.saveVotes(blkInfo.VoteStates, blkInfo.AccountVoteds, blkInfo.Block.Header.Height)
	if err != nil {
		return err
	}
	err = this.saveUT(blkInfo.UT, blkInfo.Block.Header.LeagueId, blkInfo.Block.Header.Height)
	if err != nil {
		return err
	}
	this.setCurrentBlock(blkInfo.Block.Header)
	this.setDestroyedEnergy(blkInfo.Block.Header, blkInfo.FeeSum)
	fmt.Printf("🚫 -----S---A---V---E----A---L---L---  Height:%d  TxLen:%d blockHash:%s bonus:%d \n",
		blkInfo.Block.Header.Height,
		blkInfo.Block.Transactions.Len(),
		blkInfo.Block.Hash().String(),
		blkInfo.Block.Header.Bonus,
	)
	return nil
}
func (this *LedgerStoreImp) saveVotes(voteStates []*storelaw.VoteState, AccountVoteds []*storelaw.AccountVoted, height uint64) error {
	this.voteStore.NewBatch()
	this.voteStore.SaveVotes(voteStates)
	this.voteStore.SaveAccountVoted(AccountVoteds, height)
	err := this.voteStore.CommitTo()
	return err
}
func (this *LedgerStoreImp) saveUT(ut *big.Int, leagueId common.Address, height uint64) error {
	fmt.Println("🚫  mainchain saveUT ", ut.Uint64(), leagueId.ToString(), height)
	amount := this.utStore.GetUTByHeight(height, leagueId)
	if amount.Cmp(ut) == 0 {
		return nil
	}
	return this.utStore.Save(height, leagueId, ut)
}
func (this *LedgerStoreImp) saveFullTxs(leagueId common.Address, height uint64, txs orgtypes.Transactions) error {
	voteSources := make(orgtypes.Transactions, 0)
	for _, v := range txs {
		if v.TxType == orgtypes.VoteIncreaseUT {
			voteSources = append(voteSources, v)
		}
	}
	if voteSources.Len() > 0 {
		err := this.fullTxStore.SaveTxs(leagueId, height, voteSources)
		return err
	}
	return nil
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

func (this *LedgerStoreImp) setCurrentBlock(header *orgtypes.Header) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.currBlockHeader[header.LeagueId] = header
	if header.Height%types.HWidth == 0 {
		if this.bonusCache[header.LeagueId] == nil {
			this.bonusCache[header.LeagueId] = make(map[uint64]uint64)
		}
		this.bonusCache[header.LeagueId][header.Height] = header.Bonus
	}
	return
}

func (this *LedgerStoreImp) setDestroyedEnergy(header *orgtypes.Header, feeSum *big.Int) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if header.Height%types.HWidth == 0 {
		if this.destroyedEnergy[header.LeagueId] == nil {
			this.destroyedEnergy[header.LeagueId] = big.NewInt(0)
		}
		this.destroyedEnergy[header.LeagueId].SetUint64(0)
	}
	rat := new(big.Rat).Set(orgtypes.BonusRate)
	rat.Mul(rat, new(big.Rat).SetInt(feeSum))
	rfTmp := big.NewFloat(0).SetRat(rat)
	integer, _ := rfTmp.Uint64()
	this.destroyedEnergy[header.LeagueId].Add(this.destroyedEnergy[header.LeagueId], big.NewInt(int64(integer)))
}

////////////////////Ledger4Validation///////////////////
func (this *LedgerStoreImp) GetBlockHashByHeight4V(leagueId common.Address, height uint64) (common.Hash, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.blockStore.GetBlockHash(leagueId, height)
}
func (this *LedgerStoreImp) GetCurrentBlockHeight4V(leagueId common.Address) uint64 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if this.currBlockHeader[leagueId] == nil {
		return 0
	}
	return this.currBlockHeader[leagueId].Height
}
func (this *LedgerStoreImp) GetPrevAccount4V(height uint64, account, leagueId common.Address) (storelaw.AccountStater, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.accountStore.GetPrev(height, account, leagueId)
}

func (this *LedgerStoreImp) GetCurrentHeaderHash4V(leagueId common.Address) common.Hash {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if this.currBlockHeader[leagueId] == nil {
		return common.Hash{}
	}
	return this.currBlockHeader[leagueId].Hash()
}
func (this *LedgerStoreImp) ContainTx4V(txHash common.Hash) bool {
	return true
}

func (this *LedgerStoreImp) GetVoteState(voteId common.Hash, height uint64) (*storelaw.VoteState, error) {
	return this.voteStore.GetVoteState(voteId, height)
}
func (this *LedgerStoreImp) AlreadyVoted(voteId common.Hash, account common.Address) bool {
	return this.voteStore.AlreadyVoted(voteId, account)
}
func (this *LedgerStoreImp) GetUTByHeight(height uint64, leagueId common.Address) *big.Int {
	return this.utStore.GetUTByHeight(height, leagueId)
	/*this.utStore.GetUTByHeight(height, leagueId)
	fmt.Println("💗 💗 💗 💗 💗 💗 p1", height, leagueId.ToString())
	ls, err := this.lightLedger.GetLeagueByHeight(height, leagueId)
	fmt.Println("💗 💗 💗 💗 💗 💗 p1.5", ls, err)
	rate := uint32(0)
	if err != nil {
		return big.NewInt(0)
		fmt.Println("💗 💗 💗 💗 💗 💗 p2")
	} else {
		rate, _, _, err = this.lightLedger.GetMetaData(leagueId)
		fmt.Println("💗 💗 💗 💗 💗 💗 p3", rate, err)
	}
	fmt.Printf("💗 💗 💗 💗 💗 💗 p4 %v", ls)
	ut := big.NewInt(0)
	ut.Mul(ls.Data.FrozenBox, big.NewInt(0).SetUint64(uint64(rate)))*/
}
func (this *LedgerStoreImp) GetBonus(leagueId common.Address) map[uint64]uint64 {
	return this.bonusCache[leagueId]
}
func (this *LedgerStoreImp) GetTransaction(txHash common.Hash, leagueId common.Address) (*orgtypes.Transaction, uint64, error) {
	return this.fullTxStore.GetTx(leagueId, txHash)
}
func (this *LedgerStoreImp) GetHeaderBonus(leagueId common.Address) *big.Int {
	b := this.destroyedEnergy[leagueId]
	if b == nil {
		return big.NewInt(0)
	}
	return b
}
func (this *LedgerStoreImp) GetAccountRange(start, end uint64, account, leagueId common.Address) (storelaw.AccountStaters, error) {
	return this.accountStore.GetAccountRange(start, end, account, leagueId)
}

////////////////////Ledger4Validation///////////////////
