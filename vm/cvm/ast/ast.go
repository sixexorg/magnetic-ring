package ast

import (
	"github.com/sixexorg/magnetic-ring/common/sink"
)

type PositionHolder interface {
	Line() int
	SetLine(int)
	LastLine() int
	SetLastLine(int)
}

type Node struct {
	CodeLine     int
	CodeLastline int
}

func (n *Node) Line() int {
	return n.CodeLine
}

func (n *Node) SetLine(line int) {
	n.CodeLine = line
}

func (n *Node) LastLine() int {
	return n.CodeLastline
}

func (n *Node) SetLastLine(line int) {
	n.CodeLastline = line
}

func Serialize(chunk []Stmt) []byte {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteVarUint(uint64(len(chunk)))
	for _, stmt := range chunk {
		stmt.Serialization(sk)
	}
	return sk.Bytes()
}
func Deserizlize(data []byte) []Stmt {

	source := sink.NewZeroCopySource(data)
	length, _, _, _ := source.NextVarUint()
	chunk := make([]Stmt, length)
	var i uint64
	for i = 0; i < length; i++ {
		chunk[i], _ = StmtDeserialize(source)
	}
	return chunk
}
