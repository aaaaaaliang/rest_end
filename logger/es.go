package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"time"
)

type LogEntry struct {
	Service     string                 `json:"service"`
	Level       string                 `json:"level"`
	Time        string                 `json:"time"`
	Message     string                 `json:"message"`
	Type        string                 `json:"type,omitempty"` // ✅ 新增字段
	TraceID     string                 `json:"trace_id,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	LogCategory string                 `json:"log_category,omitempty"` // ✅ 新增字段：日志分类

}

func SendLogToESCtx(ctx context.Context, level, logType, logCategory, msg string, fields map[string]interface{}) {
	traceID := TraceIDFromContext(ctx)

	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["trace_id"] = traceID

	if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
		fields["user_code"] = ginCtx.GetString("user")
		fields["user_name"] = ginCtx.GetString("user_name")
		fields["user_role"] = ginCtx.GetString("user_role")
	}

	entry := LogEntry{
		Service:     "rest-api",
		Level:       level,
		Time:        time.Now().Format(time.RFC3339),
		Message:     msg,
		Type:        logType,
		LogCategory: logCategory, // ✅ 新增字段赋值
		TraceID:     traceID,
		Fields:      fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("❌ 日志序列化失败: %v\n", err)
		return
	}

	res, err := config.ESClient.Index("system-logs", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("❌ 写入 Elasticsearch 失败: %v\n", err)
		return
	}
	defer res.Body.Close()
}
