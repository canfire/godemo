package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/adk"
	"github.com/eino-contrib/ollama/api"
)

func main() {
	Demo()
}

func Demo() {
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

	// 创建多个专业agent
	generalagent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "GeneralAssistant",
		Description: "通用助理，可以处理各种问题，也可以将任务转移给专业 Agent",
		Instruction: `你是一个通用助手。你可以：
		1.直接回答简单问题
		2.将复杂的技术问题转移给 TechExpert
		3.将数学问题转移给 MathExpert`,
		Model: chatModel,
	})
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	techAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TechExpert",
		Description: "技术专家，专门处理编程和技术问题",
		Instruction: "你是一个技术专家，请详细解答编程和技术相关的问题",
		Model:       chatModel,
	})
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	mathAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "MathExpert",
		Description: "数学专家，专门处理数学问题",
		Instruction: "你是一个数学专家，请详细解答数学相关的问题",
		Model:       chatModel,
	})
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	generalAgentWithSubs, err := adk.SetSubAgents(ctx, generalagent, []adk.Agent{techAgent, mathAgent})
	if err != nil {
		log.Fatalf("设置子 Agent 失败: %v", err)
	}

	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           generalAgentWithSubs,
		EnableStreaming: false,
		// CheckPointStore: store,
	})

	ques := []string{
		"你好，今天天气怎么样",
		"Go语言如何实现并发",
		"如何计算圆的面积",
	}

	for _, q := range ques {
		fmt.Println("用户问题======================", q)
		iter := r.Query(ctx, q)
		for {
			event, ok := iter.Next()
			if !ok {
				fmt.Println("======================")
				break
			}
			if event.Err != nil {
				log.Fatalf("Query failed: %v", event.Err)
			}
			// log.Printf("event: %v", event)
			if event.Output != nil && event.Output.MessageOutput != nil {
				msg := event.Output.MessageOutput.Message
				if msg != nil {
					if len(msg.ToolCalls) > 0 {
						for _, tc := range msg.ToolCalls {
							if tc.Function.Name == "transfer_to_agent" {
								log.Printf("任务转移: %s -> %s\n", event.AgentName, tc.Function.Arguments)
							}
						}
					} else if msg.Content != "" {
						log.Printf("助手[%s]: %s\n", event.AgentName, msg.Content)
					}
				}
			}
		}
	}

}
