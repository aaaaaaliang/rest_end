package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"rest/config"
	"rest/response"
	"rest/utils"
)

type ESQueryMap map[string]interface{}

func listOrder(c *gin.Context) {
	// 1️⃣ 解析请求参数
	type Req struct {
		Index   int    `form:"index" json:"index" binding:"required"` // 当前页码
		Size    int    `form:"size" json:"size" binding:"required"`   // 每页条数
		Status  int    `form:"status" json:"status"`                  // 订单状态
		All     bool   `form:"all" json:"all"`                        // 是否查询所有用户
		Keyword string `form:"keyword" json:"keyword"`                // 关键字搜索
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	userCode := utils.GetUser(c)
	log.Println("req", req, userCode)
	// 2️⃣ 构造 Elasticsearch 查询
	query := ESQueryMap{
		"from": (req.Index - 1) * req.Size, // 偏移量
		"size": req.Size,                   // 每页条数
		"query": ESQueryMap{
			"bool": ESQueryMap{
				"must":   []ESQueryMap{}, // 精确匹配
				"filter": []ESQueryMap{}, // 过滤条件
			},
		},
		"sort": []ESQueryMap{
			{"created": map[string]string{"order": "desc"}}, // 按时间降序
		},
	}

	if !req.All {
		query["query"].(ESQueryMap)["bool"].(ESQueryMap)["filter"] = append(
			query["query"].(ESQueryMap)["bool"].(ESQueryMap)["filter"].([]ESQueryMap),
			ESQueryMap{"term": map[string]string{"user_code.keyword": userCode}},
		)
	}

	// 4️⃣ 过滤 `status`
	if req.Status != 0 {
		query["query"].(ESQueryMap)["bool"].(ESQueryMap)["filter"] = append(
			query["query"].(ESQueryMap)["bool"].(ESQueryMap)["filter"].([]ESQueryMap),
			ESQueryMap{"term": map[string]int{"status": req.Status}}, // 精确匹配
		)
	}

	// 5️⃣ **全文搜索 `keyword`**
	if req.Keyword != "" {
		query["query"].(ESQueryMap)["bool"].(ESQueryMap)["must"] = append(
			query["query"].(ESQueryMap)["bool"].(ESQueryMap)["must"].([]ESQueryMap),
			ESQueryMap{
				"bool": ESQueryMap{
					"should": []ESQueryMap{
						// 搜索 user_name 和 remark   multi_match只支持普通字段
						{
							"multi_match": ESQueryMap{
								"query":  req.Keyword,
								"fields": []string{"user_name", "remark"},
							},
						},
						{
							"match": map[string]string{"order_detail.product_name": req.Keyword},
						},
					},
				},
			},
		)
	}

	// 6️⃣ 发送查询请求
	queryJSON, _ := json.Marshal(query)
	//esURL := "http://localhost:9200/orders/_search"
	esURL := fmt.Sprintf("%v/orders/_search", config.G.ES.Url)

	reqES, _ := http.NewRequest("POST", esURL, bytes.NewBuffer(queryJSON))
	reqES.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(reqES)
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("查询 Elasticsearch 失败: %v", err))
		return
	}
	defer resp.Body.Close()

	// 7️⃣ 解析 Elasticsearch 返回的数据
	var esResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&esResult); err != nil {
		log.Println("解析 Elasticsearch 结果失败:", err)
		response.SuccessWithTotal(c, response.ServerError, nil, 0)
		return
	}

	log.Println("*******************", esResult)

	// 确保 `hits` 结构存在
	hitsData, exists := esResult["hits"]
	if !exists {
		log.Println("Elasticsearch 返回数据异常：hits 不存在")
		response.SuccessWithTotal(c, response.SuccessCode, []interface{}{}, 0)
		return
	}

	hitsMap, ok := hitsData.(map[string]interface{})
	if !ok {
		log.Println("Elasticsearch 返回数据 hits 解析失败")
		response.SuccessWithTotal(c, response.SuccessCode, []interface{}{}, 0)
		return
	}

	// 获取 `hits.hits`
	hitsArray, ok := hitsMap["hits"].([]interface{})
	if !ok {
		log.Println("Elasticsearch 返回数据 hits.hits 解析失败")
		response.SuccessWithTotal(c, response.SuccessCode, []interface{}{}, 0)
		return
	}

	// 解析 `_source`
	var res []map[string]interface{}
	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			log.Println("Elasticsearch hit 解析失败")
			continue
		}

		source, exists := hitMap["_source"]
		if !exists {
			log.Println("Elasticsearch hit 缺少 _source")
			continue
		}

		sourceMap, ok := source.(map[string]interface{})
		if !ok {
			log.Println("Elasticsearch _source 解析失败")
			continue
		}

		res = append(res, sourceMap)
	}

	// 解析 `total`
	totalValue := 0
	if totalData, exists := hitsMap["total"]; exists {
		if totalMap, ok := totalData.(map[string]interface{}); ok {
			if value, exists := totalMap["value"]; exists {
				if valueFloat, ok := value.(float64); ok {
					totalValue = int(valueFloat)
				}
			}
		}
	}

	// ✅ 返回数据
	response.SuccessWithTotal(c, response.SuccessCode, res, totalValue)
}
