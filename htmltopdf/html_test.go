package htmltopdf

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
)

func TestPlayWright(t *testing.T) {
	ctx := context.Background()
	result, err := GetPDF(ctx, "", "asdjalsj")
	if err != nil {
		t.Fatalf("GetPDF failed: %v", err)
	}
	// 保存 PDF
	if err := SavePDF("contract.pdf", result.PDFBytes); err != nil {
		log.Fatalf("Error saving PDF: %v", err)
	}
	// 输出锚点信息
	fmt.Println("=== PDF 锚点信息 ===")
	for _, anchor := range result.Anchors {
		fmt.Printf("名称: %s\n", anchor.Name)
		fmt.Printf("  页码: %d\n", anchor.Page)
		fmt.Printf("  X坐标: %.2f px\n", anchor.X)
		fmt.Printf("  Y坐标: %.2f px (页内相对位置)\n", anchor.Y)
		fmt.Printf("  宽度: %.2f px\n", anchor.Width)
		fmt.Printf("  高度: %.2f px\n", anchor.Height)
		fmt.Println("---")
	}
	// 保存锚点信息到 JSON
	anchorsJSON, _ := json.MarshalIndent(result.Anchors, "", "  ")
	if err := os.WriteFile("anchors.json", anchorsJSON, 0644); err != nil {
		log.Fatalf("Error saving anchors: %v", err)
	}

	fmt.Println("✓ PDF 生成成功: contract.pdf")
	fmt.Println("✓ 锚点信息已保存: anchors.json")
}

func TestPhotoPDF(t *testing.T) {
	PhotoPDF()
}
