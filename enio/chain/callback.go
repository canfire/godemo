package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	callbackHelper "github.com/cloudwego/eino/utils/callbacks"
	"github.com/eino-contrib/ollama/api"
)

func CallBackDemo() {
	ctx := context.Background()

	modelHandler := &callbackHelper.ModelCallbackHandler{
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
			fmt.Println("模型的思考过程为: ")
			fmt.Println(output.Message.Content)
			return ctx
		},
	}

	toolHander := &callbackHelper.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			fmt.Println("开始执行工具，参数：", input.ArgumentsInJSON)
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			fmt.Println("工具执行完成，输出：", output.Response)
			return ctx
		},
	}

	handler := callbackHelper.NewHandlerHelper().ChatModel(modelHandler).Tool(toolHander).Handler()

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

	info, err := timeTool.Info(ctx)
	if err != nil {
		log.Fatalf("获取工具信息失败: %v", err)
	}
	infos := []*schema.ToolInfo{info}
	err = chatModel.BindTools(infos)
	if err != nil {
		log.Fatalf("绑定工具失败: %v", err)
	}

	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{
			timeTool,
		},
	})
	if err != nil {
		log.Fatalf("创建 ToolsNode 失败: %v", err)
	}

	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()

	chain.AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
		AppendToolsNode(toolsNode, compose.WithNodeName("tools"))

	r, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译失败：%v", err)
	}

	out, err := r.Invoke(ctx, []*schema.Message{
		schema.UserMessage("现在几点了"),
	}, compose.WithCallbacks(handler))
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}
	// fmt.Printf("输出: %s\n", out)
	for _, v := range out {
		fmt.Printf("输出: %s\n", v.Content)
	}
}

type TimeParams struct {
	Format string `json:"format"`
}

type TimeResult struct {
	Time string `json:"time"`
}

func GetCurrentTime(ctx context.Context, params TimeParams) (TimeResult, error) {
	now := time.Now()
	var result TimeResult
	switch params.Format {
	case "data":
		result.Time = now.Format("2006-01-02")
	case "time":
		result.Time = now.Format("15:04:05")
	default:
		result.Time = now.Format("2006-01-02 15:04:05")

	}
	return result, nil
}

func ToolNewToolDemo() {
	ctx := context.Background()
	timeTool := utils.NewTool[TimeParams, TimeResult](&schema.ToolInfo{
		Name: "get_current_time",
		Desc: "获取当前时间",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"format": {
				Type:     schema.String,
				Desc:     "时间格式: data(日期),time(时间),default(日期时间)",
				Enum:     []string{"data", "time", "default"},
				Required: false,
			},
		}),
	}, GetCurrentTime)

	out, err := timeTool.InvokableRun(ctx, `{"format": "default"}`)
	if err != nil {
		log.Fatalf("InvokableRun failed: %v", err)
	}
	fmt.Printf("out: %v\n", out)

}

// NewTimeTool 创建一个工具
func NewTimeTool() tool.InvokableTool {
	timeTool := utils.NewTool[TimeParams, TimeResult](&schema.ToolInfo{
		Name: "get_current_time",
		Desc: "获取当前时间",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"format": {
				Type:     schema.String,
				Desc:     "时间格式: data(日期),time(时间),default(日期时间)",
				Enum:     []string{"data", "time", "default"},
				Required: false,
			},
		}),
	}, GetCurrentTime)
	return timeTool
}
