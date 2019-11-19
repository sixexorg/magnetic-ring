package errors

import (
	"errors"
)

const callStackDepth = 10

type DetailError interface {
	error
	ErrCoder
	CallStacker
	GetRoot() error
}

func NewErr(errmsg string) error {
	return errors.New(errmsg)
}

func NewDetailErr(err error, errcode ErrCode, errmsg string) DetailError {
	if err == nil {
		return nil
	}

	p2perr, ok := err.(p2pError)
	if !ok {
		p2perr.root = err
		p2perr.errmsg = err.Error()
		p2perr.callstack = getCallStack(0, callStackDepth)
		p2perr.code = errcode

	}
	if errmsg != "" {
		p2perr.errmsg = errmsg + ": " + p2perr.errmsg
	}

	return p2perr
}

func RootErr(err error) error {
	if err, ok := err.(DetailError); ok {
		return err.GetRoot()
	}
	return err
}
