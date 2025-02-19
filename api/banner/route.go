package banner

import "github.com/gin-gonic/gin"

func RegisterBannerRoutes(group *gin.RouterGroup) {
	group.POST("/banner", addBanner)
	group.DELETE("/banner", deleteBanner)
	group.PUT("/banner", updateBanner)
	group.GET("/banner", getBanners)
}
