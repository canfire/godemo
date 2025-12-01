package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func main() {
	fstring()
}

func fstring() {
	// 1. 创建上下文
	ctx := context.Background()

	temp := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个{role}"),
		schema.UserMessage("{question}"),
	)

	v := map[string]interface{}{
		"role":     "AI 助手",
		"question": "你好，今天西安天气怎么样",
	}
	msg, err := temp.Format(ctx, v)
	if err != nil {
		log.Fatalf("格式化消息失败: %v", err)
	}

	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:8b",
	})

	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 4. 调用模型生成响应
	response, err := chatModel.Stream(ctx, msg)
	if err != nil {
		log.Fatalf("生成响应失败: %v", err)
	}
	defer response.Close()
	// c := &schema.Message{}

	for {
		chunk, err := response.Recv()
		if err == io.EOF {
			// fmt.Printf("\n %+v", c)
			// fmt.Printf("\n --------- %+v", c.ResponseMeta)
			// fmt.Printf("\n -------------------- %+v", c.ResponseMeta.Usage)
			break
		}
		if err != nil {
			log.Fatalf("接收响应失败: %v", err)
		}
		fmt.Print(chunk.ReasoningContent)
		fmt.Print(chunk.Content)
		// c = chunk
	}

}
