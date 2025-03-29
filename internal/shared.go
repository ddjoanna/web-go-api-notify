package shared

const (
	LOG_FORMAT_JSON = "json"
	LOG_FORMAT_TEXT = "text"
)

type Config struct {
	GrpcPort               int
	PostgresHost           string
	PostgresPort           int
	PostgresUser           string
	PostgresPassword       string
	PostgresDb             string
	PostgresSchema         string
	DbMaxIdleConns         int
	DbMaxOpenConns         int
	OtlpEndpoint           string
	OtlpServiceName        string
	LogFormat              string
	ScheduleLimitDays      int
	AESKey                 string
	SmsProvider            string
	SmsProviderBatchLimit  int
	SmsProviderToken       string
	MailProvider           string
	MailProviderBatchLimit int
	MailProviderToken      string
}
