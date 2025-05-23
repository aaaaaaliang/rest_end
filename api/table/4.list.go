package table

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func listTables(c *gin.Context) {
	type Req struct {
		Index  int    `form:"index" binding:"required,min=1"`        // 当前页码
		Size   int    `form:"size" binding:"required,min=1,max=100"` // 每页条数
		Query  string `form:"query"`                                 // 模糊搜索关键词（匹配 location / remark）
		Status *int   `form:"status"`                                // 可选：过滤状态（1=可用，0=不可用）
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	var tables []model.TableInfo
	db := config.DB.Table(model.TableInfo{})

	// 状态筛选（只允许 0 和 1）
	if req.Status != nil && (*req.Status == 0 || *req.Status == 1) {
		db = db.And("status = ?", *req.Status)
	}

	// 模糊搜索 location 或 remark
	if req.Query != "" {
		likeStr := "%" + req.Query + "%"
		db = db.And("(location LIKE ? OR remark LIKE ?)", likeStr, likeStr)
	}

	count, err := db.Limit(req.Size, (req.Index-1)*req.Size).FindAndCount(&tables)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.SuccessWithTotal(c, response.SuccessCode, tables, int(count))
}
