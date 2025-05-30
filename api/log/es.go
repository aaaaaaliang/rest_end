package log

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"rest/config"
	"rest/response"
	"rest/utils"
)

func searchLog(c *gin.Context) {
	type LogSearchReq struct {
		UserCode    string `form:"user_code"`
		UserName    string `form:"user_name"`
		UserRole    string `form:"user_role"`
		Type        string `form:"type"`
		Level       string `form:"level"`
		Message     string `form:"message"`
		StartTime   string `form:"start_time"`
		EndTime     string `form:"end_time"`
		Page        int    `form:"page,default=1"`
		Size        int    `form:"size,default=10"`
		LogCategory string `form:"log_category"`
	}
	var req LogSearchReq
	if !utils.ValidationQuery(c, &req) {
		return
	}

	// 构造 DSL 查询体
	must := []map[string]interface{}{}

	if req.UserCode != "" {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"fields.user_code": req.UserCode,
			},
		})
	}
	if req.UserName != "" {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"fields.user_name": req.UserName,
			},
		})
	}

	if req.LogCategory != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"log_category": req.LogCategory,
			},
		})
	}

	if req.UserRole != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"fields.user_role": req.UserRole,
			},
		})
	}
	if req.Type != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"type": req.Type,
			},
		})
	}
	if req.Level != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"level": req.Level,
			},
		})
	}
	if req.Message != "" {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"message": req.Message,
			},
		})
	}

	// 时间范围过滤
	if req.StartTime != "" || req.EndTime != "" {
		rangeFilter := map[string]interface{}{
			"time": map[string]string{},
		}
		if req.StartTime != "" {
			rangeFilter["time"].(map[string]string)["gte"] = req.StartTime
		}
		if req.EndTime != "" {
			rangeFilter["time"].(map[string]string)["lte"] = req.EndTime
		}
		must = append(must, map[string]interface{}{
			"range": rangeFilter,
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
		"from": (req.Page - 1) * req.Size,
		"size": req.Size,
		"sort": []map[string]interface{}{
			{"time": map[string]string{"order": "desc"}},
		},
	}

	queryBytes, _ := json.Marshal(query)
	res, err := config.ESClient.Search(
		config.ESClient.Search.WithIndex("system-logs"),
		config.ESClient.Search.WithBody(bytes.NewReader(queryBytes)),
		config.ESClient.Search.WithContext(context.Background()),
	)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var parsed map[string]interface{}
	_ = json.Unmarshal(body, &parsed)

	response.SuccessWithData(c, response.SuccessCode, parsed)
}
