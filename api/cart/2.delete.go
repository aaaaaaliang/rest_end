package cart

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/logger"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// 删除购物车项
func deleteCart(c *gin.Context) {
	type Req struct {
		Code []string `json:"code" binding:"required"` // 要删除的购物车项 code
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)

	// 执行删除操作
	affectRow, err := config.DB.Where("user_code = ?", userCode).In("code", req.Code).Delete(&model.UserCart{})
	if err != nil {
		// ❌ 删除失败，记录错误日志
		logger.SendLogToESCtx(c.Request.Context(), "ERROR", "cart", "error", "cart.delete.fail", map[string]interface{}{
			"user_code": userCode,
			"codes":     req.Code,
			"err":       err.Error(),
		})
		response.Success(c, response.DeleteFail, err)
		return
	}

	if affectRow == 0 {
		// ⚠️ 没有匹配项
		logger.SendLogToESCtx(c.Request.Context(), "WARN", "cart", "error", "cart.delete.not_found", map[string]interface{}{
			"user_code": userCode,
			"codes":     req.Code,
		})
		response.Success(c, response.NotFound, errors.New("购物车项不存在"))
		return
	}

	// ✅ 删除成功日志
	logger.SendLogToESCtx(c.Request.Context(), "INFO", "cart", "operation", "cart.delete.success", map[string]interface{}{
		"user_code":  userCode,
		"codes":      req.Code,
		"delete_num": affectRow,
	})

	response.Success(c, response.SuccessCode)
}
