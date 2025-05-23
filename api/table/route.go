package table

import "github.com/gin-gonic/gin"

func RegisterTableRoutes(group *gin.RouterGroup) {
	group.POST("/table", addTable)
	group.GET("/table", listTables)
	group.PUT("/table", updateTable)
	group.DELETE("/table", deleteTable)
}
