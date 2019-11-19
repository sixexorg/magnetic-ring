package main

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/sixexorg/magnetic-ring/vm/cvm"
	"github.com/sixexorg/magnetic-ring/vm/cvm/ast"
	"github.com/sixexorg/magnetic-ring/vm/cvm/parse"
	"github.com/urfave/cli"
	"io"
	"os"
	"path"
)

const (
	contractDir = "./contract"
)

func init() {
	gob.Register(ast.ExprBase{})
	gob.Register(ast.LocalAssignStmt{})
	gob.Register(ast.AssignStmt{})
	gob.Register(ast.TableExpr{})
	gob.Register(ast.FuncCallStmt{})
	gob.Register(ast.NumberExpr{})
	gob.Register(ast.StringExpr{})
	gob.Register(ast.NilExpr{})
	gob.Register(ast.IdentExpr{})
	gob.Register(ast.AttrGetExpr{})
	gob.Register(ast.FuncCallExpr{})
	gob.Register(ast.ParList{})
	gob.Register(ast.ArithmeticOpExpr{})
	gob.Register(ast.FunctionExpr{})
	gob.Register(ast.ReturnStmt{})
	gob.Register(ast.LogicalOpExpr{})
	gob.Register(ast.IfStmt{})
	gob.Register(ast.GenericForStmt{})
	gob.Register(ast.NumberForStmt{})
	gob.Register(ast.Comma3Expr{})
	gob.Register(ast.TrueExpr{})
	gob.Register(ast.FalseExpr{})
	gob.Register(ast.WhileStmt{})
	gob.Register(ast.UnaryLenOpExpr{})
	gob.Register(ast.UnaryMinusOpExpr{})
	gob.Register(ast.UnaryNotOpExpr{})
	gob.Register(ast.BreakStmt{})
	gob.Register(ast.RelationalOpExpr{})
	gob.Register(ast.FuncDefStmt{})
	gob.Register(ast.StringConcatOpExpr{})
}

func main() {
	app := newApp()
	if err := app.Run(os.Args); err != nil {
		//logger.Error("start app failed. %v", err)
	}

	// L := cvm.NewState()
	// defer L.Close()
	// if err := L.DoFile("vm/cvm/cmd/test.lua"); err != nil {
	// 	panic(err)
	// }

	return
}

func address() string {
	var buf [32]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func deploy(ctx *cli.Context) (err error) {
	filePath := ctx.Args().Get(0)
	fp, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file failed. casue: %v\n", err)
		return err
	}
	defer fp.Close()

	chunk, err := parse.Parse(fp, "")
	if err != nil {
		fmt.Printf("Abstract syntax tree: %v\n", err)
		return
	}

	for _, stmt := range chunk {
		if callStmt, ok := stmt.(*ast.FuncCallStmt); ok {
			expr := callStmt.Expr.(*ast.FuncCallExpr)
			if expr.Func.(*ast.IdentExpr).Value == "require" {
				continue
			}
			fmt.Printf("function call is forbiden. start line: %d, end line: %d\n", callStmt.Line(), callStmt.LastLine())
			return err
		}
	}
	proto, err := cvm.Compile(chunk, "")
	if err != nil {
		fmt.Printf("compile failed. cause: %v\n", err)
		return err
	}
	vm := cvm.NewState()
	lfunc := vm.NewFunctionFromProto(proto)
	vm.Push(lfunc)
	vm.Call(0, 1)
	ret := vm.Get(vm.GetTop())
	vm.Pop(vm.GetTop())
	vm.SetGlobal("contract", ret)

	fcStmt := &ast.FuncCallStmt{
		Expr: &ast.FuncCallExpr{
			Func: &ast.IdentExpr{
				Value: "init",
			},
		},
	}

	internalProto, err := cvm.Compile([]ast.Stmt{fcStmt}, "internal")
	if err != nil {
		fmt.Printf("Built-in initialization error %v\n", err)
		return err
	}

	lfunc = vm.NewFunctionFromProto(internalProto)
	vm.Push(lfunc)
	if err = vm.PCall(0, cvm.MultRet, nil); err != nil {
		fmt.Printf("Failed to execute script %v\n", err)
		return err
	}

	if _, err := os.Stat(contractDir); os.IsNotExist(err) {
		if err = os.Mkdir(contractDir, 0755); err != nil {
			os.Exit(1)
		}
	}

	data := ast.Serialize(chunk)
	if err != nil {
		fmt.Printf("contrace serializion failed. cause: %v\n", err)
		return nil
	}
	address := address()
	fpParsed, err := os.Create(path.Join(contractDir, address))
	if err != nil {
		fmt.Printf("create file failed. cause: %v\n", err)
		return nil
	}
	defer fpParsed.Close()

	_, err = fpParsed.Write(data)
	if err != nil {
		fmt.Printf("Save error: %v\n", err)
		return err
	}

	// After the parsing and saving, you still need to initialize the work.
	// Initialization should be done before saving to a file. Make sure the initialization is working.
	// The contract is successfully deployed after the operation is successful.

	fmt.Printf("Successful contract deployment: %s\n", address)
	return nil
}

func invoke(ctx *cli.Context) (err error) {
	filePath := ctx.Args().Get(0)
	funcName := ctx.Args().Get(1)
	fpParsed, err := os.Open(path.Join(contractDir, filePath))
	if err != nil {
		fmt.Printf("create file failed. cause: %v\n", err)
		return nil
	}
	defer fpParsed.Close()

	chunk2 := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := fpParsed.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if 0 == n {
			break
		}
		chunk2 = append(chunk2, buf[:n]...)
	}

	chunk := ast.Deserizlize(chunk2)
	if err != nil {
		fmt.Printf("deserialize chunk failed. cause: %v\n", err)
		return err
	}
	fmt.Printf("Reload %v\n", chunk)

	proto, err := cvm.Compile(chunk, filePath)
	if err != nil {
		fmt.Printf("compile failed. cause: %v\n", err)
		return err
	}
	vm := cvm.NewState()
	lfunc := vm.NewFunctionFromProto(proto)
	vm.Push(lfunc)
	vm.Call(0, 1)
	ret := vm.Get(vm.GetTop())
	vm.Pop(vm.GetTop())
	vm.SetGlobal("contract", ret)

	fcStmt := &ast.FuncCallStmt{
		Expr: &ast.FuncCallExpr{
			Func: &ast.IdentExpr{
				Value: funcName,
			},
		},
	}

	internalProto, err := cvm.Compile([]ast.Stmt{fcStmt}, "internal")
	if err != nil {
		fmt.Printf("Built-in initialization error %v\n", err)
		return err
	}

	lfunc = vm.NewFunctionFromProto(internalProto)
	vm.Push(lfunc)
	if err = vm.PCall(0, cvm.MultRet, nil); err != nil {
		fmt.Printf("Failed to execute script %v\n", err)
		return err
	}

	return nil
}

func check(ctx *cli.Context) (err error) {
	filePath := ctx.Args().Get(0)
	fp, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file failed when check your contract. casue: %v\n", err)
		return err
	}
	defer fp.Close()

	chunk, err := parse.Parse(fp, "")
	if err != nil {
		fmt.Printf("Abstract syntax tree failed: %v\n", err)
		return
	}

	defInit := false
	for _, stmt := range chunk {
		if callStmt, ok := stmt.(*ast.FuncCallStmt); ok {
			expr := callStmt.Expr.(*ast.FuncCallExpr)
			if expr.Func.(*ast.IdentExpr).Value == "require" {
				continue
			}
			fmt.Printf("function call is forbiden. start line: %d, end line: %d\n", callStmt.Line(), callStmt.LastLine())
			return err
		}
		if funcDef, ok := stmt.(*ast.FuncDefStmt); ok {
			if expr, ok := funcDef.Name.Func.(*ast.IdentExpr); ok {
				if expr.Value == "init" {
					defInit = true
				}
			}
		}
	}
	if !defInit {
		fmt.Printf("need define function `init`")
		return err
	}
	return nil
}
func newApp() *cli.App {
	app := cli.NewApp()
	//app.Action = run

	app.Version = "0.0.1"
	app.Name = "cvm"
	app.Usage = "command line interface"
	app.Author = "crystal.org"
	app.Copyright = "Copyright 2017-2019 The crystal.la Authors"
	app.Email = "develop@crystal.la"
	app.Description = "LUA-based virtual machine"

	app.Commands = []cli.Command{
		{
			Name:   "check",
			Action: check,
			Usage:  "Check a stmart contract before deploy it",
		},
		{
			Name:   "deploy",
			Action: deploy,
			Usage:  "Deploy a stmart contract",
		},
		{
			Name:   "invoke",
			Action: invoke,
			Usage:  "Invoke a stmart contract after deploy it",
		},
	}
	return app
}
