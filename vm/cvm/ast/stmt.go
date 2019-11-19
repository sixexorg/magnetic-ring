package ast

import (
	"github.com/sixexorg/magnetic-ring/common/sink"
)

const (
	stmtAssign      byte = 0x01
	stmtLocalAssign byte = 0x02
	stmtFuncCall    byte = 0x03
	stmtDoBlock     byte = 0x04
	stmtWhile       byte = 0x05
	stmtRepeat      byte = 0x06
	stmtIf          byte = 0x07
	stmtNumberFor   byte = 0x08
	stmtGenericFor  byte = 0x09
	stmtFuncDef     byte = 0x0A
	stmtReturn      byte = 0x0B
	stmtBreak       byte = 0x0C
)

type Stmt interface {
	PositionHolder
	Serialization(sk *sink.ZeroCopySink)
}

type StmtBase struct {
	Node
}

type AssignStmt struct {
	StmtBase

	Lhs []Expr
	Rhs []Expr
}

type LocalAssignStmt struct {
	StmtBase

	Names []string
	Exprs []Expr
}

type FuncCallStmt struct {
	StmtBase

	Expr Expr
}

type DoBlockStmt struct {
	StmtBase

	Stmts []Stmt
}

type WhileStmt struct {
	StmtBase

	Condition Expr
	Stmts     []Stmt
}

type RepeatStmt struct {
	StmtBase

	Condition Expr
	Stmts     []Stmt
}

type IfStmt struct {
	StmtBase

	Condition Expr
	Then      []Stmt
	Else      []Stmt
}

type NumberForStmt struct {
	StmtBase

	Name  string
	Init  Expr
	Limit Expr
	Step  Expr
	Stmts []Stmt
}

type GenericForStmt struct {
	StmtBase

	Names []string
	Exprs []Expr
	Stmts []Stmt
}

type FuncDefStmt struct {
	StmtBase

	Name *FuncName
	Func *FunctionExpr
}

type ReturnStmt struct {
	StmtBase

	Exprs []Expr
}

type BreakStmt struct {
	StmtBase
}

func (this *AssignStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtAssign)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteVarUint(uint64(len(this.Lhs)))
	for _, expr := range this.Lhs {
		expr.Serialization(sk)
	}
	sk.WriteVarUint(uint64(len(this.Rhs)))
	for _, expr := range this.Rhs {
		expr.Serialization(sk)
	}
}
func (this *LocalAssignStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtLocalAssign)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteVarUint(uint64(len(this.Names)))
	for _, name := range this.Names {
		sk.WriteString(name)
	}
	sk.WriteVarUint(uint64(len(this.Exprs)))
	for _, expr := range this.Exprs {
		expr.Serialization(sk)
	}
}
func (this *FuncCallStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtFuncCall)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	this.Expr.Serialization(sk)
}
func (this *DoBlockStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtDoBlock)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteVarUint(uint64(len(this.Stmts)))
	for _, stmt := range this.Stmts {
		stmt.Serialization(sk)
	}
}
func (this *WhileStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtWhile)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	this.Condition.Serialization(sk)
	sk.WriteVarUint(uint64(len(this.Stmts)))
	for _, stmt := range this.Stmts {
		stmt.Serialization(sk)
	}
}
func (this *RepeatStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtRepeat)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	// Condition
	this.Condition.Serialization(sk)
	// Stmts
	sk.WriteVarUint(uint64(len(this.Stmts)))
	for _, stmt := range this.Stmts {
		stmt.Serialization(sk)
	}
}
func (this *IfStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtIf)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	// Condition
	this.Condition.Serialization(sk)
	// Then
	sk.WriteVarUint(uint64(len(this.Then)))
	for _, stmt := range this.Then {
		stmt.Serialization(sk)
	}
	// Else
	sk.WriteVarUint(uint64(len(this.Else)))
	for _, stmt := range this.Else {
		stmt.Serialization(sk)
	}
}
func (this *NumberForStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtNumberFor)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	// Name
	sk.WriteString(this.Name)
	// Init
	this.Init.Serialization(sk)
	// Limit
	this.Limit.Serialization(sk)
	// Step
	this.Step.Serialization(sk)
	// Stmts
	for _, stmt := range this.Stmts {
		stmt.Serialization(sk)
	}
}
func (this *GenericForStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtGenericFor)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	// Names
	sk.WriteVarUint(uint64(len(this.Names)))
	for _, name := range this.Names {
		sk.WriteString(name)
	}
	// Exprs
	sk.WriteVarUint(uint64(len(this.Exprs)))
	for _, expr := range this.Exprs {
		expr.Serialization(sk)
	}
	// Stmts
	sk.WriteVarUint(uint64(len(this.Stmts)))
	for _, stmt := range this.Stmts {
		stmt.Serialization(sk)
	}
}
func (this *FuncDefStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtFuncDef)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	// Name
	this.Name.Serialization(sk)
	// Func
	this.Func.Serialization(sk)
}
func (this *ReturnStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtReturn)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteVarUint(uint64(len(this.Exprs)))
	for _, expr := range this.Exprs {
		expr.Serialization(sk)
	}
}
func (this *BreakStmt) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(stmtBreak)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
}

func StmtDeserialize(source *sink.ZeroCopySource) (stmt Stmt, err error) {

	t, _ := source.NextByte()
	codeLine, _, _, _ := source.NextVarUint()
	codeLastLine, _, _, _ := source.NextVarUint()
	var i uint64
	switch t {

	case stmtAssign:
		len, _, _, _ := source.NextVarUint()
		lhs := make([]Expr, len)
		for i = 0; i < len; i++ {
			lhs[i], err = ExprDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		len, _, _, _ = source.NextVarUint()
		rhs := make([]Expr, len)
		for i = 0; i < len; i++ {
			rhs[i], err = ExprDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		stmt = &AssignStmt{Lhs: lhs, Rhs: rhs}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtLocalAssign:
		len, _, _, _ := source.NextVarUint()
		names := make([]string, len)
		for i = 0; i < len; i++ {
			names[i], _, _, _ = source.NextString()
			if err != nil {
				return nil, err
			}
		}
		len, _, _, _ = source.NextVarUint()
		exprs := make([]Expr, len)
		for i = 0; i < len; i++ {
			s, err := ExprDeserialize(source)
			// exprs[i], err = ExprDeserialize(source)
			exprs[i] = s
			if err != nil {
				return nil, err
			}
		}
		stmt = &LocalAssignStmt{Names: names, Exprs: exprs}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtFuncCall:
		expr, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		stmt = &FuncCallStmt{Expr: expr}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtDoBlock:
		len, _, _, _ := source.NextVarUint()
		stmts := make([]Stmt, len)
		for i = 0; i < len; i++ {
			stmts[i], err = StmtDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		stmt = &DoBlockStmt{Stmts: stmts}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtWhile:
		condition, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		len, _, _, _ := source.NextVarUint()
		stmts := make([]Stmt, len)
		for i = 0; i < len; i++ {
			stmts[i], err = StmtDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		stmt = &WhileStmt{Condition: condition, Stmts: stmts}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtRepeat:
		condition, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		len, _, _, _ := source.NextVarUint()
		stmts := make([]Stmt, len)
		for i = 0; i < len; i++ {
			stmts[i], err = StmtDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		stmt = &RepeatStmt{Condition: condition, Stmts: stmts}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtIf:
		condition, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		len, _, _, _ := source.NextVarUint()
		then := make([]Stmt, len)
		i = 0
		for i = 0; i < len; i++ {
			then[i], err = StmtDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		len, _, _, _ = source.NextVarUint()
		else_ := make([]Stmt, len)
		i = 0
		for i = 0; i < len; i++ {
			else_[i], err = StmtDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		stmt = &IfStmt{Condition: condition, Then: then, Else: else_}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtNumberFor:
		name, _, _, _ := source.NextString()
		init, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		limit, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		step, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		len, _, _, _ := source.NextVarUint()
		stmts := make([]Stmt, len)
		for i = 0; i < len; i++ {
			stmts[i], err = StmtDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		stmt = &NumberForStmt{Name: name, Init: init, Limit: limit, Step: step, Stmts: stmts}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtGenericFor:
		len, _, _, _ := source.NextVarUint()
		names := make([]string, len)
		for i = 0; i < len; i++ {
			names[i], _, _, _ = source.NextString()
		}
		len, _, _, _ = source.NextVarUint()
		exprs := make([]Expr, len)
		for i = 0; i < len; i++ {
			exprs[i], err = ExprDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		len, _, _, _ = source.NextVarUint()
		stmts := make([]Stmt, len)
		for i = 0; i < len; i++ {
			stmts[i], err = StmtDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		stmt = &GenericForStmt{Names: names, Exprs: exprs, Stmts: stmts}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtFuncDef:
		name := &FuncName{}
		name.DeSerialization(source)

		funcExpr := &FunctionExpr{}
		funcExpr.DeSerialization(source)
		if err != nil {
			return nil, err
		}
		stmt = &FuncDefStmt{Name: name, Func: funcExpr}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtReturn:
		len, _, _, _ := source.NextVarUint()
		exprs := make([]Expr, len)
		for i = 0; i < len; i++ {
			exprs[i], err = ExprDeserialize(source)
		}
		stmt = &ReturnStmt{Exprs: exprs}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	case stmtBreak:
		stmt = &BreakStmt{}
		stmt.SetLine(int(codeLine))
		stmt.SetLastLine(int(codeLastLine))
	}
	return stmt, err
}
