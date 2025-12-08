package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func main() {
	// chainTest()
	// chainStreamTest()
	// LambdaDemo()
	// ParallerDemo()
	// BranchDemo()
	CallBackDemo()
}

func chainTest() {
	// 1. 创建上下文
	ctx := context.Background()

	temp := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个{role}"),
		schema.UserMessage("{question}"),
	)

	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:8b",
	})

	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 输入 map[string]any ，输出 *schema.Message
	chain := compose.NewChain[map[string]any, *schema.Message]()

	chain.AppendChatTemplate(temp).AppendChatModel(chatModel)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译链条失败: %v", err)
	}
	v := map[string]interface{}{
		"role":     "AI 助手",
		"question": "你好，今天西安天气怎么样",
	}
	out, err := runnable.Invoke(ctx, v)
	if err != nil {
		log.Fatalf("生成响应失败: %v", err)
	}
	fmt.Printf("out: %v\n", out.Content)
}

func chainStreamTest() {
	// 1. 创建上下文
	ctx := context.Background()

	temp := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个{role}"),
		schema.UserMessage("{question}"),
	)

	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:8b",
	})

	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 输入 map[string]any ，输出 *schema.Message
	chain := compose.NewChain[map[string]any, *schema.Message]()

	chain.AppendChatTemplate(temp).AppendChatModel(chatModel)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译链条失败: %v", err)
	}
	v := map[string]interface{}{
		"role":     "AI 助手",
		"question": "你好，今天西安天气怎么样",
	}
	out, err := runnable.Stream(ctx, v)
	if err != nil {
		log.Fatalf("生成响应失败: %v", err)
	}
	defer out.Close()
	for {
		chunk, err := out.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("接收响应失败: %v", err)
		}
		fmt.Print(chunk.ReasoningContent)
		fmt.Print(chunk.Content)
	}
	fmt.Println()
}
