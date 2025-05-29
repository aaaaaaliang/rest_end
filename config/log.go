package config

import (
	"os"
	xlog "rest/utils"
)

var Log xlog.Logger

func InitLogger() {
	Log = xlog.New("rest-api")
	// 设置日志格式
	if os.Getenv("ENV") == "prod" {
		Log.SetFormat(xlog.FormatJSON)
	} else {
		Log.SetFormat(xlog.FormatText)
	}
	// 设置日志级别
	Log.SetLevel(xlog.LevelDebug)

	//输出到文件
	//f, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//if err == nil {
	//	Log.SetOutput(f)
	//}
}
