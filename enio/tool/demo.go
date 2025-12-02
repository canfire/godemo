package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/ollama/api"
)

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
	timeTool := NewTimeTool()
	cu := &CalculatorTool{}
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{
			timeTool,
			cu,
		},
	})
	if err != nil {
		log.Fatalf("创建 ToolsNode 失败: %v", err)
	}
	cuInfo, _ := cu.Info(ctx)
	timeinfo, _ := timeTool.Info(ctx)
	toolsInfo := []*schema.ToolInfo{
		timeinfo,
		cuInfo,
	}
	testCase := []string{
		"几点了",
		"100+100等于多少",
		"几点了,并且告诉我100+100等于多少",
		// "介绍下你自己",
	}

	for _, v := range testCase {
		msg := []*schema.Message{
			schema.UserMessage(v),
		}
		res, err := chatModel.Generate(ctx, msg,
			model.WithTools(toolsInfo),
		)
		if err != nil {
			log.Fatalf("生成响应失败: %v", err)
		}
		if len(res.ToolCalls) > 0 {
			fmt.Println("=========模型返回================", res.Content)
			for _, a := range res.ToolCalls {
				fmt.Println("========调用工具=========\n", a.Function.Name)
				fmt.Println("==========参数===========\n", a.Function.Arguments)
			}
			tres, _ := toolsNode.Invoke(ctx, res)
			fmt.Println("========调用结果==========\n", tres)
			fmt.Println("========第二次回答==========")
			msg := []*schema.Message{
				schema.UserMessage(v),
				schema.AssistantMessage(res.Content, res.ToolCalls),
			}
			for _, as := range tres {
				msg = append(msg, schema.ToolMessage(as.Content, res.ToolCallID))
			}
			stream, err := chatModel.Stream(ctx, msg)
			if err != nil {
				log.Fatalf("生成响应失败: %v", err)
			}
			defer stream.Close()
			for {
				chunk, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalf("接收响应失败: %v", err)
				}
				fmt.Print(chunk.ReasoningContent)
				fmt.Print(chunk.Content)
			}
			fmt.Println("")
		} else {
			fmt.Println("========直接回答==========", res)
		}
	}

}
