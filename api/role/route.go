package role

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(group *gin.RouterGroup) {
	group.GET("/role", listRoles)                          // 获取角色列表
	group.POST("/role", addRole)                           // 创建新角色
	group.PUT("/role", updateRole)                         // 更新角色信息
	group.DELETE("/role", deleteRole)                      // 删除角色
	group.GET("/role/permission", getRolePermissions)      // 获取某个角色的权限
	group.GET("/role/permissions", listPermissions)        // 所有角色的权限
	group.POST("/role/assign", assignRolePermissions)      // 给角色分配权限
	group.DELETE("/role/permission", removeRolePermission) // 移除角色的某个权限
	group.GET("/role/public", getPublicPermissions)
}
