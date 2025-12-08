package main

import (
	"context"
	"log"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/ollama/api"
)

var _ compose.CheckPointStore = (*MyCheckPointStore)(nil)

type MyCheckPointStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMyCheckPointStore() *MyCheckPointStore {
	return &MyCheckPointStore{
		data: make(map[string][]byte),
	}
}

func (m *MyCheckPointStore) Get(ctx context.Context, checkPointKey string) ([]byte, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.data[checkPointKey]
	return data, ok, nil
}

func (m *MyCheckPointStore) Set(ctx context.Context, checkPointKey string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[checkPointKey] = data
	return nil
}

func main() {
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
		Name:        "BookRecommender",
		Description: "书籍推荐 Agent",
		Instruction: "你是一个书籍推荐专家，请根据用户需求推荐合适的书籍",
		Model:       chatModel,
	})
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	store := NewMyCheckPointStore()

	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
		CheckPointStore: store,
	})

	checkPointStoreID := "session_001"
	query := "我想学习golang，请推荐我一些书籍"

	iter := r.Run(ctx, []adk.Message{
		schema.UserMessage(query),
	}, adk.WithCheckPointID(checkPointStoreID))

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if err != nil {
			log.Fatalf("接收响应失败: %v", err)
		}
		log.Printf("event: %v", event)

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg := event.Output.MessageOutput.Message
			if msg != nil {
				log.Printf("助手: %v\n", msg.Content)
			}
		}

	}

	// resunmeIter, err := r.Resume(ctx, checkPointStoreID)
	// if err != nil {
	// 	log.Fatalf("恢复会话失败: %v", err)
	// }
	// ... 恢复处理函数
}
