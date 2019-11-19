package poa

import (
	//"bytes"
	"crypto"
	"errors"
	"fmt"
	"math/big"
	//"math/rand"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	mycrypte "github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/orgcontainer"
	"github.com/sixexorg/magnetic-ring/rlp"
	"github.com/sixexorg/magnetic-ring/store/orgchain/storages"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
	_ "golang.org/x/crypto/sha3"
)

const (
	inmemorySnapshots  = 128  // Number of recent vote snapshots to keep in memory
	inmemorySignatures = 4096 // Number of recent block signatures to keep in memory

	wiggleTime = 500 * time.Millisecond // Random delay (per signer) to allow concurrent signers
)

var (
	epochLength = uint64(30000) // Default number of blocks after which to checkpoint and reset the pending votes
	diffInTurn  = big.NewInt(2) // Block difficulty for in-turn signatures
	diffNoTurn  = big.NewInt(1) // Block difficulty for out-of-turn signatures
)

var (
	errUnknownBlock                 = errors.New("unknown block")
	errInvalidCheckpointBeneficiary = errors.New("beneficiary in checkpoint block non-zero")
	errInvalidVote                  = errors.New("vote nonce not 0x00..0 or 0xff..f")
	errInvalidCheckpointVote        = errors.New("vote nonce in checkpoint block non-zero")
	errMissingVanity                = errors.New("extra-data 32 byte vanity prefix missing")
	errMissingSignature             = errors.New("extra-data 65 byte suffix signature missing")
	errExtraSigners                 = errors.New("non-checkpoint block contains extra signer list")
	errInvalidCheckpointSigners     = errors.New("invalid signer list on checkpoint block")
	errInvalidMixDigest             = errors.New("non-zero mix digest")
	errInvalidUncleHash             = errors.New("non empty uncle hash")
	errInvalidDifficulty            = errors.New("invalid difficulty")
	ErrInvalidTimestamp             = errors.New("invalid timestamp")
	errInvalidVotingChain           = errors.New("invalid voting chain")
	errUnauthorized                 = errors.New("unauthorized")
	errWaitTransactions             = errors.New("waiting for transactions")
)

func sigHash(header *types.Header) (hash common.Hash) {
	sha3256 := crypto.SHA3_256.New()
	rlp.Encode(sha3256, []interface{}{
		header.PrevBlockHash,
		header.Coinbase,
		header.BlockRoot,
		header.TxRoot,
		header.ReceiptsRoot,
		header.Difficulty,
		header.Height,
		uint64(header.Timestamp),
		header.Extra, // Yes, this will panic if extra is too short
	})
	sha3256.Sum(hash[:0])
	return hash
}

func ecrecover(header *types.Header, sigcache *lru.ARCCache) (common.Address, error) {
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address.(common.Address), nil
	}
	signature := header.Extra

	pubkey, _ := mycrypte.UnmarshalPubkey([]byte(header.Coinbase[:]))
	if fg, err := pubkey.Verify(hash[:], signature); !fg && err != nil {
		return common.Address{}, err
	}

	sigcache.Add(hash, header.Coinbase)
	return header.Coinbase, nil
}

type Clique struct {
	config *config.CliqueConfig
	ledger *storages.LedgerStoreImp

	recents    *lru.ARCCache
	signatures *lru.ARCCache

	account account.Account
	signer  common.Address
	//txpool      *common2.SubTxPool
	orgContent *orgcontainer.Container
	lock       sync.RWMutex
}

func New(config *config.CliqueConfig, account account.Account, orgCtx *orgcontainer.Container, ledger *storages.LedgerStoreImp) *Clique {
	conf := *config
	if conf.Epoch == 0 {
		conf.Epoch = epochLength
	}
	recents, _ := lru.NewARC(inmemorySnapshots)
	signatures, _ := lru.NewARC(inmemorySignatures)

	return &Clique{
		config:     &conf,
		account:    account,
		recents:    recents,
		signatures: signatures,
		//txpool:     txpool,
		orgContent: orgCtx,
		ledger:     ledger,
	}
}

func (c *Clique) Author(header *types.Header) (common.Address, error) {
	return ecrecover(header, c.signatures)
}

func (c *Clique) VerifyHeader(header *types.Header, seal bool) error {
	return c.verifyHeader(header)
}

func (c *Clique) verifyHeader(header *types.Header) error {
	number := header.Height

	if header.Timestamp > uint64(time.Now().Unix()) {
		return ErrFutureBlock
	}

	checkpoint := (number % c.config.Epoch) == 0
	if checkpoint && header.Coinbase != (common.Address{}) {
		return errInvalidCheckpointBeneficiary
	}

	if number > 0 {
		if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
			return errInvalidDifficulty
		}
	}

	return c.verifySeal(header)
}

func (c *Clique) snapshot(number uint64, hash common.Hash) (*Snapshot, error) {
	var (
		headers []*types.Header
		snap    *Snapshot
	)
	for snap == nil {
		// If an in-memory snapshot was found, use that
		if s, ok := c.recents.Get(hash); ok {
			snap = s.(*Snapshot)
			break
		}

		if number == 1 {
			genesis, _ := c.ledger.GetBlockByHeight(1)
			//if err := c.VerifyHeader(genesis.Header, false); err != nil {
			//	return nil, err
			//}
			var signers []common.Address
			signers = append(signers, c.account.Address())
			snap = newSnapshot(c.config, c.signatures, 1, genesis.Hash(), signers)
			log.Trace("Stored genesis voting snapshot to disk")
			break
		}

		var header *types.Header
		blk, _ := c.ledger.GetBlockByHeight(number)
		//fmt.Println("øøøøøøøøøøøøøø", blk.Header.Coinbase.ToString(), blk.Header.Height)
		header = blk.Header
		if header == nil {
			return nil, ErrUnknownAncestor
		}
		headers = append(headers, header)
		number, hash = number-1, header.PrevBlockHash
	}

	snap, err := snap.apply(headers)
	if err != nil {
		return nil, err
	}
	c.recents.Add(snap.Hash, snap)

	return snap, err
}

func (c *Clique) VerifySeal(header *types.Header) error {
	return c.verifySeal(header)
}

func (c *Clique) verifySeal(header *types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Height
	if number == 0 {
		return errUnknownBlock
	}
	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := c.snapshot(number-1, header.PrevBlockHash)
	if err != nil {
		return err
	}

	// Resolve the authorization key and check against signers
	//signer, err := ecrecover(header, c.signatures)
	//if err != nil {
	//	return err
	//}
	signer := header.Coinbase
	if _, ok := snap.Signers[signer]; !ok {
		return errUnauthorized
	}
	for seen, recent := range snap.Recents {
		if recent == signer {
			// Signer is among recents, only fail if the current block doesn't shift it out
			if limit := uint64(len(snap.Signers)/2 + 1); seen > number-limit {
				return errUnauthorized
			}
		}
	}
	// Ensure that the difficulty corresponds to the turn-ness of the signer
	inturn := snap.inturn(header.Height, signer)
	if inturn && header.Difficulty.Cmp(diffInTurn) != 0 {
		return errInvalidDifficulty
	}
	if !inturn && header.Difficulty.Cmp(diffNoTurn) != 0 {
		return errInvalidDifficulty
	}
	return nil
}

func (c *Clique) Prepare(header *types.Header) error {
	//header.Coinbase = common.Address{}
	number := header.Height

	// Assemble the voting snapshot to check which votes make sense
	snap, err := c.snapshot(number-1, header.PrevBlockHash)
	if err != nil {
		return err
	}
	// Set the correct difficulty
	header.Difficulty = diffNoTurn
	if snap.inturn(header.Height, c.signer) {
		header.Difficulty = diffInTurn
	}
	//parent, _ := c.ledger.GetBlockByHeight(number - 1)
	//if parent == nil {
	//	return ErrUnknownAncestor
	//}
	//header.Timestamp = uint64(time.Unix(int64(header.Timestamp), 0).Add(time.Millisecond *500).Unix())
	//timeNow := uint64(time.Now().Unix())
	//if header.Timestamp < timeNow {
	//	header.Timestamp = timeNow
	//}
	//if time.Now().After(t1) {
	//	header.Timestamp = uint64(time.Now().Unix())
	//}
	//fmt.Println("++++++++++++++++++++++++++++++++++++",// t.Format("15:04:05.000"), "t1:", t1.Format("15:04:05.000"),
	//	"height:", header.Height, "timestamp:", time.Unix(0, int64(header.Timestamp)).Format("15:04:05.000"))
	return nil
}

func (c *Clique) Finalize(blkInfo *storelaw.OrgBlockInfo) (*types.Block, error) {
	// save all

	//txpool := common2.SubTxPool{}
	//
	////block := txpool.GenerateBlock(common.Address{},2,true)
	//blk,_,acctstate,_ := txpool.Execute(common.Address{})

	//if err := c.ledger.SaveAll(blkInfo); err != nil{
	//
	//}

	return blkInfo.Block, nil
}

func (c *Clique) Seal(block *types.Block, stop <-chan struct{}) (*types.Block, error) {
	header := block.Header
	number := header.Height
	if number == 0 {
		return nil, errUnknownBlock
	}

	if c.config.Period == 0 || len(block.Transactions) == 0 { //if c.config.Period == 0 && len(block.Transactions) == 0 {
		return nil, errWaitTransactions
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	snap, err := c.snapshot(number-1, header.PrevBlockHash)
	if err != nil {
		return nil, err
	}

	if _, authorized := snap.Signers[c.account.Address()]; !authorized {
		return nil, errUnauthorized
	}
	for seen, recent := range snap.Recents {
		if recent == c.account.Address() {
			if limit := uint64(len(snap.Signers)/2 + 1); number < limit || seen > number-limit {
				log.Info("Signed recently, must wait for others")
				<-stop
				return nil, nil
			}
		}
	}
	delay := time.Unix(0, int64(header.Timestamp)).Sub(time.Now())
	//if header.Difficulty.Cmp(diffNoTurn) == 0 {
	//	wiggle := time.Duration(len(snap.Signers)/2+1) * wiggleTime
	//	delay += time.Duration(rand.Int63n(int64(wiggle)))
	//
	//	log.Trace("Out-of-turn signing requested", "wiggle", wiggle)
	//}
	log.Info("Waiting for slot to sign and propagate", "delay", delay)
	select {
	case <-stop:
		return nil, nil
	case <-time.After(delay):
	}
	//time.Sleep(time.Duration(rand.Int63n(int64(10))))
	aa := sigHash(header)
	sighash, err := c.account.Sign(aa[:])
	if err != nil {
		fmt.Println(">>>sigFn err", err, time.Now().String())
		return nil, err
	}
	copy(header.Extra, sighash)
	return block, nil
}
