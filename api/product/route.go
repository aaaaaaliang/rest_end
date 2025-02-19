package product

import "github.com/gin-gonic/gin"

func RegisterProductRoutes(group *gin.RouterGroup) {
	group.POST("/product", addProduct)
	group.DELETE("/product", deleteProduct)
	group.PUT("/product", updateProduct)
	group.GET("/product", listProducts)
}
