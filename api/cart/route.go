package cart

import "github.com/gin-gonic/gin"

// RegisterCartRoutes 注册购物车相关路由
func RegisterCartRoutes(group *gin.RouterGroup) {
	group.POST("/cart", addCart)           // 添加购物车
	group.POST("/cart/delete", deleteCart) // 删除购物车项
	group.GET("/cart", listCart)           // 获取购物车列表
}
