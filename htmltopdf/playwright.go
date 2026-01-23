package htmltopdf

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/playwright-community/playwright-go"
	"github.com/skip2/go-qrcode"
)

func GetPDF(ctx context.Context, html string, qrContent string) (*PDFResult, error) {
	browser, err := GetBrowser(ctx)
	if err != nil {
		log.Fatalf("failed to get browser %v", err)
	}
	context, _ := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  794, // A4 @ 96 DPI
			Height: 1123,
		},
	})
	// userDataJSON := `{"userInfo":{"jwt_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjIjo0LCJ0IjoxNzY5MDY0NTYyLCJlIjoxNzcwNzkyNTYyLCJpIjoiZG90cGVuIiwiYSI6InVzZXIiLCJsIjo4fQ.JYnBGjR0Sq4aDhYRKUBpVRhfYnx_tErR-WR9Po4bdGA","uin":[{"uin":{"ID":4,"CreatedAt":"2026-01-08T17:30:31.807+08:00","UpdatedAt":"2026-01-13T10:40:32.992+08:00","DeletedAt":null,"Name":"测试前端123","UserID":2,"SubjectType":"company","SubjectID":1,"UinStatus":"normal","Issuer":"dotpen"},"company_name":"点笔教育","role":"sys_teacher","company_status":"passed"}],"user_info":{"identify":"1001","name":"测试前端","id":2,"created_at":"0001-01-01T00:00:00Z","uin":4}},"personalUserInfo":null}`

	userDataMap := map[string]interface{}{
		"userInfo": map[string]interface{}{
			"jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjIjo0LCJ0IjoxNzY5MDY0NTYyLCJlIjoxNzcwNzkyNTYyLCJpIjoiZG90cGVuIiwiYSI6InVzZXIiLCJsIjo4fQ.JYnBGjR0Sq4aDhYRKUBpVRhfYnx_tErR-WR9Po4bdGA",
		},
	}
	userDataJSONBytes, _ := json.Marshal(userDataMap)
	userDataJSON := string(userDataJSONBytes)

	// page, err := browser.NewPage(playwright.BrowserNewPageOptions{
	// 	Viewport: &playwright.Size{
	// 		Width:  794, // A4 @ 96 DPI
	// 		Height: 1123,
	// 	},
	// })
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("failed to create page %v", err)
	}
	_, _ = page.Goto("http://172.16.0.71:5173/preview/homework?id=32")

	_, err = page.Evaluate(fmt.Sprintf(`() => {localStorage.setItem("dotpen-yygu-user", '%s') }`, userDataJSON))
	if err != nil {
		log.Fatalf("failed to create page %v", err)
	}

	_, _ = page.Goto("http://172.16.0.71:5173/preview/homework?id=32", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})

	// time.Sleep(time.Hour)
	// Wait for MathJax to finish typesetting to avoid blank formulas in PDF
	if _, err := page.Evaluate(`() => {
    if (!window.MathJax) return null;
    if (MathJax.startup && MathJax.typesetPromise) {
      return MathJax.startup.promise.then(() => MathJax.typesetPromise());
    }
    if (MathJax.Hub && MathJax.Hub.Queue) {
      return new Promise((resolve, reject) => {
        try {
          MathJax.Hub.Queue(["Typeset", MathJax.Hub, () => resolve(null)]);
        } catch (e) {
          reject(e);
        }
      });
    }
    return null;
  }`); err != nil {
		return nil, fmt.Errorf("failed to wait for mathjax: %w", err)
	}

	if _, err := page.WaitForSelector("mjx-container", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(5000)}); err != nil {
		return nil, fmt.Errorf("mathjax not rendered: %w", err)
	}

	// 计算锚点时，统一基于实际打印布局（print media），并考虑滚动容器的偏移
	if err := page.EmulateMedia(playwright.PageEmulateMediaOptions{Media: playwright.MediaPrint}); err != nil {
		return nil, fmt.Errorf("failed to emulate print media: %w", err)
	}

	result, err := page.Evaluate(`() => {
	const res = [];
	const headerHeight = 70; // 页头高度（与 PDF Margin.Top 一致）
	const pageHeight = 1123 - headerHeight; // A4@96DPI 减去页头高度，实际内容区高度

	// 页面主体是一个 overflow-auto 的滚动容器，需把容器的滚动量计入绝对坐标
	const scrollContainer = document.querySelector('.overflow-auto')
		|| document.scrollingElement
		|| document.documentElement
		|| document.body;
	const containerRect = scrollContainer.getBoundingClientRect();
	const scrollTop = scrollContainer.scrollTop || window.pageYOffset || 0;
	const scrollLeft = scrollContainer.scrollLeft || window.pageXOffset || 0;

	// 获取所有锚点元素，并按原始Y坐标排序
	const elements = Array.from(document.querySelectorAll('[data-pdf-anchor]'))
		.map(el => {
			const rect = el.getBoundingClientRect();
			return {
				el,
				rect,
				originalTop: rect.top - containerRect.top + scrollTop,
				originalLeft: rect.left - containerRect.left + scrollLeft,
			};
		})
		.sort((a, b) => a.originalTop - b.originalTop);

	// 模拟分页过程，计算 break-inside-avoid 产生的累计偏移
	let cumulativeOffset = 0;

	elements.forEach(({el, rect, originalTop, originalLeft}) => {
		// 当前元素调整后的顶部位置
		const adjustedTop = originalTop + cumulativeOffset;
		// 在当前页内的位置
		const positionInPage = adjustedTop % pageHeight;
		// 元素底部是否会超出当前页
		const wouldCrossPage = positionInPage + rect.height > pageHeight;

		// 检查元素是否设置了 break-inside: avoid（Tailwind 的 break-inside-avoid 类）
		const style = getComputedStyle(el);
		const hasBreakInsideAvoid = style.breakInside === 'avoid' || 
									style.pageBreakInside === 'avoid';

		let finalTop = adjustedTop;

		// 如果元素会跨页且需要避免跨页，则推到下一页顶部
		if (wouldCrossPage && hasBreakInsideAvoid && positionInPage > 0) {
			const gapToFill = pageHeight - positionInPage;
			cumulativeOffset += gapToFill;
			finalTop = originalTop + cumulativeOffset;
		}

		const pageNumber = Math.floor(finalTop / pageHeight) + 1;
		// 页内 Y 坐标需要加上页头高度偏移
		const pageRelativeY = (finalTop % pageHeight) + headerHeight;

		res.push({
			name: el.dataset.pdfAnchor,
			page: pageNumber,
			x: originalLeft,
			y: pageRelativeY,
			width: rect.width,
			height: rect.height,
		});
	});

	return res;
}`)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate anchor script: %w", err)
	}

	// 解析锚点数据
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var anchors []Anchor
	if err := json.Unmarshal(jsonData, &anchors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal anchors: %w", err)
	}

	// 生成二维码 base64 图片
	qrBase64 := ""
	if qrContent != "" {
		qrPNG, err := qrcode.Encode(qrContent, qrcode.Medium, 60) // 60x60 像素
		if err != nil {
			return nil, fmt.Errorf("failed to generate qrcode: %w", err)
		}
		qrBase64 = base64.StdEncoding.EncodeToString(qrPNG)
	}

	// 构建 HeaderTemplate，左上角放二维码和页码
	headerTemplate := `<span></span>` // 默认空白
	if qrBase64 != "" {
		headerTemplate = fmt.Sprintf(`
			<div style="width: 100%%; padding: 8px 16px; display: flex; align-items: center; font-size: 10px; color: #666;">
				<img src="data:image/png;base64,%s" style="width: 50px; height: 50px; margin-right: 8px;" />
				<span>第 <span class="pageNumber"></span> 页 / 共 <span class="totalPages"></span> 页</span>
			</div>
		`, qrBase64)
	}

	// 生成 PDF
	pdfBytes, err := page.PDF(playwright.PagePdfOptions{
		Format:              playwright.String("A4"),
		Scale:               playwright.Float(1),
		PrintBackground:     playwright.Bool(true),
		DisplayHeaderFooter: playwright.Bool(true),
		HeaderTemplate:      playwright.String(headerTemplate),
		FooterTemplate:      playwright.String(`<span></span>`), // 空白底部
		Margin: &playwright.Margin{
			Top:    playwright.String("70px"), // 顶部留出二维码和页码空间
			Bottom: playwright.String("0"),
			Left:   playwright.String("0"),
			Right:  playwright.String("0"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return &PDFResult{
		PDFBytes: pdfBytes,
		Anchors:  anchors,
	}, nil
}

type Anchor struct {
	Name   string  `json:"name"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Page   int     `json:"page"`
}
type PDFResult struct {
	PDFBytes []byte
	Anchors  []Anchor
}

func GetBrowser(ctx context.Context) (playwright.Browser, error) {
	runOpt := &playwright.RunOptions{
		Browsers: []string{"chromium"},
		Verbose:  true,
	}
	err := playwright.Install(runOpt)
	if err != nil {
		log.Fatalf("failed to install playwright %v", err)
		return nil, err
	}
	println("Playwright installed successfully")
	pw, err := playwright.Run(runOpt)
	if err != nil {
		log.Fatalf("failed to run playwright %v", err)
		return nil, err
	}
	launchOptions := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // 展示浏览器界面
	}
	browser, err := pw.Chromium.Launch(launchOptions)
	if err != nil {
		log.Fatalf("failed to launch browser %v", err)
		return nil, err
	}
	return browser, nil
}

// SavePDF 保存 PDF 文件
func SavePDF(filename string, data []byte) error {
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to save pdf: %w", err)
	}
	return nil
}
