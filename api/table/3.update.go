package table

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func updateTable(c *gin.Context) {
	var req model.TableInfo
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)
	req.Updater = userCode

	if _, err := config.DB.Where("code = ?", req.Code).Update(&req); err != nil {
		response.Success(c, response.UpdateFail, err)
		return
	}
	response.Success(c, response.SuccessCode)
}
