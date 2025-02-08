package model

import "github.com/google/uuid"

type BasicModel struct {
	Id      int64  `json:"id" xorm:"pk autoincr"`
	Code    string `json:"code" xorm:"varchar(70) index"`
	Created int64  `json:"created" xorm:"created"`
	Updated int64  `json:"updated" xorm:"updated"`
	Deleted int64  `json:"deleted" xorm:"deleted"`
	Creator string `json:"creator" xorm:"varchar(70)"`
	Updater string `json:"updater" xorm:"varchar(70)"`
}

func (b *BasicModel) BeforeInsert() {
	if b.Code == "" {
		b.Code = uuid.New().String()
	}
}

// Annex 附件材料
type Annex struct {
	// Code 文件标识
	Code string `json:"code"  binding:"max=70"`
	// Name 文件名称
	Name string `json:"name" binding:"max=255"`
}

//// SystemBackend 系统后端接口表
//type SystemBackend struct {
//	Id int64 `json:"id,omitempty" xorm:"pk autoincr"`
//	// Name 后端api名称
//	Name string `json:"name" xorm:"varchar(200) comment('后端接口的名称,尽可能的有意义')"`
//	// Code 后端api标识, 非随机生成 生成方式( hash(app + method + path))
//	Code string `json:"code" xorm:"varchar(70) comment('后端接口的唯一标识')"`
//	// App 应用名称
//	App string `json:"app" xorm:"varchar(200) comment('接口所属应用, 用以支持对子应用的控制')"`
//	// Protocol 协议
//	Protocol string `json:"protocol" xorm:"varchar(200) comment('接口所属协议')"`
//	// Path 后端API路径
//	Path string `json:"path" xorm:"varchar(200) comment('后端接口的路径')"`
//	// Method 请求的方法
//	Method string `json:"method" xorm:"varchar(100) comment('接口请求方法')"`
//	// Sort 显示排序
//	Sort int `json:"sort" xorm:"default 100 comment('接口显示排序')"`
//	// Enabled 接口是否已启用
//	Enabled bool `json:"enable" xorm:"default true comment('接口是否启用')"`
//	// Public 是否为公开可访问api
//	Public bool `json:"public" xorm:"default false comment('接口是否公开')"`
//	// Category 分类名称
//	Category string `json:"category" xorm:"varchar(200) comment('接口分类名称')"`
//	// Ancestor 根节点
//	Ancestor string `json:"ancestor" xorm:"varchar(200) comment('接口所属的根节点名称')"`
//	Created  int    `json:"created" xorm:"created comment('接口创建时间')"`
//	Updated  int    `json:"updated" xorm:"updated comment('接口更新时间')"`
//	// 业务处理中间件
//	Handles []gin.HandlerFunc `json:"-" xorm:"-"`
//}
//
//func (s *SystemBackend) ExistWithCode() (backend SystemBackend, exist bool) {
//	if s.Code == "" {
//		return
//	}
//	exist, _ = public.DB.Where("code = ?", s.Code).Get(&backend)
//	if !exist {
//		return *s, exist
//	}
//	return
//}
//
//func (s *SystemBackend) UpdateWithCols(cols []string) error {
//	if s.Code == "" {
//		return errors.New("code is empty")
//	}
//	_, err := public.DB.Cols(cols...).Where("code = ?", s.Code).Update(s)
//	return err
//}
//
//func (s *SystemBackend) Insert(session ...*xorm.Session) error {
//	if len(session) > 0 {
//		_, err := session[0].Insert(s)
//		return err
//	}
//	_, err := public.DB.Insert(s)
//	return err
//}
