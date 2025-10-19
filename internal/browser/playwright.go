package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

type Config struct {
	UserDataDir string
	Headless    bool
	SlowMo      int
}

type Browser struct {
	pw     *pw.Playwright
	br     pw.Browser
	ctx    pw.BrowserContext
	page   pw.Page
	config Config
}

func NewPlaywright(ctx context.Context, cfg Config) (*Browser, error) {
	// playwright browsers should be installed via `npx playwright install chromium`
	pwrt, err := pw.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright run: %w", err)
	}

	launch, err := pwrt.Chromium.Launch(pw.BrowserTypeLaunchOptions{
		Headless: pw.Bool(cfg.Headless),
		SlowMo:   pw.Float(float64(cfg.SlowMo)),
	})
	if err != nil {
		return nil, fmt.Errorf("launch: %w", err)
	}

	_ = os.MkdirAll(cfg.UserDataDir, 0o755)
	userDir, _ := filepath.Abs(cfg.UserDataDir)
	_ = userDir

	bctx, err := launch.NewContext(pw.BrowserNewContextOptions{
		Viewport: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("context: %w", err)
	}

	page, err := bctx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("page: %w", err)
	}

	page.SetDefaultTimeout(30000)
	page.SetDefaultNavigationTimeout(45000)

	return &Browser{pw: pwrt, br: launch, ctx: bctx, page: page, config: cfg}, nil
}

func (b *Browser) Page() pw.Page { return b.page }

func (b *Browser) Screenshot(path string) error {
	_, err := b.page.Screenshot(pw.PageScreenshotOptions{Path: pw.String(path), FullPage: pw.Bool(true)})
	return err
}

func (b *Browser) DumpHTML(path string) error {
	content, err := b.page.Content()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func (b *Browser) Goto(url string) error {
	_, err := b.page.Goto(url, pw.PageGotoOptions{WaitUntil: pw.WaitUntilStateDomcontentloaded})
	return err
}

func (b *Browser) WaitIdle() { time.Sleep(800 * time.Millisecond) }

func (b *Browser) Close() {
	if b.ctx != nil {
		_ = b.ctx.Close()
	}
	if b.br != nil {
		_ = b.br.Close()
	}
	if b.pw != nil {
		_ = b.pw.Stop()
	}
}
