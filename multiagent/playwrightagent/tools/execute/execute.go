package execute

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var _ tool.InvokableTool = (*ExecuteTool)(nil)

type ExecuteTool struct {
}

// ExecuteRequest 请求参数
type ExecuteRequest struct {
	Command string `json:"command"`
}

type ExecuteResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func (c *ExecuteResponse) String() string {
	jsonStr, _ := json.Marshal(c)
	return string(jsonStr)
}

func (c *ExecuteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "execute_tool",
		Desc: "智能在当前页面进行点击或输入操作，无法进行其他操作",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"command": {
				Type:     schema.String,
				Desc:     "在当前页面需要如何操作的指令",
				Required: true,
			},
		}),
	}, nil
}

func (c *ExecuteTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var request ExecuteRequest
	result := ExecuteResponse{}
	err := json.Unmarshal([]byte(argumentsInJSON), &request)
	if err != nil {
		return "", fmt.Errorf("unmarshal arguments failed: %w", err)
	}
	println("++++ExecuteTool+++++++++++", request.Command)
	elInfos, err := ExtractPageElements(ctx, request.Command)
	if err != nil {
		log.Fatalf("ExtractPageElements err: %v", err)
		result.Error = err.Error()
		return "", err
	}
	req, err := GetOperation(ctx, elInfos, request.Command)
	if err != nil {
		log.Fatalf("GetOperation err: %v", err)
		result.Error = err.Error()
		return "", err
	}
	err = Execute(ctx, req, elInfos)
	if err != nil {
		log.Fatalf("Execute err: %v", err)
		result.Error = err.Error()
		return "", err
	}
	time.Sleep(time.Second * 2)
	result.Message = "执行成功"
	return result.String(), nil
}
