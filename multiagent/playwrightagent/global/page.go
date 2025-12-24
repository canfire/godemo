package global

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/playwright-community/playwright-go"
)

var (
	gbrowser playwright.Browser
	gpage    playwright.Page
)

func GetBrowser(ctx context.Context) (playwright.Browser, error) {
	if gbrowser != nil {
		return gbrowser, nil
	}
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
	gbrowser = browser
	return browser, nil
}

func GetPage(ctx context.Context) (playwright.Page, error) {
	if gpage != nil {
		return gpage, nil
	}
	browser, err := GetBrowser(ctx)
	if err != nil {
		return nil, err
	}
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1620,
			Height: 980,
		},
		NoViewport: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("failed to create new browser context %v", err)
		return nil, err
	}
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("failed to create new page %v", err)
		return nil, err
	}
	_, err = page.Goto("https://www.baidu.com")
	if err != nil {
		log.Fatalf("failed to navigate to login page: %v", err)
	}
	gpage = page
	return page, nil
}

// GetScreenshotBase64 获取页面截图的base64编码
func GetScreenshotBase64(page playwright.Page) (string, error) {
	imgbyte, err := page.Screenshot()
	if err != nil {
		log.Fatalf("failed to get screenshot: %v", err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(imgbyte), nil
}

// GetScreenshotByte 获取页面截图的base64编码
func GetScreenshotByte(page playwright.Page) ([]byte, error) {
	imgbyte, err := page.Screenshot()
	if err != nil {
		log.Fatalf("failed to get screenshot: %v", err)
		return nil, err
	}
	return imgbyte, nil
}
