package excel

import "github.com/xuri/excelize/v2"

// ExcelFile 封装excelize.File，对外暴露简化后的方法
type ExcelFile struct {
	file *excelize.File // 底层excelize文件实例
}

// 内部配置结构体：严格对齐官方DocProperties字段，设置默认值
type docPropsConfig struct {
	title          string // 标题
	subject        string // 主题
	creator        string // 创建者（作者），默认系统用户名
	keywords       string // 关键词
	description    string // 描述/备注
	lastModifiedBy string // 最后修改者
	category       string // 分类
	contentStatus  string // 内容状态（草稿/发布等）
	created        string // 创建时间（ISO 8601，如 2026-03-06T12:00:00Z）
	identifier     string // 标识符
	modified       string // 修改时间（ISO 8601）
	revision       string // 修订版本
	language       string // 语言（zh-CN/en-US等）
	version        string // 版本号
}
