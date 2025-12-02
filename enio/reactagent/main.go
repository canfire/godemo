package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/ollama/api"
)

func main() {
	ReactDemo()
}

func ReactDemo() {
	ctx := context.Background()
	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:8b",
		Thinking: &api.ThinkValue{
			Value: false,
		},
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	timeTool := NewTimeTool()
	cu := &CalculatorTool{}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{
				timeTool,
				cu,
			},
		},
	})
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	msg := []*schema.Message{
		schema.UserMessage("现在几点了"),
	}

	res, err := agent.Generate(ctx, msg)
	if err != nil {
		log.Fatalf("Generate failed: %v", err)
	}
	fmt.Printf("res: %v\n", res.Content)
}
