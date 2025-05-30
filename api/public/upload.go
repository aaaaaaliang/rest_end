package public

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"path/filepath"
	"rest/config"
	"rest/response"
	"time"
)

//func uploadFile(c *gin.Context) {
//	// 获取上传的文件
//	file, err := c.FormFile("file")
//	if err != nil {
//		logger.Println("文件获取失败:", err)
//		response.Success(c, response.BadRequest, fmt.Errorf("获取上传文件失败: %v", err))
//		return
//	}
//
//	// 获取文件扩展名
//	ext := filepath.Ext(file.Filename)
//	if !isAllowedFileType(ext) {
//		response.Success(c, response.BadRequest, fmt.Errorf("文件类型不支持: %s", file.Filename))
//		return
//	}
//
//	// 生成唯一文件名
//	fileName := generateFileName(ext)
//	savePath := filepath.Join("./uploads", fileName)
//
//	// 保存文件
//	if err := c.SaveUploadedFile(file, savePath); err != nil {
//		logger.Println("文件保存失败:", err)
//		response.Success(c, response.ServerError, fmt.Errorf("文件保存失败: %v", err))
//		return
//	}
//
//	url := fmt.Sprintf("%s", config.G.Uploads.Url)
//
//	// 返回成功信息
//	response.SuccessWithData(c, response.SuccessCode, gin.H{
//		"filename": file.Filename,                               // 原始文件名
//		"filepath": savePath,                                    // 服务器存储路径
//		"url":      fmt.Sprintf("%s/uploads/%s", url, fileName), // 访问 URL
//	})
//}

// uploadFile 处理文件上传
func uploadFile(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		log.Println("文件获取失败:", err)
		response.Success(c, response.BadRequest, fmt.Errorf("获取上传文件失败: %v", err))
		return
	}

	// 获取文件扩展名
	ext := filepath.Ext(file.Filename)
	if !isAllowedFileType(ext) {
		response.Success(c, response.BadRequest, fmt.Errorf("文件类型不支持: %s", file.Filename))
		return
	}

	// 生成唯一文件名
	fileName := generateFileName(ext)
	savePath := filepath.Join("./uploads", fileName)

	// 确保目录存在
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		log.Println("创建目录失败:", err)
		response.Success(c, response.ServerError, fmt.Errorf("创建目录失败: %v", err))
		return
	}

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		log.Println("文件保存失败:", err)
		response.Success(c, response.ServerError, fmt.Errorf("文件保存失败: %v", err))
		return
	}

	// 设置可访问的 URL
	url := fmt.Sprintf("%s/uploads/%s", config.G.Uploads.Url, fileName)

	// 返回成功信息
	response.SuccessWithData(c, response.SuccessCode, gin.H{
		"filename": file.Filename, // 原始文件名
		"filepath": savePath,      // 服务器存储路径
		"url":      url,           // 返回可访问的 URL
	})
}

func isAllowedFileType(ext string) bool {
	allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt", ".webp"}
	for _, allowed := range allowedTypes {
		if ext == allowed {
			return true
		}
	}
	return false
}

// generateFileName 生成唯一文件名
func generateFileName(ext string) string {
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
}
