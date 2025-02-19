package category

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 新增分类
func createCategory(c *gin.Context) {
	type Req struct {
		ParentCode   string `json:"parent_code" binding:"max=70"`
		CategoryName string `json:"category_name" binding:"required,max=255"`
		Sort         int    `json:"sort" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SuccessWithData(c, response.BadRequest, err.Error())
		return
	}

	if req.ParentCode != "" {
		var parent model.Category
		if has, _ := config.DB.Where("code = ?", req.ParentCode).Get(&parent); !has {
			response.Success(c, response.QueryFail)
			return
		}
	}

	category := model.Category{
		CategoryName: req.CategoryName,
		ParentCode:   &req.ParentCode,
		Sort:         req.Sort,
	}

	if _, err := config.DB.Insert(&category); err != nil {
		response.Success(c, response.CreateFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
