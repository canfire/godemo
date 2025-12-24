package agents

import (
	"context"

	"github.com/canfire/godemo/multiagent/util/model"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
)

// NewPlanner 新建计划
func NewPlanner(ctx context.Context) (adk.Agent, error) {
	return planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: model.NewVLChatModel(ctx),
	})
}
