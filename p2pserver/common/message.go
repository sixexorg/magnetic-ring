package common

import (
	"errors"
	"fmt"

	comm "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	ct "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/p2pserver/discover"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
)

// "version" 握手用
type VersionPayload struct {
	Version      uint32
	Services     uint64
	TimeStamp    uint64
	SyncPort     uint16
	HttpInfoPort uint16
	ConsPort     uint16
	Cap          [32]byte
	Nonce        uint64
	StartHeight  uint64
	Relay        uint8
	IsConsensus  bool
}
type Version struct {
	P VersionPayload
	N discover.Node
}

//Serialize message payload
func (this *Version) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}
func (this *Version) CmdType() string {
	return VERSION_TYPE
}

//Deserialize message payload
func (this *Version) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// "verack"
type VerACK struct {
	IsConsensus bool
}

//Serialize message payload
func (this *VerACK) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *VerACK) CmdType() string {
	return VERACK_TYPE
}

//Deserialize message payload
func (this *VerACK) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// "getaddr"
type AddrReq struct{}

//Serialize message payload
func (this AddrReq) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *AddrReq) CmdType() string {
	return GetADDR_TYPE
}

//Deserialize message payload
func (this *AddrReq) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// "addr"
type Addr struct {
	NodeAddrs []PeerAddr
}

//Serialize message payload
func (this Addr) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *Addr) CmdType() string {
	return ADDR_TYPE
}

func (this *Addr) Deserialization(source *sink.ZeroCopySource) error {

	return nil
}

// "ping"

type OrgPIPOInfo struct {
	OrgID  comm.Address
	Height uint64
}

type Ping struct {
	Height   uint64
	InfoType string
	OrgInfo  []OrgPIPOInfo
}

//Serialize message payload
func (this Ping) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *Ping) CmdType() string {
	return PING_TYPE
}

//Deserialize message payload
func (this *Ping) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// "pong"
type Pong struct {
	Height   uint64
	InfoType string
	OrgInfo  []OrgPIPOInfo
}

//Serialize message payload
func (this Pong) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this Pong) CmdType() string {
	return PONG_TYPE
}

//Deserialize message payload
func (this *Pong) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// "getheaders"
type HeadersReq struct {
	Len       uint8     //no use
	HashStart comm.Hash //no use
	HashEnd   comm.Hash //no use
	Height    uint64    //
	OrgID     comm.Address
	SyncType  string
}

//Serialize message payload
func (this *HeadersReq) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *HeadersReq) CmdType() string {
	return GET_HEADERS_TYPE
}

//Deserialize message payload
func (this *HeadersReq) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// "headers"
type BlkHeader struct {
	BlkHdr    []*ct.Header
	BlkOrgHdr []*orgtypes.Header
	OrgID     comm.Address
	SyncType  string
}

//
type BlkP2PHeader struct {
	BlkHdr    []*ct.Header
	BlkOrgHdr []*orgtypes.Header
	OrgID     comm.Address
	SyncType  string
}

func BlkP2PHeaderToBlkHeader(bh *BlkP2PHeader) (*BlkHeader, error) {
	headers := bh.BlkHdr
	hds := make([]*ct.Header, len(headers))
	headerorgs := bh.BlkOrgHdr
	hdorgs := make([]*orgtypes.Header, len(headerorgs))

	if bh.SyncType == SYNC_DATA_MAIN {
		for index, head := range headers {
			hds[index] = head
		}
	} else if bh.SyncType == SYNC_DATA_ORG {
		for index, head := range headerorgs {
			hdorgs[index] = head
		}
	}
	return &BlkHeader{
		BlkHdr:    hds,
		BlkOrgHdr: hdorgs,
		OrgID:     bh.OrgID,
		SyncType:  bh.SyncType,
	}, nil
}

//Serialize message payload
func (this BlkHeader) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *BlkHeader) CmdType() string {
	return HEADERS_TYPE
}

//Deserialize message payload
func (this *BlkHeader) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

func (this BlkP2PHeader) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *BlkP2PHeader) CmdType() string {
	return HEADERS_TYPE
}

//Deserialize message payload
func (this *BlkP2PHeader) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// "inv"
var LastInvHash comm.Hash

type InvPayload struct {
	InvType comm.InventoryType
	Blk     []comm.Hash
	Heis    []uint64
}

type Inv struct {
	P InvPayload
}

func (this Inv) invType() comm.InventoryType {
	return this.P.InvType
}

func (this *Inv) CmdType() string {
	return INV_TYPE
}

//Serialize message payload
func (this Inv) Serialization(sink *sink.ZeroCopySink) error {

	return nil
}

//Deserialize message payload
func (this *Inv) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// getdata
type DataReq struct {
	DataType comm.InventoryType
	Hash     comm.Hash
	OrgID    comm.Address
	SyncType string
}

//Serialize message payload
func (this DataReq) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *DataReq) CmdType() string {
	return GET_DATA_TYPE
}

//Deserialize message payload
func (this *DataReq) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// block
type P2PTransaction struct {
	Raw []byte // raw transaction data
}

type P2PBlock struct {
	Header       *ct.Header //*P2PHeader
	Transactions []*P2PTransaction
	Sigs         *ct.SigData
}

type P2POrgBlock struct {
	Header       *orgtypes.Header //*P2POrgHeader
	Transactions []*P2PTransaction
}

type Block struct {
	Blk      *ct.Block
	BlkOrg   *orgtypes.Block
	OrgID    comm.Address
	SyncType string
}

func TransToP2Ptrans(trans []*ct.Transaction) []*P2PTransaction {
	txs := make([]*P2PTransaction, len(trans))
	for index, tran := range trans {
		txs[index] = new(P2PTransaction)
		txs[index].Raw = tran.Raw
	}
	return txs
}

func OrgTransToP2Ptrans(trans []*orgtypes.Transaction) []*P2PTransaction {
	txs := make([]*P2PTransaction, len(trans))
	for index, tran := range trans {
		txs[index] = new(P2PTransaction)
		txs[index].Raw = tran.Raw
	}
	return txs
}

func P2PtransToTrans(trans []*P2PTransaction) []*ct.Transaction {
	txs := make([]*ct.Transaction, len(trans))
	for index, tran := range trans {

		// txs[index],_ = ct.TransactionFromRawBytes(tran.Raw)
		source := sink.NewZeroCopySource(tran.Raw)
		tx := &ct.Transaction{}
		tx.Deserialization(source)
		txs[index] = tx
	}
	return txs
}

func P2PtransToOrgTrans(trans []*P2PTransaction) []*orgtypes.Transaction {
	txs := make([]*orgtypes.Transaction, len(trans))
	for index, tran := range trans {
		source := sink.NewZeroCopySource(tran.Raw)
		tx := &orgtypes.Transaction{}
		tx.Deserialization(source)
		txs[index] = tx
	}
	return txs
}

type TrsBlock struct {
	Blk      *P2PBlock
	BlkOrg   *P2POrgBlock
	OrgID    comm.Address
	SyncType string
}

func BlockToTrsBlock(tb *Block) *TrsBlock {
	blk := new(P2PBlock)
	blkorg := new(P2POrgBlock)

	if tb.SyncType == SYNC_DATA_MAIN {
		block := tb.Blk
		blk.Header = block.Header
		blk.Transactions = TransToP2Ptrans(block.Transactions)
		blk.Sigs = tb.Blk.Sigs
		// init org
		blkorg.Header = &orgtypes.Header{} //&P2POrgHeader{}

	} else if tb.SyncType == SYNC_DATA_ORG || tb.SyncType == SYNC_DATA_A_TO_STELLAR || tb.SyncType == SYNC_DATA_STELLAR_TO_STELLAR {
		block := tb.BlkOrg
		blkorg.Header = block.Header
		blkorg.Transactions = OrgTransToP2Ptrans(block.Transactions)
		// init main
		blk.Header = &ct.Header{} //&P2PHeader{}
		blk.Sigs = &ct.SigData{}
		fmt.Println("🌐 blockToTrsBlock type ", tb.SyncType)
	}
	return &TrsBlock{
		Blk:      blk,
		BlkOrg:   blkorg,
		OrgID:    tb.OrgID,
		SyncType: tb.SyncType,
	}
}

func TrsBlockToBlock(tb *TrsBlock) (*Block, error) {
	blk := new(ct.Block)
	blkorg := new(orgtypes.Block)
	var err error
	if tb.SyncType == SYNC_DATA_MAIN {
		block := tb.Blk
		blk.Header = block.Header
		blk.Sigs = tb.Blk.Sigs
		if err != nil {
			return nil, err
		}
		blk.Transactions = P2PtransToTrans(block.Transactions)
	} else if tb.SyncType == SYNC_DATA_ORG || tb.SyncType == SYNC_DATA_A_TO_STELLAR || tb.SyncType == SYNC_DATA_STELLAR_TO_STELLAR {
		block := tb.BlkOrg
		blkorg.Header = block.Header
		blkorg.Transactions = P2PtransToOrgTrans(block.Transactions)
	}
	return &Block{
		Blk:      blk,
		BlkOrg:   blkorg,
		OrgID:    tb.OrgID,
		SyncType: tb.SyncType,
	}, nil
}

func (this *TrsBlock) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *TrsBlock) CmdType() string {
	return BLOCK_TYPE
}

//Deserialize message payload
func (this *TrsBlock) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

//Serialize message payload
func (this *Block) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *Block) CmdType() string {
	return BLOCK_TYPE
}

//Deserialize message payload
func (this *Block) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

//consensus

// notfoud
type NotFound struct {
	Hash comm.Hash
}

//Serialize message payload
func (this NotFound) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this NotFound) CmdType() string {
	return NOT_FOUND_TYPE
}

//Deserialize message payload
func (this *NotFound) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// tx
// Transaction message
type Trn struct {
	Txn      *ct.Transaction
	OrgTxn   *orgtypes.Transaction
	OrgID    comm.Address
	SyncType string
}

type P2PTrn struct {
	Txn      *P2PTransaction
	OrgTxn   *P2PTransaction
	OrgID    comm.Address
	SyncType string
}

func TrnToP2PTrn(trn *Trn) *P2PTrn {
	res := new(P2PTransaction)
	orgres := new(P2PTransaction)

	if trn.SyncType == SYNC_DATA_MAIN {
		res.Raw = trn.Txn.Raw
	} else if trn.SyncType == SYNC_DATA_ORG {
		orgres.Raw = trn.OrgTxn.Raw
	}

	res.Raw = trn.Txn.Raw
	return &P2PTrn{
		Txn:      res,
		OrgTxn:   orgres,
		OrgID:    trn.OrgID,
		SyncType: trn.SyncType,
	}
}

func P2PTrnToTrn(trn *P2PTrn) (*Trn, error) {
	var err error
	res := new(ct.Transaction)
	orgres := new(orgtypes.Transaction)

	if trn.SyncType == SYNC_DATA_MAIN {
		source := sink.NewZeroCopySource(trn.Txn.Raw)
		err = res.Deserialization(source)
		// res,err = ct.TransactionFromRawBytes(trn.Txn.Raw)
	} else if trn.SyncType == SYNC_DATA_ORG {
		source := sink.NewZeroCopySource(trn.OrgTxn.Raw)
		err = orgres.Deserialization(source)
	}

	return &Trn{
		Txn:      res,
		OrgTxn:   orgres,
		OrgID:    trn.OrgID,
		SyncType: trn.SyncType,
	}, err
}

//Serialize message payload
func (this Trn) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *Trn) CmdType() string {
	return TX_TYPE
}

//Deserialize message payload
func (this *Trn) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

//Serialize message payload
func (this P2PTrn) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *P2PTrn) CmdType() string {
	return TX_TYPE
}

//Deserialize message payload
func (this *P2PTrn) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// disconnect
type Disconnected struct{}

//Serialize message payload
func (this Disconnected) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this Disconnected) CmdType() string {
	return DISCONNECT_TYPE
}

//Deserialize message payload
func (this *Disconnected) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

///////
type PingSpc struct {
	BNodeASrc bool // true:Source from node A false:Source from stellar
}

//Serialize message payload
func (this PingSpc) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *PingSpc) CmdType() string {
	return PINGSPC_TYPE
}

//Deserialize message payload
func (this *PingSpc) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

// "pong"
type PongSpc struct {
	BNodeASrc bool // true:Source from node A false:Source from stellar
	// true:Is the node requested by the other party // fale:Otherwise, the opposite
	BOtherSideReq bool
}

//Serialize message payload
func (this PongSpc) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this PongSpc) CmdType() string {
	return PONGSPC_TYPE
}

//Deserialize message payload
func (this *PongSpc) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

type EarthNotifyMsg struct {
	Height uint64
	Hash   comm.Hash
}

//Serialize message payload
func (this EarthNotifyMsg) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this EarthNotifyMsg) CmdType() string {
	return EARTHNOTIFYHASH
}

//Deserialize message payload
func (this *EarthNotifyMsg) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

//////

// message
//MsgPayload in link channel
type MsgPayload struct {
	Id          uint64  //peer ID
	Addr        string  //link address
	PayloadSize uint32  //payload size
	Payload     Message //msg payload
}

type Message interface {
	Serialization(sink *sink.ZeroCopySink) error
	Deserialization(source *sink.ZeroCopySource) error
	CmdType() string
}

func MakeEmptyMessage(cmdType string) (Message, error) {
	switch cmdType {
	case PING_TYPE:
		return &Ping{}, nil
	case VERSION_TYPE:
		return &Version{}, nil
	case VERACK_TYPE:
		return &VerACK{}, nil
	case ADDR_TYPE:
		return &Addr{}, nil
	case GetADDR_TYPE:
		return &AddrReq{}, nil
	case PONG_TYPE:
		return &Pong{}, nil
	case GET_HEADERS_TYPE:
		return &HeadersReq{}, nil
	case HEADERS_TYPE:
		// return &BlkHeader{}, nil
		return &BlkP2PHeader{}, nil
	case INV_TYPE:
		return &Inv{}, nil
	case GET_DATA_TYPE:
		return &DataReq{}, nil
	case BLOCK_TYPE:
		// return &Block{}, nil
		return &TrsBlock{}, nil
	case TX_TYPE:
		// return &Trn{}, nil
		return &P2PTrn{}, nil
	case CONSENSUS_TYPE:
		// return &Consensus{}, nil
		return &P2PConsPld{}, nil
	case NOT_FOUND_TYPE:
		return &NotFound{}, nil
	case DISCONNECT_TYPE:
		return &Disconnected{}, nil
	case GET_BLOCKS_TYPE:
		return &BlocksReq{}, nil
	case PINGSPC_TYPE:
		return &PingSpc{}, nil
	case PONGSPC_TYPE:
		return &PongSpc{}, nil
	case EARTHNOTIFYHASH:
		return &EarthNotifyMsg{}, nil
	case NODE_LEAGUE_HEIGHT:
		return &NodeLHMsg{}, nil
	case EXTDATA_RSP_TYPE:
		return &ExtDataResponse{}, nil
	case EXTDATA_REQ_TYPE:
		return &ExtDataRequest{}, nil
	default:
		return nil, errors.New("unsupported cmd type:" + cmdType)
	}

}

// getblocks
type BlocksReq struct {
	HeaderHashCount uint8
	HashStart       comm.Hash
	HashStop        comm.Hash
}

//Serialize message payload
func (this *BlocksReq) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *BlocksReq) CmdType() string {
	return GET_BLOCKS_TYPE
}

//Deserialize message payload
func (this *BlocksReq) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

//=========================================================
type RdrBlock struct {
	Blk      *P2PBlock
	BlkOrg   *P2POrgBlock
	OrgID    comm.Address
	SyncType string
}

func BlockToRdrBlock(tb *Block) *RdrBlock {
	blk := new(P2PBlock)
	blkorg := new(P2POrgBlock)

	if tb.SyncType == SYNC_DATA_MAIN {
		block := tb.Blk
		blk.Header = block.Header
		blk.Transactions = TransToP2Ptrans(block.Transactions)
		blk.Sigs = tb.Blk.Sigs
		// init org
		blkorg.Header = &orgtypes.Header{} //&P2POrgHeader{}
	} else if tb.SyncType == SYNC_DATA_ORG {
		block := tb.BlkOrg
		blkorg.Header = block.Header
		blkorg.Transactions = OrgTransToP2Ptrans(block.Transactions)
		// init main
		blk.Header = &ct.Header{} //&P2PHeader{}
		blk.Sigs = &ct.SigData{}
	}
	return &RdrBlock{
		Blk:      blk,
		BlkOrg:   blkorg,
		OrgID:    tb.OrgID,
		SyncType: tb.SyncType,
	}
}

func RdrBlockToBlock(tb *RdrBlock) (*Block, error) {
	blk := new(ct.Block)
	blkorg := new(orgtypes.Block)
	if tb.SyncType == SYNC_DATA_MAIN {
		block := tb.Blk
		blk.Header = block.Header
		blk.Sigs = tb.Blk.Sigs
		blk.Transactions = P2PtransToTrans(block.Transactions)
	} else if tb.SyncType == SYNC_DATA_ORG {
		block := tb.BlkOrg
		blkorg.Header = block.Header
		blkorg.Transactions = P2PtransToOrgTrans(block.Transactions)
	}
	return &Block{
		Blk:      blk,
		BlkOrg:   blkorg,
		OrgID:    tb.OrgID,
		SyncType: tb.SyncType,
	}, nil
}

func (this *RdrBlock) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *RdrBlock) CmdType() string {
	return GET_DATA_TYPE
}

//Deserialize message payload
func (this *RdrBlock) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

type NodeLHMsg struct {
	NodeId   string
	LeagueId comm.Address
	Height   uint64
}

//Serialize message payload
func (this NodeLHMsg) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this NodeLHMsg) CmdType() string {
	return NODE_LEAGUE_HEIGHT
}

//Deserialize message payload
func (this *NodeLHMsg) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

func (nlh *NodeLHMsg) Unique() string {
	return fmt.Sprintf("%s_%s_%d", nlh.LeagueId, nlh.NodeId, nlh.Height)
}

type ExtDataRequest struct {
	LeagueId comm.Address //if IsReq=true, must have
	Height   uint64       //if IsReq=true, must have
}

func (this *ExtDataRequest) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *ExtDataRequest) CmdType() string {
	return EXTDATA_REQ_TYPE
}

//Deserialize message payload
func (this *ExtDataRequest) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}

type ExtDataResponse struct {
	*extstates.ExtData
}

func (this *ExtDataResponse) Serialization(sink *sink.ZeroCopySink) error {
	return nil
}

func (this *ExtDataResponse) CmdType() string {
	return EXTDATA_RSP_TYPE
}

//Deserialize message payload
func (this *ExtDataResponse) Deserialization(source *sink.ZeroCopySource) error {
	return nil
}
