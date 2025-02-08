package cart

import "github.com/gin-gonic/gin"

type Router struct{}

var Routing Router

func (r Router) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/add", r.Add)
	group.GET("/list", r.List)
	// 其他购物车相关路由...
}

func (r Router) Add(c *gin.Context) {
	// 处理添加购物车
}

func (r Router) List(c *gin.Context) {
	// 处理购物车列表
} 