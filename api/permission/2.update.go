package permission

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func updatePermission(c *gin.Context) {
	var req model.APIPermission
	if !utils.ValidationJson(c, &req) {
		return
	}

	req.Updater = utils.GetUser(c)

	_, err := config.DB.Where("code = ?", req.Code).Update(&req)
	if err != nil {
		response.Success(c, response.UpdateFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
