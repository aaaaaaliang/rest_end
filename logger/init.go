package logger

import (
	"os"
)

var Log Logger

func InitLogger() {
	Log = New("rest-api")
	// 设置日志格式
	if os.Getenv("ENV") == "prod" {
		Log.SetFormat(FormatJSON)
	} else {
		Log.SetFormat(FormatText)
	}
	// 设置日志级别
	Log.SetLevel(LevelDebug)
	//输出到文件
	//f, err := os.OpenFile("logs/app.logger", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//if err == nil {
	//	Log.SetOutput(f)
	//}
}
