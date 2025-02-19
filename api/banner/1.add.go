package banner

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 添加轮播图
func addBanner(c *gin.Context) {
	type Req struct {
		Image string `json:"image" binding:"required"`
		Title string `json:"title"`
		Sort  int    `json:"sort" binding:"required"` // 排序
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	banner := model.Banner{
		Sort:  req.Sort,
		Image: req.Image,
		Title: req.Title,
	}

	if _, err := config.DB.Insert(&banner); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
