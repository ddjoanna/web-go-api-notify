package error

import errorpb "proto/pkg/notify/v1/error"

type BaseError struct {
	MsgCode errorpb.ErrorReasonCode
	Msg     string
}

func (e *BaseError) Error() string {
	return e.Msg
}

type ValueError struct {
	BaseError
}

type KeyError struct {
	BaseError
}

type DuplicateError struct {
	BaseError
}

type NotFoundError struct {
	BaseError
}

type ServerError struct {
	BaseError
}
