package table

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func addTable(c *gin.Context) {
	var req model.TableInfo
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)
	req.Creator = userCode

	if _, err := config.DB.Insert(&req); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	response.Success(c, response.SuccessCode)
}
