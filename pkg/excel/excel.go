package excel

import (
	"fmt"
	"os/user"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelFile 封装excelize.File
type ExcelFile struct {
	file *excelize.File
}

// -------------------------- 1. 核心文档属性（DocProps）相关定义 --------------------------
type docPropsConfig struct {
	title          string
	subject        string
	creator        string
	keywords       string
	description    string
	lastModifiedBy string
	category       string
	contentStatus  string
	created        string
	identifier     string
	modified       string
	revision       string
	language       string
	version        string
}

type DocPropOption func(*docPropsConfig)

// 预定义DocProps选项函数
func WithTitle(title string) DocPropOption {
	return func(c *docPropsConfig) { c.title = title }
}

func WithSubject(subject string) DocPropOption {
	return func(c *docPropsConfig) { c.subject = subject }
}

func WithCreator(creator string) DocPropOption {
	return func(c *docPropsConfig) { c.creator = creator }
}

func WithKeywords(keywords string) DocPropOption {
	return func(c *docPropsConfig) { c.keywords = keywords }
}

func WithDescription(desc string) DocPropOption {
	return func(c *docPropsConfig) { c.description = desc }
}

func WithLastModifiedBy(user string) DocPropOption {
	return func(c *docPropsConfig) { c.lastModifiedBy = user }
}

func WithCategory(cate string) DocPropOption {
	return func(c *docPropsConfig) { c.category = cate }
}

// -------------------------- 2. 应用属性（AppProps）相关定义--------------------------
// appPropsConfig 内部应用属性配置
type appPropsConfig struct {
	application       string // 创建文档的应用程序名称
	scaleCrop         bool   // 文档缩略图显示方式（true=缩放，false=剪裁）
	docSecurity       int    // 文档安全级别（1-4）
	company           string // 关联公司名称
	linksUpToDate     bool   // 超链接是否最新
	hyperlinksChanged bool   // 是否需要更新超链接关系
	appVersion        string // 应用版本（格式：XX.YYYY）
}

// AppPropOption 应用属性选项函数类型
type AppPropOption func(*appPropsConfig)

// 预定义AppProps选项函数
func WithApplication(appName string) AppPropOption {
	return func(c *appPropsConfig) { c.application = appName }
}

func WithScaleCrop(scale bool) AppPropOption {
	return func(c *appPropsConfig) { c.scaleCrop = scale }
}

// WithDocSecurity 设置文档安全级别（有效值1-4）
// 1-文档受密码保护
// 2-建议以只读方式打开
// 3-强制以只读方式打开
// 4-文档批注被锁定
func WithDocSecurity(level int) AppPropOption {
	return func(c *appPropsConfig) {
		// 简单校验安全级别范围
		if level < 1 || level > 4 {
			level = 0 // 非法值则设为0（表示无安全设置）
		}
		c.docSecurity = level
	}
}

func WithCompany(company string) AppPropOption {
	return func(c *appPropsConfig) { c.company = company }
}

func WithLinksUpToDate(isUpToDate bool) AppPropOption {
	return func(c *appPropsConfig) { c.linksUpToDate = isUpToDate }
}

func WithHyperlinksChanged(changed bool) AppPropOption {
	return func(c *appPropsConfig) { c.hyperlinksChanged = changed }
}

// WithAppVersion 设置应用版本（格式建议为 XX.YYYY，如 "1.2025"）
func WithAppVersion(version string) AppPropOption {
	return func(c *appPropsConfig) { c.appVersion = version }
}

// -------------------------- 3. 核心方法实现 --------------------------
// NewFile 创建新Excel文件，自动设置默认文档属性
func NewFile() *ExcelFile {
	ef := &ExcelFile{
		file: excelize.NewFile(),
	}
	// 自动设置默认核心文档属性
	_ = ef.SetDocProps()
	// 自动设置默认应用属性
	_ = ef.SetAppProps()
	return ef
}

// OpenFile 打开已有文件，不修改原有属性
func OpenFile(path string) (*ExcelFile, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %w", err)
	}
	return &ExcelFile{file: f}, nil
}

// SetDocProps 设置核心文档属性（含默认值）
func (e *ExcelFile) SetDocProps(opts ...DocPropOption) error {
	// 1. 初始化默认配置
	config := &docPropsConfig{
		creator:        "miladystack",
		language:       "zh-CN",
		contentStatus:  "已完成",
		created:        time.Now().Format(time.RFC3339),
		lastModifiedBy: "miladystack",
	}

	// 2. 优先用系统用户名覆盖
	if u, err := user.Current(); err == nil {
		config.creator = u.Username
		config.lastModifiedBy = u.Username
	}

	// 3. 应用用户选项
	for _, opt := range opts {
		opt(config)
	}

	// 4. 映射到原生配置
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

	if err := e.file.SetDocProps(excelProps); err != nil {
		return fmt.Errorf("设置核心文档属性失败: %w", err)
	}
	return nil
}

// SetAppProps 设置应用属性
func (e *ExcelFile) SetAppProps(opts ...AppPropOption) error {
	// 1. 初始化默认配置
	config := &appPropsConfig{
		application: "miladystack-excel", // 默认应用名称
		company:     "miladystack",       // 默认公司名称
		appVersion:  "1.0000",            // 默认版本（符合XX.YYYY格式）
		scaleCrop:   false,               // 默认缩略图剪裁显示
		docSecurity: 0,                   // 默认无安全设置
	}

	// 2. 应用用户传入的选项（覆盖默认值）
	for _, opt := range opts {
		opt(config)
	}

	// 3. 映射到excelize原生AppProperties
	excelAppProps := &excelize.AppProperties{
		Application:       config.application,
		ScaleCrop:         config.scaleCrop,
		DocSecurity:       config.docSecurity,
		Company:           config.company,
		LinksUpToDate:     config.linksUpToDate,
		HyperlinksChanged: config.hyperlinksChanged,
		AppVersion:        config.appVersion,
	}

	// 4. 调用原生SetAppProps方法
	if err := e.file.SetAppProps(excelAppProps); err != nil {
		return fmt.Errorf("设置应用属性失败: %w", err)
	}
	return nil
}

// FreezePanes 冻结窗格
func (e *ExcelFile) FreezePanes(sheetName string, freezeRows int, freezeCols int) error {
	if freezeRows < 0 || freezeCols < 0 {
		return fmt.Errorf("冻结行数/列数不能为负数")
	}

	colName, err := excelize.ColumnNumberToName(freezeCols + 1)
	if err != nil {
		return fmt.Errorf("列数转换失败: %w", err)
	}
	topLeftCell := fmt.Sprintf("%s%d", colName, freezeRows+1)

	return e.file.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      freezeCols,
		YSplit:      freezeRows,
		TopLeftCell: topLeftCell,
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{
			{SQRef: topLeftCell, ActiveCell: topLeftCell, Pane: "bottomLeft"},
		},
	})
}

// Save 保存文件
func (e *ExcelFile) Save(path string) error {
	defer e.file.Close()
	if err := e.file.SaveAs(path); err != nil {
		return fmt.Errorf("保存Excel文件失败: %w", err)
	}
	return nil
}
