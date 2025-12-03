package main

import (
	"context"
	"log"

	"github.com/canfire/godemo/enio/agent/tools"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/eino-contrib/ollama/api"
)

func AgentChatModel() {
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

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SimpleAssistant",
		Description: "一个简单的助手Agent，能够回答用户问题",
		Instruction: "你是一个友好的助手，能够回答用户问题", // 指令
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{}, // 不使用工具
	})
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})
	query := "什么是golang"
	iter := runner.Query(ctx, query)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatalf("Query failed: %v", event.Err)
		}
		log.Printf("event: %v", event)

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg := event.Output.MessageOutput.Message
			if msg != nil {
				log.Printf("助手: %v\n", msg.Content)
			}
		}
	}
}

func AgentChatModelWithTools() {
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

	timeTool := tools.NewTimeTool()

	calculatorTool := &tools.CalculatorTool{}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ToolAssistant",
		Description: "可以使用工具的助手Agent",
		Instruction: "你是一个智能助手，可以调用工具帮用户解决问题", // 指令
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{
					timeTool,
					calculatorTool,
				},
			},
		},
		MaxIterations: 10, // 最多执行10工具调用
	})
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})

	queries := []string{
		"现在几点了",
		"帮我计算123 + 456 * 789",
	}

	for _, v := range queries {
		println("用户问题======================", v)
		iter := runner.Query(ctx, v)
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}
			if event.Err != nil {
				log.Fatalf("Query failed: %v", event.Err)
			}
			log.Printf("event: %v", event)

			if event.Output != nil && event.Output.MessageOutput != nil {
				msg := event.Output.MessageOutput.Message
				if msg != nil {
					log.Printf("助手: %v\n", msg.Content)
				}
			}
		}
	}
}
