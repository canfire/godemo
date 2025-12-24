package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/canfire/godemo/multiagent/playwrightagent/global"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var _ tool.InvokableTool = (*GoTOTool)(nil)

type GoTOTool struct {
}

// GoTORequest 请求参数
type GoTORequest struct {
	Url string `json:"url"`
}

type GoTOResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func (c *GoTOResponse) String() string {
	jsonStr, _ := json.Marshal(c)
	return string(jsonStr)
}

func (c *GoTOTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "redirection_tool",
		Desc: "跳转到指定网址页面，如跳转到 https://www.baidu.com 等页面",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type:     schema.String,
				Desc:     "当前网页需要跳转到的页面地址，如 https://www.baidu.com 等，必须携带http或https",
				Required: true,
			},
		}),
	}, nil
}

func (c *GoTOTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var request GoTORequest
	result := GoTOResponse{}
	err := json.Unmarshal([]byte(argumentsInJSON), &request)
	if err != nil {
		return "", fmt.Errorf("unmarshal arguments failed: %w", err)
	}
	println("++++++++GoTOTool+++++++", request.Url)
	page, err := global.GetPage(ctx)
	if err != nil {
		return "", fmt.Errorf("get page failed: %w", err)
	}
	_, err = page.Goto(request.Url)
	if err != nil {
		result.Error = err.Error()
		result.Message = "跳转失败"
		return result.String(), nil
	}
	result.Message = "跳转成功"
	return result.String(), nil
}
