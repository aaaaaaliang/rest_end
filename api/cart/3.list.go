package cart

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/logger"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func listCart(c *gin.Context) {
	type Req struct {
		Index int `form:"index" json:"index" binding:"required,min=1"`
		Size  int `form:"size" json:"size" binding:"required,min=1,max=100"`
	}

	type Annex struct {
		Code string `json:"code"` // 图片 URL
		Name string `json:"name"`
	}

	type Res struct {
		Code         string  `json:"code"`
		ProductCode  string  `json:"product_code"`
		ProductName  string  `json:"product_name"`
		ProductPrice float64 `json:"product_price"`
		SelectNum    int     `json:"select_num"`
		TotalPrice   float64 `json:"total_price"`
		Picture      Annex   `json:"picture"`
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	userCode := utils.GetUser(c)
	var rawRes []struct {
		Code         string  `json:"code"`
		ProductCode  string  `json:"product_code"`
		ProductName  string  `json:"product_name"`
		ProductPrice float64 `json:"product_price"`
		SelectNum    int     `json:"select_num"`
		TotalPrice   float64 `json:"total_price"`
		Picture      string  `json:"picture"` // JSON 字符串
	}

	count, err := config.DB.Table(new(model.UserCart)).Alias("c").
		Join("INNER", []interface{}{new(model.Products), "p"}, "c.product_code = p.code").
		Select("c.code AS code, p.products_name AS product_name, p.code AS product_code, "+
			" p.price AS product_price, c.product_num AS select_num, c.total_price AS total_price, p.picture as picture").
		Where("c.user_code = ? AND c.is_ordered = ?", userCode, false).
		Limit(req.Size, (req.Index-1)*req.Size).
		FindAndCount(&rawRes)

	if err != nil {
		logger.SendLogToESCtx(c.Request.Context(), "ERROR", "cart", "error", "cart.list.query_fail", map[string]interface{}{
			"user_code": userCode,
			"index":     req.Index,
			"size":      req.Size,
			"err":       err.Error(),
		})
		response.Success(c, response.ServerError, err)
		return
	}

	var res []Res
	for _, item := range rawRes {
		var picture Annex
		if err := json.Unmarshal([]byte(item.Picture), &picture); err != nil {
			picture = Annex{} // 解析失败时，默认空对象
		}
		res = append(res, Res{
			Code:         item.Code,
			ProductCode:  item.ProductCode,
			ProductName:  item.ProductName,
			ProductPrice: item.ProductPrice,
			SelectNum:    item.SelectNum,
			TotalPrice:   item.TotalPrice,
			Picture:      picture,
		})
	}

	// ✅ 日志记录查询成功
	logger.SendLogToESCtx(c.Request.Context(), "INFO", "cart", "operation", "cart.list.success", map[string]interface{}{
		"user_code": userCode,
		"index":     req.Index,
		"size":      req.Size,
		"count":     count,
	})

	response.SuccessWithTotal(c, response.SuccessCode, res, int(count))
}
