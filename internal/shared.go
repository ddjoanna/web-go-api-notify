package shared

const (
	LOG_FORMAT_JSON = "json"
	LOG_FORMAT_TEXT = "text"
)

const (
	KafkaTopicSms      = "notify-sms"
	KafkaTopicMail     = "notify-mail"
	KafkaGroupIdNotify = "notify"
)

type Config struct {
	GrpcPort                      int
	PostgresHost                  string
	PostgresPort                  int
	PostgresUser                  string
	PostgresPassword              string
	PostgresDb                    string
	PostgresSchema                string
	DbMaxIdleConns                int
	DbMaxOpenConns                int
	OtlpEndpoint                  string
	OtlpServiceName               string
	LogFormat                     string
	ScheduleLimitDays             int
	AESKey                        string
	SmsProvider                   string
	SmsProviderBatchLimit         int
	MailProvider                  string
	MailProviderBatchLimit        int
	KafkaBrokers                  string
	KafkaVersion                  string
	KafkaConsumerGroupInstanceNum int
	JobName                       string
	SendgridToken                 string
	MitakeUserName                string
	MitakePassword                string
}
