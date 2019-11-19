package poa

import (
	"bytes"
	//"encoding/json"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	//"github.com/ethereum/go-ethereum/ethdb"
	//"github.com/ethereum/go-ethereum/params"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sixexorg/magnetic-ring/config"
	"fmt"
)

type Snapshot struct {
	config   *config.CliqueConfig //
	sigcache *lru.ARCCache        //

	Number  uint64                      `json:"number"`
	Hash    common.Hash                 `json:"hash"`
	Signers map[common.Address]struct{} `json:"signers"`
	Recents map[uint64]common.Address   `json:"recents"`
}

func newSnapshot(config *config.CliqueConfig, sigcache *lru.ARCCache, number uint64, hash common.Hash, signers []common.Address) *Snapshot {
	snap := &Snapshot{
		config:   config,
		sigcache: sigcache,
		Number:   number,
		Hash:     hash,
		Signers:  make(map[common.Address]struct{}),
		Recents:  make(map[uint64]common.Address),
	}
	for _, signer := range signers {
		snap.Signers[signer] = struct{}{}
	}
	return snap
}

func (s *Snapshot) copy() *Snapshot {
	cpy := &Snapshot{
		config:   s.config,
		sigcache: s.sigcache,
		Number:   s.Number,
		Hash:     s.Hash,
		Signers:  make(map[common.Address]struct{}),
		Recents:  make(map[uint64]common.Address),
	}

	for signer := range s.Signers {
		cpy.Signers[signer] = struct{}{}
	}

	for block, signer := range s.Recents {
		cpy.Recents[block] = signer
	}

	return cpy
}

func (s *Snapshot) apply(headers []*types.Header) (*Snapshot, error) {
	if len(headers) == 0 {
		return s, nil
	}

	for i := 0; i < len(headers)-1; i++ {
		if headers[i+1].Height != headers[i].Height+1 {
			return nil, errInvalidVotingChain
		}
	}

	if headers[0].Height != s.Number+1 {
		return nil, errInvalidVotingChain
	}

	snap := s.copy()

	for _, header := range headers {
		number := header.Height
		if limit := uint64(len(snap.Signers)/2 + 1); number >= limit {
			delete(snap.Recents, number-limit)
		}

		//signer, err := ecrecover(header, s.sigcache)
		signer := header.Coinbase
		//if err != nil {
		//	return nil, err
		//}

		if _, ok := snap.Signers[signer]; !ok {
			for a, _ := range  snap.Signers{
				fmt.Println("-------------!!!!!!!!!!!!!!!!!!!!!!!!!", a.ToString())
			}
			fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!", header.Height, signer.ToString())
			return nil, errUnauthorized
		}

		for _, recent := range snap.Recents {
			if recent == signer {
				return nil, errUnauthorized
			}
		}

		snap.Recents[number] = signer
	}
	snap.Number += uint64(len(headers))
	snap.Hash = headers[len(headers)-1].Hash()

	return snap, nil
}

// signers retrieves the list of authorized signers in ascending order.
func (s *Snapshot) signers() []common.Address {
	signers := make([]common.Address, 0, len(s.Signers))
	for signer := range s.Signers {
		signers = append(signers, signer)
	}
	for i := 0; i < len(signers); i++ {
		for j := i + 1; j < len(signers); j++ {
			if bytes.Compare(signers[i][:], signers[j][:]) > 0 {
				signers[i], signers[j] = signers[j], signers[i]
			}
		}
	}
	return signers
}

// inturn returns if a signer at a given block height is in-turn or not.
func (s *Snapshot) inturn(number uint64, signer common.Address) bool {
	signers, offset := s.signers(), 0
	for offset < len(signers) && signers[offset] != signer {
		offset++
	}
	return (number % uint64(len(signers))) == uint64(offset)
}
