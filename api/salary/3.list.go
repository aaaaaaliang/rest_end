package salary

import (
	"github.com/gin-gonic/gin"
	"log"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 获取薪资记录（根据起始时间和结束时间查询）
func getSalaryRecords(c *gin.Context) {
	type Req struct {
		Index    int    `form:"index" binding:"required,min=1"` // 页码
		Size     int    `form:"size" binding:"required,min=1"`  // 每页大小
		UserCode string `form:"user_code"`                      // 用户code，可以为空，查询所有
		//StartTime int64  `form:"start_time"`                     // 起始时间  前端传递时间戳
		//EndTime   int64  `form:"end_time"`                       // 结束时间
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	log.Println("--------", req)
	// 基本查询条件
	query := config.DB.Table(model.SalaryRecord{})

	//// 如果传递了起始时间，则加上起始时间过滤
	//if req.StartTime > 0 {
	//	query = query.Where("pay_date >= ?", req.StartTime)
	//}
	//
	//// 如果传递了结束时间，则加上结束时间过滤
	//if req.EndTime > 0 {
	//	query = query.Where("pay_date <= ?", req.EndTime)
	//}

	// 如果指定了用户code，添加额外的查询条件
	if req.UserCode != "" {
		query = query.Where("user_code = ?", req.UserCode)
	}

	// 查询数据
	var salaryRecords []model.SalaryRecord
	count, err := query.Limit(req.Size, (req.Index-1)*req.Size).FindAndCount(&salaryRecords)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 返回薪资记录
	response.SuccessWithTotal(c, response.SuccessCode, salaryRecords, int(count))
}
