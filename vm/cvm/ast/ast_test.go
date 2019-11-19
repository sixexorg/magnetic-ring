package ast_test

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/vm/cvm/ast"
	"github.com/sixexorg/magnetic-ring/vm/cvm/parse"
	"os"
	"testing"
)

func init() {

}

func TestStmtsSerialize(t *testing.T) {
	fp, err := os.Open("ast_test.lua")
	if err != nil {
		fmt.Printf("open file failed. casue: %v\n", err)
		return
	}
	defer fp.Close()

	chunk, err := parse.Parse(fp, "")
	if err != nil {
		fmt.Printf("Abstract syntax tree: %v\n", err)
		return
	}

	data := ast.Serialize(chunk)
	if err != nil {
		fmt.Printf("contrace serializion failed. cause: %v\n", err)
		return
	}

	dChunk := ast.Deserizlize(data)

	for index, value := range dChunk {
		fmt.Printf("%#v\n", chunk[index])
		fmt.Printf("%#v\n\n", value)
	}
	assert.Equal(t, dChunk, chunk)
}
