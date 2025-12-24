package tools

import (
	"context"

	"github.com/canfire/godemo/multiagent/playwrightagent/tools/execute"
	"github.com/cloudwego/eino/components/tool"
)

func GetAllTools(ctx context.Context) ([]tool.BaseTool, error) {
	goToTool := &GoTOTool{}
	askForClarificationTool := NewAskForClarificationTool()
	execTool := &execute.ExecuteTool{}
	return []tool.BaseTool{goToTool, execTool, askForClarificationTool}, nil
}
