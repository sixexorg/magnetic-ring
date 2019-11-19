package ast

import (
	"github.com/sixexorg/magnetic-ring/common/sink"
)

type Expr interface {
	PositionHolder
	Serialization(sk *sink.ZeroCopySink)
}

type ExprBase struct {
	Node
}

/* ConstExprs {{{ */

type ConstExpr interface {
	Expr
	constExprMarker()
}

type ConstExprBase struct {
	ExprBase
}

func (expr *ConstExprBase) constExprMarker() {}

type TrueExpr struct {
	ConstExprBase
}

type FalseExpr struct {
	ConstExprBase
}

type NilExpr struct {
	ConstExprBase
}

type NumberExpr struct {
	ConstExprBase

	Value string
}

type StringExpr struct {
	ConstExprBase

	Value string
}

/* ConstExprs }}} */

type Comma3Expr struct {
	ExprBase
}

type IdentExpr struct {
	ExprBase

	Value string
}

type AttrGetExpr struct {
	ExprBase

	Object Expr
	Key    Expr
}

type TableExpr struct {
	ExprBase

	Fields []*Field
}

type FuncCallExpr struct {
	ExprBase

	Func      Expr
	Receiver  Expr
	Method    string
	Args      []Expr
	AdjustRet bool
}

type LogicalOpExpr struct {
	ExprBase

	Operator string
	Lhs      Expr
	Rhs      Expr
}

type RelationalOpExpr struct {
	ExprBase

	Operator string
	Lhs      Expr
	Rhs      Expr
}

type StringConcatOpExpr struct {
	ExprBase

	Lhs Expr
	Rhs Expr
}

type ArithmeticOpExpr struct {
	ExprBase

	Operator string
	Lhs      Expr
	Rhs      Expr
}

type UnaryMinusOpExpr struct {
	ExprBase
	Expr Expr
}

type UnaryNotOpExpr struct {
	ExprBase
	Expr Expr
}

type UnaryLenOpExpr struct {
	ExprBase
	Expr Expr
}

type FunctionExpr struct {
	ExprBase

	ParList *ParList
	Stmts   []Stmt
}

const (
	exprTrue           byte = 0x01
	exprFalse          byte = 0x02
	exprNil            byte = 0x03
	exprNumber         byte = 0x04
	exprString         byte = 0x05
	exprComma3         byte = 0x06
	exprIdent          byte = 0x07
	exprAttrGet        byte = 0x08
	exprTable          byte = 0x09
	exprFuncCall       byte = 0x0A
	exprLogicalOp      byte = 0x0B
	exprRelationalOp   byte = 0x0C
	exprStringConcatOp byte = 0x0D
	exprArithmeticOp   byte = 0x0E
	exprUnaryMinusOp   byte = 0x0F
	exprUnaryNotOp     byte = 0x10
	exprUnaryLenOp     byte = 0x11
	exprFunction       byte = 0x12

	nilReceiver byte = 0x13
)

func (this *TrueExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprTrue)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
}
func (this *FalseExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprFalse)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
}
func (this *NilExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprNil)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
}
func (this *NumberExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprNumber)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteString(this.Value)
}
func (this *StringExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprString)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteString(this.Value)
}
func (this *Comma3Expr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprComma3)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
}
func (this *IdentExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprIdent)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteString(this.Value)
}
func (this *AttrGetExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprAttrGet)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	this.Object.Serialization(sk)
	this.Key.Serialization(sk)
}
func (this *TableExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprTable)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteVarUint(uint64(len(this.Fields)))
	for _, field := range this.Fields {
		field.Serialization(sk)
	}
}
func (this *FuncCallExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprFuncCall)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	this.Func.Serialization(sk)
	if this.Receiver == nil {
		sk.WriteByte(nilReceiver)
	} else {
		this.Receiver.Serialization(sk)
	}
	sk.WriteString(this.Method)
	sk.WriteVarUint(uint64(len(this.Args)))
	for _, arg := range this.Args {
		arg.Serialization(sk)
	}
	sk.WriteBool(this.AdjustRet)
}
func (this *LogicalOpExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprLogicalOp)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteString(this.Operator)
	this.Lhs.Serialization(sk)
	this.Rhs.Serialization(sk)
}
func (this *RelationalOpExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprRelationalOp)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteString(this.Operator)
	this.Lhs.Serialization(sk)
	this.Rhs.Serialization(sk)
}
func (this *StringConcatOpExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprStringConcatOp)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	this.Lhs.Serialization(sk)
	this.Rhs.Serialization(sk)
}
func (this *ArithmeticOpExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprArithmeticOp)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	sk.WriteString(this.Operator)
	this.Lhs.Serialization(sk)
	this.Rhs.Serialization(sk)
}
func (this *UnaryMinusOpExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprUnaryMinusOp)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	this.Expr.Serialization(sk)
}
func (this *UnaryNotOpExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprUnaryNotOp)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	this.Expr.Serialization(sk)
}
func (this *UnaryLenOpExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprUnaryLenOp)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	this.Expr.Serialization(sk)
}
func (this *FunctionExpr) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteByte(exprFunction)
	sk.WriteVarUint(uint64(this.CodeLine))
	sk.WriteVarUint(uint64(this.CodeLastline))
	this.ParList.Serialization(sk)
	sk.WriteVarUint(uint64(len(this.Stmts)))
	for _, stmt := range this.Stmts {
		stmt.Serialization(sk)
	}
}
func (this *FunctionExpr) DeSerialization(source *sink.ZeroCopySource) (err error) {
	source.NextByte()
	codeLine, _, _, _ := source.NextVarUint()
	this.SetLine(int(codeLine))
	codeLastLine, _, _, _ := source.NextVarUint()
	this.SetLastLine(int(codeLastLine))
	this.ParList = &ParList{}
	this.ParList.DeSerialization(source)
	length, _, _, _ := source.NextVarUint()
	this.Stmts = make([]Stmt, length)
	var i uint64
	for i = 0; i < length; i++ {
		this.Stmts[i], err = StmtDeserialize(source)
		if err != nil {
			return err
		}
	}
	return nil
}

func ExprDeserialize(source *sink.ZeroCopySource) (expr Expr, err error) {

	t, _ := source.NextByte()
	var i uint64
	switch t {
	case exprTrue:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		expr = &TrueExpr{}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprFalse:
		expr = &FalseExpr{}
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprNil:
		expr := &NilExpr{}
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprNumber:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		v, _, _, _ := source.NextString()
		expr = &NumberExpr{
			Value: v,
		}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprString:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		v, _, _, _ := source.NextString()
		expr = &StringExpr{
			Value: v,
		}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprComma3:
		expr = &Comma3Expr{}
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprIdent:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		v, _, _, _ := source.NextString()
		expr = &IdentExpr{
			Value: v,
		}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprAttrGet:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		object, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		key, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		expr = &AttrGetExpr{Object: object, Key: key}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprTable:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		length, _, _, _ := source.NextVarUint()
		fields := make([]*Field, length)
		i = 0
		for i = 0; i < length; i++ {
			fields[i] = &Field{}
			err = fields[i].DeSerialization(source)
			if err != nil {
				return nil, err
			}
		}
		expr = &TableExpr{Fields: fields}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprFuncCall:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		f, err := ExprDeserialize(source)
		r, err := ExprDeserialize(source)
		m, _, _, _ := source.NextString()
		length, _, _, _ := source.NextVarUint()
		args := make([]Expr, length)
		i = 0
		for i = 0; i < length; i++ {
			args[i], err = ExprDeserialize(source)
		}
		adjustRet, _, _ := source.NextBool()
		if err != nil {
			return nil, err
		}
		expr = &FuncCallExpr{
			Func:      f,
			Receiver:  r,
			Method:    m,
			Args:      args,
			AdjustRet: adjustRet,
		}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprLogicalOp:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		operator, _, _, _ := source.NextString()
		lhs, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		rhs, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		expr = &LogicalOpExpr{Operator: operator, Lhs: lhs, Rhs: rhs}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprRelationalOp:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		operator, _, _, _ := source.NextString()
		lhs, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		rhs, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		expr = &RelationalOpExpr{Operator: operator, Lhs: lhs, Rhs: rhs}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprStringConcatOp:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		lhs, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		rhs, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		expr = &StringConcatOpExpr{Lhs: lhs, Rhs: rhs}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprArithmeticOp:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		operator, _, _, _ := source.NextString()
		lhs, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		rhs, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		expr := &ArithmeticOpExpr{Operator: operator, Lhs: lhs, Rhs: rhs}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprUnaryMinusOp:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		e, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		expr := &UnaryMinusOpExpr{Expr: e}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprUnaryNotOp:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		e, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		expr = &UnaryNotOpExpr{Expr: e}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprUnaryLenOp:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		e, err := ExprDeserialize(source)
		if err != nil {
			return nil, err
		}
		expr = &UnaryLenOpExpr{Expr: e}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case exprFunction:
		codeLine, _, _, _ := source.NextVarUint()
		codeLastLine, _, _, _ := source.NextVarUint()
		parList := &ParList{}
		parList.DeSerialization(source)
		length, _, _, _ := source.NextVarUint()
		stmts := make([]Stmt, length)
		i = 0
		for i = 0; i < length; i++ {
			stmts[i], err = StmtDeserialize(source)
			if err != nil {
				return nil, err
			}
		}
		expr = &FunctionExpr{ParList: parList, Stmts: stmts}
		expr.SetLine(int(codeLine))
		expr.SetLastLine(int(codeLastLine))
	case nilReceiver:
		expr = nil
	}
	return expr, nil
}
