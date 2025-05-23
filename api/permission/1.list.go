package permission

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func listPermissions(c *gin.Context) {
	type Req struct {
		Index   int    `form:"index" binding:"required,min=1"`
		Size    int    `form:"size" binding:"required,min=1,max=100"`
		Keyword string `form:"keyword"`
	}
	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	session := config.DB.Table(model.APIPermission{})
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		session = session.Where("name LIKE ? OR path LIKE ? OR description LIKE ?", like, like, like)
	}

	var list []model.APIPermission
	count, err := session.Limit(req.Size, (req.Index-1)*req.Size).FindAndCount(&list)
	if err != nil {
		response.SuccessWithTotal(c, response.ServerError, nil, 0)
		return
	}
	response.SuccessWithTotal(c, response.SuccessCode, list, int(count))
}
