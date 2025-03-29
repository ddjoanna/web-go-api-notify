package service

import (
	cus_error "notify-service/internal/errors"
	errorpb "proto/pkg/notify/v1/error"
)

type BaseService struct{}

// 通用錯誤處理
func (s *BaseService) newError(msg string, msgCode errorpb.ErrorReasonCode) *cus_error.BaseError {
	return &cus_error.BaseError{Msg: msg, MsgCode: msgCode}
}

func (s *BaseService) ValueError(msg string, msgCode errorpb.ErrorReasonCode) *cus_error.ValueError {
	return &cus_error.ValueError{BaseError: *s.newError(msg, msgCode)}
}

func (s *BaseService) KeyError(msg string, msgCode errorpb.ErrorReasonCode) *cus_error.KeyError {
	return &cus_error.KeyError{BaseError: *s.newError(msg, msgCode)}
}

func (s *BaseService) DuplicateError(msg string, msgCode errorpb.ErrorReasonCode) *cus_error.DuplicateError {
	return &cus_error.DuplicateError{BaseError: *s.newError(msg, msgCode)}
}

func (s *BaseService) NotFoundError(msg string, msgCode errorpb.ErrorReasonCode) *cus_error.NotFoundError {
	return &cus_error.NotFoundError{BaseError: *s.newError(msg, msgCode)}
}

func (s *BaseService) ServerError(msg string, msgCode errorpb.ErrorReasonCode) *cus_error.ServerError {
	return &cus_error.ServerError{BaseError: *s.newError(msg, msgCode)}
}
