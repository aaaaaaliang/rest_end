package order

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func updateOrder(c *gin.Context) {
	type Req struct {
		Code    string `json:"code" binding:"required"` // 订单编号
		Status  int    `json:"status"`                  // 更新订单状态
		Remark  string `json:"remark"`                  // 更新备注
		Version int    `json:"version"`                 // 乐观锁版本号
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	var o model.UserOrder
	// 查找订单
	if _, err := config.DB.Where("code = ?", req.Code).Get(&o); err != nil {
		response.Success(c, response.QueryFail, fmt.Errorf("updateOrder error: %v", err))
		return
	}

	// 检查订单是否已被锁定
	if o.Status == 2 {
		response.Success(c, response.UpdateFail, errors.New("订单已锁定"))
		return
	}

	// 检查版本号是否一致
	if o.Version != req.Version {
		response.Success(c, response.UpdateFail, errors.New("数据已被其他操作修改"))
		return
	}

	// 构造更新模型
	order := &model.UserOrder{
		Status:  req.Status,
		Remark:  req.Remark,
		Version: o.Version + 1, // 更新版本号
	}

	// 执行更新操作
	if affectRow, err := config.DB.Where("code = ?", req.Code).Update(order); err != nil || affectRow != 1 {
		response.Success(c, response.UpdateFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
