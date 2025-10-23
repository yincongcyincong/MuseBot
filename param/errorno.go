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
	
	CodeServerFail       = 110003
	CodeMethodNotFound   = 210003
	CodeCallUserFuncFail = 210004
	CodeDBWriteFail      = 210005
	CodeDBQueryFail      = 210006
	CodeLoginFail        = 210007
	CodeNotLogin         = 210008
	CodeConfigError      = 210010
	
	CodeTxtFileOnly = 300000
)

const (
	MsgSuccess = "success"
	MsgUnknown = "err unknown"
	
	MsgParamError       = "param error"
	MsgServerFail       = "request server fail"
	MsgDBWriteFail      = "db write fail"
	MsgMethodNotFound   = "method not found"
	MsgCallUserFuncFail = "call user func fail"
	MsgDBQueryFail      = "db query fail"
	MsgLoginFail        = "login fail"
	MsgNotLogin         = "not login"
	MsgConfigError      = "config error"
	
	MsgTxtFileOnly = "only support txt file"
)

var (
	ErrSuccess = New(CodeSuccess, MsgSuccess)
	ErrUnknown = New(CodeUnknown, MsgUnknown)
	
	ErrParamError = New(CodeParamError, MsgParamError)
	
	ErrMethodNotFound   = New(CodeMethodNotFound, MsgMethodNotFound)
	ErrServerFail       = New(CodeServerFail, MsgServerFail)
	ErrCallUserFuncFail = New(CodeCallUserFuncFail, MsgCallUserFuncFail)
	ErrDBQueryFail      = New(CodeDBQueryFail, MsgDBQueryFail)
	ErrDBWriteFail      = New(CodeDBWriteFail, MsgDBWriteFail)
)
