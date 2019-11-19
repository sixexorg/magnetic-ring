package mainchain

import (
	"math/big"
	"sync"
	"time"

	"fmt"

	"log"

	"github.com/ahmetb/go-linq"
	//"github.com/vechain/thor/block"
	"github.com/sixexorg/magnetic-ring/common"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
	"github.com/sixexorg/magnetic-ring/store/orgchain/genesis"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

type LeagueConsumer struct {
	m        sync.RWMutex
	pubkey   crypto.PublicKey //init from genesis block
	leagueId common.Address
	//currentBlock *orgtypes.Block                   //the block in analysis
	currBlockHash common.Hash                                                               //1
	curHeight     uint64                                                                    //1
	lockerr       error                             // for league lock  and this signal signal indicates that the consumer will not validate new blocks
	lastHeight    uint64                            //lastheight change when determine how much height to cut
	blockPool     map[uint64]*orgtypes.Block        //the receive block into here and remove the map key which less than lastHeight
	mainTxUsed    common.HashArray                  //Cache the main chain block reference of the current compute block，it's persisted to the database at the end
	mainTxsNew    map[uint64]maintypes.Transactions //newly produced main chain tx，Delete after generateMainTx(func) or checkLeagueResult(func)
	receiveSign   chan uint64                       //notice end
	adapter       *ConsumerAdapter
	willdo        uint64
	//needRemove    []uint64
	executing bool
}

//receiveBlock is shunt the block to its corresponding location, which requires special handling if it is a genesis block
//func (this *LeagueConsumer) receiveBlock(block *orgtypes.Block, teller Teller, initTeller InitTellerfunc) {
func (this *LeagueConsumer) receiveBlock(block *orgtypes.Block) {
	if block.Header.Height <= this.currentHeight() {
		return
	}
	if block.Header.Height == 1 {
		if this.pubkey != nil {
			return
		}
		//todo Verify the circle tx of the public chain and initialize the creation data
	} else {
		if !this.authenticate(block) {
			return
		}
	}
	fmt.Println("📡 receiveBlock")
	this.fillBlockPool(block)
}
func (this *LeagueConsumer) commitLastHeight() {
	if this.willdo != 0 {
		willdo := this.willdo
		last := this.getLastHeight()
		this.setLastHeight(willdo)
		this.removeMainNewBefore(last, willdo)
		fmt.Println("📡 commitLastHeight ", last, willdo)
	}
}
func (this *LeagueConsumer) rollbackWillDo() {
	this.m.Lock()
	defer this.m.Unlock()
	this.willdo = 0
}

func (this *LeagueConsumer) generateMainTx(right uint64) (mtc *MainTxCradle, err error) {
	if this.getLastHeight() >= right {
		return nil, errors.ERR_RADAR_NOT_FOUNT
	}
	if this.currentHeight() >= right {
		return this.generateMainTxPart1(this.getLastHeight()+1, right)
	}
	return nil, errors.ERR_RADAR_NOT_FOUNT
	/*	if this.currentHeight() < right {
			if this.lockerr != nil {
				//TODO Lock circle
				return this.generateMainTxPart1(this.getLastHeight()+1, this.currentHeight(), maintypes.LockLeague)

			} else {
				return nil, errors.ERR_RADAR_M_HEIGHT_BLOCKING
			}
		} else {
			return this.generateMainTxPart1(this.getLastHeight()+1, right, maintypes.ConsensusLeague)
		}*/
}
func (this *LeagueConsumer) generateMainTxPart1(left, right uint64) (*MainTxCradle, error) {
	hashes, gasSum, err := this.adapter.Ledger.GetBlockHashSpan(this.leagueId, left, right)
	if err != nil {
		return nil, err
	}
	fmt.Printf("📡 🔭  generateMainTxPart1 left:%d right:%d last:%d hashes:%v \n", left, right, this.getLastHeight(), hashes)
	tp := maintypes.ConsensusLeague
	if this.lockerr != nil {
		if this.currentHeight() == right {
			tp = maintypes.LockLeague
		}
	}

	mtx := &maintypes.Transaction{
		Version: maintypes.MainTxVersion,
		TxType:  tp,
		TxData: &maintypes.TxData{
			StartHeight: this.getLastHeight() + 1,
			EndHeight:   right,
			BlockRoot:   hashes.GetHashRoot(),
			LeagueId:    this.leagueId,
			Fee:         gasSum,
		},
	}
	if tp == maintypes.LockLeague {
		mtx.TxData.UnlockEnergy = big.NewInt(10000)
	}
	txCount := 0
	for i := left; i <= right; i++ {
		txCount += this.mainTxsNew[i].Len()
	}
	referenceTxs := make(maintypes.Transactions, 0, txCount)
	//removeK := make([]uint64, 0, right-left+1)
	for i := left; i <= right; i++ {
		referenceTxs = append(referenceTxs, this.mainTxsNew[i]...)
	}
	mtc := &MainTxCradle{
		LeagueId:     this.leagueId,
		Check:        mtx,
		ReferenceTxs: referenceTxs,
	}
	//this.needRemove = removeK
	this.willdo = right
	//TODO update
	//this.setLastHeight(right)
	//this.removeMainNewBefore(removeK)
	return mtc, nil
}

//RegisterConsume is controls the life cycle of a single round operation
func (this *LeagueConsumer) registerConsume(setStopSignal func(leagueId common.Address, tx *maintypes.Transaction), getLightHouse func() bool) {
	if this.executing {
		return
	}
	this.start()
	go func() {
		//TODO FOR TEST
		//this.clearHistorical()
		for height := range this.receiveSign {
			if height == this.currentHeight()+1 {
				err := this.consume()
				if err != nil {
					this.lockerr = err
					return
				}
			} else if height == 0 {
				tx, err := this.createMainChainTx()
				if err != nil {
					//todo log
					log.Println("func registerConsume createMainChainTx err ", err.Error())
				}
				setStopSignal(this.leagueId, tx)
				this.stop()
				return
			}
		}
	}()

	go func() {
		//todo This is the first use here. If you need to optimize the stop later, you can optimize it globally. Otherwise, this poll will increase in disguise.
		for {
			if getLightHouse() {
				this.receiveSign <- 0
			}
			time.Sleep(time.Second)
		}
	}()
}

func (this *LeagueConsumer) consume() error {
	//fmt.Println("🔆 📡 consume 0 start block height")
	block := this.findBlockPool(this.currentHeight() + 1)
	if block != nil {
		fmt.Println("🔆 📡 consume 1 start block height", block.Header.Height)
		blkInfo, err := this.verifyBlock(block)
		if err != nil {
			fmt.Println("🔆 📡 consume 2 lock", block.Header.Height, ",err:", err)
			//log.Println("func consume err ", err.Error())
			/*if block.Header.Height == 1 {
				return
			}*/
			return err
		}
		fmt.Println("🔆 📡 consume 2.5 saveall", blkInfo.Block.Header.Height)
		//todo Confirm tx generation, necessary data storage
		err = this.saveAll(blkInfo)
		fmt.Println("🔆 📡 consume 3 saveall", err)
		if err != nil {
			//log.Println("func consume err ", err.Error())
			return err
		}
		//this.blockHashes = append(this.blockHashes, block.Hash())
		this.setCurrentBlock(block)
		fmt.Println("🔆 📡 consume 4 ok")
	}
	return nil
}

//VerifyBlock is verify the block and compute all the details, which must be exactly the same as the circle validation
func (this *LeagueConsumer) verifyBlock(block *orgtypes.Block) (*storelaw.OrgBlockInfo, error) {
	fmt.Println("🚫 verifyBlock bonus ", block.Header.Bonus)
	if this.currentHeight()+1 != block.Header.Height {
		return nil, errors.ERR_EXTERNAL_HEIGHT_NOT_MATCH
	}
	if this.currentBlockHash() != block.Header.PrevBlockHash {
		return nil, fmt.Errorf("%s,the real prevHash is %s ,the wrong is %s",
			errors.ERR_EXTERNAL_PREVBLOCKHASH_DIFF,
			this.currentBlockHash().String(),
			block.Header.PrevBlockHash.String())
	}
	//todo more detailed description of the error
	if block.Header.Height == 1 {
		tx, height, err := this.adapter.FuncGenesis(block.Header.LeagueId)
		if err != nil {
			return nil, err
		}

		blockRef, err := this.adapter.FuncBlk(height)
		if err != nil {
			return nil, err
		}

		blockGenesis, sts := genesis.GensisBlock(tx, blockRef.Timestamp)
		this.leagueId = block.Header.LeagueId
		stsExt := make(storelaw.AccountStaters, 1)
		ut := big.NewInt(0)
		for _, v := range sts {
			stsExt[0] = &extstates.LeagueAccountState{
				Address:  v.Account(),
				LeagueId: block.Header.LeagueId,
				Height:   block.Header.Height,
				Data: &extstates.Account{
					Nonce:       v.Nonce(),
					Balance:     v.Balance(),
					EnergyBalance: v.Energy(),
					BonusHeight: 0,
				},
			}
			ut.Add(ut, v.Balance())
		}

		/*blockGenesis.Header.Coinbase = block.Header.Coinbase
		blockGenesis.Header.Extra = block.Header.Extra*/

		blkInfo := &storelaw.OrgBlockInfo{
			Block:     blockGenesis,
			AccStates: stsExt,
			UT:        ut,
			FeeSum:    big.NewInt(0),
		}
		return blkInfo, nil
	} else {
		var errCheck error
		this.initializeMainTxHeight(block.Header.Height, block.Transactions.Len()/5)
		for _, v := range block.Transactions {
			errCheck = this.checkReferencedTransaction(v)
			if errCheck != nil {
				return nil, errCheck
			}
			if v.TxType != orgtypes.VoteIncreaseUT {
				this.orgTxBirthMainTx(v, block.Header.Height, block.Header.LeagueId, nil)
			}
		}
		blkInfo, err := this.adapter.FuncValidate(block, this.adapter.Ledger)

		fmt.Printf("🚫 cmp ⭕ \n"+
			"🚫❓ version %v %v \n"+
			"🚫❓ prevHash %s %s \n"+
			"🚫❓ leagueId %s %s \n"+
			"🚫❓ txRoot   %s %s \n"+
			"🚫❓ stateRoot %s %s \n"+
			"🚫❓ receipt %s %s \n"+
			"🚫❓ timestamp %d %d \n"+
			"🚫❓ height %d %d \n"+
			"🚫❓ difficulty %d %d \n"+
			"🚫❓ coinbase %s %s \n"+
			"🚫❓ extra %s %s\n"+
			"🚫❓ hash %s %s \n",
			blkInfo.Block.Header.Version, block.Header.Version,
			blkInfo.Block.Header.PrevBlockHash.String(), block.Header.PrevBlockHash.String(),
			blkInfo.Block.Header.LeagueId.ToString(), block.Header.LeagueId.ToString(),
			blkInfo.Block.Header.TxRoot.String(), block.Header.TxRoot.String(),
			blkInfo.Block.Header.StateRoot.String(), block.Header.StateRoot.String(),
			blkInfo.Block.Header.ReceiptsRoot.String(), block.Header.ReceiptsRoot.String(),
			blkInfo.Block.Header.Timestamp, block.Header.Timestamp,
			blkInfo.Block.Header.Height, block.Header.Height,
			blkInfo.Block.Header.Difficulty.Uint64(), block.Header.Difficulty.Uint64(),
			blkInfo.Block.Header.Coinbase.ToString(), block.Header.Coinbase.ToString(),
			string(blkInfo.Block.Header.Extra), string(block.Header.Extra),
			blkInfo.Block.Hash().String(), block.Hash().String(),
		)
		if err != nil {
			return nil, err
		}
		fmt.Println("🚫  voteFirstPass num ", blkInfo.VoteFirstPass.Len())
		for _, v := range blkInfo.VoteFirstPass {
			ff := func(leagueId common.Address) (rate uint32) {
				tx, _, _ := this.adapter.Ledger.LightLedger.GetTxByLeagueId(leagueId)
				return tx.TxData.Rate
			}
			this.orgTxBirthMainTx(v, block.Header.Height, block.Header.LeagueId, ff)

		}
		return blkInfo, nil
	}
}
func (this *LeagueConsumer) setBlockCurrentBySync(hash common.Hash, height uint64) {
	this.m.Lock()
	defer this.m.Unlock()
	if this.curHeight+1 == height {
		this.curHeight = height
		this.currBlockHash = hash
	}
}

func (this *LeagueConsumer) saveAll(blkInfo *storelaw.OrgBlockInfo) error {
	err := this.adapter.Ledger.SaveAll(blkInfo, this.mainTxUsed)
	if err != nil {
		return err
	}
	this.mainTxUsed = make(common.HashArray, 0, 50)
	this.clearBlockPool(blkInfo.Block.Header.Height)
	return nil
}

func (this *LeagueConsumer) createMainChainTx() (*maintypes.Transaction, error) {
	var (
		tx  *maintypes.Transaction
		err error = nil
	)
	left := this.getLastHeight()
	right := this.currentHeight()
	if right > left {
		hashes, gasSum, err := this.adapter.Ledger.GetBlockHashSpan(this.leagueId, left+1, right)
		if err != nil {
			return nil, err
		}
		//log.Println("createMainChainTx func: offset=>", l)
		txdata := &maintypes.TxData{
			LeagueId:    this.leagueId,
			StartHeight: left + 1,
			EndHeight:   right,
			BlockRoot:   hashes.GetHashRoot(),
			Energy:        gasSum,
		}
		if this.lockerr != nil {
			txdata.UnlockEnergy = unlockEnergy
			txdata.Msg = common.Hash{} //TODO
			tx, err = maintypes.NewTransaction(maintypes.LockLeague, 0x01, txdata)

		} else {
			tx, err = maintypes.NewTransaction(maintypes.ConsensusLeague, 0x01, txdata)
		}
	}
	return tx, err
}

func (this *LeagueConsumer) checkGenesisBlock(block *orgtypes.Block) {
	if this.currentHeight() != 1 {
		return
	}
	//this.pubkey = &ecdsa.PrivateKey{}
	//TODO get tx from mainchain

}

//checkReferencedTransaction is only check the transaction from the main chain to the league
func (this *LeagueConsumer) checkReferencedTransaction(tx *orgtypes.Transaction) error {
	if containMainType(tx.TxType) {

		if this.contains(tx.TxData.TxHash) {
			return errors.ERR_EXTERNAL_MAIN_TX_USED
		}
		bl := this.adapter.Ledger.MainTxUsedExist(tx.Hash())
		if bl {
			return errors.ERR_EXTERNAL_MAIN_TX_USED
		}
		this.mainTxUsed = append(this.mainTxUsed, tx.TxData.TxHash)
		txMain, _, err := this.adapter.FuncTx(tx.TxData.TxHash)
		if err != nil {
			return err
		}

		err = checkOrgTx(txMain, tx)
		return err
		/*//todo need receipt
		if tx.TxType == orgtypes.Join &&
			txMain.TxType == maintypes.JoinLeague &&
			txMain.TxData.From == tx.TxData.From {
			return nil

		} else if tx.TxType == orgtypes.EnergyFromMain &&
			txMain.TxType == maintypes.EnergyToLeague &&
			txMain.TxData.From == tx.TxData.From &&
			txMain.TxData.Energy.Cmp(tx.TxData.Energy) == 0 {
			return nil
		}
		return errors.ERR_EXTERNAL_TX_REFERENCE_WRONG*/
	}
	return nil
}
func (this *LeagueConsumer) contains(txHash common.Hash) bool {
	return linq.From(this.mainTxUsed).Contains(txHash)
}

//TODO Verify that the block is the one that the circle itself gave.
func (this *LeagueConsumer) authenticate(block *orgtypes.Block) bool {
	return true
}

func (this *LeagueConsumer) setCurrentBlock(block *orgtypes.Block) {
	this.m.Lock()
	defer this.m.Unlock()
	this.currBlockHash = block.Hash()
	this.curHeight = block.Header.Height
	fmt.Println("🚫 setCurrentBlock height ", block.Header.Height)
	/*this.adapter.initTeller()
	this.adapter.P2pTlr.Tell(&common.NodeLH{
		NodeId:   this.adapter.NodeId,
		LeagueId: this.leagueId,
		Height:   block.Header.Height,
	})*/
	//fmt.Printf("🚫 📫  send nodeLH nodeId:%s leagueId:%s height:%d\n", this.adapter.NodeId, this.leagueId.ToString(), block.Header.Height)
}
func (this *LeagueConsumer) currentHeight() uint64 {
	this.m.RLock()
	defer this.m.RUnlock()
	return this.curHeight
}
func (this *LeagueConsumer) setLastHeight(height uint64) {
	this.m.Lock()
	defer this.m.Unlock()
	this.lastHeight = height
	this.willdo = 0
}
func (this *LeagueConsumer) getLastHeight() uint64 {
	this.m.RLock()
	defer this.m.RUnlock()
	return this.lastHeight
}

func (this *LeagueConsumer) currentBlockHash() common.Hash {
	this.m.RLock()
	defer this.m.RUnlock()
	return this.currBlockHash
}
func (this *LeagueConsumer) UnlockLeague() {
	if this.lockerr != nil {
		this.lockerr = nil
	}
}

func (this *LeagueConsumer) start() {
	this.executing = true
}
func (this *LeagueConsumer) stop() {
	this.executing = false
}
func containMainType(key orgtypes.TransactionType) bool {
	_, ok := txTypeForReference[key]
	return ok
}

func containTransferToMainTx(key orgtypes.TransactionType) bool {
	_, ok := txTypeForTransfer[key]
	return ok
}
func (this *LeagueConsumer) initializeMainTxHeight(height uint64, cap int) {
	this.mainTxsNew[height] = make(maintypes.Transactions, 0, cap)
}
func (this *LeagueConsumer) orgTxBirthMainTx(tx *orgtypes.Transaction, height uint64, leagueId common.Address, f func(leagueId common.Address) (rate uint32)) {
	mainTx := orgTxBirthMainTx(tx, height, leagueId, f)
	if mainTx != nil {
		fmt.Println("🚫 orgTxBirthMainTx append txType ", mainTx.TxType, "height ", height)
		this.appendMainNew(mainTx, height)
	}
}
func (this *LeagueConsumer) appendMainNew(mainTx *maintypes.Transaction, height uint64) {
	this.m.Lock()
	defer this.m.Unlock()
	this.mainTxsNew[height] = append(this.mainTxsNew[height], mainTx)
}
func (this *LeagueConsumer) removeMainNewBefore(start, end uint64) {
	this.m.Lock()
	defer this.m.Unlock()
	for i := start; i <= end; i++ {
		delete(this.mainTxsNew, i)
	}
}
func (this *LeagueConsumer) fillBlockPool(block *orgtypes.Block) {
	this.m.Lock()
	defer this.m.Unlock()
	if this.blockPool[block.Header.Height] == nil {
		this.blockPool[block.Header.Height] = block
		/*opd := &p2pcommon.OrgPendingData{
			BANodeSrc: false,
			OrgId:     block.Header.LeagueId,
			Block:     block,
		}*/
		this.receiveSign <- block.Header.Height
		/*	this.adapter.initTeller()
			this.adapter.P2pTlr.Tell(opd)*/

		/*	actor, err := bactor.GetActorPid(bactor.P2PACTOR)
			if err != nil {
				panic(err)
			}
			actor.Tell(opd)*/
	}
}
func (this *LeagueConsumer) findBlockPool(height uint64) *orgtypes.Block {
	this.m.RLock()
	defer this.m.RUnlock()
	return this.blockPool[height]
}
func (this *LeagueConsumer) clearBlockPool(targetHeight uint64) {
	this.m.Lock()
	defer this.m.Unlock()
	for k, _ := range this.blockPool {
		if k <= targetHeight {
			delete(this.blockPool, k)
		}
	}
}
