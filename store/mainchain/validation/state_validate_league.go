package validation

import (
	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

func (s *StateValidate) VerifyLeague(oplogs []*OpLog) (opIndes []int, err error) {
	opIndes = make([]int, 0, len(oplogs))
	for k, v := range oplogs {
		opTmp := v
		err = s.verifyLeague(opTmp)
		if err == nil {
			//already exec
			opIndes = append(opIndes, k)
		} else {
			break
		}
	}
	return opIndes, err
}

//verifyLeague is check leagueState and memberState
//in league
func (s *StateValidate) verifyLeague(oplog *OpLog) error {
	s.M.Lock()
	defer s.M.Unlock()
	{
		if oplog.method < LeagueMaxLine && oplog.method > LeagueMinLine {
			switch oplog.method {
			case League_create:
				modelCreate := oplog.data.(*OPCreateLeague)
				err := s.ContainsLeague(modelCreate.League, modelCreate.Symbol)
				if err != nil {
					return err
				}
				leagueState := &states.LeagueState{
					Address: modelCreate.League,
					Height:  s.TargetHeight,
					Symbol:  modelCreate.Symbol,
					MinBox:  modelCreate.Minbox,
					Creator: modelCreate.Creator,
					Rate:    modelCreate.Rate,
					Private: modelCreate.Private,
					Data: &states.League{
						Nonce:      1,
						FrozenBox:  modelCreate.FrozenBox,
						MemberRoot: common.Hash{}, //todo merkle root

					},
				}
				s.DirtyLeagueState[leagueState.Address] = leagueState
				s.AppendLeagueForCreate(modelCreate.TxHash, leagueState.Address)
				return nil
			case League_member_apply, League_member_add:
				opaa := oplog.data.(*OPAddressAdress)
				err := s.getLeagueState(opaa.Second)
				fmt.Printf("🔆 verifyLeague p1 leagueId:%s account:%s %v\n", opaa.Second.ToString(), opaa.First.ToString(), s.DirtyLMemberState[opaa.Second][opaa.First])
				if err != nil {
					//fmt.Printf("🔆 verifyLeague err1 %v\n", err)
					return err
				}
				//TODO Verify that the league exists,and coherence of state
				bl, st := s.getMemberCondition(opaa.Second, opaa.First, oplog.method)
				if !bl {
					fmt.Println("🔆 verifyLeague err2")
					return errors.ERR_STATE_MEMBER_ILLEGAL_STATUS
				}
				if s.DirtyLMemberState[opaa.Second] == nil {
					s.DirtyLMemberState[opaa.Second] = make(map[common.Address]*states.LeagueMember)
				}
				if s.DirtyLMemberState[opaa.Second][opaa.First] == nil {
					s.DirtyLMemberState[opaa.Second][opaa.First] = &states.LeagueMember{
						LeagueId: opaa.Second,
						Height:   s.TargetHeight,
						Data: &states.LeagueAccount{
							Account: opaa.First,
							Status:  st,
						},
					}
					fmt.Printf("🔆 verifyLeague p2 %v\n", s.DirtyLMemberState[opaa.Second][opaa.First])
					return nil
				}
				return errors.ERR_STATE_LEAGUE_ONCE
			case League_member_remove:
				opaa := oplog.data.(*OPAddressAdress)
				err := s.getLeagueState(opaa.Second)
				//fmt.Printf("🔆 verifyLeague p1 leagueId:%s account:%s %v\n", opaa.Second.ToString(), opaa.First.ToString(), s.DirtyLMemberState[opaa.Second][opaa.First])
				if err != nil {
					return err
				}
				bl := s.getMemberRecycleMinbox(opaa.Second, opaa.First)
				if !bl {
					//fmt.Println("🔆 verifyLeague err2")
					return errors.ERR_STATE_MEMBER_ILLEGAL_STATUS
				}
				if s.DirtyLMemberState[opaa.Second] == nil {
					s.DirtyLMemberState[opaa.Second] = make(map[common.Address]*states.LeagueMember)
				}
				if s.DirtyLMemberState[opaa.Second][opaa.First] == nil {
					s.DirtyLMemberState[opaa.Second][opaa.First] = &states.LeagueMember{
						LeagueId: opaa.Second,
						Height:   s.TargetHeight,
						Data: &states.LeagueAccount{
							Account: opaa.First,
							Status:  states.LAS_Exit,
						},
					}
					fmt.Printf("🔆 verifyLeague p3 %v\n", s.DirtyLMemberState[opaa.Second][opaa.First])
					return nil
				}
				return errors.ERR_STATE_LEAGUE_ONCE
			case League_minbox:
				opaa := oplog.data.(*OPAddressUint64)
				err := s.getLeagueState(opaa.Address)
				if err != nil {
					return err
				}
				if s.MemoLeagueState[opaa.Address].MinBox != opaa.Num {
					return errors.ERR_LEAGUE_MINBOX_DIFF
				}
				return nil
			/*case League_nonce_add:
			an := oplog.data.(*OPAddressUint64)
			err := s.getLeagueState(an.Address)
			if err != nil {
				return err
			}
			thisNonce := s.DirtyLeagueState[an.Address].Data.Nonce
			if thisNonce+1 == an.Num {
				s.DirtyLeagueState[an.Address].Data.Nonce = an.Num
				return nil
			} else if thisNonce+1 < an.Num {
				return errors.ERR_STATE_LEAGUE_NONCE_BIGGER
			}
			return errors.ERR_STATE_LEAGUE_NONCE_DISCONTINUITY*/
			case League_raise:
				an := oplog.data.(*OPAddressBigInt)
				err := s.getLeagueState(an.Address)
				if err != nil {
					return err
				}
				return nil

			}
		}
		return nil
	}
}
func (s *StateValidate) RollbackLeague(oplogs []*OpLog, opIndes []int) {
	for _, v := range opIndes {
		s.rollbackLeague(oplogs[v])
	}
}
func (s *StateValidate) rollbackLeague(oplog *OpLog) {
	s.M.Lock()
	defer s.M.Unlock()
	{
		if oplog.method == League_create {
			modelCreate := oplog.data.(*OPCreateLeague)
			s.RemoveElementInLeagueForCreate(modelCreate.League)
		} else if oplog.method == League_member_add || oplog.method == League_member_apply {
			aa := oplog.data.(*OPAddressAdress)
			s.DirtyLMemberState[aa.Second][aa.First] = nil
		}
	}
}

func (s *StateValidate) getLeagueState(address common.Address) error {
	if s.DirtyAccountState[address] != nil {
		return nil
	}
	as, err := s.ledgerStore.GetLeagueByHeight(s.TargetHeight-1, address)
	if err != nil {
		return err
	}
	fmt.Println("☀️ getLeagueState ", as.Address.ToString(), as.Rate, as.Data.FrozenBox.Uint64())
	as.Height = s.TargetHeight
	asDirty := &states.LeagueState{}
	common.DeepCopy(&asDirty, as)
	s.MemoLeagueState[address] = as
	s.DirtyLeagueState[address] = asDirty
	return nil
}
func (s *StateValidate) getMemberCondition(leagueId, account common.Address, method opCode) (bool, states.LeagueAccountStatus) {
	if method == League_member_apply {
		//fmt.Printf("🔆 getMemberCondition p1 \n")
		return s.getMemberApply(leagueId, account)
	} else if method == League_member_add {
		//fmt.Printf("🔆 getMemberCondition p2 \n")
		return s.getMemberAdd(leagueId, account)
	}
	return false, states.LAS_Apply
}

func (s *StateValidate) getMemberApply(leagueId, account common.Address) (bool, states.LeagueAccountStatus) {
	memberState, err := s.ledgerStore.GetLeagueMemberByHeight(s.TargetHeight-1, leagueId, account)
	isPrivate := s.MemoLeagueState[leagueId].IsPrivate()
	if err == errors.ERR_DB_NOT_FOUND || memberState.Data.Status == states.LAS_Exit {
		if isPrivate {
			return true, states.LAS_Apply
		} else {
			return true, states.LAS_Normal
		}
	}
	return false, states.LAS_Exit
}
func (s *StateValidate) getMemberAdd(leagueId, account common.Address) (bool, states.LeagueAccountStatus) {
	isPrivate := s.MemoLeagueState[leagueId].IsPrivate()
	//fmt.Printf("🔆 getMemberAdd p1 %v\n", isPrivate)
	if !isPrivate {
		return false, states.LAS_Apply
	}
	memberState, err := s.ledgerStore.GetLeagueMemberByHeight(s.TargetHeight-1, leagueId, account)
	//fmt.Printf("🔆 getMemberAdd  p2 %v %v\n", memberState, err)
	if err != nil {
		return false, states.LAS_Apply
	}
	if memberState.Data.Status == states.LAS_Apply {
		return true, states.LAS_Normal
	}
	return false, states.LAS_Exit
}
func (s *StateValidate) getMemberRecycleMinbox(leagueId, account common.Address) bool {
	isPrivate := s.MemoLeagueState[leagueId].IsPrivate()
	if !isPrivate {
		return false
	}
	memberState, err := s.ledgerStore.GetLeagueMemberByHeight(s.TargetHeight-1, leagueId, account)
	//fmt.Printf("🔆 getMemberAdd  p2 %v %v\n", memberState, err)
	if err != nil {
		return false
	}
	if memberState.Data.Status == states.LAS_Apply {
		return true
	}
	return false
}

func (s *StateValidate) AppendLeagueForCreate(txHash common.Hash, address common.Address) {
	s.LeagueForCreate = append(s.LeagueForCreate, &storages.LeagueKV{address, txHash})
}
func (s *StateValidate) checkSymbol(symbol common.Symbol) error {
	if _, ok := s.SymbolForCreate[symbol]; ok {
		return errors.ERR_STATE_SYMBOL_EXISTS
	}
	err := s.ledgerStore.SymbolExists(symbol)
	if err != nil {
		return err
	}
	s.SymbolForCreate[symbol] = struct{}{}
	return nil
}
func (s *StateValidate) ContainsLeague(address common.Address, symbol common.Symbol) error {
	for _, v := range s.LeagueForCreate {
		if address == v.K {
			return errors.ERR_STATE_LEAGUE_EXISTS
		}
	}
	err := s.checkSymbol(symbol)
	if err != nil {
		return err
	}
	return nil
}
