package category

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xormplus/xorm"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 删除分类及其所有子分类（包括子孙分类）
func deleteCategory(c *gin.Context) {
	type Req struct {
		Code string `json:"code" binding:"required,max=70"`
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequestCode, errors.New("参数错误"))
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
		response.Success(c, response.DeleteFail, errors.New("删除分类及子分类失败"))
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

// 递归删除分类及其所有子分类
func deleteCategoryAndChildren(session *xorm.Session, code string) error {
	// 删除当前分类
	affectRow, err := session.Where("code = ?", code).Delete(&model.Category{})
	if err != nil || affectRow != 1 {
		return fmt.Errorf("删除当前分类失败: %v", err)
	}

	// 查询当前分类下的所有子分类
	var children []model.Category
	if err = session.Where("parent_code = ?", code).Find(&children); err != nil {
		return errors.New("查询子分类失败: " + err.Error())
	}

	// 递归删除每一个子分类及其后代
	for _, child := range children {
		if _, err = session.Where("code = ?", child.Code).Delete(&model.Category{}); err != nil {
			return errors.New("删除子分类失败: " + err.Error())
		}

		// 递归删除子分类的子分类
		err = deleteCategoryAndChildren(session, child.Code)
		if err != nil {
			return err
		}
	}
	return nil
}
