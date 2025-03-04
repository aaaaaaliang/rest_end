package banner

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// 获取轮播图列表
func getBanners(c *gin.Context) {
	type Req struct {
		Index int `form:"index" json:"index" binding:"required"` // 当前页码
		Size  int `form:"size" json:"size" binding:"required"`   // 每页条数
		//All   bool `form:"all" json:"all"`                        // 是否查看全部
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}
	var banners []model.Banner
	count, err := config.DB.Asc("created").Limit(req.Size, (req.Index-1)*req.Size).Desc("created").FindAndCount(&banners)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	response.SuccessWithTotal(c, response.SuccessCode, banners, int(count))
}
