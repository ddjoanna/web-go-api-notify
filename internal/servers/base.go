package server

import (
	cus_error "notify-service/internal/errors"

	errorpb "proto/pkg/notify/v1/error"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BaseServer struct{}

func (s *BaseServer) HandleError(err error) error {

	code, msgCode := s.parseCode(err)
	st := status.New(code, err.Error())
	st, _ = st.WithDetails(&errdetails.ErrorInfo{
		Reason: errorpb.ErrorReasonCode_name[int32(msgCode)],
	})
	return st.Err()
}

func (s *BaseServer) parseCode(err error) (codes.Code, errorpb.ErrorReasonCode) {
	switch err := err.(type) {
	case *cus_error.ValueError:
		return codes.InvalidArgument, err.MsgCode

	case *cus_error.KeyError:
		return codes.InvalidArgument, err.MsgCode

	case *cus_error.DuplicateError:
		return codes.AlreadyExists, err.MsgCode

	case *cus_error.NotFoundError:
		return codes.NotFound, err.MsgCode

	case *cus_error.ServerError:
		return codes.Internal, err.MsgCode

	default:
		return codes.Internal, errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL
	}
}
