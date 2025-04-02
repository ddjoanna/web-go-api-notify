package smser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-ini/ini"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	shared "notify-service/internal"
)

const (
	CallbackURLForMitake = ""
	MitakeAPIDomain      = "https://smsapi.mitake.com.tw"
	SendBatchPath        = "/api/mtk/SmBulkSend"
)

type MitakeSmser struct {
	config *shared.Config
	resty  *resty.Client
}

func NewMitakeSmser(config *shared.Config, resty *resty.Client) *MitakeSmser {
	return &MitakeSmser{config: config, resty: resty}
}

func (s *MitakeSmser) SendBatchSms(ctx context.Context, request SmsBatchRequest) SmsBatchResponse {
	queryParams := map[string]string{
		"username":        s.config.MitakeUserName,
		"password":        s.config.MitakePassword,
		"Encoding_PostIn": "UTF-8",
		"objectID":        request.Message.MessageId,
	}

	var payload strings.Builder
	for _, receiver := range request.Receivers {
		payload.WriteString(formatPayloadString(receiver, request.Message.Message))
	}

	url := MitakeAPIDomain + SendBatchPath

	resp, err := s.resty.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParams(queryParams).
		SetBody(payload.String()).
		Post(url)

	if err != nil {
		log.WithContext(ctx).Error("Request failed", "error", err.Error())
		return newFailedResponse(ctx, request, "REQUEST_FAILED")
	}

	if resp.IsError() {
		log.WithContext(ctx).Error("Unexpected status code", "status_code", resp.StatusCode(), "body", resp.String())
		return newFailedResponse(ctx, request, "UNEXPECTED_STATUS_CODE")
	}

	providerResponses, err := parseMitakeResponse(resp.String())
	if err != nil {
		log.WithContext(ctx).Error("Failed to parse response body", err.Error())
		return newFailedResponse(ctx, request, "RESPONSE_PARSE_ERROR")
	}

	smsResponses := convertToSmsResponse(providerResponses)
	return newResponse(request, smsResponses)
}

func newFailedResponse(ctx context.Context, request SmsBatchRequest, errorCode string) SmsBatchResponse {
	log.WithContext(ctx).Errorf("failed to send sms, error: %s", errorCode)
	return SmsBatchResponse{
		Status:      string(MitakeStatus_FAILED),
		MessageId:   request.Message.MessageId,
		SmsResponse: nil,
	}
}

func newResponse(request SmsBatchRequest, smsResponses []SmsResponse) SmsBatchResponse {
	status := string(MitakeStatus_FAILED)
	for _, smsResponse := range smsResponses {
		if smsResponse.TraceId == "DEFAULT" {
			status = smsResponse.Status
			break
		}
	}
	return SmsBatchResponse{
		Status:      status,
		MessageId:   request.Message.MessageId,
		SmsResponse: smsResponses,
	}
}

func formatPayloadString(receiver SmsReceiver, message string) string {
	return fmt.Sprintf(
		"%s$$%s$$%s$$%s$$%s$$%s$$%s\r\n",
		receiver.TargetId,
		receiver.Receiver,
		"",
		"",
		"",
		CallbackURLForMitake,
		message,
	)
}

type MitakeResponse struct {
	ProviderTraceId string `json:"provider_trace_id"`
	StatusCode      string `json:"status_code"`
	AccountPoint    string `json:"account_point"`
}

func parseMitakeResponse(content string) (map[string]MitakeResponse, error) {
	cfg, err := ini.Load([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("failed to load ini content: %v", err)
	}

	result := make(map[string]MitakeResponse)
	for _, section := range cfg.Sections() {
		var msgData MitakeResponse
		targetId := section.Name()
		msgData.ProviderTraceId = section.Key("msgid").String()
		msgData.StatusCode = section.Key("statuscode").String()
		msgData.AccountPoint = section.Key("AccountPoint").String()
		result[targetId] = msgData
	}
	return result, nil
}

func convertToSmsResponse(providerResponses map[string]MitakeResponse) []SmsResponse {
	smsResponse := make([]SmsResponse, 0, len(providerResponses))
	for targetId, response := range providerResponses {
		providerResponse := map[string]interface{}{
			"ProviderTraceId": response.ProviderTraceId,
			"StatusCode":      response.StatusCode,
			"StatusMessage":   MitakeCodeReason[response.StatusCode],
			"AccountPoint":    response.AccountPoint,
		}
		providerResponseJson, err := json.Marshal(providerResponse)
		if err != nil {
			log.Error("Failed to marshal provider response", err.Error())
			continue
		}
		mistakeStatus := MitakeCodeStatus[response.StatusCode]
		smsResponse = append(smsResponse, SmsResponse{
			Status:           string(mistakeStatus),
			TraceId:          targetId,
			ProviderResponse: string(providerResponseJson),
		})
	}
	return smsResponse
}

var MitakeCodeReason = map[string]string{
	"*": "系統發生錯誤，請聯絡三竹資訊窗口人員",
	"a": "簡訊發送功能暫時停止服務，請稍候再試",
	"b": "簡訊發送功能暫時停止服務，請稍候再試",
	"c": "請輸入帳號",
	"d": "請輸入密碼",
	"e": "帳號、密碼錯誤",
	"f": "帳號已過期",
	"h": "帳號已被停用",
	"k": "無效的連線位址",
	"m": "必須變更密碼，在變更密碼前，無法使用簡訊發送服務",
	"n": "密碼已逾期，在變更密碼前，將無法使用簡訊發送服務",
	"p": "沒有權限使用外部Http程式",
	"r": "系統暫停服務，請稍後再試",
	"s": "帳務處理失敗，無法發送簡訊",
	"t": "簡訊已過期",
	"u": "簡訊內容不得為空白",
	"v": "無效的手機號碼",
	"0": "預約傳送中",
	"1": "已送達業者",
	"2": "已送達業者",
	"3": "已送達業者",
	"4": "已送達手機",
	"5": "內容有錯誤",
	"6": "門號有錯誤",
	"7": "簡訊已停用",
	"8": "逾時無送達",
	"9": "預約已取消",
}

var MitakeCodeStatus = map[string]MitakeStatus{
	"*": MitakeStatus_FAILED,
	"a": MitakeStatus_FAILED,
	"b": MitakeStatus_FAILED,
	"c": MitakeStatus_FAILED,
	"d": MitakeStatus_FAILED,
	"e": MitakeStatus_FAILED,
	"f": MitakeStatus_FAILED,
	"h": MitakeStatus_FAILED,
	"k": MitakeStatus_FAILED,
	"m": MitakeStatus_FAILED,
	"n": MitakeStatus_FAILED,
	"p": MitakeStatus_FAILED,
	"r": MitakeStatus_FAILED,
	"s": MitakeStatus_FAILED,
	"t": MitakeStatus_FAILED,
	"u": MitakeStatus_FAILED,
	"v": MitakeStatus_FAILED,
	"0": MitakeStatus_SENT,
	"1": MitakeStatus_SENT,
	"2": MitakeStatus_SENT,
	"3": MitakeStatus_SENT,
	"4": MitakeStatus_SENT,
	"5": MitakeStatus_FAILED,
	"6": MitakeStatus_FAILED,
	"7": MitakeStatus_FAILED,
	"8": MitakeStatus_FAILED,
	"9": MitakeStatus_FAILED,
}
