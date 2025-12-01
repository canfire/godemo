package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// 创建多个并行节点，合并结果

func ParallerDemo() {
	// 1. 创建上下文
	ctx := context.Background()
	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:8b",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	paraller := compose.NewParallel()

	paraller.AddLambda("keywords", compose.InvokableLambda(
		func(ctx context.Context, input map[string]any) (string, error) {
			fmt.Println("并行任务1,提取关键词")
			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("请提取文本中的关键词，用逗号分隔"),
				schema.UserMessage("{text}"),
			)
			messages, _ := template.Format(ctx, input)
			resp, err := chatModel.Generate(ctx, messages)
			if err != nil {
				return "", err
			}
			return resp.Content, nil
		},
	))

	paraller.AddLambda("sentiment", compose.InvokableLambda(
		func(ctx context.Context, input map[string]any) (string, error) {
			fmt.Println("并行任务2,情感分析")
			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("清分析文本的情感倾向（正面/中性/负面）。"),
				schema.UserMessage("{text}"),
			)
			messages, _ := template.Format(ctx, input)
			resp, err := chatModel.Generate(ctx, messages)
			if err != nil {
				return "", err
			}
			return resp.Content, nil
		},
	))

	paraller.AddLambda("summary", compose.InvokableLambda(
		func(ctx context.Context, input map[string]any) (string, error) {
			fmt.Println("并行任务3,生成摘要")
			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("请用一句话总结文本内容"),
				schema.UserMessage("{text}"),
			)
			messages, _ := template.Format(ctx, input)
			resp, err := chatModel.Generate(ctx, messages)
			if err != nil {
				return "", err
			}
			return resp.Content, nil
		},
	))

	chain := compose.NewChain[string, map[string]any]()

	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, text string) (map[string]any, error) {
		return map[string]any{
			"text": text,
		}, nil
	})).
		AppendParallel(paraller).
		// 处理并行结果
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			fmt.Printf("并行结果: %v\n", input)
			return input, nil
		}))

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译链条失败: %v", err)
	}
	text := "最后，保持回答友好和帮助性，确保用户知道如何获取所需信息，并邀请他们提出更多问题，以提供更全面的帮助"
	out, err := runnable.Invoke(ctx, text)
	if err != nil {
		log.Fatalf("生成响应失败: %v", err)
	}
	fmt.Printf("out: %+v\n", out)
}
