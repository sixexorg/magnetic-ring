package dpoa

import (
	"encoding/json"
	"fmt"
	"time"

	//"github.com/ontio/ontology-crypto/keypair"
	"github.com/sixexorg/magnetic-ring/common"
	//"github.com/sixexorg/magnetic-ring/core/ledger"
	//"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	//"github.com/sixexorg/magnetic-ring/core/signature"
	//"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
)

type ConsensusMsgPayload struct {
	Type    comm.MsgType `json:"type"`
	Len     uint32  `json:"len"`
	Payload []byte  `json:"payload"`
}

func DeserializeVbftMsg(msgPayload []byte) (comm.ConsensusMsg, error) {

	m := &ConsensusMsgPayload{}
	if err := json.Unmarshal(msgPayload, m); err != nil {
		return nil, fmt.Errorf("unmarshal consensus msg payload: %s", err)
	}
	if m.Len < uint32(len(m.Payload)) {
		return nil, fmt.Errorf("invalid payload length: %d", m.Len)
	}

	switch m.Type {
	case comm.ConsensePrepare:
		t := &comm.PrepareMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.ConsensePromise:
		t := &comm.PromiseMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.ConsenseProposer:
		t := &comm.ProposerMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.ConsenseAccept:
		t := &comm.AcceptMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.ConsenseDone:
		t := &comm.ConsedoneMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.ViewTimeout:
		t := &comm.ViewtimeoutMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.EvilEvent:
		t := &comm.EvilMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	//case EarthFetchSigs:
	//	t := &EarthSigsFetchMsg{}
	//	if err := json.Unmarshal(m.Payload, t); err != nil {
	//		return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
	//	}
	//	return t, nil
	case comm.EarthFetchRspSigs:
		t := &comm.EarthSigsFetchRspMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.EventFetchMessage:
		t := &comm.EventFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.EventFetchRspMessage:
		t := &comm.EventFetchRspMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.EvilFetchMessage:
		t := &comm.EvilFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.EvilFetchRspMessage:
		t := &comm.EvilFetchRspMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.TimeoutFetchMessage:
		t := &comm.TimeoutFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.TimeoutFetchRspMessage:
		t := &comm.TimeoutFetchRspMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.PeerHandshakeMessage:
		t := &comm.PeerHandshakeMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.PeerHeartbeatMessage:
		t := &comm.PeerHeartbeatMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.BlockInfoFetchMessage:
		t := &comm.BlockInfoFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.BlockInfoFetchRespMessage:
		t := &comm.BlockInfoFetchRespMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.BlockFetchMessage:
		t := &comm.BlockFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case comm.BlockFetchRespMessage:
		t := &comm.BlockFetchRespMsg{}
		if err := t.Deserialize(m.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	}

	return nil, fmt.Errorf("unknown msg type: %d", m.Type)
}

func SerializeVbftMsg(msg comm.ConsensusMsg) ([]byte, error) {

	payload, err := msg.Serialize()
	if err != nil {
		return nil, err
	}

	return json.Marshal(&ConsensusMsgPayload{
		Type:    msg.Type(),
		Len:     uint32(len(payload)),
		Payload: payload,
	})
}

func constructHandshakeMsg(self *Server) (*comm.PeerHandshakeMsg, error) {
	blkNum := self.store.db.GetCurrentBlockHeight()
	block, _ := self.store.getSealedBlock(blkNum)
	if block == nil {
		return nil, fmt.Errorf("failed to get sealed block, current block: %d", self.GetCurrentBlockNo())
	}
	msg := &comm.PeerHandshakeMsg{
		CommittedBlockNumber: blkNum,
		//CommittedBlockHash:   blockhash,
		//CommittedBlockLeader: block.getProposer(),
		//ChainConfig:          self.config,
	}

	return msg, nil
}

func constructHeartbeatMsg(store *BlockStore)  (*comm.PeerHeartbeatMsg, error) {

	blkNum := store.db.GetCurrentBlockHeight()
	block, _ := store.getSealedBlock(blkNum)
	if block == nil {
		return nil, fmt.Errorf("failed to get sealed block, current block: %d", store.db.GetCurrentBlockHeight())
	}
	//fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@#########################", store.db.GetCurrentBlockHeight(), time.Now().String())
	str := time.Now().String()
	msg := &comm.PeerHeartbeatMsg{
		CommittedBlockNumber: store.db.GetCurrentBlockHeight(),
		TimeStamp:            []byte(str),
	}

	return msg, nil
}

func constructBlockFetchMsg(blkNum uint64) (*comm.BlockFetchMsg, error) {
	msg := &comm.BlockFetchMsg{
		BlockNum: blkNum,
	}
	return msg, nil
}

func constructBlockInfoFetchMsg(startBlkNum uint64) (*comm.BlockInfoFetchMsg, error) {

	msg := &comm.BlockInfoFetchMsg{
		StartBlockNum: startBlkNum,
	}
	return msg, nil
}

func constructEventFetchRspMessage(blkNum uint64, evtimeout, evevil []interface{}) *comm.EventFetchRspMsg {
	msg := &comm.EventFetchRspMsg{
		BlkNum    :blkNum,
		EvilEv    :evevil,
		TimeoutEv :evtimeout,
	}

	return msg
}

func constructEvilFetchMsg(blkNum uint64) *comm.EvilFetchMsg {
	msg := &comm.EvilFetchMsg{
		BlkNum    :blkNum,
	}

	return msg
}

func constructEarthSigsFetchRspMsg(blkNum uint64, pubKey string, sig []byte, blkhash common.Hash) *comm.EarthSigsFetchRspMsg {
	return &comm.EarthSigsFetchRspMsg{
		BlkNum : blkNum,
		PubKey : pubKey,
		Sigs   : sig,
		BlkHash: blkhash,
	}
}


func constructBlockFetchRespMsg(blkNum uint64, blk *comm.Block, blkHash common.Hash) (*comm.BlockFetchRespMsg, error) {
	msg := &comm.BlockFetchRespMsg{
		BlockNumber: blkNum,
		BlockHash:   blkHash,
		BlockData:   blk,
	}
	return msg, nil
}


func constructBlockInfoFetchRespMsg(blockInfos []*comm.BlockInfo_) (*comm.BlockInfoFetchRespMsg, error) {
	msg := &comm.BlockInfoFetchRespMsg{
		Blocks: blockInfos,
	}
	return msg, nil
}