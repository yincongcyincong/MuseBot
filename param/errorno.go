package param

type Error interface {
	error
	Errno() uint32
}

func New(code uint32, text string) Error {
	return &errorErrno{code, text}
}

type errorErrno struct {
	c uint32
	s string
}

func (e *errorErrno) Error() string {
	return e.s
}

func (e *errorErrno) Errno() uint32 {
	return e.c
}

const (
	CodeSuccess = 0
	CodeUnknown = 1
	
	CodeParamError = 210009
	
	CodeCallServiceFail  = 110003
	CodeMethodNotFound   = 210003
	CodeCallUserFuncFail = 210004
	CodeDBWriteFail      = 210005
	CodeDBQueryFail      = 210006
)

const (
	MsgSuccess = "success"
	MsgUnknown = "err unknown"
	
	MsgParamError       = "param error"
	MsgCallServiceFail  = "call service fail"
	MsgDBWriteFail      = "db write fail"
	MsgMethodNotFound   = "method not found"
	MsgCallUserFuncFail = "call user func fail"
	MsgDBQueryFail      = "db query fail"
)

var (
	ErrSuccess = New(CodeSuccess, MsgSuccess)
	ErrUnknown = New(CodeUnknown, MsgUnknown)
	
	ErrParamError = New(CodeParamError, MsgParamError)
	
	ErrMethodNotFound   = New(CodeMethodNotFound, MsgMethodNotFound)
	ErrCallServiceFail  = New(CodeCallServiceFail, MsgCallServiceFail)
	ErrCallUserFuncFail = New(CodeCallUserFuncFail, MsgCallUserFuncFail)
	ErrDBQueryFail      = New(CodeDBQueryFail, MsgDBQueryFail)
	ErrDBWriteFail      = New(CodeDBWriteFail, MsgDBWriteFail)
)
