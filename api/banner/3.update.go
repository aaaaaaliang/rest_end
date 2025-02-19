package banner

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 更新轮播图
func updateBanner(c *gin.Context) {
	type Req struct {
		Code  string `json:"code" binding:"required"`  // 轮播图ID
		Image string `json:"image" binding:"required"` // 轮播图图片地址
		Title string `json:"title"`                    // 轮播图标题（可选）
		Sort  int    `json:"sort" binding:"required"`  // 排序
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	// 构造更新数据
	banner := model.Banner{
		Image: req.Image,
		Title: req.Title,
		Sort:  req.Sort,
	}

	// 更新数据库中的数据
	affected, err := config.DB.Where("code = ?", req.Code).Update(&banner)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 检查是否有记录被更新
	if affected == 0 {
		response.Success(c, response.NotFound, errors.New("轮播图不存在或未更新"))
		return
	}

	response.Success(c, response.SuccessCode)
}
