package comm

import (
	"encoding/json"
	"math/big"
	"github.com/sixexorg/magnetic-ring/common"
	//"github.com/ontio/ontology-crypto/keypair"
	p2pmsg "github.com/sixexorg/magnetic-ring/p2pserver/common"
	"github.com/sixexorg/magnetic-ring/crypto"
)

type P2pMsgPayload struct {
	FromPeer string
	Payload  *p2pmsg.ConsensusPayload
}


type  EarthSigsFetchRspMsg struct {
	BlkNum   uint64
	PubKey   string
	Sigs     []byte
	BlkHash  common.Hash
}

func (e *EarthSigsFetchRspMsg) Type() MsgType {
	return EarthFetchRspSigs
}

func (e *EarthSigsFetchRspMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (e *EarthSigsFetchRspMsg) GetBlockNum() uint64 {
	return e.BlkNum
}

func (e *EarthSigsFetchRspMsg) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

type VrfData struct {
	BlockNum uint64 `json:"block_num"`
	PrevVrf  []byte `json:"prev_vrf"`
}

type  TimeoutFetchRspMsg struct {
	BlkNum   uint64
	PubKey   string
	Sigs     []byte
	SigsArr  [][]byte
}

func (e *TimeoutFetchRspMsg) Type() MsgType {
	return TimeoutFetchRspMessage
}

func (e *TimeoutFetchRspMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (e *TimeoutFetchRspMsg) GetBlockNum() uint64 {
	return e.BlkNum
}

func (e *TimeoutFetchRspMsg) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

type  TimeoutFetchMsg struct {
	BlkNum   uint64
	PubKey   string
	Sigs     []byte
}

func (e *TimeoutFetchMsg) Type() MsgType {
	return TimeoutFetchMessage
}

func (e *TimeoutFetchMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (e *TimeoutFetchMsg) GetBlockNum() uint64 {
	return e.BlkNum
}

func (e *TimeoutFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(e)
}


type ViewData struct {
	BlkNum 		   uint64
	View 		   uint32
}

type ViewtimeoutMsg struct {
	RawData      *ViewData
	Signature    []byte
	PubKey       string
}

func (v *ViewtimeoutMsg) Type() MsgType {
	return ViewTimeout
}

func (v *ViewtimeoutMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (v *ViewtimeoutMsg) GetBlockNum() uint64 {
	return v.RawData.BlkNum
}

func (v *ViewtimeoutMsg) Serialize() ([]byte, error) {
	return json.Marshal(v)
}

type EvilMsg struct {
	BlkNum   uint64
	EvilSigs [][]byte
}

func (e *EvilMsg) Type() MsgType {
	return EvilEvent
}

func (e *EvilMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (e *EvilMsg) GetBlockNum() uint64 {
	return e.BlkNum
}

func (e *EvilMsg) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

type  EventFetchMsg struct {
	BlkNum   uint64
}

func (e *EventFetchMsg) Type() MsgType {
	return EventFetchMessage
}

func (e *EventFetchMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (e *EventFetchMsg) GetBlockNum() uint64 {
	return e.BlkNum
}

func (e *EventFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

type TimeoutDesc struct {
	BlkNum uint32
	View   uint32
	PubKey string
	Sig    []byte
}


type  EventFetchRspMsg struct {
	BlkNum    uint64
	TimeoutEv []interface{}
	EvilEv    []interface{}
}

func (e *EventFetchRspMsg) Type() MsgType {
	return EventFetchRspMessage
}

func (e *EventFetchRspMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (e *EventFetchRspMsg) GetBlockNum() uint64 {
	return e.BlkNum
}

func (e *EventFetchRspMsg) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

type  EvilFetchMsg struct {
	BlkNum   uint64
}

func (e *EvilFetchMsg) Type() MsgType {
	return EvilFetchMessage
}

func (e *EvilFetchMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (e *EvilFetchMsg) GetBlockNum() uint64 {
	return e.BlkNum
}

func (e *EvilFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

type  EvilFetchRspMsg struct {
	BlkNum   uint64
	Sigs     [][]byte
}

func (e *EvilFetchRspMsg) Type() MsgType {
	return EvilFetchRspMessage
}

func (e *EvilFetchRspMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (e *EvilFetchRspMsg) GetBlockNum() uint64 {
	return e.BlkNum
}

func (e *EvilFetchRspMsg) Serialize() ([]byte, error) {
	return json.Marshal(e)
}


type ConsedoneMsg struct {
	BlockNumber    uint64
	View           uint32
	BlockData      []byte
	PublicKey      string
	SigData        []byte
}

func (p *ConsedoneMsg)Type() MsgType{
	return ConsenseDone
}
func (p *ConsedoneMsg)Verify(pub crypto.PublicKey) error{
	return nil
}
func (p *ConsedoneMsg)GetBlockNum() uint64{
	return p.BlockNumber
}

func (p *ConsedoneMsg)Serialize() ([]byte, error){
	return json.Marshal(p)
}

type P1a struct {
	Ballot       Ballot
	ID           uint32
	BlockNumber  uint64
	View         uint32
	TxHashs      []common.Hash
	TotalGas     *big.Int
}

type PrepareMsg struct {
	Msg    P1a `json:"msg"`
	Sig   []byte
}

func (p *PrepareMsg)Type() MsgType{
	return ConsensePrepare
}
func (p *PrepareMsg)Verify(pub crypto.PublicKey) error{
	return nil
}
func (p *PrepareMsg)GetBlockNum() uint64{
	return p.Msg.BlockNumber
}

func (p *PrepareMsg)Serialize() ([]byte, error){
	return json.Marshal(p)
}

type P1b struct {
	Ballot Ballot
	ID     uint32               // node id
	Ab	  Ballot
	Av    []byte
	BlockNumber  uint64
	View         uint32
}

type PromiseMsg struct {
	Msg P1b `json:"msg"`
	Sig []byte
}

func (p *PromiseMsg)Type() MsgType{
	return ConsensePromise
}
func (p *PromiseMsg)Verify(pub crypto.PublicKey) error{
	return nil
}
func (p *PromiseMsg)GetBlockNum() uint64{
	return p.Msg.BlockNumber
}

func (p *PromiseMsg)Serialize() ([]byte, error){
	return json.Marshal(p)
}

type P2a struct {
	ID      uint32
	Ballot  Ballot
	Slot    int
	Av      []byte

	BlockNumber  uint64
	View         uint32
}

type ProposerMsg struct {
	Msg P2a `json:"msg"`
	Sig []byte
}

func (p *ProposerMsg)Type() MsgType{
	return ConsenseProposer
}
func (p *ProposerMsg)Verify(pub crypto.PublicKey) error{
	return nil
}
func (p *ProposerMsg)GetBlockNum() uint64{
	return p.Msg.BlockNumber
}

func (p *ProposerMsg)Serialize() ([]byte, error){
	return json.Marshal(p)
}

type P2b struct {
	Ballot Ballot
	ID     uint32 // from node id
	BlockNumber  uint64
	View         uint32
	Signature []byte
}

type AcceptMsg struct {
	Msg P2b `json:"msg"`
	Sig []byte
}

func (p *AcceptMsg)Type() MsgType{
	return ConsenseAccept
}
func (p *AcceptMsg)Verify(pub crypto.PublicKey) error{
	return nil
}
func (p *AcceptMsg)GetBlockNum() uint64{
	return p.Msg.BlockNumber
}

func (p *AcceptMsg)Serialize() ([]byte, error){
	return json.Marshal(p)
}


type BlockFetchMsg struct {
	BlockNum uint64 `json:"block_num"`
}

func (msg *BlockFetchMsg) Type() MsgType {
	return BlockFetchMessage
}

func (msg *BlockFetchMsg) Verify(pub crypto.PublicKey) error {
	return nil
}

func (msg *BlockFetchMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *BlockFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}
