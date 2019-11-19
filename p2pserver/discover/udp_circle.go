package discover

import (
	"errors"
	"fmt"
	"net"
	"time"

	comm "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/log"
)

const (
	maxNeighborsCircle = 10
	REQPEERORGNUM      = 200
)

// Errors
var (
	errCirclePacketTooSmall = errors.New("circle data too small")
	// errBadHash          = errors.New("bad hash")
	// errExpired          = errors.New("expired")
	// errUnsolicitedReply = errors.New("unsolicited reply")
	errUnknownCircle = errors.New("unknown circle")
	// errTimeout          = errors.New("RPC timeout")
	// errClockWarp        = errors.New("reply deadline too far in the future")
	// errClosed           = errors.New("socket closed")
)

// RPC request structures
type (
	sendcircle struct {
		Version    uint
		From, To   rpcEndpoint
		SrcOrgID   comm.Address
		OwnID      uint64
		BAdd       bool
		Expiration uint64
	}

	// pong is the reply to ping.
	rtncircle struct {
		// This field should mirror the UDP envelope address
		// of the ping packet, which provides a way to discover the
		// the external address (after NAT).
		To rpcEndpoint

		ReplyTok   []byte // This contains the hash of the ping packet.
		Expiration uint64 // Absolute timestamp at which the packet becomes invalid.
	}

	sendConnect struct {
		Version         uint
		From, To        rpcEndpoint
		OwnID, RemoteID uint64
		BConnect        bool
		BStellarNode    bool
		Expiration      uint64
	}

	rtnConnect struct {
		// This field should mirror the UDP envelope address
		// of the ping packet, which provides a way to discover the
		// the external address (after NAT).
		To rpcEndpoint

		ReplyTok   []byte // This contains the hash of the ping packet.
		Expiration uint64 // Absolute timestamp at which the packet becomes invalid.
	}

	findcircle struct {
		// Target     CirclseID // doesn't need to be an actual public key
		Target     comm.Address
		Expiration uint64
	}

	neighborscircle struct {
		Nodes      []rpcNode
		Num        uint64
		Expiration uint64
	}

	reqPeerOrg struct {
		Target     NodeID // doesn't need to be an actual public key
		Expiration uint64
	}

	rntPeerOrg struct {
		Num        uint64
		OwnID      uint64
		Orgs       []comm.Address
		Expiration uint64
	}

	reqPeerConnect struct {
		Target     NodeID // doesn't need to be an actual public key
		Expiration uint64
	}

	// reply to findnode
	rntPeerConnect struct {
		Num        uint64
		OwnID      uint64
		RemoteID   []uint64
		Expiration uint64
	}

	connectCircuOrg struct {
		From        rpcEndpoint
		OrgID       comm.Address
		ConnectNode rpcNode
		Expiration  uint64
	}

	rtnCircuOrg struct {
		Expiration uint64
	}
)

func (t *udp) sendcircle(ndoe, srcnode *Node, SrcOrgID comm.Address, ownid uint64, badd bool) error {
	// TODO: maybe check for ReplyTo field in callback to measure RTT
	//fmt.Println(" ****** udp sendcircle start ...... badd:",badd)
	toid, toaddr := ndoe.ID, ndoe.addr()
	errc := t.pending(toid, rtncirclePacket, func(interface{}) bool { return true })
	t.send(toaddr, sendcirclePacket, &sendcircle{
		Version: Version,
		From:    t.ourEndpoint,
		// To:         makeEndpoint(toaddr, 0), // TODO: maybe use known TCP port from DB
		BAdd:       badd,
		OwnID:      ownid,
		SrcOrgID:   SrcOrgID,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	})
	errs := <-errc
	return errs
}

func (t *udp) findcircle(toid NodeID, toaddr *net.UDPAddr, target comm.Address) ([]*Node, error) {
	//fmt.Println(" ******* udp findcircle ...... ")
	nodes := make([]*Node, 0)
	nreceived := uint64(0)
	total := uint64(0)
	errc := t.pending(toid, neighborscirclePacket, func(r interface{}) bool {
		reply := r.(*neighborscircle)
		//fmt.Println(" ***** len(reply.Nodes):",len(reply.Nodes))
		//fmt.Println(" ***** reply.Nodes:",reply.Nodes)
		total = reply.Num
		for _, rn := range reply.Nodes {
			nreceived++
			// n, err := t.nodeFromRPC(toaddr, rn)
			n, err := t.nodeFromRPC2(toaddr, rn)

			if err != nil {
				log.Trace("Invalid neighbor node received", "ip", rn.IP, "addr", toaddr, "err", err)
				//fmt.Println(" *** findcircle received", "ip", rn.IP, "addr", toaddr, "err", err)
				continue
			}
			nodes = append(nodes, n)
		}
		return nreceived >= total
	})
	t.send(toaddr, findcirclePacket, &findcircle{
		Target:     target,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	})
	err := <-errc
	return nodes, err
}

func (t *udp) sendConnectInfo(ndoe, srcnode *Node, ownid, remoteid uint64, bconnect, bstellar bool) error {
	// TODO: maybe check for ReplyTo field in callback to measure RTT
	//fmt.Println(" ****** udp sendConnect start ...... bconnect:",bconnect)
	//defer fmt.Println(" ****** udp sendConnect end ...... bconnect:", bconnect)

	toid, toaddr := ndoe.ID, ndoe.addr()
	errc := t.pending(toid, rtnconnectPacket, func(interface{}) bool { return true })
	eerr := t.send(toaddr, sendConnectPacket, &sendConnect{
		Version:      Version,
		From:         t.ourEndpoint,
		OwnID:        ownid,
		RemoteID:     remoteid,
		BConnect:     bconnect,
		BStellarNode: bstellar,
		Expiration:   uint64(time.Now().Add(expiration).Unix()),
	})

	if eerr!=nil {
		fmt.Println("🐞 🐞 🐞 🐞 🐞 🐞 🐞 🐞-->",eerr)
	}

	return <-errc
}

func (t *udp) reqPeerOrg(toid NodeID, toaddr *net.UDPAddr, node *Node) (*rtnTabPeerOrg, error) {
	//fmt.Println(" ******* udp reqPeerOrg ...... ")
	orgs := new(rtnTabPeerOrg)
	orgs.orgsId = make([]comm.Address, 0)
	orgs.node = node

	receiveNum := uint64(0)
	total := uint64(0)

	errc := t.pending(toid, rntPeerOrgPacket, func(r interface{}) bool {
		reply := r.(*rntPeerOrg)
		//fmt.Println(" ***** len(reply.Nodes):",len(reply.Orgs))
		total = reply.Num
		orgs.ownId = reply.OwnID
		// for _, rn := range reply.Orgs {
		orgs.orgsId = append(orgs.orgsId, reply.Orgs...)
		receiveNum = receiveNum + uint64(len(reply.Orgs))
		// }
		return receiveNum >= total
	})
	t.send(toaddr, reqPeerOrgPacket, &reqPeerOrg{
		Target:     toid,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	})
	err := <-errc
	//fmt.Println(" *8888888* udp reqPeerConnect orgs:",*orgs)
	return orgs, err
}

func (t *udp) reqPeerConnect(toid NodeID, toaddr *net.UDPAddr, node *Node) (*rtnTabPeerConnect, error) {
	//fmt.Println(" ******* udp reqPeerConnect ...... ")
	receiveNum := uint64(0)
	total := uint64(0)
	result := new(rtnTabPeerConnect)
	result.remotesID = make([]uint64, 0)
	result.node = node

	errc := t.pending(toid, rntPeerConnectPacket, func(r interface{}) bool {
		reply := r.(*rntPeerConnect)
		//fmt.Println(" ***** reqPeerConnect len(reply.Peer):",len(reply.RemoteID))
		total = reply.Num
		result.ownId = reply.OwnID
		for _, rn := range reply.RemoteID {
			result.remotesID = append(result.remotesID, rn)
			receiveNum = receiveNum + 1
		}
		return receiveNum >= total
	})
	t.send(toaddr, reqPeerConnectPacket, &reqPeerConnect{
		Target:     toid,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	})
	err := <-errc
	//fmt.Println(" *8888888* udp reqPeerConnect result:",*result)
	return result, err
}

func (t *udp) connectCircuOrg(ndoe, connectnode *Node, orgid comm.Address) error {
	// TODO: maybe check for ReplyTo field in callback to measure RTT
	//fmt.Println(" ****** udp connectCircuOrg start ...... ")
	defer fmt.Println(" ****** udp connectCircuOrg end ...... ")
	toid, toaddr := ndoe.ID, ndoe.addr()
	errc := t.pending(toid, rtnCircuOrgPacket, func(interface{}) bool { return true })
	t.send(toaddr, connectCircuOrgPacket, &connectCircuOrg{
		From:        t.ourEndpoint,
		OrgID:       orgid,
		ConnectNode: nodeToRPC(connectnode), //
		Expiration:  uint64(time.Now().Add(expiration).Unix()),
	})
	return <-errc
}

func (req *sendcircle) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	t.send(from, rtncirclePacket, &rtncircle{
		To:         makeEndpoint(from, req.From.TCP),
		ReplyTok:   mac,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	})

	// // add circle
	// fmt.Println(" ***** sendcircle handle from:",from)
	// fmt.Println(" ***** sendcircle handle fromID:",fromID)
	// fmt.Println(" ***** sendcircle handle req.From:",req.From)
	n := NewNode(fromID, from.IP, uint16(from.Port), req.From.TCP)
	if req.SrcOrgID == t.stellarNodeID {
		go t.stellarNode(n, req.SrcOrgID, req.OwnID, req.BAdd)
	} else {
		go t.setOrgs(n, req.SrcOrgID, req.OwnID, req.BAdd)
	}

	return nil
}

func (req *sendcircle) name() string { return "sendcircle/v4" }

func (req *rtncircle) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.handleReply(fromID, rtncirclePacket, req) {
		return errUnsolicitedReply
	}
	return nil
}

func (req *rtncircle) name() string { return "rtncircle/v4" }

func (req *findcircle) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	//fmt.Println(" ****** findcircle handle ...... ")
	if expired(req.Expiration) {
		return errExpired
	}
	// if t.db.node(fromID) == nil {
	// 	return errUnknownNode
	// }

	var err error
	targetData := make([]*Node, 0)
	if req.Target == t.stellarNodeID {
		targetData, err = t.getStellar()
	} else {
		targetData, err = t.getCircle(req.Target, maxNeighborsCircle)
	}

	if err != nil {
		return errUnknownCircle
	}

	if len(targetData) <= 0 {
		return errCirclePacketTooSmall
	}

	p := neighborscircle{
		Num:        uint64(len(targetData)),
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}
	// Send neighbors in chunks with at most maxNeighbors per packet
	// to stay below the 1280 byte limit.
	for i, n := range targetData {
		// errip := netutil.CheckRelayIP(from.IP, n.IP)
		// if errip != nil {
		// 	fmt.Printf(" ***** CheckRelayIP from.IP:%v,n.IP:%v\n",from.IP,n.IP)
		// 	fmt.Println(" ****** CheckRelayIP error errip:",errip)
		// 	continue
		// }
		p.Nodes = append(p.Nodes, nodeToRPC(n))
		if len(p.Nodes) == maxNeighbors || i == len(targetData)-1 {
			t.send(from, neighborscirclePacket, &p)
			p.Nodes = p.Nodes[:0]
		}
	}
	return nil
}

func (req *findcircle) name() string { return "FINDCIRCLE/v4" }

func (req *neighborscircle) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	//fmt.Println(" ******** neighborscircle handle len(req.Nodes):",len(req.Nodes))
	err := t.handleReply(fromID, neighborscirclePacket, req)
	if !err {
		//fmt.Println(" ****** neighborscircle handle err:",err)
		return errUnsolicitedReply
	}
	return nil
}

func (req *neighborscircle) name() string { return "NEIGHBORSCIRCLE/v4" }

func (req *sendConnect) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	t.send(from, rtnconnectPacket, &rtnConnect{
		To:         makeEndpoint(from, req.From.TCP),
		ReplyTok:   mac,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	})

	// add circle
	// fmt.Println(" ***** sendConnect handle from:",from)
	// fmt.Println(" ***** sendConnect handle fromID:",fromID)
	// fmt.Println(" ***** sendConnect handle req.From:",req.From)
	if !req.BStellarNode {
		n := NewNode(fromID, from.IP, uint16(from.Port), req.From.TCP)
		go t.setConnect(req.OwnID, req.RemoteID, req.BConnect, n) //
		if !req.BConnect {
			go t.setConnect(req.RemoteID, req.OwnID, req.BConnect, n) //
		} else if req.BConnect {
			// go t.setConnect(req.RemoteID,req.OwnID,req.BConnect,nil) //
		}
	} else {
		t.stellarInspectDis(req.RemoteID)
	}

	return nil
}

func (req *sendConnect) name() string { return "sendConnect/v4" }

func (req *rtnConnect) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.handleReply(fromID, rtnconnectPacket, req) {
		return errUnsolicitedReply
	}
	return nil
}

func (req *rtnConnect) name() string { return "rtnConnect/v4" }

func (req *reqPeerOrg) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	//fmt.Println(" ****** reqPeerOrg handle ...... ")
	if expired(req.Expiration) {
		return errExpired
	}

	ownID, orgsId := t.callPeerOrgInfo()

	p := rntPeerOrg{
		Expiration: uint64(time.Now().Add(expiration).Unix()),
		Num:        uint64(len(orgsId)),
		OwnID:      ownID,
	}
	// Send neighbors in chunks with at most maxNeighbors per packet
	// to stay below the 1280 byte limit.
	if len(orgsId) > 0 {
		for i, n := range orgsId {
			p.Orgs = append(p.Orgs, n)
			if len(p.Orgs) == maxPeerOrg || i == len(orgsId)-1 {
				t.send(from, rntPeerOrgPacket, &p)
				p.Orgs = p.Orgs[:0]
			}
		}
	} else if len(orgsId) <= 0 {
		t.send(from, rntPeerOrgPacket, &p)
	}
	return nil
}

func (req *reqPeerOrg) name() string { return "reqPeerOrg/v4" }

func (req *rntPeerOrg) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	// fmt.Println(" ******** rntPeerOrg handle len(req.Nodes):",len(req.))
	err := t.handleReply(fromID, rntPeerOrgPacket, req)
	if !err {
		//fmt.Println(" ****** rntPeerOrg handle err:",err)
		return errUnsolicitedReply
	}
	return nil
}

func (req *rntPeerOrg) name() string { return "rntPeerOrg/v4" }

func (req *reqPeerConnect) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	fmt.Println(" ****** reqPeerConnect handle ...... ")
	if expired(req.Expiration) {
		return errExpired
	}
	// targetData := new(rtnTabPeerConnect)
	// var err error
	ownID, remotesID := t.callPeerConnectInfo()

	p := rntPeerConnect{
		Expiration: uint64(time.Now().Add(expiration).Unix()),
		Num:        uint64(len(remotesID)),
		OwnID:      ownID,
	}
	// Send neighbors in chunks with at most maxNeighbors per packet
	// to stay below the 1280 byte limit.
	if len(remotesID) > 0 {
		for i, n := range remotesID {
			p.RemoteID = append(p.RemoteID, n)
			if len(p.RemoteID) == maxPeerOrg || i == len(remotesID)-1 {
				t.send(from, rntPeerConnectPacket, &p)
				p.RemoteID = p.RemoteID[:0]
			}
		}
	} else if len(remotesID) <= 0 {
		t.send(from, rntPeerConnectPacket, &p)
	}
	return nil
}

func (req *reqPeerConnect) name() string { return "reqPeerConnect/v4" }

func (req *rntPeerConnect) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	// fmt.Println(" ******** rntPeerOrg handle len(req.Nodes):",len(req.))
	err := t.handleReply(fromID, rntPeerConnectPacket, req)
	if !err {
		fmt.Println(" ****** rntPeerConnect handle err:", err)
		return errUnsolicitedReply
	}
	return nil
}

func (req *rntPeerConnect) name() string { return "rntPeerConnect/v4" }

// circu
func (req *connectCircuOrg) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	fmt.Println(" ****** connectCircuOrg handle ...... ")
	if expired(req.Expiration) {
		return errExpired
	}

	n := NewNode(fromID, from.IP, uint16(from.Port), req.From.TCP) // bootnode
	destnode, err := t.nodeFromRPC(n.addr(), req.ConnectNode)      //dest node
	if err == nil {
		t.callSendConnectOrg(destnode, req.OrgID)
	}

	p := rtnCircuOrg{Expiration: uint64(time.Now().Add(expiration).Unix())}

	t.send(from, rtnCircuOrgPacket, &p)
	return nil
}

func (req *connectCircuOrg) name() string { return "connectCircuOrg/v4" }

func (req *rtnCircuOrg) handle(t *udp, from *net.UDPAddr, fromID NodeID, mac []byte) error {
	if expired(req.Expiration) {
		return errExpired
	}
	// fmt.Println(" ******** rntPeerOrg handle len(req.Nodes):",len(req.))
	err := t.handleReply(fromID, rtnCircuOrgPacket, req)
	if !err {
		fmt.Println(" ****** rtnCircuOrg handle err:", err)
		return errUnsolicitedReply
	}
	return nil
}

func (req *rtnCircuOrg) name() string { return "rtnCircuOrg/v4" }
