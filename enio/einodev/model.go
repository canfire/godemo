package einodev

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/ollama/api"
)

type ChatModelImpl struct {
	config *ChatModelConfig
}

type ChatModelConfig struct {
}

// newChatModel component initialization function of node 'CustomChatModel2' in graph 'testGraph'
func newChatModel(ctx context.Context) (cm model.ChatModel, err error) {
	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:8b",
		Thinking: &api.ThinkValue{
			Value: false,
		},
	})
	return chatModel, nil
}

func (impl *ChatModelImpl) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	panic("implement me")
}

func (impl *ChatModelImpl) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	panic("implement me")
}

func (impl *ChatModelImpl) BindTools(tools []*schema.ToolInfo) error {
	panic("implement me")
}
