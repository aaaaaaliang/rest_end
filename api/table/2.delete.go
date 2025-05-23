package table

import (
	"github.com/gin-gonic/gin"
	"log"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func deleteTable(c *gin.Context) {
	log.Println("deleteTable")
	type Req struct {
		Code string `json:"code" binding:"required" form:"code"`
	}
	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}
	log.Println("code", req.Code)

	_, err := config.DB.Where("code = ?", req.Code).Delete(&model.TableInfo{})
	if err != nil {
		response.Success(c, response.DeleteFail, err)
		return
	}
	response.Success(c, response.SuccessCode)
}
