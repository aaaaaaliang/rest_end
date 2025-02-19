package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"rest/config"
	"rest/model"
	"rest/response"
	"strings"
)

func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiPath := c.Request.URL.Path
		method := c.Request.Method

		// **1. 查询 API 是否为公共接口**
		var apiPermission model.APIPermission
		has, err := config.DB.Where("path = ? AND method = ?", apiPath, method).Get(&apiPermission)
		if err != nil {
			response.Success(c, response.Unauthorized, errors.New("权限查询失败"))
			c.Abort()
			return
		}

		// **如果 API 记录不存在，或者 public == 0，直接放行**
		if !has || apiPermission.Public == 0 {
			c.Next()
			return
		}

		// **2. 从 Cookie 获取 Token**
		token, err := c.Cookie("access_token")
		if err != nil {
			response.Success(c, response.Unauthorized, fmt.Errorf("未登录 %v", err))
			c.Abort()
			return
		}
		log.Println("access_token", token)

		// **3. 解析 Token 获取 user_code**
		userCode, err := config.ParseJWT(token)
		if err != nil || userCode == "" {
			response.Success(c, response.Unauthorized, fmt.Errorf("用户code未拿到 %v", err))
			c.Abort()
			return
		}
		c.Set("user", userCode)

		// **5. 处理登录 API**
		if apiPermission.Public == 1 {
			// **public == 1 表示只需要登录即可访问**
			c.Next()
			return
		}

		// **4. 查询用户的权限**
		var userPermissions []struct {
			Path   string `xorm:"path"`
			Method string `xorm:"method"`
		}

		// 使用 DISTINCT 确保去重
		err = config.DB.Table(model.UserRole{}).Alias("ur").
			Join("INNER", []interface{}{model.RolePermission{}, "rp"}, "ur.role_code = rp.role_code").
			Join("INNER", []interface{}{model.APIPermission{}, "p"}, "rp.permission_code = p.code").
			Where("ur.user_code = ?", userCode).
			Where("p.public = 2").
			Distinct("p.path", "p.method").
			Select("p.path, p.method").
			Find(&userPermissions)

		log.Println("userCode", userCode)
		log.Println("该用户权限:", userPermissions)
		if err != nil {
			response.Success(c, response.Unauthorized, fmt.Errorf("权限查询失败 %v", err))
			c.Abort()
			return
		}

		// **5. 处理 API 路径匹配**
		allowed := false
		for _, perm := range userPermissions {
			if perm.Path == apiPath && strings.EqualFold(perm.Method, method) {
				allowed = true
				break
			}
		}

		// **6. 没有权限，返回 403**
		if !allowed {
			response.Success(c, response.Unauthorized, fmt.Errorf("无访问权限 %v", err))
			c.Abort()
			return
		}

		// **7. 通过权限校验，继续处理请求**
		c.Next()
	}
}
