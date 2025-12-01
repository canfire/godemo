package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func ToolsCallDemo() {
	ctx := context.Background()
	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:8b",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

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

	cu := &CalculatorTool{}
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{
			cu,
			timeTool,
		},
	})

	cuInfo, _ := cu.Info(ctx)
	timeinfo, _ := timeTool.Info(ctx)

	toolsInfo := []*schema.ToolInfo{
		cuInfo,
		timeinfo,
	}

	if err != nil {
		log.Fatalf("创建 ToolsNode 失败: %v", err)
	}

	testCase := []string{
		"几点了",
	}
	for _, v := range testCase {
		fmt.Println("===========================" + v)

		msg := []*schema.Message{
			schema.UserMessage(v),
		}
		res, err := chatModel.Generate(ctx, msg,
			model.WithTools(toolsInfo),
		)
		if err != nil {
			log.Fatalf("生成响应失败: %v", err)
		}
		// fmt.Println(res)
		if len(res.ToolCalls) > 0 {
			for _, a := range res.ToolCalls {
				fmt.Println(a.Function.Name)
				fmt.Println(a.Function.Arguments)
			}
			res, _ := toolsNode.Invoke(ctx, res)
			fmt.Println("==================", res)
		} else {
			fmt.Println("========直接回答==========", res)
		}

	}
}
