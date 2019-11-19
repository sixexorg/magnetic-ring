package states

import (
	"io"

	"github.com/sixexorg/magnetic-ring/common"
)

type StateKV interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
	GetKey() []byte	//
	Hash() common.Hash	//root hash
}
