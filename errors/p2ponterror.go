package errors

type p2pError struct {
	errmsg    string
	callstack *CallStack
	root      error
	code      ErrCode
}

func (e p2pError) Error() string {
	return e.errmsg
}

func (e p2pError) GetErrCode() ErrCode {
	return e.code
}

func (e p2pError) GetRoot() error {
	return e.root
}

func (e p2pError) GetCallStack() *CallStack {
	return e.callstack
}
