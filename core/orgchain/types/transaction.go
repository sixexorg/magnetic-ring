package types

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/log"
	"io"
	"math/big"
	"sort"

	"github.com/sixexorg/magnetic-ring/common"
)

type TransactionType byte
type VoteAttitude uint8
type VoteType byte

type SigMap map[common.Address][]byte

type SigPack struct {
	SigData []*SigMap
}

func (s *SigPack) ToBytes() []byte {
	return nil
}

type Payload interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
}
type Transaction struct {
	Version         byte
	TxType          TransactionType `json:"tx_type"`
	Payload         Payload
	Raw             []byte //data from payload
	TransactionHash common.Hash
	TxData          *TxData `json:"tx_data"`
	Filled          bool

	Templt *common.SigTmplt
	Sigs   *common.SigPack
}

type Transactions []*Transaction

func (s Transactions) Len() int      { return len(s) }
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s Transactions) Less(i, j int) bool {
	return s[i].TransactionHash.String() < s[j].TransactionHash.String()
}

func (s Transactions) GetHashRoot() common.Hash {
	sort.Sort(s)
	hashes := make([]common.Hash, 0, s.Len())
	for _, v := range s {
		hashes = append(hashes, v.Hash())
	}
	return common.ComputeMerkleRoot(hashes)
}

type TxByNonce Transactions

func (s TxByNonce) Len() int      { return len(s) }
func (s TxByNonce) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s TxByNonce) Less(i, j int) bool {
	return s[i].TxData.Nonce < s[j].TxData.Nonce
}
func NewTransactionRaw(typ TransactionType, version byte, txData *TxData) (*Transaction, error) {
	tx := &Transaction{
		Version: version,
		TxType:  typ,
		TxData:  txData,
	}
	return tx, nil
}
func (tx *Transaction) Hash() common.Hash {
	tx.ToRaw()
	tx.TransactionHash, _ = common.ParseHashFromBytes(common.Sha256(tx.Raw, tx.Sigs.ToBytes()))
	return tx.TransactionHash
}
func (tx *Transaction) Type() TransactionType {
	return tx.TxType
}

func (tx *Transaction) Nonce() uint64 {
	return tx.TxData.Nonce
}

func (tx *Transaction) GasLimit() *big.Int {
	return tx.TxData.GasLimit
}

func (tx *Transaction) Fee() *big.Int {
	return tx.TxData.Fee
}

func (tx *Transaction) VerifySignature() (bool, error) {
	return true, nil
	tmplt := tx.Templt
	sigdata := tx.Sigs
	if len(sigdata.SigData) == 0 || len(tmplt.SigData) == 0 {
		return false, nil
	}

	if len(tmplt.SigData) != len(sigdata.SigData) {
		return false, nil
	}

	tx.Sigs = new(common.SigPack)

	err := tx.ToRaw()
	if err != nil {
		return false, err
	}



	hashbuf := common.Sha256(tx.Raw)
	for index, lst := range tmplt.SigData {
		layer := sigdata.SigData[index]
		for _, unit := range *lst {
			m := unit.M
			if len(*layer) < int(m) {
				fmt.Println("sign err 222222")

				return false, nil
			}
			var vm int = 0
			for _, pbuf := range unit.Pubks {
				pubkey, err := crypto.UnmarshalPubkey(pbuf[:])
				if err != nil {
					fmt.Println("sign err 333333")

					return false, err
				}

				for _, sig := range *layer {
					rt, err := pubkey.Verify(hashbuf, sig.Val)
					if err != nil {
						fmt.Println("sign err 444444")

						log.Info("verify tx err", "verifySignature", err)
						continue
					}

					if !rt {
						continue
					}
					vm++
				}
			}

			if vm < int(m) {
				fmt.Println("sign org err 777777")

				return false, nil
			}

		}
	}

	tx.Sigs = sigdata

	return true, nil
}
