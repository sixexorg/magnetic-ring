package storages

import (
	"fmt"
	"os"
	"sync"

	"path"

	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/log"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/account_level"
	"github.com/sixexorg/magnetic-ring/store/mainchain/actor"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	//scom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
	"encoding/binary"
	"github.com/sixexorg/magnetic-ring/node"
	"github.com/sixexorg/magnetic-ring/config"
)

const (
	SYSTEM_VERSION          = byte(1)      //Version of ledger store
	HEADER_INDEX_BATCH_SIZE = uint64(2000) //Bath size of saving header index
	cycle                   = int(1)
)

var (
	//Storage save path.
	DBDirBlock          = "block"
	DBDirState          = "states"
	DBDirLeague         = "league"
	DBDirMember         = "member"
	DBDirBirthCert      = "birth"
	DBDirReceipt        = "receipt"
	DBDirLvl            = "lvl"
	MerkleTreeStorePath = "merkle_tree.db"
)

//LedgerStoreImp is main store struct fo ledger
type LedgerStoreImp struct {
	blockStore *BlockStore //BlockStore for saving block & transaction data
	//stateStore *StateStore //StateStore for saving state data, like balance, smart contract execution result, and so on.
	lvlManager     *account_level.LevelManager
	accountStore   *AccountStore // Latest balance
	leagueStore    *LeagueStore  // How many funds are frozen in the circle, creating circles, issuing additional operations, etc.
	memberStore    *MemberStore  // User status on the main chain
	birthCertStore *LeagueBirthCertStore  // Org trading
	receiptStore   *ReceiptStore  // Transaction receipt

	storedIndexCount uint64                        //record the count of have saved block index
	currBlockHeight  uint64                        //Current block height                                1
	currBlockHash    common.Hash                   //Current block hash                                  1
	headerCache      map[common.Hash]*types.Header //BlockHash => Header
	headerIndex      map[uint64]common.Hash        //Header index, Mapping header height => block hash
	savingBlock      bool                          //is saving block now

	bonusCache map[uint64][]uint64
	lock       sync.RWMutex
}

var ledgerStore *LedgerStoreImp

func (this *LedgerStoreImp) GetAccountStore() *AccountStore {
	return this.accountStore
}

func GetLedgerStore() *LedgerStoreImp {
	/*	if ledgerStore == nil {
		return nil, errors.ERR_SETUP_MAIN_LEDGER_UNINITIALIZED
	}*/
	return ledgerStore
}

//NewLedgerStore return LedgerStoreImp instance
func NewLedgerStore(dataDir string) (*LedgerStoreImp, error) {
	if ledgerStore != nil {
		return ledgerStore, nil
	}

	ledgerStore = &LedgerStoreImp{
		headerIndex: make(map[uint64]common.Hash),
		headerCache: make(map[common.Hash]*types.Header, 0),
		bonusCache:  make(map[uint64][]uint64),
	}
	blockStore, err := NewBlockStore(path.Join(dataDir, DBDirBlock), true)
	if err != nil {
		return nil, fmt.Errorf("NewBlockStore error %s", err)
	}
	ledgerStore.blockStore = blockStore

	accountStore, err := NewAccoutStore(path.Join(dataDir, DBDirState), false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewAccoutStore error %s\n", err)
		return nil, err
	}

	ledgerStore.accountStore = accountStore

	leagueStore, err := NewLeagueStore(path.Join(dataDir, DBDirLeague))
	if err != nil {
		return nil, fmt.Errorf("NewLeagueStore error %s", err)
	}
	ledgerStore.leagueStore = leagueStore

	memberStore, err := NewMemberStore(path.Join(dataDir, DBDirMember))
	if err != nil {
		return nil, fmt.Errorf("NewMemberStore error %s", err)
	}
	ledgerStore.memberStore = memberStore

	birthStore, err := NewLeagueBirthCertStore(path.Join(dataDir, DBDirBirthCert))
	if err != nil {
		return nil, fmt.Errorf("NewMemberStore error %s", err)
	}
	ledgerStore.birthCertStore = birthStore
	receiptStore, err := NewReceiptStore(path.Join(dataDir, DBDirReceipt), false)
	if err != nil {
		return nil, fmt.Errorf("NewReceiptStore error %s", err)
	}
	ledgerStore.receiptStore = receiptStore
	/*stateStore, err := NewStateStore(fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirState),
		fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), MerkleTreeStorePath))
	if err != nil {
		return nil, fmt.Errorf("NewStateStore error %s", err)
	}
	ledgerStore.stateStore = stateStore*/

	/*lvlManager, err := account_level.NewLevelManager(cycle, types.HWidth, path.Join(dataDir, DBDirLvl))
	if err != nil {
		return nil, fmt.Errorf("NewLevelManager error %s", err)
	}
	ledgerStore.lvlManager = lvlManager*/
	ledgerStore.restoreVars()

	return ledgerStore, nil
}

func (this *LedgerStoreImp) RestoreVars() {
	this.restoreVars()
}

func (this *LedgerStoreImp) restoreVars() {
	//this.currBlockHash, this.currBlockHeight, _ = this.blockStore.GetCurrentBlock()
	//this.blockStore.GetCurrentBlock()
	var err error
	var blk *types.Block
	prefix := make([]byte, 1, 1)
	prefix[0] = 0x00
	iter := this.blockStore.store.NewIterator(prefix)
	if iter.Last() {
		key := iter.Key()
		this.currBlockHeight = binary.LittleEndian.Uint64(key[1:])
		value := iter.Value() //hash
		this.currBlockHash, err = common.ParseHashFromBytes(value)
		if err != nil {
			return
		}
		blk, _ = this.GetBlockByHash(this.currBlockHash)
	}
	this.storedIndexCount = this.currBlockHeight
	log.Info("LedgerStoreImp restoreVars currBlockHash：%v, currBlockHeight:%v, .storedIndexCount:%v",
		this.currBlockHash.String(), this.currBlockHeight, this.storedIndexCount)
	fmt.Println("------=========:", this.currBlockHash, this.currBlockHeight, this.storedIndexCount)


	if this.currBlockHeight > 1{
		stars := make([]string, 0, len(config.GlobalConfig.Genesis.Stars))
		for _, star := range config.GlobalConfig.Genesis.Stars{
			stars = append(stars, star.Nodekey)
		}
		node.PushStars(stars)
	}

	if actor.ConsensusPid != nil {
		fmt.Println()
		actor.ConsensusPid.Tell(blk)
	}

	//txpoolPid, err := bactor.GetActorPid(bactor.TXPOOLACTOR)
	//
	//if err != nil {
	//	log.Error("txpoolactor not found")
	//} else {
	//	if blockInfo.Block.Header.Height > 1 {
	//		hi := bactor.HeightChange{blockInfo.Height()}
	//		txpoolPid.Tell(hi)
	//	}
	//}

	radarActor, err := bactor.GetActorPid(bactor.MAINRADARACTOR)
	if err != nil {
		log.Error("radarActor not found")
	} else {
		radarActor.Tell(this.currBlockHeight)
	}
}

func (this *LedgerStoreImp) ContainBlock(blockHash common.Hash) (bool, error) {
	return this.blockStore.ContainBlock(blockHash)
}
func (this *LedgerStoreImp) ContainTx(txHash common.Hash) bool {
	result, _ := this.blockStore.ContainTransaction(txHash)
	return result
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

//GetBlockHash return the block hash by block height
func (this *LedgerStoreImp) GetBlockHeaderHashByHeight(height uint64) common.Hash {
	return this.getHeaderIndex(height)
}

func (this *LedgerStoreImp) GetBlockHeaderByHeight(height uint64) (*types.Header, error) {
	blockHash, err := this.blockStore.GetBlockHash(height)
	if err != nil {
		return nil, err
	}
	return this.blockStore.GetHeader(blockHash)
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

/*func (this *LedgerStoreImp) setCurrentBlock(height uint64, blockHash common.Hash) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.currBlockHash = blockHash
	this.currBlockHeight = height
	return
}*/
func (this *LedgerStoreImp) setCurrentBlock(header *types.Header) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.currBlockHeight < header.Height {
		this.currBlockHash = header.Hash()
		this.currBlockHeight = header.Height
	}
/*	if header.Height%types.HWidth == 0 {
		this.bonusCache[header.Height] = []uint64{
			header.Lv1,
			header.Lv2,
			header.Lv3,
			header.Lv4,
			header.Lv5,
			header.Lv6,
			header.Lv7,
			header.Lv8,
			header.Lv9,
		}
	}*/
	return
}

//GetCurrentBlock return the current block height, and block hash.
//Current block means the latest block in store.
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
func (this *LedgerStoreImp) GetAccount(account common.Address) (*states.AccountState, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.accountStore.GetPrev(this.GetCurrentBlockHeight(), account)
}
func (this *LedgerStoreImp) GetAccountByHeight(height uint64, account common.Address) (*states.AccountState, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.accountStore.GetPrev(height, account)
}
func (this *LedgerStoreImp) GetLeagueByHeight(height uint64, leagueId common.Address) (*states.LeagueState, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.leagueStore.GetPrev(height, leagueId)
}

/*func (this *LedgerStoreImp) GetMetaData(leagueId common.Address) (rate uint32, creator common.Address, minBox uint64, err error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.leagueStore.GetMetaData(leagueId)
}*/
func (this *LedgerStoreImp) GetLeagueMemberByHeight(height uint64, leagueId, account common.Address) (*states.LeagueMember, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.memberStore.GetPrev(height, leagueId, account)

}
func (this *LedgerStoreImp) SymbolExists(symbol common.Symbol) error {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.leagueStore.SymbolExists(symbol)
}

func (this *LedgerStoreImp) GetTxByHash(txHash common.Hash) (*types.Transaction, uint64, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.blockStore.GetTransaction(txHash)
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
	/*if prevHeader.Timestamp >= header.Timestamp {
		return fmt.Errorf("block timestamp is incorrect")
	}*/
	return nil
}

func (this *LedgerStoreImp) SaveAll(
	blockInfo *BlockInfo) error {
	if this.isSavingBlock() {
		//hash already saved or is saving
		return nil
	}
	defer this.resetSavingBlock()
	/*wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		this.lvlManager.ReceiveAccountStates(blockInfo.AccountStates, blockInfo.GasDestroy, blockInfo.Height())
		wg.Done()
	}()*/

	err := this.addBlock(blockInfo.Block)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}
	if err = this.addAccountState(blockInfo.AccountStates); err != nil {

		return fmt.Errorf("saveAccountState error %s", err)
	}
	if err = this.addLeagueMembers(blockInfo.Members); err != nil {
		return fmt.Errorf("saveLeagueMemberState error %s", err)
	}
	if err = this.addLeagueState(blockInfo.LeagueStates); err != nil {
		return fmt.Errorf("saveLeagueState error %s", err)
	}
	if err = this.addBirthCerts(blockInfo.LeagueKVs); err != nil {
		return fmt.Errorf("saveBirthCert error %s", err)
	}
	if err = this.saveReceipts(blockInfo.Receipts); err != nil {
		return fmt.Errorf("saveReceipt error %s", err)
	}
	this.delHeaderCache(blockInfo.Block.Hash())
	this.setCurrentBlock(blockInfo.Block.Header)

	if actor.ConsensusPid != nil {
		actor.ConsensusPid.Tell(blockInfo.Block)
	}

	txpoolPid, err := bactor.GetActorPid(bactor.TXPOOLACTOR)

	if err != nil {
		log.Error("txpoolactor not found")
	} else {
		if blockInfo.Block.Header.Height > 1 {
			hi := bactor.HeightChange{blockInfo.Height()}
			txpoolPid.Tell(hi)
		}
	}

	radarActor, err := bactor.GetActorPid(bactor.MAINRADARACTOR)
	if err != nil {
		log.Error("radarActor not found")
	} else {
		radarActor.Tell(blockInfo.Height())
	}
	/*	for _, v := range blockInfo.AccountStates {
		fmt.Println("save accountState", v.Address.ToString(), v.Data.Nonce)
	}*/
	log.Info("func storages saveAll", "blockHeight", blockInfo.Block.Header.Height, "txlen", blockInfo.Block.Transactions.Len())
	fmt.Printf("🔆 -----S---A---V---E----A---L---L---  Height:%d TxLen:%d Hash:%s lv1:%d lv2:%d lv3:%d lv4:%d lv5:%d lv6:%d lv7:%d lv8:%d lv9:%d\n",
		blockInfo.Block.Header.Height,
		blockInfo.Block.Transactions.Len(),
		blockInfo.Block.Hash().String(),
		//blockInfo.Block.Header.Lv1, blockInfo.Block.Header.Lv2, blockInfo.Block.Header.Lv3, blockInfo.Block.Header.Lv4, blockInfo.Block.Header.Lv5,
		//blockInfo.Block.Header.Lv6, blockInfo.Block.Header.Lv7, blockInfo.Block.Header.Lv8, blockInfo.Block.Header.Lv9,
	)

	/*
		if blockInfo.Block.Sigs != nil {
			for _, v := range blockInfo.Block.Sigs.FailerSigs {
				p, _ := GetNode(stars, int(byte2Int(v[0:2])))
				fmt.Println("BlockStore computeFairs", p, byte2Int(v[0:2]))
			}
			for _, v := range blockInfo.Block.Sigs.ProcSigs {
				p, _ := GetNode(stars, int(byte2Int(v[0:2])))
				fmt.Println("BlockStore computeProc", p, byte2Int(v[0:2]))
			}
			for _, v := range blockInfo.Block.Sigs.TimeoutSigs {
				p, _ := GetNode(stars, int(byte2Int(v[0:2])))
				fmt.Println("BlockStore computeTimeout", p, byte2Int(v[0:2]))
			}
		}*/
	//this.Print(blockInfo.Block.Header.Height)
	//wg.Wait()
	return nil
}
func (this *LedgerStoreImp) Print(height uint64) {
	blk, err := this.GetBlockByHeight(height)
	if err != nil {
		fmt.Println("🐷🐷print err", err)
	}
	for k, v := range blk.Transactions {
		if v.TxType == types.TransferEnergy || v.TxType == types.TransferBox {
			fmt.Println("🐼🐼num:", k, ",hash:", v.Hash())
		}
	}
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
func (this *LedgerStoreImp) addLeagueMembers(states states.LeagueMembers) error {
	this.memberStore.NewBatch()
	err := this.memberStore.BatchSave(states)
	if err != nil {
		return err
	}
	if err = this.memberStore.CommitTo(); err != nil {
		return err
	}
	return nil
}
func (this *LedgerStoreImp) addLeagueState(states states.LeagueStates) error {
	this.leagueStore.NewBatch()
	err := this.leagueStore.BatchSave(states)
	if err != nil {
		return err
	}
	if err = this.leagueStore.CommitTo(); err != nil {
		return err
	}
	return nil
}
func (this *LedgerStoreImp) addBirthCerts(kvs []*LeagueKV) error {
	this.birthCertStore.NewBatch()
	for _, v := range kvs {
		this.birthCertStore.BatchPutBirthCert(v.K, v.V)
	}
	return this.birthCertStore.CommitTo()
}
func (this *LedgerStoreImp) addAccountState(accountStates states.AccountStates) error {
	this.accountStore.NewBatch()
	err := this.accountStore.BatchSave(accountStates)
	if err != nil {
		return err
	}
	if err = this.accountStore.CommitTo(); err != nil {
		return err
	}
	return nil
}

func (this *LedgerStoreImp) addBlock(block *types.Block) error {

	currBlockHeight := this.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		return fmt.Errorf("The block height is less than the current height")
	}
	nextBlockHeight := currBlockHeight + 1
	if blockHeight != nextBlockHeight {
		return fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
	}
	err := this.verifyHeader(block.Header)
	if err != nil {
		return fmt.Errorf("verifyHeader error %s", err)
	}
	return this.saveBlock(block)
}

func (this *LedgerStoreImp) GetTxHashesByAccount(account common.Address) {
	this.lock.RLock()
	defer this.lock.RUnlock()

}
func (this *LedgerStoreImp) GetAccountStateHeights(account common.Address) {

}

func (this *LedgerStoreImp) GetNextHeaderProperty(height uint64) ([]uint64, error) {
	return this.lvlManager.GetNextHeaderProperty(height)

}
func (this *LedgerStoreImp) saveBlock(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	this.blockStore.NewBatch()
	//err := this.saveHeaderIndexList() // 2000高度存一次 暂时考虑废弃
	//if err != nil {
	//	return fmt.Errorf("saveHeaderIndexList error %s", err)
	//}
	var err error
	this.blockStore.SaveBlockHash(blockHeight, blockHash) // 高度-块哈希
	if err = this.blockStore.SaveBlock(block); err != nil {
		return fmt.Errorf("SaveBlock height %d hash %s error %s", blockHeight, blockHash.String(), err)
	}
	err = this.blockStore.CommitTo()
	if err != nil {
		return err
	}
	this.setHeaderIndex(blockHeight, blockHash)
	return nil
}

func (this *LedgerStoreImp) SaveAccounts(states states.AccountStates) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.accountStore.NewBatch()
	err := this.accountStore.BatchSave(states)
	if err != nil {
		return err
	}
	err = this.accountStore.CommitTo()
	return err
}
func (this *LedgerStoreImp) saveReceipts(receipts types.Receipts) error {
	this.receiptStore.NewBatch()
	this.receiptStore.BatchSave(receipts)
	err := this.receiptStore.CommitTo()
	return err
}
func (this *LedgerStoreImp) SaveLeagues(states states.LeagueStates, members states.LeagueMembers) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.leagueStore.NewBatch()
	err := this.leagueStore.BatchSave(states)
	if err != nil {
		return err
	}
	if err = this.leagueStore.CommitTo(); err != nil {
		return err
	}

	this.memberStore.NewBatch()
	if err = this.memberStore.BatchSave(members); err != nil {
		return err
	}
	if err = this.memberStore.CommitTo(); err != nil {
		return err
	}
	return nil
}

func (this *LedgerStoreImp) GetTxByLeagueId(leagueId common.Address) (*types.Transaction, uint64, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	txHash, err := this.birthCertStore.GetBirthCert(leagueId)
	if err != nil {
		return nil, 0, err
	}
	tx, height, err := this.blockStore.GetTransaction(txHash)
	if err != nil {
		return nil, 0, err
	}
	return tx, height, nil
}
func (this *LedgerStoreImp) SaveBlockForMockTest(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	this.setHeaderIndex(blockHeight, blockHash)
	this.blockStore.NewBatch()
	this.blockStore.SaveBlockHash(blockHeight, blockHash)
	err := this.blockStore.SaveBlock(block)
	if err != nil {
		return fmt.Errorf("SaveBlock height %d hash %s error %s", blockHeight, blockHash.String(), err)
	}
	err = this.blockStore.CommitTo()
	if err != nil {
		return err
	}
	this.delHeaderCache(block.Hash())
	this.setCurrentBlock(block.Header)
	return nil
}
func (this *LedgerStoreImp) SaveBirthCertForMockTest(leagueId common.Address, txHash common.Hash) error {
	this.birthCertStore.NewBatch()
	this.birthCertStore.BatchPutBirthCert(leagueId, txHash)
	err := this.birthCertStore.CommitTo()
	return err
}

func (this *LedgerStoreImp) GetAccountRange(start, end uint64, account common.Address) (states.AccountStates, error) {
	return this.accountStore.GetAccountRange(start, end, account)
}
func (this *LedgerStoreImp) GetAccountLvlRange(start, end uint64, account common.Address) []common.FTreer {
	return this.lvlManager.LevelStore.GetAccountLvlRange(start, end, account)
}
func (this *LedgerStoreImp) GetAccountLevel(height uint64, account common.Address) account_level.EasyLevel {
	return this.lvlManager.LevelStore.GetAccountLevel(height, account)
}
func (this *LedgerStoreImp) GetBouns() map[uint64][]uint64 {
	return this.bonusCache
}
