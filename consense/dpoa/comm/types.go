package comm

import (
	"encoding/json"
	//"github.com/ontio/ontology-crypto/keypair"
	//"github.com/sixexorg/magnetic-ring/core/types"
	"github.com/sixexorg/magnetic-ring/common"

	//"github.com/sixexorg/magnetic-ring/consensus/vbft/config"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	//"github.com/sixexorg/magnetic-ring/common/config"
	"bytes"
	//"strings"

	"fmt"
	"math"

	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/rlp"
)

// 用户角色
type roletype int

const (
	Observer roletype = iota
	Processer
	Failor
)

type MsgType uint8

const (
	BlockProposalMessage MsgType = iota
	BlockEndorseMessage
	BlockCommitMessage
	PeerHandshakeMessage
	PeerHeartbeatMessage
	BlockInfoFetchMessage
	BlockInfoFetchRespMessage
	ProposalFetchMessage
	BlockFetchMessage
	BlockFetchRespMessage

	ConsensePrepare
	ConsensePromise
	ConsenseProposer
	ConsenseAccept
	ConsenseDone
	ViewTimeout
	EvilEvent
	EarthFetchRspSigs
	EventFetchMessage
	EventFetchRspMessage
	EvilFetchMessage
	EvilFetchRspMessage
	TimeoutFetchMessage
	TimeoutFetchRspMessage
)

type ParticiConfig struct {
	BlkNum      uint64
	View        uint32
	ProcNodes   []string
	ObserNodes  []string
	FailsNodes  []string
	PartiRaw    []string
	StarsSorted []string
}

type ConsenseNotify struct {
	BlkNum    uint64
	ProcNodes []string
	Istart    bool
}

type ConsensusMsg interface {
	Type() MsgType
	Verify(pub crypto.PublicKey) error
	GetBlockNum() uint64
	Serialize() ([]byte, error)
}

type PeerHeartbeatMsg struct {
	CommittedBlockNumber uint64 `json:"committed_block_number"`
	TimeStamp            []byte `json:"timestamp"`
}

func (msg *PeerHeartbeatMsg) Type() MsgType {
	return PeerHeartbeatMessage
}

func (msg *PeerHeartbeatMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (msg *PeerHeartbeatMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *PeerHeartbeatMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type PeerHandshakeMsg struct {
	CommittedBlockNumber uint64 `json:"committed_block_number"`
	//CommittedBlockHash   common.Hash  `json:"committed_block_hash"`
	//CommittedBlockLeader uint32       `json:"committed_block_leader"`
	//ChainConfig          *ChainConfig `json:"chain_config"`
}

func (msg *PeerHandshakeMsg) Type() MsgType {
	return PeerHandshakeMessage
}

func (msg *PeerHandshakeMsg) Verify(pub crypto.PublicKey) error {

	return nil
}

func (msg *PeerHandshakeMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *PeerHandshakeMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type BlockFetchRespMsg struct {
	BlockNumber uint64      `json:"block_number"`
	BlockHash   common.Hash `json:"block_hash"`
	BlockData   *Block      `json:"block_data"`
}

func (msg *BlockFetchRespMsg) Type() MsgType {
	return BlockFetchRespMessage
}

func (msg *BlockFetchRespMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (msg *BlockFetchRespMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *BlockFetchRespMsg) Serialize() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(buffer, msg.BlockNumber)
	msg.BlockHash.Serialize(buffer)
	blockbuff, err := msg.BlockData.Serialize()
	if err != nil {
		return nil, err
	}
	buffer.Write(blockbuff)
	return buffer.Bytes(), nil
}

func (msg *BlockFetchRespMsg) Deserialize(data []byte) error {
	buffer := bytes.NewBuffer(data)
	blocknum, err := serialization.ReadUint64(buffer)
	if err != nil {
		return err
	}
	msg.BlockNumber = blocknum
	err = msg.BlockHash.Deserialize(buffer)
	if err != nil {
		return err
	}
	blk := &Block{}
	if err := blk.Deserialize(buffer.Bytes()); err != nil {
		return fmt.Errorf("unmarshal block type: %s", err)
	}
	msg.BlockData = blk
	return nil
}

type BlockInfoFetchMsg struct {
	StartBlockNum uint64 `json:"start_block_num"`
}

func (msg *BlockInfoFetchMsg) Type() MsgType {
	return BlockInfoFetchMessage
}

func (msg *BlockInfoFetchMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (msg *BlockInfoFetchMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *BlockInfoFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type BlockInfo_ struct {
	BlockNum   uint64            `json:"block_num"`
	Proposer   uint32            `json:"proposer"`
	Signatures map[uint32][]byte `json:"signatures"`
}

type BlockInfoFetchRespMsg struct {
	Blocks []*BlockInfo_ `json:"blocks"`
}

func (msg *BlockInfoFetchRespMsg) Type() MsgType {
	return BlockInfoFetchRespMessage
}

func (msg *BlockInfoFetchRespMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (msg *BlockInfoFetchRespMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *BlockInfoFetchRespMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type ChainConfig struct {
	Peers []string `json:"peers"`
}

//type VbftBlockInfo struct {
//	Proposer           string       `json:"leader"`
//	VrfValue           []byte       `json:"vrf_value"`
//	VrfProof           []byte       `json:"vrf_proof"`
//	LastConfigBlockNum uint32       `json:"last_config_block_num"`
//	NewChainConfig     *ChainConfig `json:"new_chain_config"`
//}

type Block struct {
	Block *types.Block
	//EmptyBlock *types.Block
	Info *VbftBlockInfo
}

func (blk *Block) GetViews() uint32 {
	return blk.Info.View
}

func (blk *Block) GetProposer() string {
	return blk.Info.Miner
}

func (blk *Block) GetBlockNum() uint64 {
	return blk.Block.Header.Height
}

//func (blk *Block) GetEmptyBlockNum() uint64 {
//	return blk.EmptyBlock.Header.Height
//}

func (blk *Block) GetPrevBlockHash() common.Hash {
	return blk.Block.Header.PrevBlockHash
}

func (blk *Block) GetLastConfigBlockNum() uint64 {
	return blk.Info.LastConfigBlockNum
}

func (blk *Block) GetNewChainConfig() *ChainConfig {
	//return blk.Info.NewChainConfig
	return nil
}

//
// getVrfValue() is a helper function for participant selection.
//
func (blk *Block) GetVrfValue() []byte {
	return blk.Info.VrfValue
}

func (blk *Block) GetVrfProof() []byte {
	return blk.Info.VrfProof
}

func (blk *Block) Serialize() ([]byte, error) {

	buf := bytes.NewBuffer(nil)
	err := blk.Block.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("serialize block: %s", err)
	}
	if blk.Info.NewChainConfig == nil {
		blk.Info.NewChainConfig = &ChainConfig{
			Peers: make([]string, 0),
		}
	}
	if err = rlp.Encode(buf, blk.Info); err != nil {
		return nil, fmt.Errorf("serialize info: %s", err)
	}
	return buf.Bytes(), nil
}

func (blk *Block) Deserialize(data []byte) error {
	buff := bytes.NewBuffer(data)
	blk.Block = &types.Block{}
	err := blk.Block.Deserialize(buff)
	if err != nil {
		return err
	}
	blk.Info = &VbftBlockInfo{}
	if err = rlp.Decode(buff, blk.Info); err != nil {
		return err
	}
	return nil
}

func InitVbftBlock(block *types.Block) (*Block, error) {
	if block == nil {
		return nil, fmt.Errorf("nil block in initVbftBlock")
	}
	//if block.Header.Height == 1{
	//	blkInfo := &VbftBlockInfo{}
	//	blkInfo.VrfValue = []byte("1c9810aa9822e511d5804a9c4db9dd08497c31087b0daafa34d768a3253441fa20515e2f30f81741102af0ca3cefc4818fef16adb825fbaa8cad78647f3afb590e")
	//	blkInfo.VrfProof = []byte("c57741f934042cb8d8b087b44b161db56fc3ffd4ffb675d36cd09f83935be853d8729f3f5298d12d6fd28d45dde515a4b9d7f67682d182ba5118abf451ff1988")
	//	block.Header.ConsensusPayload, _ = json.Marshal(blkInfo)
	//}

	blkInfo := &VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, blkInfo); err != nil {
		return nil, fmt.Errorf("unmarshal blockInfo: %s", err)
	}

	return &Block{
		Block: block,
		Info:  blkInfo,
	}, nil
}

//
//type VRFValue [64]byte
//var NilVRF = VRFValue{}
//
//func (v VRFValue) Bytes() []byte {
//	return v[:]
//}
//
//func (v VRFValue) IsNil() bool {
//	return bytes.Compare(v.Bytes(), NilVRF.Bytes()) == 0
//}

func genConsensusPayload() ([]byte, error) {
	//if cfg.C == 0 {
	//	return nil, fmt.Errorf("C must larger than zero")
	//}
	//if int(cfg.K) > len(cfg.Peers) {
	//	return nil, fmt.Errorf("peer count is less than K")
	//}
	//if cfg.K < 2*cfg.C+1 {
	//	return nil, fmt.Errorf("invalid config, K: %d, C: %d", cfg.K, cfg.C)
	//}
	//if cfg.L%cfg.K != 0 || cfg.L < cfg.K*2 {
	//	return nil, fmt.Errorf("invalid config, K: %d, L: %d", cfg.K, cfg.L)
	//}
	//chainConfig, err := GenesisChainConfig(cfg, cfg.Peers, txhash, height)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// save VRF in genesis config file, to genesis block
	//vrfValue, err := hex.DecodeString(cfg.VrfValue)
	//if err != nil {
	//	return nil, fmt.Errorf("invalid config, vrf_value: %s", err)
	//}
	//
	//vrfProof, err := hex.DecodeString(cfg.VrfProof)
	//if err != nil {
	//	return nil, fmt.Errorf("invalid config, vrf_proof: %s", err)
	//}

	// Notice:
	// take genesis msg as random source,
	// don't need verify (genesisProposer, vrfValue, vrfProof)

	vbftBlockInfo := &VbftBlockInfo{
		//Proposer:           "",
		//VrfValue:           vrfValue,
		//VrfProof:           vrfProof,
		LastConfigBlockNum: math.MaxUint32,
		//NewChainConfig:     chainConfig,
	}
	return json.Marshal(vbftBlockInfo)
}

//
//func GenesisConsensusPayload(txhash common.Uint256, height uint32) ([]byte, error) {
//	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
//
//	switch consensusType {
//	case "vbft":
//		return genConsensusPayload(config.DefConfig.Genesis.VBFT, txhash, height)
//	}
//	return nil, nil
//}

type VbftBlockInfo struct {
	View               uint32       `json:"view"`
	Miner              string       `json:"miner"`
	Proposer           uint32       `json:"leader"`
	VrfValue           []byte       `json:"vrf_value"`
	VrfProof           []byte       `json:"vrf_proof"`
	LastConfigBlockNum uint64       `json:"last_config_block_num"`
	NewChainConfig     *ChainConfig `json:"new_chain_config"`
}

const (
	VRF_SIZE            = 64 // bytes
	MAX_PROPOSER_COUNT  = 32
	MAX_ENDORSER_COUNT  = 240
	MAX_COMMITTER_COUNT = 240
)

type VRFValue [VRF_SIZE]byte

var NilVRF = VRFValue{}

func (v VRFValue) Bytes() []byte {
	return v[:]
}

func (v VRFValue) IsNil() bool {
	return bytes.Compare(v.Bytes(), NilVRF.Bytes()) == 0
}

type Account interface {
	Sign(hash []byte) (sig []byte, err error)
	Verify(hash, sig []byte) (bool, error)
	Address() common.Address
}

type DataHandler interface {
	DelOrder(key string)
}
