package mailer

import (
	"context"
	"net/http"

	shared "notify-service/internal"

	"github.com/sendgrid/sendgrid-go"
	helper "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	config *shared.Config
}

func NewSendGridMailer(
	config *shared.Config,
) *SendGridMailer {
	return &SendGridMailer{
		config: config,
	}
}

func (s *SendGridMailer) SendEmail(ctx context.Context, request MailRequest) MailResponse {
	from := helper.NewEmail(request.Message.SenderName, request.Message.SenderAddress)

	message := helper.NewSingleEmail(from, request.Message.Subject, nil, request.Message.Body, request.Message.Body)

	for _, receiver := range request.Receivers {
		toEmail := helper.NewEmail("", receiver.Email)
		personalization := helper.NewPersonalization()
		personalization.AddTos(toEmail)
		message.AddPersonalizations(personalization)
	}

	client := sendgrid.NewSendClient(s.config.SendgridToken)
	response, err := client.Send(message)

	if err != nil {
		return newMailResponse(SendgidStatus_FAILED, "", err.Error())
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		traceId := getTraceId(response.Headers)
		return newMailResponse(SendgidStatus_SENT, traceId, response.Body)
	}

	return newMailResponse(SendgidStatus_FAILED, "", response.Body)
}

func newMailResponse(status SendgidStatus, traceId string, ProviderResponse string) MailResponse {
	return MailResponse{
		Status:           string(status),
		TraceId:          traceId,
		ProviderResponse: ProviderResponse,
	}
}

func getTraceId(headers http.Header) string {
	return headers.Get("X-Message-Id")
}
