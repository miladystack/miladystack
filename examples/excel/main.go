package main

import (
	"fmt"

	"github.com/miladystack/miladystack/pkg/excel"
)

func main() {
	ef := excel.NewFile()

	// 测试1：冻结前9行、0列（最常用场景）
	// err := ef.FreezePanes("Sheet1", 9, 0)
	// 测试2：冻结前2行、3列（冻结前2行+前3列）
	err := ef.FreezePanes("Sheet1", 2, 3)
	if err != nil {
		fmt.Printf("冻结窗格失败：%v\n", err)
		return
	}

	_ = ef.Save("./freeze_test.xlsx")
	fmt.Println("文件生成成功，可打开验证冻结效果")
}
