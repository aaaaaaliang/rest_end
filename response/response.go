package response

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// **✅ 1. 通用状态码**
const (
	SuccessCode  = 200 // 成功
	BadRequest   = 400 // 参数错误
	Unauthorized = 401 // 未授权
	Forbidden    = 403 // 禁止访问
	NotFound     = 404 // 资源不存在
	TooManyReq   = 429 // 请求过多
	ServerError  = 500 // 服务器内部错误
)

// **✅ 2. 增删改查状态码**
const (
	QuerySuccess  = 2000
	QueryNoData   = 2001
	QueryFail     = 4001
	CreateSuccess = 2100
	CreateFail    = 4100
	UpdateSuccess = 2200
	UpdateFail    = 4200
	DeleteSuccess = 2300
	DeleteFail    = 4300
)

// **✅ 3. 业务状态码**
const (
	UserNotFound     = 5200
	UserExists       = 5201
	BalanceNotEnough = 5300
)

// **✅ 4. 状态码消息映射**
var messages = map[int]string{
	Success:      "操作成功",
	BadRequest:   "参数错误",
	Unauthorized: "未授权",
	Forbidden:    "禁止访问",
	NotFound:     "资源不存在",
	TooManyReq:   "请求过多",
	ServerError:  "服务器错误",

	QuerySuccess: "查询成功",
	QueryNoData:  "查询成功（但无数据）",
	QueryFail:    "查询失败",

	CreateSuccess: "创建成功",
	CreateFail:    "创建失败",

	UpdateSuccess: "更新成功",
	UpdateFail:    "更新失败",

	DeleteSuccess: "删除成功",
	DeleteFail:    "删除失败",

	UserNotFound:     "用户不存在",
	UserExists:       "用户已存在",
	BalanceNotEnough: "余额不足",
}

// **✅ 5. 统一 Response 结构**
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Total     *int        `json:"total,omitempty"`
	RequestID string      `json:"request_id"`
	Err       *string     `json:"error,omitempty"`
}

// **✅ 6. 响应成功**
func Success(c *gin.Context, code int, err ...error) {
	var errMsg *string
	if len(err) > 0 && err[0] != nil {
		errStr := err[0].Error()
		errMsg = &errStr
	}
	c.JSON(http.StatusOK, Response{
		Code:      code,
		Message:   getMessage(code),
		RequestID: getRequestID(c),
		Err:       errMsg,
	})
}

// **✅ 7. 响应成功（含数据）**
func SuccessWithData(c *gin.Context, code int, data interface{}, err ...error) {
	var errMsg *string
	if len(err) > 0 && err[0] != nil {
		errStr := err[0].Error()
		errMsg = &errStr
	}
	c.JSON(http.StatusOK, Response{
		Code:      code,
		Message:   getMessage(code),
		Data:      data,
		RequestID: getRequestID(c),
		Err:       errMsg,
	})
}

// **✅ 8. 响应成功（含数据 + 总数）**
func SuccessWithTotal(c *gin.Context, code int, data interface{}, total int, err ...error) {
	var errMsg *string
	if len(err) > 0 && err[0] != nil {
		errStr := err[0].Error()
		errMsg = &errStr
	}
	c.JSON(http.StatusOK, Response{
		Code:      code,
		Message:   getMessage(code),
		Data:      data,
		Total:     &total,
		RequestID: getRequestID(c),
		Err:       errMsg,
	})
}

// **✅ 9. 失败返回**
func Resp(c *gin.Context, code int, err error) {
	errMsg := err.Error()
	c.JSON(http.StatusBadRequest, Response{
		Code:      code,
		Message:   getMessage(code),
		RequestID: getRequestID(c),
		Err:       &errMsg,
	})
}

// **✅ 10. 根据 `code` 获取 `message`**
func getMessage(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}

// **✅ 11. 生成请求 ID**
func getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return uuid.New().String()
}
