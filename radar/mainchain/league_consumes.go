package mainchain

import (
	"log"
	"math/big"
	"sync"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"

	"time"

	"bytes"

	"sort"

	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstorages"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

var (
	leagueConsumers    *LeagueConsumers
	unlockEnergy         = big.NewInt(0)
	txTypeForReference = make(map[orgtypes.TransactionType]struct{})
	txTypeForTransfer  = make(map[orgtypes.TransactionType]struct{})
)

func init() {
	txTypeForReference[orgtypes.EnergyFromMain] = struct{}{}
	txTypeForReference[orgtypes.Join] = struct{}{}

	txTypeForTransfer[orgtypes.EnergyToMain] = struct{}{}
}

type LeagueConsumers struct {
	m *sync.Mutex

	lgHtCache map[common.Address]uint64 //last get league height cache

	blockCh   chan *orgtypes.Block
	consumers map[common.Address]*LeagueConsumer // Circle address  Circle consumer

	isminer  bool
	adpter   *ConsumerAdapter
	ledgeror storelaw.Ledger4Validation

	//These properties are designed to perform switching logic，
	// but this logic is not needed now and is now used for unit tests
	wg         *sync.WaitGroup
	lighthouse bool //Notifies the program that this round of calculations will be closed
	leagueSize int                               // Number of circles
	stopSignal chan common.Address
	txs        maintypes.Transactions
}

func NewLeagueConsumers(adpter *ConsumerAdapter) *LeagueConsumers {
	leagueConsumers = &LeagueConsumers{
		m:          new(sync.Mutex),
		consumers:  make(map[common.Address]*LeagueConsumer),
		blockCh:    make(chan *orgtypes.Block, 10000), //todo Need to control the details of the upper limit, add later
		stopSignal: make(chan common.Address, 1000),
		wg:         new(sync.WaitGroup),
		adpter:     adpter,
	}
	go leagueConsumers.routing()
	//actor := NewMainRadarActor(leagueConsumers.ReceiveBlock)

	//pid, err := startActor(actor, "main radar actor")
	return leagueConsumers
}

func GetLeagueConsumersInstance() *LeagueConsumers {
	return leagueConsumers
}

func (this *LeagueConsumers) CommitLastHeight() {
	for _, v := range this.consumers {
		v.commitLastHeight()
	}
}
func (this *LeagueConsumers) RollbackWillDo() {
	for _, v := range this.consumers {
		v.rollbackWillDo()
	}
}

func (this *LeagueConsumers) GetLeagueHeight() map[common.Address]uint64 {
	re := make(map[common.Address]uint64)
	for k, v := range this.consumers {
		curHeight := v.currentHeight()
		re[k] = curHeight
	}
	this.lgHtCache = re
	return re
}
func (this *LeagueConsumers) GenerateMainTxs() (maintypes.Transactions, error) {
	return this.GenerateMainTxByLeagueHeight(nodeLHC.getNodeLHs())
}

func (this *LeagueConsumers) GenerateMainTxByLeagueHeight(earth map[common.Address]uint64, nodes []map[common.Address]uint64) (maintypes.Transactions, error) {
	local := this.GetLeagueHeight()
	lh := this.interceptHeight(local, earth, nodes...)
	for k, v := range lh {
		fmt.Println("🔗 🔗 🔗 key:", k.ToString(), " height:", v)
	}
	txs, err := this.generateMainTx(lh)
	if err != nil {
		return nil, err
	}
	/*	this.adpter.initTeller()
		this.adpter.TxPoolTlr.Tell(txpool.MustPackTxsReq{
			Txs: txs,
		})*/
	return txs, nil
}

type nodeHeight struct {
	ndIdx int
	h     uint64
}

//get

func (this *LeagueConsumers) interceptHeight(local, earth map[common.Address]uint64, nodes ...map[common.Address]uint64) map[common.Address]uint64 {
	l := len(local)
	if l == 0 {
		return nil
	}
	result := make(map[common.Address]uint64, l)
	rl := uint64(len(nodes))
	idx := rl / 2
	referenceIdx := make([]int, 0, idx+1)
	realm := this.min(local, earth)
	for ko, v := range realm {
		tmpV := v
		if tmpV == 0 {
			result[ko] = tmpV
		} else {
			if len(referenceIdx) == 0 {
				hsort := make([]*nodeHeight, 0, rl)
				for ndIdx, remote := range nodes {
					hsort = append(hsort, &nodeHeight{ndIdx: ndIdx, h: remote[ko]})
				}
				sort.Slice(hsort, func(i, j int) bool {
					return hsort[i].h < hsort[j].h
				})
				for _, rv := range hsort[idx:] {
					referenceIdx = append(referenceIdx, rv.ndIdx)
				}
				if hsort[idx].h > tmpV {
					result[ko] = tmpV
				} else {
					result[ko] = hsort[idx].h
				}
			} else {
				hs := make([]uint64, 0, idx+1)
				for _, rv := range referenceIdx {
					hs = append(hs, nodes[rv][ko])
				}
				sort.Slice(hs, func(i, j int) bool {
					return hs[i] < hs[j]
				})
				if hs[0] > tmpV {
					result[ko] = tmpV
				} else {
					result[ko] = hs[0]
				}
			}
		}
	}
	return result
}
func (this *LeagueConsumers) min(local, earth map[common.Address]uint64) map[common.Address]uint64 {
	result := make(map[common.Address]uint64, len(local))
	for k, v := range local {
		for ki, vi := range earth {
			if k.Equals(ki) {
				if v < vi {
					result[k] = v
				} else {
					result[k] = vi
				}
			}
		}
	}
	return result
}
func (this *LeagueConsumers) generateMainTx(leagueHM map[common.Address]uint64) (maintypes.Transactions, error) {
	mtcCh := make(chan *MainTxCradle, 50)
	leagueLen := len(leagueHM)
	wg := new(sync.WaitGroup)
	wg.Add(leagueLen)
	wg2 := new(sync.WaitGroup)
	for k, v := range leagueHM {
		lgcmer, ok := this.consumers[k]
		if ok {
			go func() {
				mtc, err := lgcmer.generateMainTx(v)
				if err != nil {
					//todo log
					fmt.Println("💔 ", lgcmer.leagueId.ToString(), " generate failed")
					wg.Done()
					return
				}
				wg2.Add(1)
				mtcCh <- mtc
				wg.Done()
			}()
		} else {
			wg.Done()
		}
	}
	txs := make(maintypes.Transactions, 0, leagueLen*5)
	go func() {
		for ch := range mtcCh {
			txs = append(txs, ch.Check)
			txs = append(txs, ch.ReferenceTxs...)
			wg2.Done()
		}
	}()

	wg.Wait()
	wg2.Wait()
	close(mtcCh)
	return txs, nil
}

type objTxForCheck struct {
	agg  *maintypes.Transaction
	sons maintypes.Transactions
}

/*func (this *LeagueConsumers) CheckLeaguesWithBoolResponse(txs maintypes.Transactions, timeout time.Duration) bool {
	ch := make(chan bool, 1)
	go func() {
		lsp := NewLeagueStatePipe()
		ch <- this.checkLeaguesResult(this.ConvertTxsToObjTxForCheck(txs), lsp, false)

	}()
	select {
	case re := <-ch:
		return re
	case <-time.After(timeout):
		return false
	}
}*/
func (this *LeagueConsumers) CheckLeaguesWithBoolResponse(txs maintypes.Transactions, timeout time.Duration) error {
	ch := make(chan bool, 1)
	go func() {
		lsp := NewLeagueStatePipe()
		objTx := this.ConvertTxsToObjTxForCheck(txs)
		leagueLen := len(objTx)
		this.checkLeaguesResult(objTx, lsp, true)
		for {
			select {
			case <-lsp.Successed:
				leagueLen--
				if leagueLen == 0 {
					fmt.Println("📡 ⏰ CheckLeaguesWithBoolResponse break")
					ch <- true
					break
				}
			}
		}
	}()
	select {
	case <-ch:
		fmt.Println("📡 ⏰ CheckLeaguesWithBoolResponse <- ch")
		this.CommitLastHeight()
		return nil
	case <-time.After(timeout):
		this.RollbackWillDo()
		return errors.ERR_TIME_OUT
	}
}

func (this *LeagueConsumers) CheckLeaguesResult(txs maintypes.Transactions, lsp *LeagueStatePipe) int {
	lm := this.ConvertTxsToObjTxForCheck(txs)
	this.checkLeaguesResult(lm, lsp, true)
	return len(lm)
}

func (this *LeagueConsumers) ConvertTxsToObjTxForCheck(txs maintypes.Transactions) map[common.Address]*objTxForCheck {
	leagueTxs := make(map[common.Address]*objTxForCheck)
	for _, v := range txs {
		tmpTX := v
		if leagueTxs[tmpTX.TxData.LeagueId] == nil {
			leagueTxs[tmpTX.TxData.LeagueId] = &objTxForCheck{}
		}
		if tmpTX.TxType == maintypes.ConsensusLeague || tmpTX.TxType == maintypes.LockLeague {
			leagueTxs[tmpTX.TxData.LeagueId].agg = tmpTX
		} else {
			leagueTxs[tmpTX.TxData.LeagueId].sons = append(leagueTxs[tmpTX.TxData.LeagueId].sons, tmpTX)
		}
	}
	return leagueTxs
}

func (this *LeagueConsumers) checkLeaguesResult(leagueTxs map[common.Address]*objTxForCheck, lsp *LeagueStatePipe, withState bool) bool {
	for k, v := range leagueTxs {
		go func(league common.Address, otfc *objTxForCheck) {
			if this.consumers[league] == nil {
				flag := true
				for {
					if flag {
						flag = false
						lsp.sendLeagueNeed(league, 1)
					}
					time.Sleep(time.Millisecond * 100)
					if this.consumers[league] != nil {
						fmt.Printf("🚫 🔐  check league first fill successed,the id is %s\n", league.ToString())
						break
					}
				}
			}
			this.consumers[league].checkLeagueResultPart2(otfc, this.adpter.Ledger, lsp, withState)
		}(k, v)
	}
	/*	if !withState {
		success := make(chan bool, 1)
		count := len(leagueTxs)
		go func(ch chan bool, c int) {
			for such := range lsp.Successed {
				if leagueTxs[such] != nil {
					c--
					if c == 0 {
						ch <- true
					}
				}
			}
		}(success, count)
		go func(ch chan bool) {
			for errch := range lsp.StateSignal {
				switch errch.(type) {
				case *LeagueNeed:
				case *LeagueExec:
				case *LeagueErr:
					ch <- false
				}
			}
		}(success)
		select {
		case re := <-success:
			return re
		}
	}*/
	return false
}

//checkLeagueResult is verify league height
//otfc: mainTxs create by consensus, ledger:storage,lsp:channel response,withState:callback which league height need synchronous,true for p2p,false for consensus
func (this *LeagueConsumer) checkLeagueResultPart2(otfc *objTxForCheck, ledger *extstorages.LedgerStoreImp, lsp *LeagueStatePipe, withState bool) {
	/*ch := make(chan bool, 1)
	go func() {*/
	start := otfc.agg.TxData.StartHeight
	end := otfc.agg.TxData.EndHeight
	isLock := otfc.agg.TxType == maintypes.LockLeague
	root := otfc.agg.TxData.BlockRoot
	var err error
	if this.currentHeight() >= end {
		goto DOIT
	} else {
		//todo Waiting for calculation, less logic
		//lsp.sendLeagueExec(this.leagueId, this.currentHeight())
		sendCache := make(map[uint64]bool)
		for {
			cHeight := this.currentHeight()
			fmt.Println("🚫 checkLeagueResultPart2 loop ", cHeight)
			if cHeight >= end {
				fmt.Println("🚫 checkLeagueResultPart2 break ", cHeight)
				break
			} else if withState {
				targetH := cHeight + 1
				if !sendCache[targetH] {
					sendCache[targetH] = true
					fmt.Printf("🚫 checkLeagueResultPart2 send need:%d target:%d \n", targetH, end)
					lsp.sendLeagueNeed(this.leagueId, targetH)
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
		fmt.Println("🚫 checkLeagueResultPart2 go doit")
		goto DOIT
	}
DOIT:
	fmt.Println("🚫 checkLeagueResultPart2 doit run")
	err = this.checkInterval(start, end, root, isLock, withState, ledger)
	if err != nil {
		lsp.sendError(this.leagueId, err)

	} else {
		this.willdo = otfc.agg.TxData.EndHeight
		lsp.sendOk(this.leagueId)
	}
}

func (this *LeagueConsumer) getLeagueState() uint64 {
	return this.currentHeight()
}

func (this *LeagueConsumer) checkInterval(start, end uint64, root common.Hash, islock, offsetUpdate bool, ledger *extstorages.LedgerStoreImp) error {
	hashes := make(common.HashArray, 0, end-start+1)
	needRemove := make([]uint64, 0, end-start+1)
	for i := start; i <= end; i++ {
		h, err := ledger.GetBlockHashByHeight4V(this.leagueId, i)
		if err != nil {
			//TODO Inconsistent preservation, need to be processed, a few
		}
		hashes = append(hashes, h)
		needRemove = append(needRemove, i)
	}
	hashRoot := hashes.GetHashRoot()
	if bytes.Equal(hashRoot[:], root[:]) {
		if (this.lockerr != nil) == islock {
			/*	if offsetUpdate {
				this.setLastHeight(end)
				this.removeMainNewBefore(needRemove)
			}*/
			return nil
		}
	}
	err := fmt.Errorf("diff between local and remote")
	return err
}

//SupplementLeagueConsumer is used to synchronize lost cross-chain storage
func (this *LeagueConsumers) SupplementLeagueConsumer(leagueId common.Address, blkHash common.Hash, height uint64) {
	league := this.getLeagueConsume(leagueId)
	if height == 1 {

		if !this.isminer && !this.lighthouse {
			//fmt.Println("🔆 📡 SupplementLeagueConsumer 1")
			go func(league *LeagueConsumer) {
				league.registerConsume(this.setStopSignal, this.getLightHouse)
			}(league)
		}
	}
	league.setBlockCurrentBySync(blkHash, height)
}

func (this *LeagueConsumers) ReceiveBlock(block *orgtypes.Block) {
	//fmt.Println("🔆 📡 🔆  1")
	this.blockCh <- block
	/*this.adpter.initTeller()
	this.adpter.P2pTlr.Tell(&p2pcommon.OrgPendingData{
		BANodeSrc: false,                 // true:ANode send staller false:staller send staller
		OrgId:     block.Header.LeagueId, // orgid
		Block:     block,                 // tx
	})*/
	//fmt.Println("🔆 📡 🔆  2")
}
func (this *LeagueConsumers) setStopSignal(leagueId common.Address, tx *maintypes.Transaction) {
	log.Println("setStopSignal:", leagueId.ToString())
	this.stopSignal <- leagueId
	if tx != nil {
		this.txs = append(this.txs, tx)
	}
}
func (this *LeagueConsumers) getLightHouse() bool {
	return this.lighthouse
}

//NewRound  is to actively launch validation
func (this *LeagueConsumers) NewRound() {
	this.m.Lock()
	defer this.m.Unlock()
	{
		this.isminer = true
		this.refresh()
		for _, v := range this.consumers {
			v.registerConsume(this.setStopSignal, this.getLightHouse)
		}
	}
}

//Truncate is to notifies the program that this round of calculations will be closed
func (this *LeagueConsumers) Settlement() {
	this.m.Lock()
	defer this.m.Unlock()
	{
		if this.lighthouse {
			return
		}
		this.lighthouse = true
		log.Println("leagueSize:", this.leagueSize)
		this.wg.Add(this.leagueSize)
		this.await()
		return
	}
}

//await is asynchronous waiting
func (this *LeagueConsumers) await() {
	go func() {
		for ch := range this.stopSignal {
			log.Println(ch.ToString(), " over")
			this.wg.Done()
		}
	}()
	this.wg.Wait()
}

func (this *LeagueConsumers) refresh() {
	this.lighthouse = false
	this.txs = nil
	this.stopSignal = make(chan common.Address, 1000)
}

func (this *LeagueConsumers) routing() {
	for block := range this.blockCh {
		//todo need check
		//fmt.Println("🔆 📡 routing 1")
		league := this.getLeagueConsume(block.Header.LeagueId)
		//fmt.Println("🔆 📡 routing 2", block.Header.Height)
		league.receiveBlock(block)
		//fmt.Println("🔆 📡 routing 3", this.isminer, this.lighthouse)
		if !this.isminer && !this.lighthouse {
			//fmt.Println("🔆 📡 routing 4 registerConsume")
			//league.registerConsume(this.adpter, this.setStopSignal, this.getLightHouse)
			go func(league *LeagueConsumer) {
				league.registerConsume(this.setStopSignal, this.getLightHouse)
			}(league)
		}
	}
}

func (this *LeagueConsumers) getLeagueConsume(leagueId common.Address) *LeagueConsumer {
	this.m.Lock()
	defer this.m.Unlock()
	{
		if this.consumers[leagueId] == nil {
			this.consumers[leagueId] = &LeagueConsumer{
				blockPool:   make(map[uint64]*orgtypes.Block),
				receiveSign: make(chan uint64, 100),
				leagueId:    leagueId,
				mainTxUsed:  make(common.HashArray, 0, 50),
				mainTxsNew:  make(map[uint64]maintypes.Transactions),
				adapter:     this.adpter,
			}
			//fmt.Println("blockHash len:", len(this.consumers[leagueId].blockHashes))
			this.leagueSize++
		}
		return this.consumers[leagueId]
	}
}

/*func (this *LeagueConsumers) checkMainTx(tx *maintypes.Transaction) error {
	if tx.TxType != maintypes.ConsensusLeague || tx.TxType != maintypes.LockLeague {
		return nil
	}
	if tx.TxData.StartHeight == 1 {
		if this.consumers[tx.TxData.LeagueId] != nil {
			return fmt.Errorf("maintx for league check not match.")
		}
		l := tx.TxData.EndHeight - tx.TxData.StartHeight
		asses := make([]storelaw.AccountStaters, 0, l)
		blocks := make([]*orgtypes.Block, l)
		for i := tx.TxData.StartHeight; i <= tx.TxData.EndHeight; i++ {
			block := this.consumers[tx.TxData.LeagueId].blockPool[i]
			if block == nil {
				return fmt.Errorf("maintx for league check not match.")
			}
			ass, _, err := this.consumers[tx.TxData.LeagueId].verifyBlock(block, this.adpter)
			if err != nil {
				return err
			}
			asses = append(asses, ass)
			blocks = append(blocks, block)
		}

	}
	return nil
}
*/
