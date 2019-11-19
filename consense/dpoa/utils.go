package dpoa

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/sixexorg/magnetic-ring/common"

	//"github.com/sixexorg/magnetic-ring/core/signature"
	//"github.com/sixexorg/magnetic-ring/consensus/vbft/config"
	//"github.com/ontio/ontology-crypto/vrf"
	//"github.com/ontio/ontology-crypto/keypair"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	"github.com/sixexorg/magnetic-ring/crypto"
)

func hashData(data []byte) common.Hash {
	t := sha256.Sum256(data)
	f := sha256.Sum256(t[:])
	return common.Hash(f)
}

type Partions struct {
	NoPartins   []string
	HasPations  []string
	Fails       []string
	AllPartions []string
}

type seedData struct {
	BlockNum        uint64      `json:"block_num"`
	PrevBlockSigner string      `json:"prev_block_proposer"` // TODO: change to NodeID
	BlockRoot       common.Hash `json:"block_root"`
	VrfValue        []byte      `json:"vrf_value"`
}

func getParticipantSelectionSeed(block *comm.Block) comm.VRFValue {
	data, err := json.Marshal(&seedData{
		BlockNum:        block.GetBlockNum() + 1,
		PrevBlockSigner: block.GetProposer(),
		BlockRoot:       block.Block.Header.BlockRoot,
		VrfValue:        block.GetVrfValue(),
	})
	if err != nil {
		return comm.VRFValue{}
	}

	t := sha512.Sum512(data)
	f := sha512.Sum512(t[:])
	return comm.VRFValue(f)
}

func CheckSigners(pmsg *comm.ViewtimeoutMsg, fromid string) error {
	pb, _ := json.Marshal(pmsg.RawData)

	buf, err := common.Hex2Bytes(fromid)
	if err != nil {
		return err
	}

	pub, _ := crypto.UnmarshalPubkey(buf)
	if _, err := pub.Verify(pb, pmsg.Signature); err != nil {
		return err
	}
	//pub, _ := vconfig.Pubkey(fromid)
	//if err := signature.Verify(pub, pb, pmsg.Signature); err != nil{
	//	return  err
	//}

	return nil
}

func computeVrf(sk crypto.PrivateKey, blkNum uint64, prevVrf []byte) ([]byte, []byte, error) {
	data, err := json.Marshal(&comm.VrfData{
		BlockNum: blkNum,
		PrevVrf:  prevVrf,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("computeVrf failed to marshal vrfData: %s", err)
	}
	//fmt.Println("computeVrfcomputeVrfcomputeVrfcomputeVrfcomputeVrfcomputeVrfcomputeVrfcomputeVrf",data)
	hash, proof := sk.Evaluate(data)

	return hash[:], proof, nil
}

func GetIndex(arr []string, p string) (int, error) {
	for k, v := range arr {
		if p == v {
			return k, nil
		}
	}
	return -1, errors.New(fmt.Sprintf("GetIndex arr:%v p:%v", arr, p))
}

func GetNode(arr []string, index int) (string, error) {
	for k, v := range arr {
		if k == index {
			return v, nil
		}
	}
	return "", errors.New(fmt.Sprintf("GetIndex arr:%v p:%v", arr, index))
}

func calcParticipant(vrf comm.VRFValue, dposTable []uint32, k uint32) uint32 {
	var v1, v2 uint32
	bIdx := k / 8
	bits1 := k % 8
	bits2 := 8 + bits1 // L - 8 + bits1
	if k >= 512 {
		return math.MaxUint32
	}
	// FIXME:
	// take 16bits random variable from vrf, if len(dposTable) is not power of 2,
	// this algorithm will break the fairness of vrf. to be fixed
	v1 = uint32(vrf[bIdx]) >> bits1
	if bIdx+1 < uint32(len(vrf)) {
		v2 = uint32(vrf[bIdx+1])
	} else {
		v2 = uint32(vrf[0])
	}

	v2 = v2 & ((1 << bits2) - 1)
	v := (v2 << (8 - bits1)) + v1
	v = v % uint32(len(dposTable))
	return dposTable[v]
}

func calcParticipantPeers(vrf comm.VRFValue, dposTable []string) []string {
	return dposTable
}

func int2Byte(num interface{}) []byte {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, num); err != nil {
		return nil
	}

	return buffer.Bytes()
}

func byte2Int(data []byte) uint32 {
	var buffer bytes.Buffer
	_, err := buffer.Write(data)
	if err != nil {
		return 0
	}
	var i uint16
	err = binary.Read(&buffer, binary.BigEndian, &i)
	if err != nil {
		return 0
	}
	return uint32(i)
}
