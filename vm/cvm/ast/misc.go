package ast

import "github.com/sixexorg/magnetic-ring/common/sink"

type Field struct {
	Key   Expr
	Value Expr
}

type ParList struct {
	HasVargs bool
	Names    []string
}

type FuncName struct {
	Func     Expr
	Receiver Expr
	Method   string
}

func (this *Field) Serialization(sk *sink.ZeroCopySink) {
	this.Key.Serialization(sk)
	this.Value.Serialization(sk)
}
func (this *Field) DeSerialization(source *sink.ZeroCopySource) (err error) {
	this.Key, err = ExprDeserialize(source)
	this.Value, err = ExprDeserialize(source)
	return err
}

func (this *ParList) Serialization(sk *sink.ZeroCopySink) {
	sk.WriteBool(this.HasVargs)
	sk.WriteVarUint(uint64(len(this.Names)))
	for _, name := range this.Names {
		sk.WriteString(name)
	}
}
func (this *ParList) DeSerialization(source *sink.ZeroCopySource) (err error) {
	this.HasVargs, _, _ = source.NextBool()
	len, _, _, _ := source.NextVarUint()
	this.Names = make([]string, len)
	var i uint64
	for i = 0; i < len; i++ {
		this.Names[i], _, _, _ = source.NextString()
	}
	return err
}

func (this *FuncName) Serialization(sk *sink.ZeroCopySink) {
	this.Func.Serialization(sk)
	if this.Receiver == nil {
		sk.WriteByte(nilReceiver)
	} else {
		this.Receiver.Serialization(sk)
	}
	sk.WriteString(this.Method)
}

func (this *FuncName) DeSerialization(source *sink.ZeroCopySource) (err error) {
	this.Func, err = ExprDeserialize(source)
	this.Receiver, err = ExprDeserialize(source)
	this.Method, _, _, _ = source.NextString()
	return err
}
