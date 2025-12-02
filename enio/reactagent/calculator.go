package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var _ tool.InvokableTool = (*CalculatorTool)(nil)

type CalculatorTool struct{}

type CalculatorParams struct {
	Operation string  `json:"operation"`
	A         float32 `json:"a"`
	B         float32 `json:"b"`
}

type CalculatorResult struct {
	Result float32 `json:"result"`
	Error  string  `json:"error,omitempty"`
}

func (c *CalculatorResult) String() string {
	jsonStr, _ := json.Marshal(c)
	return string(jsonStr)
}

func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行基本的数学计算（加/减/乘/除）",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"operation": {
				Type:     schema.String,
				Desc:     "计算类型: add(加法),sub(减法),mul(乘法),div(除法)",
				Required: true,
			},
			"a": {
				Type:     schema.Number,
				Desc:     "第一个数",
				Required: true,
			},
			"b": {
				Type:     schema.Number,
				Desc:     "第二个数",
				Required: true,
			},
		}),
	}, nil
}

func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params CalculatorParams
	err := json.Unmarshal([]byte(argumentsInJSON), &params)
	if err != nil {
		return "", fmt.Errorf("unmarshal arguments failed: %w", err)
	}
	result := CalculatorResult{}
	switch params.Operation {
	case "add":
		result.Result = params.A + params.B
	case "sub":
		result.Result = params.A - params.B
	case "mul":
		result.Result = params.A * params.B
	case "div":
		if params.B == 0 {
			result.Error = "除数不能为0"
			break
		}
		result.Result = params.A / params.B
	default:
		result.Error = "不支持的运算类型"
	}
	return result.String(), nil
}
