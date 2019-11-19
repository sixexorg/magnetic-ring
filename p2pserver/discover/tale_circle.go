package discover

import (
	"math/rand"
	"sync"
	"time"

	comm "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/log"
)

const (
	CONNECT_NUM       = 3
	REFRESH_NUM       = 10
	CircuConnectNum   = 3
	StellarDisChanLen = 1000
)

type OrgNodeInfo struct {
	node   *Node
	errnum int
	peerid uint64
}

// peerid -> OrgNodeInfo
type NodeMap map[uint64]bool
type PeerOrg map[comm.Address]bool

type OrgInfo struct {
	sync.Mutex
	orgSingleData    map[comm.Address]NodeMap // orgid -> NodeMap
	peerIDToOrgIDMap map[uint64]PeerOrg
	peerIDToNode     map[uint64]*OrgNodeInfo
}

// connect
type ConnPeer map[uint64]bool
type ConnectInfo struct {
	sync.Mutex
	peerToPeerMap map[uint64]ConnPeer //
	peerIDToNode  map[uint64]*OrgNodeInfo
}

func (tab *Table) SendBootNodeOrg(nodes []*Node, orgid comm.Address, ownid uint64, badd bool) ([]*Node, []error) {
	errNode, errs := tab.sendbootnodeorg(nodes, orgid, ownid, badd)
	i := 1
	for len(errNode) > 0 {
		i = i + 1
		time.Sleep(time.Second)
		errNode, errs = tab.sendbootnodeorg(errNode, orgid, ownid, badd)
		if i >= 3 {
			break
		}
	}
	return errNode, errs
}

func (tab *Table) sendbootnodeorg(nodes []*Node, orgid comm.Address, ownid uint64, badd bool) ([]*Node, []error) {
	errNode := make([]*Node, 0)
	errs := make([]error, 0)
	for _, n := range nodes {
		err := tab.net.sendcircle(n, tab.Self(), orgid, ownid, badd)
		log.Info("=====>>>>>Table sendbootnodeorg", "node:", n, "orgid", orgid.ToString(), "badd", badd, "err", err)
		if err != nil {
			errNode = append(errNode, n)
			errs = append(errs, err)
		}
	}
	return errNode, errs
}

func (tab *Table) SendBootNodeConnect(nodes []*Node, ownid, remoteid uint64, bconnect, bstellar bool) ([]*Node, []error) {
	//fmt.Println(" **** Table SendBootNodeConnect bconnect:",bconnect)
	errNode, errs := tab.sendbootnodeconnect(nodes, ownid, remoteid, bconnect, bstellar)
	i := 1
	for len(errNode) > 0 {
		i = i + 1
		time.Sleep(time.Second)
		errNode, errs = tab.sendbootnodeconnect(nodes, ownid, remoteid, bconnect, bstellar)
		if i >= 3 {
			break
		}
	}
	return errNode, errs
}

func (tab *Table) sendbootnodeconnect(nodes []*Node, ownid, remoteid uint64, bconnect, bstellar bool) ([]*Node, []error) {
	errNode := make([]*Node, 0)
	errs := make([]error, 0)
	for _, n := range nodes {
		err := tab.net.sendConnectInfo(n, tab.Self(), ownid, remoteid, bconnect, bstellar)
		if err != nil {
			errNode = append(errNode, n)
			errs = append(errs, err)
		}
	}
	return errNode, errs
}

func (tab *Table) GetOrgNodesFromBN(bns []*Node, targetID comm.Address) map[NodeID]*Node {
	results := make(map[NodeID]*Node)
	for _, n := range bns {
		nodes, err := tab.getOrgNodesFromBN(n, targetID)
		if err != nil {
			continue
		}
		for _, node := range nodes {
			results[node.ID] = node
		}
	}
	return results
}

func (tab *Table) getOrgNodesFromBN(n *Node, targetID comm.Address) ([]*Node, error) {
	//fmt.Println(" ****** Table readCircleNodes... ")
	r, err := tab.net.findcircle(n.ID, n.addr(), targetID)
	if err != nil {
		//fmt.Println(" ****** err:",err)
		return nil, err
	}
	return r, nil
}

func (tab *Table) delConnectPeer(ownid uint64) {
	tab.connectData.Lock()
	defer tab.connectData.Unlock()

	// del peerToPeerMap
	delete(tab.connectData.peerToPeerMap, ownid)
	// del peerIDToNode
	delete(tab.connectData.peerIDToNode, ownid)

	if len(tab.connectData.peerToPeerMap) <= 0 {
		tab.connectData.peerToPeerMap = make(map[uint64]ConnPeer)
		tab.connectData.peerIDToNode = make(map[uint64]*OrgNodeInfo)
	}
	//fmt.Println(" ************ Table delConnectPeer peerIDToNode:",tab.connectData.peerIDToNode)
}

// updata connect info for batch
func (tab *Table) updateConnectForBatch(batchInfo *rtnTabPeerConnect) {
	// remoteid connects is null,so delete the ownId
	if len(batchInfo.remotesID) == 0 {
		tab.delConnectPeer(batchInfo.ownId)
		return
	}
	// update remoteid connect info
	connpeer := make(ConnPeer)
	for _, peerid := range batchInfo.remotesID {
		connpeer[peerid] = true
	}

	tab.connectData.Lock()
	defer tab.connectData.Unlock()
	// update connectData
	tab.connectData.peerToPeerMap[batchInfo.ownId] = connpeer
	tab.connectData.peerIDToNode[batchInfo.ownId] = &OrgNodeInfo{node: batchInfo.node, errnum: 0, peerid: batchInfo.ownId}
}

// set connect info for single
func (tab *Table) setConnect(ownid, remoteid uint64, bconnect bool, node *Node) {
	tab.connectData.Lock()
	defer tab.connectData.Unlock()

	if bconnect {
		// add to peerToPeerMap
		peertopeerdata, peerok := tab.connectData.peerToPeerMap[ownid]
		if !peerok {
			peertopeerdata = make(ConnPeer)
		}
		peertopeerdata[remoteid] = true
		tab.connectData.peerToPeerMap[ownid] = peertopeerdata

		// add to peerIDToNode
		tab.connectData.peerIDToNode[ownid] = &OrgNodeInfo{node: node, errnum: 0, peerid: ownid}

		//fmt.Println(" *888888* Table setConnect add peerToPeerMap:",tab.connectData.peerToPeerMap)
		//fmt.Println(" *888888* Table setConnect add peerIDToNode:",tab.connectData.peerIDToNode)
	} else {
		peertopeerdata, peerok := tab.connectData.peerToPeerMap[ownid]
		if peerok {
			delete(peertopeerdata, remoteid)
			if len(peertopeerdata) <= 0 {
				// delete peer -> peer
				delete(tab.connectData.peerToPeerMap, ownid)
				// delete peer -> node
				delete(tab.connectData.peerIDToNode, ownid)

				if len(tab.connectData.peerToPeerMap) <= 0 {
					tab.connectData.peerToPeerMap = make(map[uint64]ConnPeer)
					tab.connectData.peerIDToNode = make(map[uint64]*OrgNodeInfo)
				}
			} else {
				tab.connectData.peerToPeerMap[ownid] = peertopeerdata
			}
		}
		//fmt.Println(" *888888* Table setConnect del peerToPeerMap:",tab.connectData.peerToPeerMap)
		//fmt.Println(" *888888* Table setConnect del peerIDToNode:",tab.connectData.peerIDToNode)
	}
}

func (tab *Table) delCirclePeer(ownid uint64) {
	tab.orgData.Lock()
	defer tab.orgData.Unlock()

	// del the peer form peerIDToOrgIDMap
	peerorgdata, peerorgok := tab.orgData.peerIDToOrgIDMap[ownid]
	if peerorgok {
		for circleID, _ := range peerorgdata {
			circledata, circleok := tab.orgData.orgSingleData[circleID]
			if circleok {
				delete(circledata, ownid)
				if len(circledata) == 0 {
					delete(tab.orgData.orgSingleData, circleID)
					if len(tab.orgData.orgSingleData) == 0 { // tab.orgData.orgSingleData
						tab.orgData.orgSingleData = make(map[comm.Address]NodeMap)
					}
				} else {
					tab.orgData.orgSingleData[circleID] = circledata
				}
			}
		}
	}
	// del the peer form peerIDToOrgIDMap and peerIDToNode
	delete(tab.orgData.peerIDToOrgIDMap, ownid)
	delete(tab.orgData.peerIDToNode, ownid)
	if len(tab.orgData.peerIDToOrgIDMap) == 0 { // tab.orgData.peerIDToOrgIDMap is empty
		tab.orgData.peerIDToOrgIDMap = make(map[uint64]PeerOrg)
		tab.orgData.peerIDToNode = make(map[uint64]*OrgNodeInfo)
	}
}

func (tab *Table) updateOrgForBatch(batchOrg *rtnTabPeerOrg) {
	if len(batchOrg.orgsId) <= 0 {
		tab.delCirclePeer(batchOrg.ownId)
		return
	}

	peerdata := make(PeerOrg)
	for _, orgid := range batchOrg.orgsId {
		peerdata[orgid] = true
	}

	// for add
	addorgs := make([]comm.Address, 0)
	// for del
	delorgs := make([]comm.Address, 0)

	tab.orgData.Lock()
	defer tab.orgData.Unlock()

	old, ok := tab.orgData.peerIDToOrgIDMap[batchOrg.ownId]
	if !ok {
		addorgs = batchOrg.orgsId
	} else {
		for _, orgid := range batchOrg.orgsId {
			if _, exit := old[orgid]; !exit {
				addorgs = append(addorgs, orgid)
			}
		}

		for orgid, _ := range old {
			if _, exit := peerdata[orgid]; !exit {
				delorgs = append(delorgs, orgid)
			}
		}
	}

	// add
	for _, orgid := range addorgs {
		circledata, circleok := tab.orgData.orgSingleData[orgid]
		if !circleok {
			circledata = make(NodeMap)
		}
		circledata[batchOrg.ownId] = true
		tab.orgData.orgSingleData[orgid] = circledata
	}

	// del
	for _, orgid := range delorgs {
		circledata, circleok := tab.orgData.orgSingleData[orgid]
		if circleok {
			delete(circledata, batchOrg.ownId)
			if len(circledata) == 0 {
				delete(tab.orgData.orgSingleData, orgid)
				if len(tab.orgData.orgSingleData) == 0 { // tab.orgData.orgSingleData
					tab.orgData.orgSingleData = make(map[comm.Address]NodeMap)
				}
			} else {
				tab.orgData.orgSingleData[orgid] = circledata
			}
		}
	}

	//
	tab.orgData.peerIDToOrgIDMap[batchOrg.ownId] = peerdata
	tab.orgData.peerIDToNode[batchOrg.ownId] = &OrgNodeInfo{
		node:   batchOrg.node,
		errnum: 0,
		peerid: batchOrg.ownId,
	}

}


func (tab *Table) setOrgs(node *Node, circleID comm.Address, ownid uint64, badd bool) {
	tab.orgData.Lock()
	defer tab.orgData.Unlock()

	if badd {
		// add orgSingleData
		circledata, circleok := tab.orgData.orgSingleData[circleID]
		if !circleok {
			circledata = make(NodeMap)
		}
		circledata[ownid] = true
		tab.orgData.orgSingleData[circleID] = circledata

		// add peerIDToOrgIDMap and peerIDToNode
		peerdata, peerok := tab.orgData.peerIDToOrgIDMap[ownid]
		if !peerok {
			peerdata = make(PeerOrg)
		}
		peerdata[circleID] = true
		tab.orgData.peerIDToOrgIDMap[ownid] = peerdata

		tab.orgData.peerIDToNode[ownid] = &OrgNodeInfo{
			node:   node,
			errnum: 0,
			peerid: ownid,
		}

		//fmt.Println(" *888888* Table setCircle add orgSingleData:",tab.orgData.orgSingleData)
		//fmt.Println(" *888888* Table setCircle add peerIDToOrgIDMap:",tab.orgData.peerIDToOrgIDMap)
		//fmt.Println(" *888888* Table setCircle add peerIDToNode:",tab.orgData.peerIDToNode)
	} else {
		// del orgSingleData
		circledata, circleok := tab.orgData.orgSingleData[circleID]
		if circleok {
			delete(circledata, ownid) //
			if len(circledata) == 0 {
				delete(tab.orgData.orgSingleData, circleID)
				if len(tab.orgData.orgSingleData) == 0 { // tab.orgData.orgSingleData为空
					tab.orgData.orgSingleData = make(map[comm.Address]NodeMap)
				}
			} else {
				tab.orgData.orgSingleData[circleID] = circledata
			}
		}

		// del peerIDToOrgIDMap and peerIDToNode
		peerdata, peerok := tab.orgData.peerIDToOrgIDMap[ownid]
		if peerok {
			delete(peerdata, circleID)
			if len(peerdata) == 0 {
				delete(tab.orgData.peerIDToOrgIDMap, ownid)
				delete(tab.orgData.peerIDToNode, ownid)
				if len(tab.orgData.peerIDToOrgIDMap) == 0 { // tab.orgData.peerIDToOrgIDMap为空
					tab.orgData.peerIDToOrgIDMap = make(map[uint64]PeerOrg)
					tab.orgData.peerIDToNode = make(map[uint64]*OrgNodeInfo)
				}
			} else {
				tab.orgData.peerIDToOrgIDMap[ownid] = peerdata
			}
		}
		//fmt.Println(" *888888* Table setCircle del orgSingleData:",tab.orgData.orgSingleData)
		//fmt.Println(" *888888* Table setCircle del peerIDToOrgIDMap:",tab.orgData.peerIDToOrgIDMap)
		//fmt.Println(" *888888* Table setCircle del peerIDToNode:",tab.orgData.peerIDToNode)
	}
}

func (tab *Table) getCircle(circleID comm.Address, num int) ([]*Node, error) {
	tab.orgData.Lock()
	defer tab.orgData.Unlock()
	if tab.orgData.orgSingleData == nil {
		return nil, nil
	}

	if tab.orgData.peerIDToNode == nil {
		return nil, nil
	}

	circledata, circleok := tab.orgData.orgSingleData[circleID]
	if !circleok {
		return nil, nil
	}

	result := make([]*Node, 0)

	i := 0
	for peerid, _ := range circledata {
		if i >= num {
			break
		}
		if orginfo, ok := tab.orgData.peerIDToNode[peerid]; ok {
			result = append(result, orginfo.node)
			i = i + 1
		}
	}

	return result, nil
}

// refresh Org
type rtnTabPeerOrg struct {
	node   *Node
	ownId  uint64
	orgsId []comm.Address
}

func (tab *Table) refreshOrg() {
	timer := time.NewTicker(autoRefreshOrgInterval)
	defer timer.Stop()
loop:
	for {
		select {
		case <-timer.C:
			tab.orgData.Lock()
			nodeInfos := make([]*OrgNodeInfo, 0)
			for pid, nodeinfo := range tab.orgData.peerIDToNode {
				nodeInfos = append(nodeInfos, &OrgNodeInfo{
					node:   nodeinfo.node,
					errnum: 0,
					peerid: pid,
				})
			}
			tab.orgData.Unlock()

			errinfo, resultall := tab.handleReqPeerOrg(nodeInfos)
			tab.calcNewOrgInfo(errinfo, resultall)
		case <-tab.closed:
			break loop
		}
	}
}

func (tab *Table) calcNewOrgInfo(errinfo []*OrgNodeInfo, orgNodeInfos []*rtnTabPeerOrg) {
	// update orgsdata
	for _, info := range orgNodeInfos {
		tab.updateOrgForBatch(info)
	}

	delpeer := make([]uint64, 0)

	// Count the number of refresh times greater than or equal to 10 times
	tab.orgData.Lock()
	for _, info := range errinfo {
		nodeInfo, ok := tab.orgData.peerIDToNode[info.peerid]
		if ok {
			if nodeInfo.errnum >= REFRESH_NUM-1 {
				delpeer = append(delpeer, info.peerid)
			} else {
				tab.orgData.peerIDToNode[info.peerid].errnum = nodeInfo.errnum + 1
			}
		}
	}
	tab.orgData.Unlock()

	// delete the 10 times err peer
	for _, id := range delpeer {
		tab.delCirclePeer(id)
	}
}

func (tab *Table) handleReqPeerOrg(nodeInfos []*OrgNodeInfo) ([]*OrgNodeInfo, []*rtnTabPeerOrg) {
	result := make([]*rtnTabPeerOrg, 0)
	i := len(nodeInfos)

	timesErr := make([]*OrgNodeInfo, 0)

	for i > 0 {
		if i >= REQPEERORGNUM {
			i = REQPEERORGNUM
		}
		errnodes, timeerr, resultsData := tab.handleSingleReqPeerOrg(nodeInfos[:i])
		nodeInfos = append(nodeInfos[i:], errnodes...)

		if len(timeerr) > 0 {
			timesErr = append(timesErr, timeerr...)
		}

		if len(resultsData) > 0 {
			result = append(result, resultsData...)
		}
		i = len(nodeInfos)
	}

	return timesErr, result
}

func (tab *Table) handleSingleReqPeerOrg(nodeInfos []*OrgNodeInfo) ([]*OrgNodeInfo, []*OrgNodeInfo, []*rtnTabPeerOrg) {
	reqErrNodes := make([]*OrgNodeInfo, 0)
	timesErr := make([]*OrgNodeInfo, 0)

	errnodes, resultsData := tab.reqPeerOrgAll(nodeInfos)
	if len(errnodes) > 0 {
		for _, temnode := range errnodes {
			if temnode.errnum < 3 {
				reqErrNodes = append(reqErrNodes, temnode)
			} else {
				timesErr = append(timesErr, temnode)
			}
		}
	}

	return reqErrNodes, timesErr, resultsData
}

func (tab *Table) reqPeerOrgAll(nodes []*OrgNodeInfo) (result []*OrgNodeInfo, rtnPeerOrg []*rtnTabPeerOrg) {
	rc := make(chan *OrgNodeInfo, len(nodes))
	rtnPeerOrgChan := make(chan *rtnTabPeerOrg, len(nodes))
	for i := range nodes {
		go func(n *OrgNodeInfo) {
			nn, rtn, _ := tab.reqPeerOrgSingle(n)
			rc <- nn
			rtnPeerOrgChan <- rtn
		}(nodes[i])
	}
	for range nodes {
		if n := <-rc; n != nil {
			n.errnum = n.errnum + 1
			result = append(result, n)
		}
		if rnt := <-rtnPeerOrgChan; rnt != nil {
			rtnPeerOrg = append(rtnPeerOrg, rnt)
		}
	}
	return result, rtnPeerOrg
}

func (tab *Table) reqPeerOrgSingle(n *OrgNodeInfo) (*OrgNodeInfo, *rtnTabPeerOrg, error) {
	result, err := tab.net.reqPeerOrg(n.node.ID, n.node.addr(), n.node)
	if err != nil {
		return n, nil, err
	}
	return nil, result, nil
}

// refresh Connect
type rtnTabPeerConnect struct {
	node      *Node
	ownId     uint64
	remotesID []uint64
}

func (tab *Table) refreshConnect() {
	timer := time.NewTicker(autoRefreshOrgInterval)
	defer timer.Stop()
loop:
	for {
		select {
		case <-timer.C:
			tab.connectData.Lock()
			nodeInfos := make([]*OrgNodeInfo, 0)
			//fmt.Println(" ********* peerIDToNode:",tab.connectData.peerIDToNode)
			for peerid, nodeinfo := range tab.connectData.peerIDToNode {
				nodeInfos = append(nodeInfos, &OrgNodeInfo{
					node:   nodeinfo.node,
					errnum: 0,
					peerid: peerid,
				})
			}
			tab.connectData.Unlock()

			errinfo, resultall := tab.handleReqPeerConnect(nodeInfos)
			tab.calcNewConnectInfo(errinfo, resultall)
		case <-tab.closed:
			break loop
		}
	}
}

func (tab *Table) calcNewConnectInfo(errinfo []*OrgNodeInfo, connectNodeInfos []*rtnTabPeerConnect) {
	// updata connect add
	for _, info := range connectNodeInfos {
		tab.updateConnectForBatch(info)
	}

	delpeer := make([]uint64, 0)
	// Count the number of refresh times greater than or equal to 10 times.
	tab.connectData.Lock()
	for _, info := range errinfo {
		nodeinfo, ok := tab.connectData.peerIDToNode[info.peerid]
		if ok {
			if nodeinfo.errnum >= REFRESH_NUM-1 {
				delpeer = append(delpeer, info.peerid)
			} else {
				tab.connectData.peerIDToNode[info.peerid].errnum = nodeinfo.errnum + 1
			}
		}
	}
	tab.connectData.Unlock()

	// delete the 10 times err peer
	for _, id := range delpeer {
		tab.delConnectPeer(id)
	}
}

func (tab *Table) handleReqPeerConnect(nodeInfos []*OrgNodeInfo) ([]*OrgNodeInfo, []*rtnTabPeerConnect) {
	result := make([]*rtnTabPeerConnect, 0)
	i := len(nodeInfos)
	timesErr := make([]*OrgNodeInfo, 0)
	for i > 0 {
		if i >= REQPEERORGNUM {
			i = REQPEERORGNUM
		}

		errnodes, timeerr, resultsData := tab.handleSingleReqPeerConnect(nodeInfos[:i])
		nodeInfos = append(nodeInfos[i:], errnodes...)
		if len(timeerr) > 0 {
			timesErr = append(timesErr, timeerr...)
		}
		//fmt.Println(" *888888* len(errnodes):",len(errnodes),",len(timeerr):",len(timeerr),",len(resultsData):",len(resultsData))
		if len(resultsData) > 0 {
			result = append(result, resultsData...)
		}
		i = len(nodeInfos)
		//fmt.Println(" *888888* i:",i)
	}

	return timesErr, result
}

func (tab *Table) handleSingleReqPeerConnect(nodeInfos []*OrgNodeInfo) ([]*OrgNodeInfo, []*OrgNodeInfo, []*rtnTabPeerConnect) {
	reqErrNodes := make([]*OrgNodeInfo, 0)
	timesErr := make([]*OrgNodeInfo, 0)

	errnodes, resultsData := tab.reqPeerConnectAll(nodeInfos)
	if len(errnodes) > 0 {
		for _, temnode := range errnodes {
			if temnode.errnum < 3 {
				reqErrNodes = append(reqErrNodes, temnode)
			} else {
				timesErr = append(timesErr, temnode)
			}
		}
	}

	return reqErrNodes, timesErr, resultsData
}

func (tab *Table) reqPeerConnectAll(nodes []*OrgNodeInfo) (result []*OrgNodeInfo, rtnPeerOrg []*rtnTabPeerConnect) {
	rc := make(chan *OrgNodeInfo, len(nodes))
	rtnPeerOrgChan := make(chan *rtnTabPeerConnect, len(nodes))
	// fmt.Println(" ***** Table bondall ...... ")
	for i := range nodes {
		go func(n *OrgNodeInfo) {
			nn, rtn, _ := tab.reqPeerConnectSingle(n)
			rc <- nn
			rtnPeerOrgChan <- rtn
		}(nodes[i])
	}
	for range nodes {
		if n := <-rc; n != nil {
			n.errnum = n.errnum + 1
			result = append(result, n)
		}
		if rnt := <-rtnPeerOrgChan; rnt != nil {
			rtnPeerOrg = append(rtnPeerOrg, rnt)
		}
	}
	return result, rtnPeerOrg
}

func (tab *Table) reqPeerConnectSingle(n *OrgNodeInfo) (*OrgNodeInfo, *rtnTabPeerConnect, error) {
	result, err := tab.net.reqPeerConnect(n.node.ID, n.node.addr(), n.node)
	//fmt.Println(" ******** Table reqPeerConnectSingle result:",result,",err:",err)
	if err != nil {
		return n, nil, err
	}
	return nil, result, nil
}

func (tab *Table) checkCirculation() {
	timer := time.NewTicker(autoRefreshOrgInterval)
	defer timer.Stop()
loop:
	for {
		select {
		case <-timer.C:
			tab.orgData.Lock()

			orgs := make([]comm.Address, 0)
			for orgid, _ := range tab.orgData.orgSingleData {
				orgs = append(orgs, orgid)
			}

			tab.orgData.Unlock()

			for _, orgid := range orgs {
				circu := tab.singleOrgCheck(orgid)
				if len(circu) > 1 {
					go tab.handlerCirculation(circu, orgid)
				}
			}

		case <-tab.closed:
			break loop
		}
	}
}

func (tab *Table) handlerCirculation(circu [][]uint64, orgid comm.Address) {
	index := 0
	for index < len(circu) {
		reserve := make([][]uint64, 0)
		reserve = append(reserve, circu[index+1:]...)
		for _, peerArry := range reserve {
			tab.produceArry(circu[index], peerArry, orgid)
		}
		index = index + 1
	}
}

func (tab *Table) produceArry(initiative, passive []uint64, orgid comm.Address) {
	resultInitiative := make([]uint64, 0)
	if len(initiative) <= CircuConnectNum {
		resultInitiative = initiative
	} else {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(initiative))
		resultInitiative = append(resultInitiative, initiative[index:]...)
		lenresult := len(resultInitiative)

		if lenresult < CircuConnectNum {
			resultInitiative = append(resultInitiative, initiative[:CircuConnectNum-lenresult]...)
		}
	}

	if len(passive) <= CircuConnectNum {
		for _, indexpeer := range resultInitiative {
			go tab.handlerSingleCirculation(indexpeer, passive, orgid)
		}
		return
	}

	for _, indexpeer := range resultInitiative {
		resultPassive := make([]uint64, 0)

		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(passive))

		resultPassive = append(resultPassive, passive[index:]...)
		lenresult := len(resultPassive)
		if lenresult < CircuConnectNum {
			resultPassive = append(resultPassive, passive[:CircuConnectNum-lenresult]...)
		}
		go tab.handlerSingleCirculation(indexpeer, resultPassive, orgid)
	}
}

// type OrgInfo struct {
// 	sync.Mutex
// 	orgSingleData   map[comm.Address]NodeMap // orgid -> NodeMap
// 	peerIDToOrgIDMap map[uint64]PeerOrg
// 	peerIDToNode map[uint64]*Node
// }
// func (t *udp) connectCircuOrg(ndoe,connectnode *Node,orgid comm.Address) error {
func (tab *Table) handlerSingleCirculation(initiative uint64, passive []uint64, orgid comm.Address) {
	initiativeNode := tab.getNodeFromOrg(initiative)
	if initiativeNode == nil {
		return
	}

	for _, peerid := range passive {
		passiveNode := tab.getNodeFromOrg(peerid)
		if passiveNode == nil {
			continue
		}
		tab.net.connectCircuOrg(initiativeNode, passiveNode, orgid)
	}
}

func (tab *Table) getNodeFromOrg(id uint64) *Node {
	tab.orgData.Lock()
	defer tab.orgData.Unlock()

	nodeInfo, ok := tab.orgData.peerIDToNode[id]
	if !ok {
		return nil
	}
	return nodeInfo.node
}

func (tab *Table) singleOrgCheck(orgid comm.Address) [][]uint64 {
	peers := make(map[uint64]bool)
	tab.orgData.Lock()
	orgpeers := tab.orgData.orgSingleData[orgid]
	for peerid, _ := range orgpeers {
		peers[peerid] = false
	}
	tab.orgData.Unlock()

	result := make([][]uint64, 0)
	for len(peers) > 0 {
		temarry, tempeer := tab.check(peers)
		if len(temarry) > 0 {
			result = append(result, temarry)
		}
		peers = tempeer
	}
	//fmt.Println(" ******* result:",result)
	return result
}

func (tab *Table) check(peers map[uint64]bool) ([]uint64, map[uint64]bool) {
	var randPeerID uint64
	bhave := false

	tab.connectData.Lock()
	for temid, _ := range peers {
		if _, ok := tab.connectData.peerToPeerMap[temid]; ok {
			randPeerID = temid
			bhave = true
			break
		}
	}
	tab.connectData.Unlock()

	if bhave == false {
		for temid, _ := range peers {
			randPeerID = temid
		}
	}

	orgs := make([]uint64, 0)
	orgs = append(orgs, randPeerID)
	delete(peers, randPeerID)

	if bhave == false {
		return orgs, peers
	}

	index := 0 //

	for index < len(orgs) {
		randPeerID = orgs[index]
		tab.connectData.Lock()
		for {
			bbreak := true

			if remotemap, ok := tab.connectData.peerToPeerMap[randPeerID]; ok {
				for remoteid, _ := range remotemap {
					if _, exit := peers[remoteid]; exit {
						orgs = append(orgs, remoteid)
						delete(peers, remoteid)
						bbreak = false
					}
				}
			}

			if bbreak {
				break
			}
			if len(peers) == 0 {
				break
			}
		}
		tab.connectData.Unlock()
		index = index + 1
		if index >= len(orgs) {
			break
		}
	}
	return orgs, peers
}

/*
func (tab *Table) circleprintf() {
	num := 0
	for num < 20 {
		// fmt.Println(" 888888888 ********** 88888888 ********* ")

		tab.circlemutex.RLock()
		for circleID,circledata := range tab.circleDta {
			if len(circledata) > 0 {
				for nodeid,_ := range circledata {
					fmt.Println(" **** circleID:",circleID,",nodeid:",nodeid)
				}
			}
		}
		tab.circlemutex.RUnlock()
		time.Sleep(15*time.Second)
		num = num + 1
	}

}
*/

type StellarInfo struct {
	sync.Mutex
	stellarmap map[uint64]*Node
}

func (tab *Table) stellarNode(node *Node, circleID comm.Address, ownid uint64, badd bool) {
	tab.stellarData.Lock()
	defer tab.stellarData.Unlock()

	//fmt.Println(" ****** Table stellarNode circleID,ownid,badd",circleID,ownid,badd)

	if tab.stellarData.stellarmap == nil {
		tab.stellarData.stellarmap = make(map[uint64]*Node)
	}

	if badd {
		tab.stellarData.stellarmap[ownid] = node
	} else {
		delete(tab.stellarData.stellarmap, ownid)
	}
}

func (tab *Table) getStellar() ([]*Node, error) {
	tab.stellarData.Lock()
	defer tab.stellarData.Unlock()

	if tab.stellarData.stellarmap == nil {
		tab.stellarData.stellarmap = make(map[uint64]*Node)
	}

	result := make([]*Node, 0)
	for _, node := range tab.stellarData.stellarmap {
		result = append(result, node)
	}
	//fmt.Println(" **** Table getStellar result:",result)
	return result, nil
}

func (tab *Table) stellarInspectDis(inspectpeer uint64) {
	//fmt.Println(" ****** Table stellarInspectDis inspectpeer:",inspectpeer)
	tab.stellarDisChan <- inspectpeer
}

func (tab *Table) refreshStellarDis() {
	type disInfo struct {
		count     int
		firsttime int64
	}
	inspectpeer := make(map[uint64]*disInfo)

	timer := time.NewTicker(10 * time.Second)
	defer timer.Stop()
loop:
	for {
		select {
		case disid := <-tab.stellarDisChan:

			tab.stellarData.Lock()

			stellarlen := len(tab.stellarData.stellarmap)
			if _, ok := tab.stellarData.stellarmap[disid]; ok {
				if info, exit := inspectpeer[disid]; exit {
					if info.count >= (stellarlen / 2) {
						//fmt.Println(" ******* info.count,stellarlen,",info.count,stellarlen)
						delete(tab.stellarData.stellarmap, disid)
					} else {
						inspectpeer[disid].count = info.count + 1
					}
				} else {
					inspectpeer[disid] = &disInfo{
						count:     1,
						firsttime: time.Now().Unix(),
					}
				}
			}
			tab.stellarData.Unlock()
		case <-timer.C:
			dispeer := make([]uint64, 0)
			for peerid, info := range inspectpeer {
				if time.Now().Unix()-info.firsttime >= 5*60 {
					dispeer = append(dispeer, peerid)
				}
			}

			for _, peerid := range dispeer {
				delete(inspectpeer, peerid)
			}

		case <-tab.closeReq:
			break loop
		}
	}
}
