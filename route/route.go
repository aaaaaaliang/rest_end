package route

import (
	"github.com/gin-gonic/gin"
	"rest/api/category"
	"rest/api/user"
)

//// Route 存储 API 路由信息
//type Route struct {
//	Path    string        // API 路径
//	Method  string        // HTTP 方法（GET, POST, PUT, DELETE）
//	Handler reflect.Value // 处理函数（反射调用）
//}
//
//// Routes 存储所有自动注册的路由
//var Routes []Route
//
//// Register **支持 API 直接注册，自动解析 HTTP 方法**
//func Register(apis ...interface{}) {
//	for _, api := range apis {
//		registerAPI(api)
//	}
//}
//
//// registerAPI 解析 API 方法，并自动注册 RESTful 路由
//func registerAPI(api interface{}) {
//	apiType := reflect.TypeOf(api)
//	module := strings.ToLower(apiType.Name()) // 取 API 结构体名称（小写）
//
//	apiValue := reflect.ValueOf(api)
//
//	// 遍历 API 结构体的方法
//	for i := 0; i < apiValue.NumMethod(); i++ {
//		method := apiValue.Method(i)         // 获取方法
//		methodName := apiType.Method(i).Name // 获取方法名
//
//		// **通过方法名前缀确定 HTTP 方法**
//		httpMethod := getHTTPMethodFromPrefix(methodName)
//		if httpMethod == "" {
//			continue // 非 RESTful 方法，跳过
//		}
//
//		// **自动生成 RESTful API 路径**
//		path := generateRoutePath(methodName, module)
//
//		// 存储路由信息
//		route := Route{
//			Path:    path,
//			Method:  httpMethod,
//			Handler: method,
//		}
//		Routes = append(Routes, route)
//
//		fmt.Printf("✔️  注册路由: %s %s\n", httpMethod, path)
//	}
//}
//
//// getHTTPMethodFromPrefix **根据方法前缀解析 HTTP 方法**
//func getHTTPMethodFromPrefix(methodName string) string {
//	lowerMethod := strings.ToLower(methodName)
//
//	switch {
//	case strings.HasPrefix(lowerMethod, "get"):
//		return "GET"
//	case strings.HasPrefix(lowerMethod, "post"):
//		return "POST"
//	case strings.HasPrefix(lowerMethod, "put"):
//		return "PUT"
//	case strings.HasPrefix(lowerMethod, "delete"):
//		return "DELETE"
//	default:
//		return "" // 不符合 RESTful 规则
//	}
//}
//
//// generateRoutePath **根据方法名自动生成 RESTful 路径**
//func generateRoutePath(methodName, module string) string {
//	lowerMethod := strings.ToLower(methodName)
//
//	if strings.Contains(lowerMethod, "byid") {
//		return fmt.Sprintf("/%s/:id", module) // `GetUserByID` -> `/user/:id`
//	}
//
//	return fmt.Sprintf("/%s", module) // `GetUsers` -> `/users`
//}
//
//// Bind **绑定所有注册的路由到 Gin**
//func Bind(e *gin.Engine) {
//	for _, route := range Routes {
//		// 绑定 API 处理函数
//		switch route.Method {
//		case "GET":
//			e.GET(route.Path, match(route.Handler))
//		case "POST":
//			e.POST(route.Path, match(route.Handler))
//		case "PUT":
//			e.PUT(route.Path, match(route.Handler))
//		case "DELETE":
//			e.DELETE(route.Path, match(route.Handler))
//		}
//	}
//}
//
//// match **调用 API 处理函数**
//func match(handler reflect.Value) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		args := []reflect.Value{reflect.ValueOf(c)} // 传递 *gin.Context
//		handler.Call(args)                          // 反射调用 API 方法
//	}
//}

// RegisterRoutes **统一注册所有 API 路由**
func RegisterRoutes(r *gin.Engine) {
	apiGroup := r.Group("/api") // 统一 API 前缀

	// 注册用户 API
	user.RegisterUserRoutes(apiGroup)
	category.RegisterCategoryRoutes(apiGroup)

	//	// 注册订单 API
	//	api.RegisterOrderRoutes(apiGroup)
	//}
}
