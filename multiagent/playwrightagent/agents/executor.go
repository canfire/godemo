package agents

import (
	"context"

	"github.com/canfire/godemo/multiagent/playwrightagent/tools"
	"github.com/canfire/godemo/multiagent/util/model"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

var executorPrompt = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`你是一个自动网页工具agent，请遵循给定的计划，仔细彻底地执行你的任务。
使用可用工具来执行每个规划步骤。
对于网页跳转，请使用 redirection_tool 工具。
对于页面点击输入操作，请使用 execute_tool 工具。
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

func NewExecutor(ctx context.Context) (adk.Agent, error) {
	// Get travel tools for the executor
	travelTools, err := tools.GetAllTools(ctx)
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
