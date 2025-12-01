package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/ollama/api"
)

func Ollama() {
	// 1. 创建上下文
	ctx := context.Background()

	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:8b",
	})

	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 3. 准备消息
	messages := []*schema.Message{
		schema.SystemMessage("你是一个友好的 AI 助手"),
		schema.UserMessage("你好，请介绍一下 Eino 框架"),
	}

	// 4. 调用模型生成响应
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成响应失败: %v", err)
	}

	// 5. 输出结果
	fmt.Printf("AI 思考: %s\n", response.ReasoningContent)
	fmt.Printf("AI 响应: %s\n", response.Content)

	// 6. 输出 token 使用情况
	if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		fmt.Printf("\nToken 使用统计:\n")
		fmt.Printf("  输入 Token: %d\n", response.ResponseMeta.Usage.PromptTokens)
		fmt.Printf("  输出 Token: %d\n", response.ResponseMeta.Usage.CompletionTokens)
		fmt.Printf("  总计 Token: %d\n", response.ResponseMeta.Usage.TotalTokens)
	}
}

func OllamaStream() {
	// 1. 创建上下文
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

	// 3. 准备消息
	messages := []*schema.Message{
		schema.SystemMessage("你是一个友好的 AI 助手"),
		schema.UserMessage("你好，请介绍一下golang chan 的用法"),
	}

	// 4. 调用模型生成响应
	response, err := chatModel.Stream(ctx, messages)
	if err != nil {
		log.Fatalf("生成响应失败: %v", err)
	}
	defer response.Close()
	c := &schema.Message{}

	for {
		chunk, err := response.Recv()
		if err == io.EOF {
			fmt.Printf("\n %+v", c)
			fmt.Printf("\n --------- %+v", c.ResponseMeta)
			fmt.Printf("\n -------------------- %+v", c.ResponseMeta.Usage)
			break
		}
		if err != nil {
			log.Fatalf("接收响应失败: %v", err)
		}
		fmt.Print(chunk.ReasoningContent)
		fmt.Print(chunk.Content)
		c = chunk
	}

}
