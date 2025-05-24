package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"rest/config"
	"rest/response"
	"rest/utils"
	"strings"
)

type ESQueryMap map[string]interface{}

func listOrder(c *gin.Context) {
	type Req struct {
		Index   int    `form:"index" json:"index" binding:"required"`
		Size    int    `form:"size" json:"size" binding:"required"`
		Status  int    `form:"status" json:"status"`
		All     bool   `form:"all" json:"all"`
		Keyword string `form:"keyword" json:"keyword"`
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	userCode := utils.GetUser(c)
	log.Println("req", req, userCode)

	query := ESQueryMap{
		"from": (req.Index - 1) * req.Size,
		"size": req.Size,
		"query": ESQueryMap{
			"bool": ESQueryMap{
				"must":   []ESQueryMap{},
				"filter": []ESQueryMap{},
			},
		},
		"sort": []ESQueryMap{
			{"created": map[string]string{"order": "desc"}},
		},
	}

	if !req.All {
		query["query"].(ESQueryMap)["bool"].(ESQueryMap)["filter"] = append(
			query["query"].(ESQueryMap)["bool"].(ESQueryMap)["filter"].([]ESQueryMap),
			ESQueryMap{"term": ESQueryMap{"user_code": userCode}},
		)
	}

	if req.Status != 0 {
		query["query"].(ESQueryMap)["bool"].(ESQueryMap)["filter"] = append(
			query["query"].(ESQueryMap)["bool"].(ESQueryMap)["filter"].([]ESQueryMap),
			ESQueryMap{"term": ESQueryMap{"status": req.Status}},
		)
	}

	if req.Keyword != "" {
		query["query"].(ESQueryMap)["bool"].(ESQueryMap)["must"] = append(
			query["query"].(ESQueryMap)["bool"].(ESQueryMap)["must"].([]ESQueryMap),
			ESQueryMap{
				"bool": ESQueryMap{
					"should": []ESQueryMap{
						{"multi_match": ESQueryMap{
							"query":  req.Keyword,
							"fields": []string{"user_name", "remark"},
						}},
						{"match": ESQueryMap{"order_detail.product_name": req.Keyword}},
					},
				},
			},
		)
	}

	esQuery, _ := json.Marshal(query)
	esClient := config.ESClient
	res, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("orders"),
		esClient.Search.WithBody(strings.NewReader(string(esQuery))),
		esClient.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("Elasticsearch 查询失败: %v", err))
		return
	}
	defer res.Body.Close()

	var esResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esResult); err != nil {
		log.Println("解析 Elasticsearch 结果失败:", err)
		response.SuccessWithTotal(c, response.ServerError, nil, 0)
		return
	}

	hitsData, exists := esResult["hits"]
	if !exists {
		response.SuccessWithTotal(c, response.SuccessCode, []interface{}{}, 0)
		return
	}

	hitsMap, ok := hitsData.(map[string]interface{})
	if !ok {
		response.SuccessWithTotal(c, response.SuccessCode, []interface{}{}, 0)
		return
	}

	hitsArray, ok := hitsMap["hits"].([]interface{})
	if !ok {
		response.SuccessWithTotal(c, response.SuccessCode, []interface{}{}, 0)
		return
	}

	var resList []map[string]interface{}
	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}
		source, exists := hitMap["_source"]
		if !exists {
			continue
		}
		sourceMap, ok := source.(map[string]interface{})
		if !ok {
			continue
		}
		resList = append(resList, sourceMap)
	}

	totalValue := 0
	if totalMap, ok := hitsMap["total"].(map[string]interface{}); ok {
		if value, ok := totalMap["value"].(float64); ok {
			totalValue = int(value)
		}
	}

	response.SuccessWithTotal(c, response.SuccessCode, resList, totalValue)
}
