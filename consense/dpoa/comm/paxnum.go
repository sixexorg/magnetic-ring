package comm

import (
	"strconv"
	"strings"
	"fmt"

	"github.com/sixexorg/magnetic-ring/log"
)

type ID string

func NewID(zone, node int) ID {
	if zone < 0 {
		zone = -zone
	}
	if node < 0 {
		node = -node
	}

	return ID(strconv.Itoa(zone) + "." + strconv.Itoa(node))
}

func (i ID) Zone() int {
	if !strings.Contains(string(i), ".") {
		//log.Warningf("id %s does not contain \".\"\n", i)
		return 0
	}
	s := strings.Split(string(i), ".")[0]
	zone, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Error("Failed to convert Zone %s to int\n", s)
	}
	return int(zone)
}

func (i ID) Node() int {
	var s string
	if !strings.Contains(string(i), ".") {
		//log.Warningf("id %s does not contain \".\"\n", i)
		s = string(i)
	} else {
		s = strings.Split(string(i), ".")[1]
	}
	node, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Error("Failed to convert Node %s to int\n", s)
	}
	return int(node)
}

type Ballot uint64

func NewBallot(n int, id ID) Ballot {
	return Ballot(n<<32 | id.Zone()<<16 | id.Node())
}

func (b Ballot) N() int {
	return int(uint64(b) >> 32)
}

func (b Ballot) ID() ID {
	zone := int(uint32(b) >> 16)
	node := int(uint16(b))
	return NewID(zone, node)
}

func (b *Ballot) Next(id ID) {
	*b = NewBallot(b.N()+1, id)
}

func (b Ballot) String() string {
	return fmt.Sprintf("%d.%s", b.N(), b.ID())
}

func NextBallot(ballot int, id ID) int {
	n := id.Zone()<<16 | id.Node()
	return (ballot>>32+1)<<32 | n
}

func LeaderID(ballot int) ID {
	zone := uint32(ballot) >> 16
	node := uint16(ballot)
	return NewID(int(zone), int(node))
}


