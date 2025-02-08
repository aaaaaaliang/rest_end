package category

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/response"
)

// 更新分类
func updateCategory(c *gin.Context) {
	type Req struct {
		Code         string `json:"code" binding:"required，max=70"`
		CategoryName string `json:"category_name" binding:"required,max=255"`
		Sort         int    `json:"sort" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SuccessWithData(c, response.BadRequestCode, err.Error())
		return
	}

	affectRow, err := config.DB.Update(map[string]interface{}{
		"category_name": req.CategoryName,
		"sort":          req.Sort,
	})
	if err != nil || affectRow != 1 {
		response.Success(c, response.CreateFail)
		return
	}

	response.Success(c, response.CreateSuccess)
}
