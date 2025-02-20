package middleware

import (
	"context"
	"encoding/json" // 添加导入 json 包
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"rest/config"
	"rest/model"
	"rest/response"
	"strings"
	"time"
)

func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个新的上下文（ctx）
		ctx := context.Background() // 创建新的上下文

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

		// **4. 查询用户的权限，先查 Redis**
		var userPermissions []struct {
			Path   string `xorm:"path"`
			Method string `xorm:"method"`
		}

		// 尝试从 Redis 获取缓存的权限
		cacheKey := fmt.Sprintf("user_permissions:%s", userCode)
		val, err := config.R.Get(ctx, cacheKey).Result()

		if errors.Is(err, redis.Nil) {
			// Redis 中没有缓存，查询数据库并缓存到 Redis
			log.Println("a1")
			err = config.DB.Table(model.UserRole{}).Alias("ur").
				Join("INNER", []interface{}{model.RolePermission{}, "rp"}, "ur.role_code = rp.role_code").
				Join("INNER", []interface{}{model.APIPermission{}, "p"}, "rp.permission_code = p.code").
				Where("ur.user_code = ?", userCode).
				Where("p.public = 2").
				Distinct("p.path", "p.method").
				Select("p.path, p.method").
				Find(&userPermissions)

			if err != nil {
				response.Success(c, response.Unauthorized, fmt.Errorf("权限查询失败 %v", err))
				c.Abort()
				return
			}

			// 将用户权限数据缓存到 Redis 中（有效期 1 小时）
			permissionsData, _ := json.Marshal(userPermissions)
			err = config.R.Set(ctx, cacheKey, permissionsData, time.Hour).Err()
			if err != nil {
				log.Println("缓存权限失败:", err)
			}

		} else if err != nil {
			// 读取 Redis 错误
			response.Success(c, response.Unauthorized, fmt.Errorf("读取缓存失败 %v", err))
			c.Abort()
			return
		} else {
			log.Println("a2")
			// Redis 中有缓存，直接解析缓存数据
			err = json.Unmarshal([]byte(val), &userPermissions)
			if err != nil {
				response.Success(c, response.Unauthorized, fmt.Errorf("解析缓存数据失败 %v", err))
				c.Abort()
				return
			}
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
			response.Success(c, response.Unauthorized, fmt.Errorf("无访问权限"))
			c.Abort()
			return
		}

		// **7. 通过权限校验，继续处理请求**
		c.Next()
	}
}
