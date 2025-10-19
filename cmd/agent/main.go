package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hang-ma/go-browser-agent/internal/agent"
	"github.com/hang-ma/go-browser-agent/internal/browser"
	"github.com/hang-ma/go-browser-agent/internal/llm"
)

func main() {
	ctx := context.Background()

	model := flag.String("model", "tngtech/deepseek-r1t2-chimera:free", "OpenRouter model (e.g., tngtech/deepseek-r1t2-chimera:free)")
	//model := flag.String("model", "gpt-4o-mini", "OpenAI model (e.g., gpt-4o-mini)")
	headless := flag.Bool("headless", false, "Run browser headless")
	slowmo := flag.Int("slowmo", 100, "SlowMo in ms for Playwright")
	flag.Parse()

	apiKey := ""
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY is not set")
		return
	}

	bw, err := browser.NewPlaywright(ctx, browser.Config{
		UserDataDir: "user_data",
		Headless:    *headless,
		SlowMo:      *slowmo,
	})
	if err != nil {
		panic(err)
	}
	defer bw.Close()

	llmClient := llm.NewOpenRouterClient(*model)
	if err != nil {
		panic(err)
	}

	core := agent.New(agent.Config{
		Browser:   bw,
		LLM:       llmClient,
		Artifacts: "run_artifacts",
	})

	fmt.Println("Агент запущен. Введите задачу (или 'exit' для выхода).")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if strings.EqualFold(text, "exit") {
			fmt.Println("Выход.")
			return
		}

		if err := core.Run(ctx, text, 12*time.Minute); err != nil {
			fmt.Println("Ошибка выполнения:", err)
		}
	}
}
