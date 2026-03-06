package main

import (
	"fmt"

	"github.com/miladystack/miladystack/pkg/excel"
)

func main() {
	ef := excel.NewFile()

	// 设置核心文档属性
	_ = ef.SetDocProps(
		excel.WithTitle("2026销售报表"),
		excel.WithCreator("张三"),
	)

	// 设置应用属性
	err := ef.SetAppProps(
		excel.WithApplication("miladystack-report"), // 应用名称
		excel.WithCompany("XX科技有限公司"),               // 公司名称
		excel.WithAppVersion("2.2026"),              // 版本号（符合XX.YYYY）
		excel.WithScaleCrop(true),                   // 缩略图缩放显示
		excel.WithDocSecurity(2),                    // 建议只读打开
		excel.WithLinksUpToDate(true),               // 超链接已更新
	)
	if err != nil {
		fmt.Printf("设置应用属性失败: %v\n", err)
		return
	}

	_ = ef.FreezePanes("Sheet1", 9, 0)
	_ = ef.Save("./report.xlsx")
	fmt.Println("Excel文件创建成功！")
}
