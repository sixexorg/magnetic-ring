package sync

import (
	"errors"
	"fmt"

	"github.com/sixexorg/magnetic-ring/p2pserver/temp"

	// "net"
	// "strconv"
	// "strings"
	// "time"

	lru "github.com/hashicorp/golang-lru"
	evtActor "github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/log"

	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	msgCommon "github.com/sixexorg/magnetic-ring/p2pserver/common"
	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
	"github.com/sixexorg/magnetic-ring/p2pserver/net/protocol"
)

//respCache cache for some response data
var respOrgCache *lru.ARCCache

// HeaderReqHandle handles the header sync req from peer
func OrgHeadersReqHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive headers request message", "Addr", data.Addr, "Id", data.Id)

	headersReq := data.Payload.(*msgCommon.HeadersReq)

	startHash := headersReq.HashStart
	stopHash := headersReq.HashEnd

	// headers, err := OrgGetHeadersFromHash(startHash, stopHash,headersReq.OrgID,headersReq.SyncType,headersReq.OrgID)

	// for test start
	headers, err := OrgGetHeadersFromHeight(headersReq.Height, headersReq.OrgID)
	// OrgGetHeadersFromHeight(srcHeight uint32) ([]*types.Header, error) {
	// for test end

	if err != nil {
		log.Warn("get headers in HeadersReqHandle ", "error", err.Error(), "startHash", startHash.String(), "stopHash", stopHash.String())
		return
	}
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in HeadersReqHandle, peer", "id", data.Id)
		return
	}
	msg := msgpack.NewHeaders(nil, headers, headersReq.OrgID, headersReq.SyncType)
	err = p2p.Send(remotePeer, msg, false)
	if err != nil {
		log.Warn("err", "err", err)
		return
	}
}


func ANodeSendToStellarPeingData(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive data req message", "Addr", data.Addr, "Id", data.Id)
	var dataReq = data.Payload.(*msgCommon.Block)
	notifyp2pactor, err := bactor.GetActorPid(bactor.MAINRADARACTOR)
	if err != nil {
		log.Error("ANodeSendToStellarPeingData GetActorPid err", "org", dataReq.OrgID, "err", err)
		return
	}
	fmt.Println("🌐 ANodeSendToStellarPeingData ", dataReq.BlkOrg)
	notifyp2pactor.Tell(dataReq.BlkOrg)
}


func StellaNodeSendToStellarPeingData(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive data req message", "Addr", data.Addr, "Id", data.Id)
	var dataReq = data.Payload.(*msgCommon.Block)
	notifyp2pactor, err := bactor.GetActorPid(bactor.MAINRADARACTOR)
	if err != nil {
		log.Error("StellaNodeSendToStellarPeingData GetActorPid err", "org", dataReq.OrgID, "err", err)
		return
	}
	notifyp2pactor.Tell(dataReq.BlkOrg)
}

// OrgDataReqHandle handles the data req(block/Transaction) from peer
func OrgDataReqHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive data req message", "Addr", data.Addr, "Id", data.Id)

	var dataReq = data.Payload.(*msgCommon.DataReq)

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in DataReqHandle")
		return
	}
	reqType := common.InventoryType(dataReq.DataType)
	hash := dataReq.Hash
	orgID := dataReq.OrgID
	synctype := dataReq.SyncType
	switch reqType {
	case common.BLOCK:
		reqID := fmt.Sprintf("%x%s", reqType, hash.String())
		data := getRespOrgCacheValue(reqID)
		var block *types.Block
		var err error
		if data != nil {
			switch data.(type) {
			case *types.Block:
				block = data.(*types.Block)
			}
		}
		if block == nil {
			if synctype == msgCommon.SYNC_DATA_ORG {
				block, err = temp.GetLedger(orgID).GetBlockByHash(hash /*,orgID*/)
			}
			if err != nil || block == nil || block.Header == nil {
				log.Debug("[p2p]can't get block send not found message by", "hash", hash)
				msg := msgpack.NewNotFound(hash)
				err := p2p.Send(remotePeer, msg, false)
				if err != nil {
					log.Warn("err", "err", err)
					return
				}
				return
			}
			saveRespOrgCache(reqID, block)
		}
		log.Debug("[p2p]block ", "height", block.Header.Height, "hash", hash)
		msg := msgpack.NewBlock(nil, block, orgID, synctype)
		err = p2p.Send(remotePeer, msg, false)
		if err != nil {
			log.Warn("err", "err", err)
			return
		}
	}
}

func OrgGetHeadersFromHeight(srcHeight uint64, orgID common.Address) ([]*types.Header, error) {
	var count uint64 = 0
	headers := []*types.Header{}
	// var startHeight uint32
	var stopHeight uint64

	curHeight := temp.GetLedger(orgID).GetCurrentHeaderHeight()
	if srcHeight == 0 {
		if curHeight > msgCommon.MAX_BLK_HDR_CNT {
			count = msgCommon.MAX_BLK_HDR_CNT
		} else {
			count = curHeight
		}
	} else {
		// bkStop, err := ledger.DefLedger.GetHeaderByHash(stopHash)
		// if err != nil || bkStop == nil {
		// 	return nil, err
		// }
		stopHeight = srcHeight
		count = curHeight - stopHeight
		if count > msgCommon.MAX_BLK_HDR_CNT {
			count = msgCommon.MAX_BLK_HDR_CNT
		}
	}

	var i uint64
	for i = 1; i <= count; i++ {
		hash, _ := temp.GetLedger(orgID).GetBlockHashByHeight(stopHeight + i)
		hd, err := temp.GetLedger(orgID).GetHeaderByHash(hash)
		if err != nil {
			log.Debug("[p2p]net_server GetBlockWithHeight failed with", "err", err.Error(), "hash", hash, "height", stopHeight+i)
			return nil, err
		}
		headers = append(headers, hd)
	}

	return headers, nil
}

//get blk hdrs from starthash to stophash
func OrgGetHeadersFromHash(startHash common.Hash, stopHash common.Hash,
	orgid common.Address, orgtype string, orgID common.Address) ([]*types.Header, error) {
	var count uint32 = 0
	headers := []*types.Header{}
	var startHeight uint32
	var stopHeight uint32
	curHeight := uint32(0)
	if orgtype == msgCommon.SYNC_DATA_MAIN {
		curHeight = uint32(temp.GetLedger(orgID).GetCurrentHeaderHeight())
	} else if orgtype == msgCommon.SYNC_DATA_ORG {
		curHeight = uint32(temp.GetLedger(orgID).GetCurrentHeaderHeight( /*orgid*/ ))
	}

	empty := common.Hash{}
	if startHash == empty {
		if stopHash == empty {
			if curHeight > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			} else {
				count = curHeight
			}
		} else {
			var bkStop *types.Header
			var err error
			if orgtype == msgCommon.SYNC_DATA_MAIN {
				bkStop, err = temp.GetLedger(orgID).GetHeaderByHash(stopHash)
			} else if orgtype == msgCommon.SYNC_DATA_ORG {
				bkStop, err = temp.GetLedger(orgID).GetHeaderByHash(stopHash /*,orgid*/)
			}

			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = uint32(bkStop.Height)
			count = curHeight - stopHeight
			if count > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			}
		}
	} else {
		var bkStart *types.Header
		var err error
		if orgtype == msgCommon.SYNC_DATA_MAIN {
			bkStart, err = temp.GetLedger(orgID).GetHeaderByHash(startHash)
		} else if orgtype == msgCommon.SYNC_DATA_ORG {
			bkStart, err = temp.GetLedger(orgID).GetHeaderByHash(startHash /*,orgid*/)
		}

		if err != nil || bkStart == nil {
			return nil, err
		}
		startHeight = uint32(bkStart.Height)
		empty := common.Hash{}
		if stopHash != empty {
			var bkStop *types.Header
			var err error
			if orgtype == msgCommon.SYNC_DATA_MAIN {
				bkStop, err = temp.GetLedger(orgID).GetHeaderByHash(stopHash)
			} else if orgtype == msgCommon.SYNC_DATA_ORG {
				bkStop, err = temp.GetLedger(orgID).GetHeaderByHash(stopHash /*,orgid*/)
			}

			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = uint32(bkStop.Height)

			// avoid unsigned integer underflow
			if startHeight < stopHeight {
				return nil, errors.New("[p2p]do not have header to send")
			}
			count = startHeight - stopHeight

			if count >= msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
				stopHeight = startHeight - msgCommon.MAX_BLK_HDR_CNT
			}
		} else {

			if startHeight > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			} else {
				count = startHeight
			}
		}
	}
	var i uint32
	for i = 1; i <= count; i++ {
		var (
			hash common.Hash
			hd   *types.Header
			err  error
		)
		if orgtype == msgCommon.SYNC_DATA_MAIN {
			hash, _ = temp.GetLedger(orgID).GetBlockHashByHeight(uint64(stopHeight + i))
			hd, err = temp.GetLedger(orgID).GetHeaderByHash(hash)
		} else if orgtype == msgCommon.SYNC_DATA_ORG {
			hash, _ = temp.GetLedger(orgID).GetBlockHashByHeight(uint64(stopHeight + i) /*,orgid*/)
			hd, err = temp.GetLedger(orgID).GetHeaderByHash(hash /*,orgid*/)
		}

		if err != nil {
			log.Debug("[p2p]net_server GetBlockWithHeight failed with", "err", err.Error(), "hash", hash, "height", stopHeight+i)
			return nil, err
		}
		headers = append(headers, hd)
	}

	return headers, nil
}

//getRespOrgCacheValue get response data from cache
func getRespOrgCacheValue(key string) interface{} {
	if respOrgCache == nil {
		return nil
	}
	data, ok := respOrgCache.Get(key)
	if ok {
		return data
	}
	return nil
}

//saveRespOrgCache save response msg to cache
func saveRespOrgCache(key string, value interface{}) bool {
	if respOrgCache == nil {
		var err error
		respOrgCache, err = lru.NewARC(msgCommon.MAX_RESP_CACHE_SIZE)
		if err != nil {
			return false
		}
	}
	respOrgCache.Add(key, value)
	return true
}
