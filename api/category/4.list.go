package category

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 查询所有分类及其子分类（树形结构）
func listAllCategories(c *gin.Context) {
	// 获取所有分类数据
	var categories []model.Category
	err := config.DB.Find(&categories)
	if err != nil {
		response.Success(c, response.QueryFail, errors.New("查询分类失败"))
		return
	}

	// 将所有分类组织成树形结构
	categoryTree, err := buildCategoryTree(categories)
	if err != nil {
		response.Success(c, response.QueryFail, errors.New("构建分类树失败"))
		return
	}

	// 返回分类树
	response.SuccessWithData(c, response.SuccessCode, categoryTree)
}

// 构建分类树，递归获取每个父分类的子分类
func buildCategoryTree(categories []model.Category) ([]model.Category, error) {
	var tree []model.Category

	// 查找所有没有父分类的分类（根节点）
	for _, category := range categories {
		// 如果 ParentCode 是 nil，说明它是根节点
		if category.ParentCode == nil || *category.ParentCode == "" {
			// 查找该根节点的所有子分类
			children, err := findChildren(category.Code, categories)
			if err != nil {
				return nil, err
			}

			// 将子分类赋值给当前根节点
			category.SubCategories = children
			tree = append(tree, category)
		}
	}

	return tree, nil
}

// 根据父分类的 Code 查找所有子分类
func findChildren(parentCode string, categories []model.Category) ([]model.Category, error) {
	var children []model.Category

	// 查找所有子分类
	for _, category := range categories {
		if category.ParentCode != nil && *category.ParentCode == parentCode {
			// 递归查找子分类的子分类
			subChildren, err := findChildren(category.Code, categories)
			if err != nil {
				return nil, err
			}

			// 将子分类的子分类赋值给当前分类
			category.SubCategories = subChildren
			children = append(children, category)
		}
	}

	return children, nil
}
