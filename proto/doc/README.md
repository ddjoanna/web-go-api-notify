# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [notify/notify.proto](#notify_notify-proto)
    - [CancelScheduledByMessageIdRequest](#notify-v1-CancelScheduledByMessageIdRequest)
    - [ListStatusWithPagingRequest](#notify-v1-ListStatusWithPagingRequest)
    - [ListStatusWithPagingResponse](#notify-v1-ListStatusWithPagingResponse)
    - [Mail](#notify-v1-Mail)
    - [PageRequest](#notify-v1-PageRequest)
    - [Paging](#notify-v1-Paging)
    - [SendMailRequest](#notify-v1-SendMailRequest)
    - [SendMailResponse](#notify-v1-SendMailResponse)
    - [SendSmsRequest](#notify-v1-SendSmsRequest)
    - [SendSmsResponse](#notify-v1-SendSmsResponse)
    - [Sms](#notify-v1-Sms)
    - [Target](#notify-v1-Target)
  
    - [MessageType](#notify-v1-MessageType)
  
    - [NotifyService](#notify-v1-NotifyService)
  
- [error/error.proto](#error_error-proto)
    - [ErrorReasonCode](#notify-v1-error-ErrorReasonCode)
  
- [Scalar Value Types](#scalar-value-types)



<a name="notify_notify-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## notify/notify.proto



<a name="notify-v1-CancelScheduledByMessageIdRequest"></a>

### CancelScheduledByMessageIdRequest
取消預約訊息請求


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message_id | [string](#string) |  | 訊息 ID |






<a name="notify-v1-ListStatusWithPagingRequest"></a>

### ListStatusWithPagingRequest
查詢發送狀態請求（支持分頁）


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message_type | [MessageType](#notify-v1-MessageType) |  | 訊息類型（SMS 或 MAIL） |
| message_id | [string](#string) |  | 訊息 ID |
| receiver | [string](#string) |  | 收件者手機號碼或電子郵件地址 |
| page | [PageRequest](#notify-v1-PageRequest) |  | 分頁請求資訊 |
| start_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | 起始時間（限制90天內） |
| end_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | 結束時間（限制90天內） |






<a name="notify-v1-ListStatusWithPagingResponse"></a>

### ListStatusWithPagingResponse
查詢發送狀態響應


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| target | [Target](#notify-v1-Target) | repeated | 訊息發送紀錄 |
| paging | [Paging](#notify-v1-Paging) |  | 分頁資訊 |






<a name="notify-v1-Mail"></a>

### Mail
郵件


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sender_address | [google.protobuf.StringValue](#google-protobuf-StringValue) |  | 寄件者電子郵件地址 |
| sender_name | [google.protobuf.StringValue](#google-protobuf-StringValue) |  | 寄件者名稱 |
| subject | [string](#string) |  | 郵件主旨 |
| body | [string](#string) |  | 郵件內容 |






<a name="notify-v1-PageRequest"></a>

### PageRequest
分頁請求資訊


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [int32](#int32) |  | 當前分頁索引 |
| size | [int32](#int32) |  | 每頁筆數 |
| sort_field | [string](#string) |  | 排序欄位 |
| sort_order | [string](#string) |  | 排序方向（ASC 或 DESC） |






<a name="notify-v1-Paging"></a>

### Paging
分頁資訊響應結構


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [int32](#int32) |  |  |
| size | [int32](#int32) |  |  |
| total | [int32](#int32) |  |  |
| sort_field | [string](#string) |  | 排序欄位 |
| sort_order | [string](#string) |  | 排序方向（ASC 或 DESC） |






<a name="notify-v1-SendMailRequest"></a>

### SendMailRequest
發送郵件請求


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mail | [Mail](#notify-v1-Mail) |  | 郵件資訊 |
| receivers | [string](#string) | repeated | 收件者電子郵件地址 |
| scheduled_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | 預約時間（限制30天內，可選填） |






<a name="notify-v1-SendMailResponse"></a>

### SendMailResponse
發送郵件響應


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message_id | [string](#string) |  | 訊息 ID |






<a name="notify-v1-SendSmsRequest"></a>

### SendSmsRequest
發送簡訊請求


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sms | [Sms](#notify-v1-Sms) |  | 簡訊 |
| receivers | [string](#string) | repeated | 收件者手機號碼 |
| scheduled_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | 預約時間（限制30天內，可選填） |






<a name="notify-v1-SendSmsResponse"></a>

### SendSmsResponse
發送簡訊響應


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message_id | [string](#string) |  | 訊息 ID |






<a name="notify-v1-Sms"></a>

### Sms
簡訊


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| body | [string](#string) |  | 簡訊內容 |






<a name="notify-v1-Target"></a>

### Target
發送記錄資訊


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message_type | [string](#string) |  | 訊息類型（SMS 或 MAIL） |
| message_id | [string](#string) |  | 訊息 ID |
| message_content | [string](#string) |  | 訊息內容（文字 或 HTML） |
| receiver | [string](#string) |  | 收件者（手機號碼或電子郵件地址） |
| status | [string](#string) |  | 寄送狀態 |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | 建立時間 |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | 更新時間 |





 


<a name="notify-v1-MessageType"></a>

### MessageType
訊息類型枚舉

| Name | Number | Description |
| ---- | ------ | ----------- |
| MESSAGE_TYPE_UNSPECIFIED | 0 |  |
| SMS | 1 |  |
| MAIL | 2 |  |


 

 


<a name="notify-v1-NotifyService"></a>

### NotifyService
通知服務

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SendSms | [SendSmsRequest](#notify-v1-SendSmsRequest) | [SendSmsResponse](#notify-v1-SendSmsResponse) | 非流式 RPC：單筆或少量簡訊發送 |
| SendBatchSms | [SendSmsRequest](#notify-v1-SendSmsRequest) stream | [SendSmsResponse](#notify-v1-SendSmsResponse) stream | 流式 RPC：批量簡訊發送 |
| SendMail | [SendMailRequest](#notify-v1-SendMailRequest) | [SendMailResponse](#notify-v1-SendMailResponse) | 非流式 RPC：單筆或少量郵件發送 |
| SendBatchMail | [SendMailRequest](#notify-v1-SendMailRequest) stream | [SendMailResponse](#notify-v1-SendMailResponse) stream | 流式 RPC：批量郵件發送 |
| CancelScheduledByMessageId | [CancelScheduledByMessageIdRequest](#notify-v1-CancelScheduledByMessageIdRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | 取消預約訊息 |
| ListStatusWithPaging | [ListStatusWithPagingRequest](#notify-v1-ListStatusWithPagingRequest) | [ListStatusWithPagingResponse](#notify-v1-ListStatusWithPagingResponse) | 查詢發送狀態 |

 



<a name="error_error-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## error/error.proto


 


<a name="notify-v1-error-ErrorReasonCode"></a>

### ErrorReasonCode


| Name | Number | Description |
| ---- | ------ | ----------- |
| ERR_COMMON_INTERNAL | 0 | 通用錯誤原因代碼: 0 ~ 999 當不需要太細緻的錯誤原因時，可以直接使用這些代碼 |
| ERR_COMMON_INVALID_ARGUMENT | 1 |  |
| ERR_NOTIFY_INVALID_RECEIVER | 1000 |  |
| ERR_NOTIFY_RECEIVER_EMPTY | 1001 |  |
| ERR_NOTIFY_SUBJECT_EMPTY | 1002 |  |
| ERR_NOTIFY_BODY_EMPTY | 1003 |  |
| ERR_NOTIFY_MESSAGE_NOT_FOUND | 1004 |  |
| ERR_NOTIFY_MESSAGE_IS_ENQUEUE_CANNOT_CANCEL | 1005 |  |


 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

