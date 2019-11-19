package account_level

import (
	"math/big"
	"sync"

	"github.com/sixexorg/magnetic-ring/errors"

	"time"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

type LvStatistic struct {
	Count  uint32
	Amount *big.Int
}

func (this *LvStatistic) sub(num *big.Int) {
	this.Amount.Sub(this.Amount, num)
	this.Count--
}
func (this *LvStatistic) add(num *big.Int) {
	this.Amount.Add(this.Amount, num)
	this.Count++
}

var lmInstance *LevelManager

type LevelManager struct {
	m      sync.RWMutex
	cycle  uint64
	height uint64
	//number of account, amount distribution
	distribution map[EasyLevel]*LvStatistic
	nextAccLvl   map[common.Address]*LvlAmount
	math         *lvlMath
	//The levels change and need to be persisted to the database
	lvlChange map[common.Address]EasyLevel

	LevelStore         *LevelStore
	nextHeaderProperty []uint64
}

func NewLevelManager(cycle int, lapWidth uint64, dbDir string) (*LevelManager, error) {
	db, err := NewLevelStore(dbDir)
	if err != nil {
		return nil, err
	}
	lm := &LevelManager{
		distribution: make(map[EasyLevel]*LvStatistic, 9),
		nextAccLvl:   make(map[common.Address]*LvlAmount),
		lvlChange:    make(map[common.Address]EasyLevel),
		math:         newLvlMath(cycle, lapWidth),
		LevelStore:   db,
	}

	for i := lv1; i <= lv9; i++ {
		lm.distribution[i] = &LvStatistic{
			Amount: big.NewInt(0),
			Count:  0,
		}
	}
	return lm, nil
}

func (this *LevelManager) GetNextHeaderProperty(nextHeight uint64) ([]uint64, error) {
	this.m.RLock()
	defer this.m.RUnlock()
	if nextHeight != this.height+1 {
		return nil, errors.ERR_PARAM_NOT_VALID
	}
FLAG:
	if this.nextHeaderProperty == nil {
		time.Sleep(time.Millisecond * 100)
		goto FLAG
	}
	return this.nextHeaderProperty, nil
}

//ReceiveAccountStates is calculated the account level after the block is saved
func (this *LevelManager) ReceiveAccountStates(ass states.AccountStates, energyUesd *big.Int, height uint64) {
	this.m.Lock()
	defer this.m.Unlock()
	this.height = height
	this.execAccountStates(ass, height)
	this.levelUpL()
	this.generateReward(height, energyUesd)
}

var (
	zeroBInt = big.NewInt(0)
)

func (this *LevelManager) execAccountStates(ass states.AccountStates, height uint64) {
	for _, v := range ass {
		fmt.Printf("🤵 ReceiveAccountStates account:%s height:%d balance:%d \n", v.Address.ToString(), v.Height, v.Data.Balance.Uint64())
		/*		if k != 3 {
				continue
			}*/
		if height == 1 {
			if v.Data.Balance.Cmp(lv9_b) != -1 {
				this.genesisLevel9(v.Address, v.Data.Balance)
				continue
			}
		}
		key := v.Address
		amount := v.Data.Balance
		lvline := rankLevel(v.Data.Balance)
		la := this.nextAccLvl[key]

		// not exists
		if la == nil {
			//no change
			if lvline == lv0 {
				continue
			}
			//achieve lv
			this.levelWillBe(key, lvline, amount)
			continue
		}
		//already exists
		im, l, _, cur := la.Lv.Decode()
		//clear level
		if lvline == 0 {
			this.levelClear(key, cur, la.Amount)
			continue
		}

		//new account is bigger
		if lvline > l {
			//bigger than next cur
			if l == cur {
				l++
			}
			this.lvlUpdate(key, im, l, lvline)
			//fmt.Println("等级", l-1, this.distribution[cur])
			this.distribution[cur].amountIncrease(amount, la.Amount)
			this.amountUpdate(key, amount)
			continue
		} else if lvline == l {
			//only amout update
			this.distribution[cur].amountIncrease(amount, la.Amount)
			this.amountUpdate(key, amount)
			continue
		} else if lvline < l {
			//lvl down
			this.levelDown(key, lvline, cur, amount, la.Amount)
			this.amountUpdate(key, amount)
			continue
		}
	}
}

func (this *LevelManager) generateReward(height uint64, energyUsed *big.Int) {
	dis := this.getLvAmountDistribution()
	rewards := this.math.reward(height, energyUsed, dis)
	//fmt.Printf("🈲️ destroy:%d ,reward:%v\n", energyUsed.Uint64(), rewards)
	this.nextHeaderProperty = rewards
}
func (this *LevelManager) getLvAmountDistribution() []*big.Int {
	s := make([]*big.Int, 0, 9)
	for i := lv1; i <= lv9; i++ {
		s = append(s, big.NewInt(0).Set(this.distribution[i].Amount))
	}
	return s
}
func (this *LevelManager) levelUpL() {
	/*	for j := EasyLevel(1); j <= lv9; j++ {
		fmt.Printf("pre lv %d count:%d amount:%d \n", j, this.distribution[j].Count, this.distribution[j].Amount.Uint64())
	}*/
	for k, v := range this.nextAccLvl {
		im, l, r, _ := v.Lv.Decode()
		//nothing to do
		if r == lv0 {
			continue
		}
		if im {
			//next level is the same as the level after next
			this.levelUp(k, l+1, r, v.Amount)
			continue
		} else {
			this.lvlUpdate(k, true, l, r)
			continue
		}
	}
	err := this.LevelStore.SaveLevels(this.height, this.lvlChange)
	if err != nil {
		panic(err)
	}
	this.lvlChange = make(map[common.Address]EasyLevel)
	//fmt.Println("-------------------")
	/*	for j := EasyLevel(1); j <= lv9; j++ {
		fmt.Printf("aft lv %d count:%d amount:%d \n", j, this.distribution[j].Count, this.distribution[j].Amount.Uint64())
	}*/
}

func (this *LevelManager) levelWillBe(account common.Address, r EasyLevel, amount *big.Int) {
	this.nextAccLvl[account] = &LvlAmount{
		Amount: amount,
		Lv:     newAccountLevel(false, lv1, r),
	}
}
func (this *LevelManager) levelUp(account common.Address, l, r EasyLevel, amount *big.Int) {
	if l > r && r != lv0 {
		l = r
		r = lv0
	}
	this.lvlUpdate(account, true, l, r)
	curLv := l
	if r == 0 {
		curLv = l
	} else {
		curLv = l - 1
	}
	prevLv := curLv - 1
	this.lvlChange[account] = curLv
	this.lvlAmoutUpdate(curLv, prevLv, amount, amount)
}

func (this *LevelManager) levelDown(account common.Address, l, prevLv EasyLevel, amount, prevAmount *big.Int) {
	this.lvlUpdate(account, true, l, 0)
	this.lvlChange[account] = l
	this.lvlAmoutUpdate(l, prevLv, amount, prevAmount)
}

func (this *LevelManager) levelClear(account common.Address, prevLv EasyLevel, prevAmount *big.Int) {
	delete(this.nextAccLvl, account)
	this.lvlChange[account] = lv0
	this.lvlAmoutUpdate(lv0, prevLv, nil, prevAmount)
}

func (this *LevelManager) lvlUpdate(account common.Address, im bool, left, right EasyLevel) {
	this.nextAccLvl[account].Lv = newAccountLevel(im, left, right)
}
func (this *LevelManager) amountUpdate(account common.Address, amount *big.Int) {
	this.nextAccLvl[account].Amount.Set(amount)
}
func (this *LvStatistic) amountIncrease(amount, prevAmount *big.Int) {
	if this != nil {
		//fmt.Println(1, this.Amount.Uint64(), amount.Uint64())
		this.Amount.Add(this.Amount, amount)
		//fmt.Println(2, this.Amount.Uint64(), prevAmount.Uint64())
		this.Amount.Sub(this.Amount, prevAmount)
		//fmt.Println(3, this.Amount.Uint64())
	}
}

func (this *LevelManager) lvlAmoutUpdate(curLv, prevLv EasyLevel, amount, prevAmount *big.Int) {
	if curLv > lv0 {
		this.distribution[curLv].add(amount)
	}
	if prevLv > lv0 {
		this.distribution[prevLv].sub(prevAmount)
	}
}

func (this *LevelManager) genesisLevel9(acc common.Address, amount *big.Int) {
	this.distribution[lv9].Count++
	this.distribution[lv9].add(amount)
	this.nextAccLvl[acc] = &LvlAmount{
		Amount: amount,
		Lv:     newAccountLevel(true, lv9, 0),
	}
	this.lvlChange[acc] = lv9
}
