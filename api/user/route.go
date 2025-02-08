package user

import "github.com/gin-gonic/gin"

// RegisterUserRoutes **在这里定义 User 相关的路由**
func RegisterUserRoutes(r *gin.RouterGroup) {
	user := r.Group("/users")
	{
		user.GET("/", getUsers)         // GET /users
		user.GET("/:id", getUserByID)   // GET /users/:id
		user.POST("/", createUser)      // POST /users
		user.PUT("/:id", updateUser)    // PUT /users/:id
		user.DELETE("/:id", deleteUser) // DELETE /users/:id
	}
}
