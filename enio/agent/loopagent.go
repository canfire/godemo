package main

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/adk"
	"github.com/eino-contrib/ollama/api"
)

func LoopAgent() {
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

	// 创建两个agent
	// 主任务解决
	mainAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "MainAgent",
		Description: "负责生成初步解决方案",
		Instruction: "你是一个问题解决专家，请根据问题生成详细的解决方案。如果解决方案需要改进，请说明需要改进的地方",
		Model:       chatModel,
		OutputKey:   "solution",
	})
	if err != nil {
		log.Fatalf("创建 MainAgent 失败: %v", err)
	}

	critiqueAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "CritiqueAgent",
		Description: "对解决方案进行批判和反馈",
		Instruction: `你是一个质量审查专家。请审查解决方案的质量，提供改进建议。
如果解决方案已经足够好请说明："已达到预期效果, 无需改进"。
可以使用{solution}获取当前解决方案。`,
		Model:     chatModel,
		OutputKey: "critique",
	})
	if err != nil {
		log.Fatalf("创建 MainAgent 失败: %v", err)
	}

	loopAgent, err := adk.NewLoopAgent(ctx, &adk.LoopAgentConfig{
		Name:        "RefletionAgent",
		Description: "迭代反思智能体，通过多轮迭代优化解决方案",
		SubAgents: []adk.Agent{
			mainAgent,
			critiqueAgent,
		},
		MaxIterations: 5, // 最多循环5次
	})
	if err != nil {
		log.Fatalf("创建 LoopAgent 失败: %v", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           loopAgent,
		EnableStreaming: false,
	})

	query := "如何设计一个高性能的分布式缓存系统"
	iter := runner.Query(ctx, query)
	i := 0
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
				if event.AgentName == "MainAgent" {
					i++
					log.Printf("== 第%v次迭代 ==生成解决方案===", i)
				} else if event.AgentName == "CritiqueAgent" {
					log.Printf("== 第%v次迭代 == 批判反馈 ===", i)
				}
				log.Printf("助手%v: %v\n", event.AgentName, msg.Content)
			}
		}
	}
}
