package main

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/adk"
	"github.com/eino-contrib/ollama/api"
)

func SequentialAgent() {
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

	analyzerAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "AnalyzerAgent",
		Description: "分析用户需求，提取关键信息",
		Instruction: "你是一个需求分析师，能够分析用户需求，提取关键信息",
		Model:       chatModel,
		OutputKey:   "analysis",
	})
	if err != nil {
		log.Fatalf("创建 AnalyzerAgent 失败: %v", err)
	}

	solutionAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SolutionAgent",
		Description: "根据分析结果提供解决方案",
		Instruction: "你是一个解决方案提供者，能够提供解决方案。可以使用{analysis}获取需求分析结果",
		Model:       chatModel,
	})
	if err != nil {
		log.Fatalf("创建 SolutionAgent 失败: %v", err)
	}

	sequentialAgent, err := adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
		Name:        "AnalyzerWorkFlow",
		Description: "需求分析和解决方案生成工作流",
		SubAgents: []adk.Agent{
			analyzerAgent,
			solutionAgent,
		},
	})
	if err != nil {
		log.Fatalf("创建 SequentialAgent 失败: %v", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           sequentialAgent,
		EnableStreaming: false,
	})

	query := "我想开发一个智能客服系统，需要支持多人对回知识库检索"
	iter := runner.Query(ctx, query)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatalf("Query failed: %v", event.Err)
		}
		log.Printf("event: %v\n", event)

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg := event.Output.MessageOutput.Message
			if msg != nil {
				log.Printf("助手: %v\n", msg.Content)
			}
		}
	}
}
