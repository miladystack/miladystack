package excel

import (
	"fmt"
	"os/user"
	"time"

	"github.com/xuri/excelize/v2"
)

// -------------------------- 1. 定义选项类型 --------------------------
// DocPropOption 文档属性配置选项类型
type DocPropOption func(*docPropsConfig)

// -------------------------- 2. 预定义便捷选项函数 --------------------------
// WithTitle 设置文档标题
func WithTitle(title string) DocPropOption {
	return func(c *docPropsConfig) {
		c.title = title
	}
}

// WithSubject 设置文档主题
func WithSubject(subject string) DocPropOption {
	return func(c *docPropsConfig) {
		c.subject = subject
	}
}

// WithCreator 设置创建者（作者），未传则默认系统用户名
func WithCreator(creator string) DocPropOption {
	return func(c *docPropsConfig) {
		c.creator = creator
	}
}

// WithKeywords 设置关键词（多个用逗号分隔）
func WithKeywords(keywords string) DocPropOption {
	return func(c *docPropsConfig) {
		c.keywords = keywords
	}
}

// WithDescription 设置文档描述/备注
func WithDescription(description string) DocPropOption {
	return func(c *docPropsConfig) {
		c.description = description
	}
}

// WithLastModifiedBy 设置最后修改者
func WithLastModifiedBy(lastModifiedBy string) DocPropOption {
	return func(c *docPropsConfig) {
		c.lastModifiedBy = lastModifiedBy
	}
}

// WithCategory 设置文档分类
func WithCategory(category string) DocPropOption {
	return func(c *docPropsConfig) {
		c.category = category
	}
}

// WithContentStatus 设置内容状态（如 "草稿"、"已发布"、"审核中"）
func WithContentStatus(status string) DocPropOption {
	return func(c *docPropsConfig) {
		c.contentStatus = status
	}
}

// WithCreated 设置创建时间（自动转为ISO 8601格式）
func WithCreated(t time.Time) DocPropOption {
	return func(c *docPropsConfig) {
		c.created = t.Format(time.RFC3339) // ISO 8601标准格式
	}
}

// WithLanguage 设置文档语言（如 "zh-CN"、"en-US"）
func WithLanguage(lang string) DocPropOption {
	return func(c *docPropsConfig) {
		c.language = lang
	}
}

// 可按需扩展：WithIdentifier、WithModified、WithRevision、WithVersion 等

// -------------------------- 3. 核心方法实现 --------------------------
// NewFile 创建新的Excel文件实例
func NewFile() *ExcelFile {
	return &ExcelFile{
		file: excelize.NewFile(),
	}
}

// OpenFile 打开已存在的Excel文件
func OpenFile(path string) (*ExcelFile, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %w", err)
	}
	return &ExcelFile{file: f}, nil
}

// SetDocProps 设置Excel文档属性
func (e *ExcelFile) SetDocProps(opts ...DocPropOption) error {
	// 1. 初始化默认配置
	config := &docPropsConfig{
		creator:       "miladystack",                   // 创作者兜底默认值
		language:      "zh-CN",                         // 默认语言：中文
		contentStatus: "已完成",                           // 默认内容状态
		created:       time.Now().Format(time.RFC3339), // 默认创建时间：当前时间
	}

	// 2. 自动填充创作者默认值（系统用户名）
	if u, err := user.Current(); err == nil {
		config.creator = u.Username
	}

	// 3. 应用所有传入的选项函数（覆盖默认值）
	for _, opt := range opts {
		opt(config)
	}

	// 4. 严格映射到excelize原生DocProperties
	excelProps := &excelize.DocProperties{
		Title:          config.title,
		Subject:        config.subject,
		Creator:        config.creator,
		Keywords:       config.keywords,
		Description:    config.description,
		LastModifiedBy: config.lastModifiedBy,
		Category:       config.category,
		ContentStatus:  config.contentStatus,
		Created:        config.created,
		Identifier:     config.identifier,
		Modified:       config.modified,
		Revision:       config.revision,
		Language:       config.language,
		Version:        config.version,
	}

	// 5. 调用官方SetDocProps方法
	if err := e.file.SetDocProps(excelProps); err != nil {
		return fmt.Errorf("设置Excel文档属性失败: %w", err)
	}
	return nil
}

// Save 保存Excel文件
func (e *ExcelFile) Save(path string) error {
	defer e.file.Close()
	if err := e.file.SaveAs(path); err != nil {
		return fmt.Errorf("保存Excel文件失败: %w", err)
	}
	return nil
}

// FreezePanes 冻结窗格
// sheetName: 工作表名
// freezeRows: 要冻结的行数（如9表示冻结前9行）
// freezeCols: 要冻结的列数（如2表示冻结前2列）
func (e *ExcelFile) FreezePanes(sheetName string, freezeRows int, freezeCols int) error {
	// 校验参数合法性
	if freezeRows < 0 || freezeCols < 0 {
		return fmt.Errorf("冻结行数/列数不能为负数")
	}

	// v2版本用ColumnNumberToName，且列数从1开始
	// 例如：freezeCols=0 → 列1 → A；freezeCols=2 → 列3 → C
	colName, err := excelize.ColumnNumberToName(freezeCols + 1)
	if err != nil {
		return fmt.Errorf("列数转换失败: %w", err)
	}
	// 计算冻结后窗格的左上角单元格（如冻结9行0列 → A10）
	topLeftCell := fmt.Sprintf("%s%d", colName, freezeRows+1)

	// 调用v2的SetPanes方法
	return e.file.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      freezeCols,   // 列拆分位置
		YSplit:      freezeRows,   // 行拆分位置
		TopLeftCell: topLeftCell,  // 修正后的左上角单元格
		ActivePane:  "bottomLeft", // 激活左下窗格（冻结后的可编辑区域）
		Selection: []excelize.Selection{
			{SQRef: topLeftCell, ActiveCell: topLeftCell, Pane: "bottomLeft"},
		},
	})
}
