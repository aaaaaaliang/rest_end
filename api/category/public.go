package category

//// CategoryTree 结构体（树形分类）
//type CategoryTree struct {
//	Code     string         `json:"code"`
//	Name     string         `json:"name"`
//	Sort     int            `json:"sort"`
//	Children []CategoryTree `json:"children"`
//}
//
//// GetCategories 获取分类列表（树形结构）
//func GetCategories(c *gin.Context) {
//	var categories []model.Category
//	err := db.Find(&categories)
//	if err != nil {
//		utils.Fail(c, http.StatusInternalServerError, utils.QueryTimeoutCode, "数据库查询错误")
//		return
//	}
//
//	// 构建树形结构
//	tree := buildCategoryTree(categories, nil)
//	utils.Success(c, tree, utils.QuerySuccessCode, "查询成功")
//}
//
//// buildCategoryTree 递归构建树形分类
//func buildCategoryTree(categories []model.Category, parentCode *string) []CategoryTree {
//	var tree []CategoryTree
//
//	for _, cat := range categories {
//		// 如果当前分类的 `parent_code` 与 `parentCode` 匹配，则属于该层级
//		if (cat.ParentCode == nil && parentCode == nil) || (cat.ParentCode != nil && parentCode != nil && *cat.ParentCode == *parentCode) {
//			node := CategoryTree{
//				Code: cat.Code,
//				Name: cat.Name,
//				Sort: cat.Sort,
//			}
//
//			// 递归查询子分类
//			children := buildCategoryTree(categories, &cat.Code)
//			if len(children) > 0 {
//				node.Children = children
//			}
//
//			tree = append(tree, node)
//		}
//	}
//	return tree
//}
