package permission

import "github.com/gin-gonic/gin"

func RegisterPermissionRoutes(group *gin.RouterGroup) {
	group.GET("/permission", listPermissions)
	group.PUT("/permission", updatePermission)
}
