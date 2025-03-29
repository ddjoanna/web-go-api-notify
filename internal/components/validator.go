package component

import (
	"encoding/json"
	"fmt"
	shared "notify-service/internal"
	model "notify-service/internal/models"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	config   *shared.Config
	validate *validator.Validate
}

func NewValidator(config *shared.Config) *Validator {
	validate := validator.New()
	validate.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		re := regexp.MustCompile(fl.Param())
		return re.MatchString(fl.Field().String())
	})
	return &Validator{
		config:   config,
		validate: validate,
	}
}

func (v *Validator) CheckSendSmsRequest(req model.SendSmsRequest) error {
	return v.validateRequest(req)
}

func (v *Validator) CheckSendMailRequest(req model.SendMailRequest) error {
	return v.validateRequest(req)
}

func (v *Validator) CheckCancelScheduledByMessageIdRequest(req model.CancelScheduledByMessageIdRequest) error {
	return v.validateRequest(req)
}

func (v *Validator) CheckListStatusWithPagingRequest(req model.ListStatusWithPagingRequest) error {
	return v.validateRequest(req)
}

// 驗證 scheduled_at 是否有效
func (v *Validator) CheckScheduledAt(scheduledAt *time.Time) error {
	if scheduledAt == nil {
		return nil
	}

	now := time.Now()
	if !scheduledAt.After(now) {
		return fmt.Errorf("scheduled_at must be in the future")
	}

	maxAllowedTime := now.Add(time.Duration(v.config.ScheduleLimitDays) * 24 * time.Hour)
	if scheduledAt.After(maxAllowedTime) {
		return fmt.Errorf("scheduled_at must be within the next %d days", v.config.ScheduleLimitDays)
	}
	return nil
}

// 驗證請求資料的通用方法
func (v *Validator) validateRequest(data interface{}) error {
	return v.handleValidationError(v.validate.Struct(data))
}

// 處理驗證錯誤
func (v *Validator) handleValidationError(err error) error {
	if err == nil {
		return nil
	}

	// 若錯誤是 ValidationErrors 類型，收集所有錯誤訊息
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		metaData := make(map[string]string)
		for _, fieldErr := range validationErrors {
			errMsg := v.getValidationErrorMessage(fieldErr.Tag())
			metaData[fieldErr.Field()] = fmt.Sprintf("The field '%s' %s.", fieldErr.Field(), errMsg)
		}

		jsonMetaData, _ := json.Marshal(metaData)

		// 返回統一的錯誤訊息
		return fmt.Errorf("%s", string(jsonMetaData))
	}

	return err
}

// 獲取對應的錯誤訊息
func (v *Validator) getValidationErrorMessage(tag string) string {
	errorMessages := map[string]string{
		"required": "is required",
		"oneof":    "must be one of the allowed values",
		"max":      "exceeds maximum allowed value",
		"min":      "is below minimum allowed value",
		"email":    "must be a valid email address",
		"boolean":  "must be a boolean",
		"gte":      "must be greater than or equal to required value",
	}

	if errMsg, exists := errorMessages[tag]; exists {
		return errMsg
	}

	// 預設錯誤訊息
	return "has an invalid value"
}
