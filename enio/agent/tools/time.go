package tools

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

type TimeParams struct {
	Format string `json:"format"`
}

type TimeResult struct {
	Time string `json:"time"`
}

func GetCurrentTime(ctx context.Context, params TimeParams) (TimeResult, error) {
	log.Printf("GetCurrentTime: %+v", params)
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
