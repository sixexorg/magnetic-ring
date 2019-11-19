package mainchain

import (
	"time"

	"bytes"
	"sort"
	"sync"

	"fmt"

	"github.com/ahmetb/go-linq"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/node"
)

var nodeLHC *nodeLHCache
var m sync.RWMutex

type nodeLHCache struct {
	//nodeId leagueId height
	nodeLH   map[string]map[common.Address]uint64
	star     []string
	lastTime time.Time
	nodeRoot common.Hash
	useable  bool

	m sync.RWMutex
}

func NewNodeLHCache() {
	nodeLHC = &nodeLHCache{
		nodeLH:   make(map[string]map[common.Address]uint64),
		lastTime: time.Now(),
		useable:  false,
	}
}

func (this *nodeLHCache) refreshStarNode(stars []string) {
	this.m.RLock()
	defer this.m.RUnlock()
	this.star = stars
}
func GetNodeLHCacheInstance() *nodeLHCache {
	return nodeLHC
}

func (this *nodeLHCache) GetNodeLeagueHeight(leagueId common.Address) map[string]uint64 {
	this.m.RLock()
	defer this.m.RUnlock()
	nodeHMap := make(map[string]uint64, len(this.nodeLH))
	for k, v := range this.nodeLH {
		nodeHMap[k] = v[leagueId]
	}
	return nodeHMap
}
func (this *nodeLHCache) checkNodeRoot(nodeIds []string) {
	m.Lock()
	defer m.Unlock()
	sort.Strings(nodeIds)
	buff := bytes.NewBuffer(nil)
	for _, v := range nodeIds {
		buff.WriteString(v)
	}
	hash := common.CalcHash(buff.Bytes())
	if bytes.Equal(this.nodeRoot.ToBytes(), hash.ToBytes()) {
		return
	}
	this.nodeRoot = hash
	this.useable = false
	earth := node.CurEarth()
	nodeIds = append(nodeIds, earth)
	existNodes := make([]string, 0, len(nodeIds))
	for k, _ := range this.nodeLH {
		needRemove := true
		for _, id := range nodeIds {
			if k == id {
				existNodes = append(existNodes, k)
				needRemove = false
				break
			}
		}
		if needRemove {
			delete(this.nodeLH, k)
		}
	}
	for _, v := range nodeIds {
		if !linq.From(existNodes).Contains(v) {
			this.nodeLH[v] = make(map[common.Address]uint64)
		}
	}
}

func (this *nodeLHCache) receiveNodeLH(nlh *common.NodeLH) {
	m.Lock()
	defer m.Unlock()
	tmp := this.nodeLH[nlh.NodeId]
	if tmp == nil {
		return
	} else if tmp != nil && tmp[nlh.LeagueId] >= nlh.Height {
		return
	}
	this.nodeLH[nlh.NodeId][nlh.LeagueId] = nlh.Height
	this.lastTime = time.Now()
	fmt.Printf("🚫 🆙  receiveNodeLH end nodeId:%s leagueId:%s height:%d\n", nlh.NodeId, nlh.LeagueId.ToString(), nlh.Height)
}

//getNodeLHs is the league processing progress in the stars and earth
func (this *nodeLHCache) getNodeLHs() (earth map[common.Address]uint64, nodes []map[common.Address]uint64) {
	m.RLock()
	defer m.RUnlock()
	m := make([]map[common.Address]uint64, 0, len(this.nodeLH)-1)
	earthId := node.CurEarth()
	earth = this.nodeLH[earthId]
	for _, v := range this.star {
		tmp := this.nodeLH[v]
		m = append(m, tmp)
	}
	/*for k, v := range this.nodeLH {
		tmp := v
		if k == earthId {
			earth = tmp
		} else {
			m = append(m, tmp)
		}
	}*/
	return earth, m
}
