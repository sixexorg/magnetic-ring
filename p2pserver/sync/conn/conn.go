package conn

import (
	"errors"
	"fmt"
	"net"
	"time"

	"sync"

	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/p2pserver/common"
)

//Link used to establish
type Link struct {
	sync.Mutex
	id        uint64
	addr      string   // The address of the node
	conn      net.Conn // Connect socket with the peer node
	rlpconn   common.MsgReadWriter
	port      uint16                  // The server port of the node
	time      time.Time               // The latest time the node activity
	recvChan  chan *common.MsgPayload //msgpayload channel
	reqRecord map[string]int64        //Map RequestId to Timestamp, using for rejecting duplicate request in specific time
}

/*
var (
	msgPool = make(map[cm1.Hash]int64)
	t       = cm1.NewTimingWheel(time.Second, 5)
	m       sync.RWMutex
)

func init() {
	go func() {
		for {
			select {
			case <-t.After(time.Second * 1):
				time := time.Now().Unix()
				for k, v := range msgPool {
					if v < time {
						delete(msgPool, k)
					}
				}
			}
		}
	}()
}
*/
func NewLink() *Link {
	link := &Link{
		reqRecord: make(map[string]int64, 0),
	}
	return link
}

//SetID set peer id to link
func (that *Link) SetID(id uint64) {
	that.id = id
}

//GetID return if from peer
func (that *Link) GetID() uint64 {
	return that.id
}

//If there is connection return true
func (that *Link) Valid() bool {
	return that.conn != nil
}

//set message channel for link layer
func (that *Link) SetChan(msgchan chan *common.MsgPayload) {
	that.recvChan = msgchan
}

//get address
func (that *Link) GetAddr() string {
	return that.addr
}

//set address
func (that *Link) SetAddr(addr string) {
	that.addr = addr
}

//set port number
func (that *Link) SetPort(p uint16) {
	that.port = p
}

//get port number
func (that *Link) GetPort() uint16 {
	return that.port
}

//get connection
func (that *Link) GetConn() net.Conn {
	return that.conn
}

//set connection
func (that *Link) SetConn(conn net.Conn) {
	that.conn = conn
}

//get connection
func (that *Link) GetRLPConn() common.MsgReadWriter {
	return that.rlpconn
}

//set connection
func (that *Link) SetRLPConn(rlpconn common.MsgReadWriter) {
	that.rlpconn = rlpconn
}

//record latest message time
func (that *Link) UpdateRXTime(t time.Time) {
	that.time = t
}

//GetRXTime return the latest message time
func (that *Link) GetRXTime() time.Time {
	return that.time
}

func (that *Link) RxHandle() error {
	rlpmsg, err := that.rlpconn.ReadMsg()
	if err != nil {
		fmt.Println("that.rlpconn.ReadMsg occured error-->", err)
		log.Warn("[p2p]error read from", "Addr", that.GetAddr(), "err", err.Error())
		return err
	}
	if rlpmsg.Size > 10*1024*1024 {
		// return errResp(ErrMsgTooLarge, "%v > %v", rlpmsg.Size, ProtocolMaxMsgSize)
		log.Info("[p2p]error ErrMsgTooLarge", "Size", rlpmsg.Size, "SizeCount", 10*1024*1024)
		return errors.New(" (()()()()")
	}
	defer rlpmsg.Discard()

	codestr := common.TransCodeUToStr(rlpmsg.Code)

	msg, err := common.MakeEmptyMessage(codestr)
	if err != nil {
		return err
	}
	if err := rlpmsg.Decode(msg); err != nil { //decode
		return err
	}
	/*	h, err := rlpmsg.DistinctSk()
		if err != nil {
			fmt.Println("distinct msg err ", err)
			return err
		}
		_, ok := msgPool[h]
		if !ok {
			msgPool[h] = time.Now().Add(time.Second * 60).Unix()
		} else {
			fmt.Println("repeat msg")
			return errors.New("repeat msg")
		}
		fmt.Println("conn rxHandle pass")*/
	if codestr == common.HEADERS_TYPE {
		msg, err = common.BlkP2PHeaderToBlkHeader(msg.(*common.BlkP2PHeader))
	} else if codestr == common.BLOCK_TYPE {
		msg, err = common.TrsBlockToBlock(msg.(*common.TrsBlock))
		// }else if codestr == common.INV_TYPE {
	} else if codestr == common.TX_TYPE {
		msg, err = common.P2PTrnToTrn(msg.(*common.P2PTrn))
	} else if codestr == common.CONSENSUS_TYPE {
		msg, err = common.P2PConsPldToConsensus(msg.(*common.P2PConsPld))
	}

	if err != nil {
		return err
	}

	t := time.Now()
	that.UpdateRXTime(t)

	if !that.needSendMsg(msg) {
		log.Debug("skip handle", "msgType", msg.CmdType(), "from", that.id)
		return nil
	}
	that.addReqRecord(msg)
	that.recvChan <- &common.MsgPayload{
		Id:          that.id,
		Addr:        that.addr,
		PayloadSize: rlpmsg.Size,
		Payload:     msg,
	}
	return nil
}

func (that *Link) Rx() {
	for {
		if err := that.RxHandle(); err != nil {
			// p.Log().Debug("Ethereum message handling failed", "err", err)
			log.Warn(" Link Rx", "err", err)
			break
		}
	}
	that.disconnectNotify()
}

//disconnectNotify push disconnect msg to channel // used
func (that *Link) disconnectNotify() {
	log.Debug("[p2p]call disconnectNotify for", "Addr", that.GetAddr())
	that.CloseConn()

	msg, _ := common.MakeEmptyMessage(common.DISCONNECT_TYPE)
	discMsg := &common.MsgPayload{
		Id:      that.id,
		Addr:    that.addr,
		Payload: msg,
	}
	that.recvChan <- discMsg
}

//close connection // used
func (that *Link) CloseConn() {
	if that.conn != nil {
		that.conn.Close()
		that.conn = nil
	}
}

func (that *Link) Tx(msg common.Message) error {
	err := common.Send(that.rlpconn, common.TransCodeStrToU(msg.CmdType()), msg)

	if err != nil {
		log.Warn("[p2p]error sending messge to", "Addr", that.GetAddr(), "msgType", msg.CmdType(), "err", err.Error())
		that.disconnectNotify()
		return err
	}
	return nil
}

//needSendMsg check whether the msg is needed to push to channel
func (that *Link) needSendMsg(msg common.Message) bool {
	that.Lock()
	defer that.Unlock()

	if msg.CmdType() != common.GET_DATA_TYPE {
		return true
	}
	var dataReq = msg.(*common.DataReq)
	reqID := fmt.Sprintf("%x%s", dataReq.DataType, dataReq.Hash.String())
	now := time.Now().Unix()

	if t, ok := that.reqRecord[reqID]; ok {
		if int(now-t) < common.REQ_INTERVAL {
			return false
		}
	}
	return true
}

//addReqRecord add request record by removing outdated request records
func (that *Link) addReqRecord(msg common.Message) {
	that.Lock()
	defer that.Unlock()

	if msg.CmdType() != common.GET_DATA_TYPE {
		return
	}
	now := time.Now().Unix()
	if len(that.reqRecord) >= common.MAX_REQ_RECORD_SIZE-1 {
		for id := range that.reqRecord {
			t := that.reqRecord[id]
			if int(now-t) > common.REQ_INTERVAL {
				delete(that.reqRecord, id)
			}
		}
	}
	var dataReq = msg.(*common.DataReq)
	reqID := fmt.Sprintf("%x%s", dataReq.DataType, dataReq.Hash.String())
	that.reqRecord[reqID] = now
}
