package validation

import (
	"math/big"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

//just sub balance
//1. check vote exist and get vote state
//2. check vote between start and end
//3. check account has voted
//4. get the total ut at the time the vote was initiated
func (s *StateValidate) verifyVote(oplog *OpLog) error {
	if oplog.method > AccountMaxLine && oplog.method < VoteMaxLine {
		aa := oplog.data.(*OPHashAddress)
		err := s.checkAccountVote(aa.Hash, aa.Address)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

type TxUT struct {
	Tx *types.Transaction
	UT *big.Int
}

func (s *StateValidate) checkAccountVote(voteId common.Hash, account common.Address) error {
	fmt.Println("⭕️ checkAccountVote p1")
	err := s.GetVoteState(voteId)
	if err != nil {
		return err
	}
	fmt.Println("⭕️ checkAccountVote p2")
	err = s.GetLeagueUTAndTx(voteId)
	if err != nil {
		return err
	}
	fmt.Println("⭕️ checkAccountVote p3")
	err = s.AccountAlreadyVoted(voteId, account)
	if err != nil {
		return err
	}
	fmt.Println("⭕️ checkAccountVote p4")
	return nil
}

func (s *StateValidate) GetVoteState(voteId common.Hash) error {
	if s.MemoVoteState[voteId] != nil {
		return nil
	}
	var (
		vs  *storelaw.VoteState
		err error
	)
	vs, err = s.ledgeror.GetVoteState(voteId, s.TargetHeight)
	if err != nil {
		fmt.Println("⭕️ GetVoteState err ", err)
		return err
	}
	fmt.Println("⭕️ GetVoteState ", vs.VoteId.String())
	s.MemoVoteState[voteId] = vs
	return nil
}

func (s *StateValidate) GetLeagueUTAndTx(hash common.Hash) error {
	if s.voteUT[hash] == nil {
		tx, height, err := s.ledgeror.GetTransaction(hash, s.leagueId)
		if err != nil {
			return err
		}
		//   -- X --  start  -- O -- currentHeight -- O -- end -- X --
		if tx.TxData.Start > s.TargetHeight || tx.TxData.End < s.TargetHeight {
			return errors.ERR_VOTE_OUT_OF_RANGE
		}
		ut := s.ledgeror.GetUTByHeight(height, s.leagueId)
		s.voteUT[hash] = &TxUT{
			Tx: tx,
			UT: ut,
		}
	}
	return nil
}
func (s *StateValidate) AccountAlreadyVoted(voteId common.Hash, account common.Address) error {
	if s.DirtyVote[voteId] != nil {
		_, ok := s.DirtyVote[voteId][account]
		if ok {
			return errors.ERR_VOTE_ALREADY_VOTED
		}
	} else {
		s.DirtyVote[voteId] = make(map[common.Address]struct{})
	}
	did := s.ledgeror.AlreadyVoted(voteId, account)
	if did {
		return errors.ERR_VOTE_ALREADY_VOTED
	}
	s.DirtyVote[voteId][account] = struct{}{}
	return nil
}
