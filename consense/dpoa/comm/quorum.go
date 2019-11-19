package comm

import (
	"sync"
)

// Quorum records each acknowledgement and check for different types of quorum satisfied
type Quorum struct {
	sync.RWMutex
	total int
	size  int
	acks  map[ID]bool
	zones map[int]int
	nacks map[ID]bool
}

// NewQuorum returns a new Quorum
func NewQuorum(total int) *Quorum {
	q := &Quorum{
		total: total,
		size:  0,
		acks:  make(map[ID]bool),
		zones: make(map[int]int),
	}
	return q
}

// ACK adds id to quorum ack records
func (q *Quorum) ACK(id ID) {
	q.Lock()
	defer q.Unlock()
	if !q.acks[id] {
		q.acks[id] = true
		q.size++
		q.zones[id.Zone()]++
	}
}

// ADD increase ack size by one
func (q *Quorum) ADD() {
	q.size++
}

// Size returns current ack size
func (q *Quorum) Size() int {
	return q.size
}

// Reset resets the quorum to empty
func (q *Quorum) Reset() {
	q.size = 0
	q.acks = make(map[ID]bool)
	q.zones = make(map[int]int)
	q.nacks = make(map[ID]bool)
}

// Majority quorum satisfied
func (q *Quorum) Majority() bool {
	/*for k, _:=range q.acks{
		fmt.Println("-<<<<<<<<<<<<<<<<", q.size, k)
	}
	fmt.Println("<<<<<<<<<<<<<<<<", q.size, q.total)*/
	return q.size > q.total/2
}

// Q1 returns true if config.Quorum type is satisfied
func (q *Quorum) Q1() bool {
	return q.Majority()
}

// Q2 returns true if config.Quorum type is satisfied
func (q *Quorum) Q2() bool {
	return q.Majority()
}
