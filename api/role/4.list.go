package role

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 获取所有角色
func listRoles(c *gin.Context) {
	var roles []model.Role
	err := config.DB.Find(&roles)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	response.SuccessWithData(c, response.SuccessCode, roles)
}

// ListPermissions 获取 API 权限列表（层级结构）
func listPermissions(c *gin.Context) {
	type PermissionResponse struct {
		Code        string                `json:"code"`
		Name        string                `json:"name"`
		Method      *string               `json:"method,omitempty"`
		Path        *string               `json:"path,omitempty"`
		Description *string               `json:"description,omitempty"` // 新增描述字段
		Children    []*PermissionResponse `json:"children,omitempty"`
	}

	var permissions []model.APIPermission
	err := config.DB.Asc("code").Where("public = 2").Find(&permissions)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 构建层级结构
	permissionMap := make(map[string]*PermissionResponse)
	var rootPermissions []*PermissionResponse

	// 先把所有权限存入 map
	for _, perm := range permissions {
		permResp := &PermissionResponse{
			Code:        perm.Code,
			Name:        perm.Name,
			Method:      perm.Method,
			Path:        perm.Path,
			Description: &perm.Description, // 加入描述字段
		}
		permissionMap[perm.Code] = permResp
	}

	// 处理层级关系
	for _, perm := range permissions {
		if perm.ParentCode == nil || *perm.ParentCode == "" {
			// 如果没有 ParentCode，说明是顶级权限
			rootPermissions = append(rootPermissions, permissionMap[perm.Code])
		} else {
			// 需要解引用 ParentCode
			parentCode := *perm.ParentCode
			if parent, exists := permissionMap[parentCode]; exists {
				// 这里要注意，直接使用指针，确保正确存入结构
				parent.Children = append(parent.Children, permissionMap[perm.Code])
			}
		}
	}

	response.SuccessWithData(c, response.SuccessCode, rootPermissions)
}

// 判断权限是否在角色的权限列表里
func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

// public的权限接口
func getPublicPermissions(c *gin.Context) {
	var permissions []model.APIPermission
	rows, err := config.DB.Table(model.APIPermission{}).Where("public in (1,2)").FindAndCount(&permissions)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	response.SuccessWithTotal(c, response.SuccessCode, permissions, int(rows))

}

// PermissionResponse 3. 构建层级结构，并标记 `checked` 状态
type PermissionResponse struct {
	Code        string                `json:"code"`
	Name        string                `json:"name"`
	Method      *string               `json:"method,omitempty"`
	Path        *string               `json:"path,omitempty"`
	Checked     bool                  `json:"checked"`
	Description *string               `json:"description,omitempty"` // 新增描述字段
	ParentCode  *string               `json:"parent_code,omitempty"` // 新增父权限字段
	Children    []*PermissionResponse `json:"children,omitempty"`
}

// 获取角色的权限（层级结构）
func getRolePermissions(c *gin.Context) {
	type Req struct {
		RoleCode string `form:"role_code" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	// 1. 查询该角色拥有的权限
	var assignedPermissions []string
	err := config.DB.Table(model.RolePermission{}).
		Where("role_code = ?", req.RoleCode).
		Cols("permission_code").
		Find(&assignedPermissions)

	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 2. 获取完整的权限列表
	var permissions []model.APIPermission
	err = config.DB.Asc("code").Find(&permissions)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	permissionMap := make(map[string]*PermissionResponse)
	var rootPermissions []*PermissionResponse

	// 先把所有权限存入 map，并标记是否 `checked`
	for _, perm := range permissions {
		permResp := &PermissionResponse{
			Code:        perm.Code,
			Name:        perm.Name,
			Method:      perm.Method,
			Path:        perm.Path,
			Description: &perm.Description,
			ParentCode:  perm.ParentCode,                          // 使用 APIPermission 中的 ParentCode
			Checked:     contains(assignedPermissions, perm.Code), // 标记是否被选中
		}
		permissionMap[perm.Code] = permResp
	}

	// 处理层级关系
	for _, perm := range permissions {
		if perm.ParentCode == nil || *perm.ParentCode == "" {
			// 如果没有 ParentCode，说明是顶级权限
			rootPermissions = append(rootPermissions, permissionMap[perm.Code])
		} else {
			// 需要解引用 ParentCode
			parentCode := *perm.ParentCode
			if parent, exists := permissionMap[parentCode]; exists {
				parent.Children = append(parent.Children, permissionMap[perm.Code])
			}
		}
	}

	// 4. 计算父级 `checked` 状态
	updateParentCheckedState(rootPermissions, permissionMap)

	response.SuccessWithData(c, response.SuccessCode, rootPermissions)
}

// 递归更新父级 `checked` 状态
func updateParentCheckedState(permissions []*PermissionResponse, permissionMap map[string]*PermissionResponse) {
	// 遍历当前权限的子权限
	for _, perm := range permissions {
		// 如果有子权限，递归处理
		if len(perm.Children) > 0 {
			// 递归更新子权限的状态
			updateParentCheckedState(perm.Children, permissionMap)
		}

		// 如果当前权限有任何子权限 `checked` 为 true，父权限也应该为 true
		if perm.Checked {
			// 更新父权限的状态
			if perm.ParentCode != nil && *perm.ParentCode != "" {
				// 获取父权限并设置父权限为true
				parentCode := *perm.ParentCode
				if parent, exists := permissionMap[parentCode]; exists {
					parent.Checked = true
				}
			}
		}
	}
}
