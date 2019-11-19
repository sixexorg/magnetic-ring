package storelaw

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
)

type Ledger4Validation interface {
	GetCurrentBlockHeight4V(leagueId common.Address) uint64
	GetCurrentHeaderHash4V(leagueId common.Address) common.Hash
	GetBlockHashByHeight4V(leagueId common.Address, height uint64) (common.Hash, error)
	GetPrevAccount4V(height uint64, account, leagueId common.Address) (AccountStater, error)
	ContainTx4V(txhash common.Hash) bool
	GetVoteState(voteId common.Hash, height uint64) (*VoteState, error)
	AlreadyVoted(voteId common.Hash, account common.Address) bool
	GetUTByHeight(height uint64, leagueId common.Address) *big.Int
	GetTransaction(txHash common.Hash, leagueId common.Address) (*orgtypes.Transaction, uint64, error)
	GetHeaderBonus(leagueId common.Address) *big.Int
	GetAccountRange(start, end uint64, account, leagueId common.Address) (AccountStaters, error)
	GetBonus(leagueId common.Address) map[uint64]uint64
}

type AccountStorer interface {
	NewBatch()
	CommitTo() error
	Get(account, leagueId common.Address) AccountStater
	GetPrev(height uint64, account, leagueId common.Address) (AccountStater, error)
	Save(state AccountStater) error
	BatchSave(states AccountStaters) error
}

type OrgBlockInfo struct {
	Block         *orgtypes.Block
	Receipts      orgtypes.Receipts
	AccStates     AccountStaters
	FeeSum        *big.Int
	VoteStates    []*VoteState
	AccountVoteds []*AccountVoted
	UT            *big.Int
	VoteFirstPass orgtypes.Transactions
}

func (this *OrgBlockInfo) Height() uint64 {
	return this.Block.Header.Height
}
