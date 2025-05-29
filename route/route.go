package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	aiv1admin "rest/api/ai/v1/admin"
	aiv1user "rest/api/ai/v1/user"
	"rest/api/banner"
	"rest/api/cart"
	"rest/api/category"
	"rest/api/coupon"
	"rest/api/dashboard"
	"rest/api/order"
	"rest/api/pay"
	"rest/api/permission"
	"rest/api/product"
	"rest/api/public"
	"rest/api/role"
	"rest/api/salary"
	"rest/api/table"
	"rest/api/user"
	handler "rest/api/ws/v1"
	"rest/config"
	"rest/model"
	"strings"
)

// **统一注册所有 API 路由**
func registerRoutes(r *gin.Engine) {
	apiGroup := r.Group("/api") // 统一 API 前缀

	// 注册用户 API
	user.RegisterUserRoutes(apiGroup)
	category.RegisterCategoryRoutes(apiGroup)
	public.RegisterPublicRoutes(apiGroup)
	product.RegisterProductRoutes(apiGroup)
	cart.RegisterCartRoutes(apiGroup)
	order.RegisterOrderRoutes(apiGroup)
	banner.RegisterBannerRoutes(apiGroup)
	role.RegisterRoleRoutes(apiGroup)
	salary.RegisterSalaryRoutes(apiGroup)
	dashboard.RegisterDashboardRoutes(apiGroup)
	pay.RegisterPayRoutes(apiGroup)
	table.RegisterTableRoutes(apiGroup)
	permission.RegisterPermissionRoutes(apiGroup)
	coupon.RegisterCouponTemplateRoutes(apiGroup)
	handler.RegisterWSRoutes(apiGroup)
	aiv1admin.RegisterAIModelRoutes(apiGroup)
	aiv1user.RegisterAIUserRoutes(apiGroup)
}

func autoRegisterAPIPermissions(router *gin.Engine) {
	routes := router.Routes()
	var permissions []model.APIPermission
	parentMap := make(map[string]string)

	log.Println("routes:", routes)
	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/debug") || strings.Contains(route.Handler, "gin.") {
			continue
		}

		// 获取顶级分类（如 `/api/user` => `user`）
		pathParts := strings.Split(strings.Trim(route.Path, "/"), "/")
		if len(pathParts) < 2 {
			continue
		}

		topLevelCode := generateCode(pathParts[1], "")
		if _, exists := parentMap[topLevelCode]; !exists {
			parentMap[topLevelCode] = topLevelCode
			permissions = append(permissions, model.APIPermission{
				BasicModel:  model.BasicModel{Code: topLevelCode},
				Name:        strings.Title(pathParts[1]) + " 管理",
				ParentCode:  nil,
				Method:      nil,
				Path:        nil,
				Description: fmt.Sprintf("%s 模块权限", strings.Title(pathParts[1])),
			})
		}

		// **提取处理函数名称（方法名）**
		methodName := extractMethodName(route.Handler)

		routeCode := generateCode(route.Path, route.Method)
		permission := model.APIPermission{
			BasicModel:  model.BasicModel{Code: routeCode},
			Name:        fmt.Sprintf("%s - %s", strings.Title(pathParts[1]), strings.ToUpper(route.Method)),
			Method:      &route.Method,
			Path:        &route.Path,
			ParentCode:  &topLevelCode,
			Description: fmt.Sprintf("处理函数: %s", methodName), // 将方法名加入 description

		}

		permissions = append(permissions, permission)
	}

	// 存入数据库（存在就更新）
	for _, perm := range permissions {
		exist, _ := config.DB.Where("code = ?", perm.Code).Exist(&model.APIPermission{})
		if exist {
			_, err := config.DB.Where("code = ?", perm.Code).Update(&perm)
			if err != nil {
				fmt.Printf("更新权限失败: %v\n", err)
			}
		} else {
			_, err := config.DB.Insert(&perm)
			if err != nil {
				fmt.Printf("插入权限失败: %v\n", err)
			}
		}
	}

	fmt.Println("API 权限自动注册完成！")
}

// 生成 `code`
func generateCode(path, method string) string {
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.Trim(path, "_")
	if method == "" {
		return path
	}
	return fmt.Sprintf("%s_%s", path, strings.ToLower(method))
}

// **提取方法名**
func extractMethodName(handler string) string {
	// Gin 的 handler 形如 "rest/api/user.createUser"
	parts := strings.Split(handler, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1] // 取最后一部分，即方法名
	}
	return handler // 兜底返回完整的 handler 名
}
