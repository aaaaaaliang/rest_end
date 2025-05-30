package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"rest/config"
	"rest/model"
	"rest/response"
	"strings"
	"time"
)

//
//func PermissionMiddleware() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		ctx := context.Background()
//
//		apiPath := c.Request.URL.Path
//		method := c.Request.Method
//
//		// 1. 查询接口权限定义
//		var apiPermission model.APIPermission
//		has, err := config.DB.Where("path = ? AND method = ?", apiPath, method).Get(&apiPermission)
//		if err != nil {
//			response.Success(c, response.Unauthorized, fmt.Errorf("查询权限失败: %v", err))
//			c.Abort()
//			return
//		}
//
//		// 2. 如果未配置该接口权限，或配置为公开接口（public == 0） => 放行
//		if !has || apiPermission.Public == 0 {
//			c.Next()
//			return
//		}
//
//		// 3️. 从 cookie 获取 access_token
//		token, err := c.Cookie("access_token")
//		if err != nil {
//			response.Success(c, response.Unauthorized, fmt.Errorf("未登录: %v", err))
//			c.Abort()
//			return
//		}
//
//		// 4️. 解析 token 获取 user_code
//		userCode, err := config.ParseJWT(token)
//		if err != nil || userCode == "" {
//			response.Success(c, response.Unauthorized, fmt.Errorf("token 解析失败: %v", err))
//			c.Abort()
//			return
//		}
//		c.Set("user", userCode)
//
//		// 超级管理员拥有所有权限
//		if userCode == "admin" {
//			c.Next()
//			return
//		}
//
//		// 5️. 如果只要求登录（public == 1）=> 登录用户已校验，直接放行
//		if apiPermission.Public == 1 {
//			c.Next()
//			return
//		}
//
//		// 6️. public == 2，角色授权 => 查缓存 or 数据库
//		cacheKey := fmt.Sprintf("user_permissions:%s", userCode)
//		val, err := config.R.Get(ctx, cacheKey).Result()
//
//		var userPermissions []struct {
//			Path   string `xorm:"path"`
//			Method string `xorm:"method"`
//		}
//
//		if errors.Is(err, redis.Nil) {
//			err = config.DB.Table(model.UserRole{}).Alias("ur").
//				Join("INNER", []interface{}{model.RolePermission{}, "rp"}, "ur.role_code = rp.role_code").
//				Join("INNER", []interface{}{model.APIPermission{}, "p"}, "rp.permission_code = p.code").
//				Where("ur.user_code = ?", userCode).
//				Where("p.public = 2").
//				Distinct("p.path", "p.method").
//				Select("p.path, p.method").
//				Find(&userPermissions)
//
//			if err != nil {
//				response.Success(c, response.Unauthorized, fmt.Errorf("权限加载失败: %v", err))
//				c.Abort()
//				return
//			}
//
//			cacheData, _ := json.Marshal(userPermissions)
//			_ = config.R.Set(ctx, cacheKey, cacheData, time.Hour).Err()
//		} else if err == nil {
//			_ = json.Unmarshal([]byte(val), &userPermissions)
//		} else {
//			response.Success(c, response.Unauthorized, fmt.Errorf("读取权限缓存失败: %v", err))
//			c.Abort()
//			return
//		}
//
//		// 7️ 权限比对
//		allowed := false
//		for _, perm := range userPermissions {
//			if perm.Path == apiPath && strings.EqualFold(perm.Method, method) {
//				allowed = true
//				break
//			}
//		}
//
//		if !allowed {
//			response.Success(c, response.Unauthorized, fmt.Errorf("无访问权限"))
//			c.Abort()
//			return
//		}
//		c.Next()
//	}
//}

func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		apiPath := c.Request.URL.Path
		method := c.Request.Method

		// 1. 查询接口权限定义
		var apiPermission model.APIPermission
		has, err := config.DB.Where("path = ? AND method = ?", apiPath, method).Get(&apiPermission)
		if err != nil {
			response.Success(c, response.Unauthorized, fmt.Errorf("查询权限失败: %v", err))
			c.Abort()
			return
		}

		// 2. 如果未配置该接口权限，或配置为公开接口（public == 0） => 放行
		if !has || apiPermission.Public == 0 {
			c.Next()
			return
		}

		// 3. 从 cookie 获取 access_token
		token, err := c.Cookie("access_token")
		if err != nil {
			response.Success(c, response.Unauthorized, fmt.Errorf("未登录: %v", err))
			c.Abort()
			return
		}

		// 4. 解析 token 获取 user_code
		userCode, err := config.ParseJWT(token)
		if err != nil || userCode == "" {
			response.Success(c, response.Unauthorized, fmt.Errorf("token 解析失败: %v", err))
			c.Abort()
			return
		}
		c.Set("user", userCode)

		// 5. 读取用户信息（真实姓名 和 角色名）缓存优先
		userInfoKey := fmt.Sprintf("user_info:%s", userCode)
		userInfoJSON, err := config.R.Get(ctx, userInfoKey).Result()

		var userInfo struct {
			UserName string `json:"user_name"`
			Role     string `json:"role"`
		}

		if errors.Is(err, redis.Nil) {
			// 联表查询用户 + 角色名
			sql := `
				SELECT u.real_name AS user_name, r.name AS role
				FROM users u
				LEFT JOIN user_role ur ON u.code = ur.user_code
				LEFT JOIN role r ON ur.role_code = r.code
				WHERE u.code = ?
				LIMIT 1
			`
			has, err := config.DB.SQL(sql, userCode).Get(&userInfo)
			if err != nil || !has {
				response.Success(c, response.Unauthorized, fmt.Errorf("用户不存在"))
				c.Abort()
				return
			}

			cacheData, _ := json.Marshal(userInfo)
			_ = config.R.Set(ctx, userInfoKey, cacheData, time.Minute*10).Err()
		} else if err == nil {
			_ = json.Unmarshal([]byte(userInfoJSON), &userInfo)
		} else {
			response.Success(c, response.Unauthorized, fmt.Errorf("读取用户信息失败: %v", err))
			c.Abort()
			return
		}

		// 注入上下文字段
		c.Set("user_name", userInfo.UserName)
		c.Set("user_role", userInfo.Role)

		// 6. 超级管理员拥有所有权限
		if userCode == "admin" {
			c.Next()
			return
		}

		// 7. 如果只要求登录即可（public == 1）=> 放行
		if apiPermission.Public == 1 {
			c.Next()
			return
		}

		// 8. public == 2，角色授权，查缓存或数据库
		cacheKey := fmt.Sprintf("user_permissions:%s", userCode)
		val, err := config.R.Get(ctx, cacheKey).Result()

		var userPermissions []struct {
			Path   string `xorm:"path"`
			Method string `xorm:"method"`
		}

		if errors.Is(err, redis.Nil) {
			err = config.DB.Table(model.UserRole{}).Alias("ur").
				Join("INNER", []interface{}{model.RolePermission{}, "rp"}, "ur.role_code = rp.role_code").
				Join("INNER", []interface{}{model.APIPermission{}, "p"}, "rp.permission_code = p.code").
				Where("ur.user_code = ?", userCode).
				Where("p.public = 2").
				Distinct("p.path", "p.method").
				Select("p.path, p.method").
				Find(&userPermissions)

			if err != nil {
				response.Success(c, response.Unauthorized, fmt.Errorf("权限加载失败: %v", err))
				c.Abort()
				return
			}

			cacheData, _ := json.Marshal(userPermissions)
			_ = config.R.Set(ctx, cacheKey, cacheData, time.Hour).Err()
		} else if err == nil {
			_ = json.Unmarshal([]byte(val), &userPermissions)
		} else {
			response.Success(c, response.Unauthorized, fmt.Errorf("读取权限缓存失败: %v", err))
			c.Abort()
			return
		}

		// 9. 权限比对
		allowed := false
		for _, perm := range userPermissions {
			if perm.Path == apiPath && strings.EqualFold(perm.Method, method) {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Success(c, response.Unauthorized, fmt.Errorf("无访问权限"))
			c.Abort()
			return
		}

		c.Next()
	}
}
