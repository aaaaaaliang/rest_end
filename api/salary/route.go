package salary

import "github.com/gin-gonic/gin"

func RegisterSalaryRoutes(group *gin.RouterGroup) {
	group.POST("/salary", payUserSalary)
	group.DELETE("/salary", deleteSalaryRecord)
	group.GET("/salary", getSalaryRecords)

}
