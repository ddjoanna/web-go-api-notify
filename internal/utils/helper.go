package util

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/snowflake"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// 將 protobuf Timestamp 轉換為 time.Time
func ConvertProtoTimestampToTime(protoTimestamp *timestamppb.Timestamp) (*time.Time, error) {
	if protoTimestamp == nil {
		return nil, nil
	}

	if err := protoTimestamp.CheckValid(); err != nil {
		return nil, fmt.Errorf("invalid protobuf timestamp: %w", err)
	}

	t := protoTimestamp.AsTime()
	return &t, nil
}

// 將 Snowflake ID 轉換為 time.Time
func ConvertSnowflakeToTime(snowflakeId string) (*time.Time, error) {
	// 將字串轉換成整數型 Snowflake ID
	idInt, err := strconv.ParseInt(snowflakeId, 10, 64)
	if err != nil {
		log.Fatalf("Failed to convert string to int: %v", err)
	}

	// 使用 snowflake.NewIDFromString 將 ID 轉換回 Snowflake ID 類型
	id := snowflake.ID(idInt)

	// 反查 ID 中的時間戳
	timestamp := id.Time()

	timestampTime := time.Unix(0, timestamp*int64(time.Millisecond))

	return &timestampTime, nil
}

func Md5(data string) string {
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ChunkArray 將切片分割為指定大小的子切片
func ChunkArray(data []string, chunkSize int) [][]string {
	if chunkSize <= 0 {
		return [][]string{}
	}

	result := make([][]string, 0, (len(data)+chunkSize-1)/chunkSize)
	for start := 0; start < len(data); start += chunkSize {
		end := min(start+chunkSize, len(data))
		result = append(result, data[start:end])
	}
	return result
}

// 輔助函數：取得兩數最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
