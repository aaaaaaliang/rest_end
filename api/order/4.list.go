package order

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func listOrder(c *gin.Context) {
	type Req struct {
		Index int  `form:"index" json:"index" binding:"required"` // 当前页码
		Size  int  `form:"size" json:"size" binding:"required"`   // 每页条数
		All   bool `form:"all" json:"all"`                        // 是否查看全部
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	userCode := utils.GetUser(c)

	var res []model.UserOrder
	db := config.DB.Limit(req.Size, (req.Index-1)*req.Size).Asc("created")

	// 如果 req.All 为 true，则查询所有；否则仅查询个人数据
	if !req.All {
		db = db.Where("user_code = ?", userCode)
	}

	num, err := db.FindAndCount(&res)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	response.SuccessWithTotal(c, response.SuccessCode, res, int(num))
}
