package orgchain

import (
	"math/big"
	"sync"

	"log"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	mainstorages "github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	"github.com/sixexorg/magnetic-ring/store/orgchain/genesis"
	"github.com/sixexorg/magnetic-ring/store/orgchain/storages"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

var (
	txTypeForReference = make(map[maintypes.TransactionType]struct{})
)

func init() {
	txTypeForReference[maintypes.JoinLeague] = struct{}{}
	txTypeForReference[maintypes.EnergyToLeague] = struct{}{}
	txTypeForReference[maintypes.LockLeague] = struct{}{}
	txTypeForReference[maintypes.ConsensusLeague] = struct{}{}
	txTypeForReference[maintypes.RaiseUT] = struct{}{}

}

type Radar struct {
	leagueId       common.Address
	cfmStartHeight uint64 //Changes can only be made after the current phase has been processed
	cfmEndHeight   uint64 //Changes can only be made after the current phase has been processed
	isPrivate      bool
	blockRoot      common.Hash //Blockroot currently required to verify
	beVerified     bool        //Whether the current verification is complete
	verifing       bool
	growing        bool //mark genesis block has been created

	//MCCMap       map[uint64]*MainChainCfm
	//adapter      *radarAdapter //ledger
	ledger       *storages.LedgerStoreImp
	leagueLocked bool //Whether the league is locked by the main chain
	mainLedger   *mainstorages.LightLedger
	//chanTxCmf  chan *MainChainCfm
	//needVerify chan *MainChainCfm
	m *sync.Mutex

	//currentHeight  uint64      //Block height verified
	//mainY        bool          //if diff
	//cmpBlocks          []*orgtypes.Block        //local blocks
	//diffHeightTxHashes map[uint64][]common.Hash //local txHashes which height is not in the current block
	//cmpLD              *league_data.LeagueData            //Main chain consensus data which needs to be compared right now
	//LeagueDataPool     map[uint64]*league_data.LeagueData //all mian chain consensus data order by height
	//diffBlocks   []*orgtypes.Block //current different blocks ,need cmp
}

func NewRadar(mainLedger *mainstorages.LightLedger) *Radar {
	radar := &Radar{
		m: new(sync.Mutex),
		//adapter:    adapter,
		beVerified: true,
		mainLedger: mainLedger,
		/*MCCMap:     make(map[uint64]*MainChainCfm),
		chanTxCmf:  make(chan *MainChainCfm),
		needVerify: make(chan *MainChainCfm),*/
	}
	return radar
}
func (this *Radar) SetStorage(ledger *storages.LedgerStoreImp) {
	this.ledger = ledger
}
func (this *Radar) SetPrivate(isPrivate bool) {
	this.isPrivate = isPrivate

}

//MonitorGenesis is monitor the transactions of the main chain and analyze the information needed for the genesis info
func (this *Radar) MonitorGenesis(txch <-chan *maintypes.Transaction, account common.Address) (
	block *orgtypes.Block, //genesis block
	rate uint32,
	frozenBox *big.Int,
	minBox uint64,
	isPrivate bool,
	sts storelaw.AccountStaters,
	//to init accountstate
) {
	this.m.Lock()
	defer this.m.Unlock()
	{
		if !this.growing {
			fmt.Println("🌲 monitorGenesis start")
			i := 0
			for tx := range txch {
				fmt.Printf("genesis consume 👻 👻 👻 👻 👻 👻 👻 👻 👻 👻  👻  tx got hash-->%s\n",tx.TransactionHash)
				i++
				if tx.TxType == maintypes.CreateLeague {
					//todo Create a circle
					if tx.TxData.From.Equals(account) {
						//TODO Need to verify receipt
						frozenBox = tx.TxData.MetaBox
						minBox = tx.TxData.MinBox
						rate = tx.TxData.Rate
						isPrivate = tx.TxData.Private
						_, height, _ := this.mainLedger.GetTxByHash(tx.Hash())
						blockRef, _ := this.mainLedger.GetBlockByHeight(height)
						block, sts = genesis.GensisBlock(tx, blockRef.Header.Timestamp)
						this.leagueId = block.Header.LeagueId
						break
					}
				}
			}
		}
		this.growing = true
		fmt.Println("⭕️ league create ok,the id is ", this.leagueId.ToString(), block.Header.Difficulty.Uint64())
		return
	}
}

func (this *Radar) GetBlocks(start, end uint64) ([]*orgtypes.Block, error) {
	blocks := []*orgtypes.Block{}
	for i := start; i <= end; i++ {
		/*blockHash, err := this.adapter.blockStore.GetBlockHash(i)
		if err != nil {
			panic(err)
		}*/
		//log.Println(blockHash.String())
		//block, err := this.adapter.blockStore.GetBlock(blockHash)
		block, err := this.ledger.GetBlockByHeight(i)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

//VerifyMainAndLeague is cmp remote and local data,
// if the two sets of data are the same then return true and
// changes the current validation height, otherwise return false
func (this *Radar) VerifyMainAndLeague(blocks []*orgtypes.Block) bool {
	this.m.Lock()
	defer this.m.Unlock()
	{
		blockhashes := make(common.HashArray, 0, len(blocks))
		for _, v := range blocks {
			blockhashes = append(blockhashes, v.Hash())
		}
		log.Println(blockhashes.GetHashRoot().String())
		log.Println(this.blockRoot.String())
		if blockhashes.GetHashRoot() == this.blockRoot {
			this.beVerified = true
			return true
		}
		return false
	}
}

func containMainType(key maintypes.TransactionType) bool {
	_, ok := txTypeForReference[key]
	return ok
}
