package category

import "github.com/gin-gonic/gin"

func RegisterCategoryRoutes(group *gin.RouterGroup) {
	group.POST("/category", createCategory)
	group.DELETE("/category", deleteCategory)
	group.PUT("/category", updateCategory)
	group.GET("/category", listAllCategories)
	//group.GET("/list", r.List)
	// 其他购物车相关路由...
}
