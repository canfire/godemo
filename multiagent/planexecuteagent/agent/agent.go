package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/canfire/godemo/multiagent/planexecuteagent/tools"
	"github.com/canfire/godemo/multiagent/util/model"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// NewPlanner 新建计划
func NewPlanner(ctx context.Context) (adk.Agent, error) {
	return planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: model.NewChatModel(ctx),
	})
}

var executorPrompt = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`你是一个勤奋且细致的旅行研究执行者，请遵循给定的计划，仔细彻底地执行你的任务。
使用可用工具来执行每个规划步骤。
对于天气查询，请使用 get_weather 工具。
对于航班搜索，请使用 search_flights 工具。
对于酒店搜索，请使用 search_hotels 工具。
对于景点研究，请使用 search_attractions 工具。
如果需要用户澄清，请使用 ask_for_clarification 工具。总之，在确认时重复问题和结果，尽量避免打扰用户。
为每个任务提供详细的结果。
可以同时调用多个工具来获得最终结果。`),
	schema.UserMessage(`## 目标
{input}
## 给定以下计划：
{plan}
## 已完成的步骤及结果
{executed_steps}
## 你的任务是执行第一个步骤，即：
{step}`))

func formatInput(in []adk.Message) string {
	return in[0].Content
}

func formatExecutedSteps(in []planexecute.ExecutedStep) string {
	var sb strings.Builder
	for idx, m := range in {
		sb.WriteString(fmt.Sprintf("## %d. Step: %v\n  Result: %v\n\n", idx+1, m.Step, m.Result))
	}
	return sb.String()
}

func NewExecutor(ctx context.Context) (adk.Agent, error) {
	// Get travel tools for the executor
	travelTools, err := tools.GetAllTravelTools(ctx)
	if err != nil {
		return nil, err
	}

	return planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model: model.NewChatModel(ctx),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: travelTools,
			},
		},

		GenInputFn: func(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
			planContent, err_ := in.Plan.MarshalJSON()
			if err_ != nil {
				return nil, err_
			}

			firstStep := in.Plan.FirstStep()

			msgs, err_ := executorPrompt.Format(ctx, map[string]any{
				"input":          formatInput(in.UserInput),
				"plan":           string(planContent),
				"executed_steps": formatExecutedSteps(in.ExecutedSteps),
				"step":           firstStep,
			})
			if err_ != nil {
				return nil, err_
			}

			return msgs, nil
		},
	})
}

func NewReplanAgent(ctx context.Context) (adk.Agent, error) {
	return planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: model.NewChatModel(ctx),
	})
}
