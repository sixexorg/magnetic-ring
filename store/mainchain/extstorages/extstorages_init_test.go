package extstorages

import (
	"math/big"
	"os"
	"testing"

	"github.com/sixexorg/magnetic-ring/common"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

var (
	leadger      *LedgerStoreImp
	store        *ExternalLeagueBlock
	accountStore *ExternalLeague
)

func TestMain(m *testing.M) {
	ledgerPath := "test"
	blockPath := "test/testBlock"
	accountPath := "test/testAccount"
	initLedger(ledgerPath)
	initLeagueBlock(blockPath)
	initAccountState(accountPath)
	os.RemoveAll("./test")
	m.Run()
}

func initLedger(dbdir string) {
	var err error
	leadger, err = NewLedgerStore(dbdir)
	if err != nil {
		panic(err)
	}
	exts := storelaw.AccountStaters{
		&extstates.LeagueAccountState{
			Address:  mock.Address_1,
			LeagueId: mock.LeagueAddress1,
			Height:   1,
			Data: &extstates.Account{
				Nonce:       1,
				Balance:     big.NewInt(100),
				EnergyBalance: big.NewInt(200),
			},
		},
		&extstates.LeagueAccountState{
			Address:  mock.Address_2,
			LeagueId: mock.LeagueAddress1,
			Height:   1,
			Data: &extstates.Account{
				Nonce:       1,
				Balance:     big.NewInt(50000),
				EnergyBalance: big.NewInt(100000),
			},
		},
		&extstates.LeagueAccountState{
			Address:  mock.Address_3,
			LeagueId: mock.LeagueAddress1,
			Height:   1,
			Data: &extstates.Account{
				Nonce:       1,
				Balance:     big.NewInt(232031),
				EnergyBalance: big.NewInt(912312),
			},
		}}
	obi := &storelaw.OrgBlockInfo{
		Block: &orgtypes.Block{
			Header: &orgtypes.Header{
				Version:       0x01,
				PrevBlockHash: common.Hash{},
				Height:        1,
				LeagueId:      mock.LeagueAddress1,
				Difficulty:    big.NewInt(10),
			},
			Transactions: mock.Block2.Transactions,
		},
		AccStates: exts,
		FeeSum:    big.NewInt(20000),
	}
	mainTxUsed := common.HashArray{common.Hash{1, 2, 3}, common.Hash{4, 5, 6}, common.Hash{7, 8, 9}}
	err = ledgerStore.SaveAll(obi, mainTxUsed)
	if err != nil {
		panic(err)
	}
}
func initAccountState(dbdir string) {
	var err error
	accountStore, err = NewExternalLeague(dbdir, false)
	if err != nil {
		panic(err)
	}
	exts := storelaw.AccountStaters{
		&extstates.LeagueAccountState{
			Address:  mock.Address_1,
			LeagueId: mock.LeagueAddress1,
			Height:   1,
			Data: &extstates.Account{
				Nonce:       1,
				Balance:     big.NewInt(100),
				EnergyBalance: big.NewInt(200),
			},
		},
		&extstates.LeagueAccountState{
			Address:  mock.Address_2,
			LeagueId: mock.LeagueAddress1,
			Height:   1,
			Data: &extstates.Account{
				Nonce:       1,
				Balance:     big.NewInt(50000),
				EnergyBalance: big.NewInt(100000),
			},
		},
		&extstates.LeagueAccountState{
			Address:  mock.Address_3,
			LeagueId: mock.LeagueAddress1,
			Height:   1,
			Data: &extstates.Account{
				Nonce:       1,
				Balance:     big.NewInt(232031),
				EnergyBalance: big.NewInt(912312),
			},
		}}
	accountStore.NewBatch()
	err = accountStore.BatchSave(exts)
	if err != nil {
		panic(err)
	}
	err = accountStore.CommitTo()
	if err != nil {
		panic(err)
	}
}
func initLeagueBlock(dbdir string) {
	var err error
	store, err = NewExternalLeagueBlock(dbdir, false)
	if err != nil {
		panic(err)
	}
	block1 := &extstates.LeagueBlockSimple{
		Header: &orgtypes.Header{
			Version:       0x01,
			PrevBlockHash: common.Hash{},
			Height:        1,
			LeagueId:      mock.LeagueAddress1,
			Difficulty:    big.NewInt(10),
		},
		EnergyUsed: big.NewInt(10),
	}

	block2 := &extstates.LeagueBlockSimple{
		Header: &orgtypes.Header{
			Version:       0x01,
			Height:        2,
			PrevBlockHash: block1.Header.Hash(),
			LeagueId:      mock.LeagueAddress1,
			Difficulty:    big.NewInt(1),
		},
		EnergyUsed: big.NewInt(20),
	}
	block3 := &extstates.LeagueBlockSimple{
		Header: &orgtypes.Header{
			Version:       0x01,
			Height:        3,
			PrevBlockHash: block2.Header.Hash(),
			LeagueId:      mock.LeagueAddress1,
			Difficulty:    big.NewInt(1),
		},
		EnergyUsed: big.NewInt(30),
	}
	block4 := &extstates.LeagueBlockSimple{
		Header: &orgtypes.Header{
			Version:       0x01,
			Height:        4,
			PrevBlockHash: block3.Header.Hash(),
			LeagueId:      mock.LeagueAddress1,
			Difficulty:    big.NewInt(1),
		},
		EnergyUsed: big.NewInt(30),
	}
	/*fmt.Println("block1 hash:", block1.Header.Hash().String())
	fmt.Println("block2 hash:", block2.Header.Hash().String())
	fmt.Println("block3 hash:", block3.Header.Hash().String())
	fmt.Println("block4 hash:", block4.Header.Hash().String())*/
	store.NewBatch()
	err = store.BatchSave([]*extstates.LeagueBlockSimple{block1, block2, block3, block4})
	if err != nil {
		panic(err)
	}
	err = store.CommitTo()
	if err != nil {
		panic(err)
	}
}
