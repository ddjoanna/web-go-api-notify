package server

import (
	"context"
	"io"
	shared "notify-service/internal"
	component "notify-service/internal/components"
	model "notify-service/internal/models"
	service "notify-service/internal/services"
	util "notify-service/internal/utils"
	notifypb "proto/pkg/notify/v1/notify"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	DEFAULT_MAIL_SENDER_ADDRESS = "notify@notify.com"
	DEFAULT_MAIL_SENDER_NAME    = "Notify"
)

type NotifyServer struct {
	notifypb.UnimplementedNotifyServiceServer
	BaseServer
	validator     *component.Validator
	notifyService *service.NotifyService
	config        *shared.Config
	aesGcm        *component.AesGcm
}

func NewNotifyServer(
	validator *component.Validator,
	notifyService *service.NotifyService,
	config *shared.Config,
	aesGcm *component.AesGcm,
) *NotifyServer {
	return &NotifyServer{
		validator:     validator,
		notifyService: notifyService,
		config:        config,
		aesGcm:        aesGcm,
	}
}

func (s NotifyServer) Register(server *grpc.Server) {
	notifypb.RegisterNotifyServiceServer(server, s)
}

func (s NotifyServer) SendSms(ctx context.Context, in *notifypb.SendSmsRequest) (*notifypb.SendSmsResponse, error) {
	scheduledAt, err := util.ConvertProtoTimestampToTime(in.ScheduledAt)
	if err != nil {
		return nil, s.HandleError(err)
	}

	if err := s.validator.CheckScheduledAt(scheduledAt); err != nil {
		return nil, s.HandleError(err)
	}

	request := model.SendSmsRequest{
		Sms: model.Sms{
			Body: in.Sms.Body,
		},
		Receivers:   in.Receivers,
		ScheduledAt: scheduledAt,
	}

	if err := s.validator.CheckSendSmsRequest(request); err != nil {
		return nil, s.HandleError(err)
	}

	message, err := s.notifyService.PublishSmsMessage(ctx, request)
	if err != nil {
		return nil, s.HandleError(err)
	}
	return &notifypb.SendSmsResponse{
		MessageId: message.Id,
	}, nil
}

func (s NotifyServer) SendBatchSms(stream notifypb.NotifyService_SendBatchSmsServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return s.HandleError(err)
		}

		scheduledAt, err := util.ConvertProtoTimestampToTime(in.ScheduledAt)
		if err != nil {
			return s.HandleError(err)
		}

		if err := s.validator.CheckScheduledAt(scheduledAt); err != nil {
			return s.HandleError(err)
		}

		request := model.SendSmsRequest{
			Sms: model.Sms{
				Body: in.Sms.Body,
			},
			Receivers:   in.Receivers,
			ScheduledAt: scheduledAt,
		}

		if err := s.validator.CheckSendSmsRequest(request); err != nil {
			return s.HandleError(err)
		}

		message, err := s.notifyService.PublishSmsMessage(context.Background(), request)
		if err != nil {
			return s.HandleError(err)
		}

		if err := stream.Send(&notifypb.SendSmsResponse{
			MessageId: message.Id,
		}); err != nil {
			return s.HandleError(err)
		}
	}

	return nil
}

func (s NotifyServer) SendMail(ctx context.Context, in *notifypb.SendMailRequest) (*notifypb.SendMailResponse, error) {
	scheduledAt, err := util.ConvertProtoTimestampToTime(in.ScheduledAt)
	if err != nil {
		return nil, s.HandleError(err)
	}

	if err := s.validator.CheckScheduledAt(scheduledAt); err != nil {
		return nil, s.HandleError(err)
	}

	senderName := in.Mail.SenderName.GetValue()
	// 默認為發件人名稱
	if senderName == "" {
		senderName = DEFAULT_MAIL_SENDER_NAME
	}

	senderAddress := in.Mail.SenderAddress.GetValue()
	// 默認為發件人電子郵件地址
	if senderAddress == "" {
		senderAddress = DEFAULT_MAIL_SENDER_ADDRESS
	}

	request := model.SendMailRequest{
		Mail: model.Mail{
			SenderName:    senderName,
			SenderAddress: senderAddress,
			Subject:       in.Mail.Subject,
			Body:          in.Mail.Body,
		},
		Receivers:   in.Receivers,
		ScheduledAt: scheduledAt,
	}

	if err := s.validator.CheckSendMailRequest(request); err != nil {
		return nil, s.HandleError(err)
	}

	message, err := s.notifyService.PublishMailMessage(ctx, request)
	if err != nil {
		return nil, s.HandleError(err)
	}
	return &notifypb.SendMailResponse{
		MessageId: message.Id,
	}, nil
}

func (s NotifyServer) SendBatchMail(stream notifypb.NotifyService_SendBatchMailServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return s.HandleError(err)
		}

		scheduledAt, err := util.ConvertProtoTimestampToTime(in.ScheduledAt)
		if err != nil {
			return s.HandleError(err)
		}

		if err := s.validator.CheckScheduledAt(scheduledAt); err != nil {
			return s.HandleError(err)
		}

		senderName := in.Mail.SenderName.GetValue()
		// 默認為發件人名稱
		if senderName == "" {
			senderName = DEFAULT_MAIL_SENDER_NAME
		}

		senderAddress := in.Mail.SenderAddress.GetValue()
		// 默認為發件人電子郵件地址
		if senderAddress == "" {
			senderAddress = DEFAULT_MAIL_SENDER_ADDRESS
		}

		request := model.SendMailRequest{
			Mail: model.Mail{
				SenderName:    senderName,
				SenderAddress: senderAddress,
				Subject:       in.Mail.Subject,
				Body:          in.Mail.Body,
			},
			Receivers:   in.Receivers,
			ScheduledAt: scheduledAt,
		}

		if err := s.validator.CheckSendMailRequest(request); err != nil {
			return s.HandleError(err)
		}

		message, err := s.notifyService.PublishMailMessage(context.Background(), request)
		if err != nil {
			return s.HandleError(err)
		}

		if err := stream.Send(&notifypb.SendMailResponse{
			MessageId: message.Id,
		}); err != nil {
			return s.HandleError(err)
		}
	}

	return nil
}

func (s NotifyServer) CancelScheduledByMessageId(ctx context.Context, in *notifypb.CancelScheduledByMessageIdRequest) (*emptypb.Empty, error) {
	request := model.CancelScheduledByMessageIdRequest{
		MessageId: in.MessageId,
	}

	if err := s.validator.CheckCancelScheduledByMessageIdRequest(request); err != nil {
		return nil, s.HandleError(err)
	}

	err := s.notifyService.CancelScheduledByMessageId(ctx, in.MessageId)
	if err != nil {
		return nil, s.HandleError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s NotifyServer) ListStatusWithPaging(ctx context.Context, in *notifypb.ListStatusWithPagingRequest) (*notifypb.ListStatusWithPagingResponse, error) {
	startAt, err := util.ConvertProtoTimestampToTime(in.StartAt)
	if err != nil {
		return nil, s.HandleError(err)
	}
	endAt, err := util.ConvertProtoTimestampToTime(in.EndAt)
	if err != nil {
		return nil, s.HandleError(err)
	}

	request := model.ListStatusWithPagingRequest{
		MessageType: string(model.ConvertMessageTypeWithProto[in.MessageType]),
		MessageId:   in.GetMessageId(),
		Receiver:    in.GetReceiver(),
		Page:        in.Page,
		StartAt:     startAt,
		EndAt:       endAt,
	}

	if err := s.validator.CheckListStatusWithPagingRequest(request); err != nil {
		return nil, s.HandleError(err)
	}

	targets, total, err := s.notifyService.ListStatusWithPaging(ctx, request)
	if err != nil {
		return nil, s.HandleError(err)
	}

	response := make([]*notifypb.Target, len(targets))
	for i, target := range targets {
		receiver, err := s.aesGcm.AesDecrypt(target.Receiver)
		if err != nil {
			return nil, s.HandleError(err)
		}
		response[i] = &notifypb.Target{
			MessageType:    string(target.Message.Type),
			MessageId:      target.MessageId,
			MessageContent: target.Message.Data,
			Receiver:       receiver,
			Status:         string(target.Status),
			CreatedAt:      timestamppb.New(target.CreatedAt),
			UpdatedAt:      timestamppb.New(target.CreatedAt),
		}
	}
	return &notifypb.ListStatusWithPagingResponse{
		Target: response,
		Paging: &notifypb.Paging{
			Index:     in.Page.Index,
			Size:      in.Page.Size,
			Total:     int32(total),
			SortField: in.Page.SortField,
			SortOrder: in.Page.SortOrder,
		},
	}, nil
}
