package category

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xormplus/xorm"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// 删除分类及其所有子分类（包括子孙分类）
func deleteCategory(c *gin.Context) {
	type Req struct {
		Code string `json:"code" binding:"required,max=70" form:"code"`
	}

	var req Req
	if ok := utils.ValidationQuery(c, &req); !ok {
		return
	}

	// 使用事务
	session := config.DB.NewSession()
	defer session.Close()

	// 开启事务
	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, errors.New("开启事务失败"))
		return
	}

	// 删除当前分类及其所有子分类（递归）
	err := deleteCategoryAndChildren(session, req.Code)
	if err != nil {
		session.Rollback()
		response.Success(c, response.DeleteFail, fmt.Errorf("删除分类及子分类失败 %v", err))
		return
	}

	// 提交事务
	if err = session.Commit(); err != nil {
		session.Rollback()
		response.Success(c, response.ServerError, errors.New("提交事务失败"))
		return
	}

	// 返回删除成功
	response.Success(c, response.SuccessCode)
}

// 递归删除分类及其所有子分类（前提是没有产品引用）
func deleteCategoryAndChildren(session *xorm.Session, code string) error {
	// 🛡️ 1. 检查是否有产品引用该分类
	count, err := session.Where("category_code = ?", code).Count(new(model.Products))
	if err != nil {
		return fmt.Errorf("检查产品引用失败: %v", err)
	}
	if count > 0 {
		return fmt.Errorf("分类已被 %d 个产品引用，无法删除", count)
	}

	// 📚 2. 查询当前分类的所有子分类
	var children []model.Category
	if err := session.Where("parent_code = ?", code).Find(&children); err != nil {
		return fmt.Errorf("查询子分类失败: %v", err)
	}

	// 🔁 3. 递归删除子分类
	for _, child := range children {
		if err := deleteCategoryAndChildren(session, child.Code); err != nil {
			return err
		}
	}

	// ❌ 4. 删除当前分类
	affected, err := session.Where("code = ?", code).Delete(new(model.Category))
	if err != nil || affected != 1 {
		return fmt.Errorf("删除当前分类失败: %v", err)
	}

	return nil
}
