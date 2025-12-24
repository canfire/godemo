package model

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
)

func NewChatModel(ctx context.Context) model.ToolCallingChatModel {
	tcm, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	// tcm, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
	// 	APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
	// 	Model:   "deepseek/deepseek-v3",
	// 	BaseURL: "https://api.yygu.cn/v3/llm.chat",
	// })
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}
	return tcm
}

func NewVLChatModel(ctx context.Context) model.ToolCallingChatModel {
	tcm, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		BaseURL:        "https://pilot.turingcm.com:18020/v1",
		Model:          "Qwen3-VL-8B-Instruct",
		EnableThinking: of(false),
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}
	return tcm
}

func of[T any](t T) *T {
	return &t
}
