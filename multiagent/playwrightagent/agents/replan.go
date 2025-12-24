package agents

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/canfire/godemo/multiagent/playwrightagent/global"
	"github.com/canfire/godemo/multiagent/util/model"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

const (
	replannerSysPrompt = `您将回顾目标的进展情况。根当前页面截图分析当前状态，并确定最佳的下一步行动。

## 您的任务
根据上述进度，您必须且仅能选择一项行动：

### 选项1：完成（若目标已完全达成）
使用以下参数调用“{respond_tool}”：
- 一个根据页面分析的全面的最终答案
- 总结目标达成情况的清晰结论
- 执行过程中的关键见解

### 选项2：继续（如需更多工作）
使用修订后的计划调用“{plan_tool}”，该计划应：
- 仅包含剩余步骤（不包括已完成的步骤）
- 融入从已执行步骤中汲取的经验教训
- 解决发现的任何漏洞或问题
- 保持逻辑上的步骤顺序

## 规划要求
你计划中的每一步都必须：
- **具体且可操作**：明确的指示，能够无歧义地执行
- **自包含式**：包含所有必要的上下文、参数和要求
- **可独立执行**：可由代理执行，无需依赖其他步骤
- **逻辑顺序排列**：以最佳顺序排列，以便高效执行
- **目标导向**：直接为实现主要目标做出贡献

## 规划指南
- 消除冗余或不必要的步骤
- 根据新信息调整策略
- 为每个步骤包含相关的约束条件、参数和成功标准

## 决策标准
- 根据图片上页面当前状态是否达成用户需求
- 最初的目标是否已完全达成？
- 是否有任何剩余的要求或子目标？
- 结果是否表明需要调整策略？
- 还需要采取哪些具体行动？`
	replannerUserPrompt = `## 目标
{input}
## 原始计划:
{plan}
## 已完成的步骤和结果
{executed_steps}
`
)

func formatInput(input []adk.Message) string {
	var sb strings.Builder
	for _, msg := range input {
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}

	return sb.String()
}

func formatExecutedSteps(results []planexecute.ExecutedStep) string {
	var sb strings.Builder
	for _, result := range results {
		sb.WriteString(fmt.Sprintf("Step: %s\nResult: %s\n\n", result.Step, result.Result))
	}

	return sb.String()
}

func NewReplanAgent(ctx context.Context) (adk.Agent, error) {
	return planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: model.NewVLChatModel(ctx),
		GenInputFn: func(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
			planContent, err := in.Plan.MarshalJSON()
			if err != nil {
				return nil, err
			}
			page, err := global.GetPage(ctx)
			if err != nil {
				log.Fatalf("%s", err.Error())
				return nil, err
			}
			imgbase64, err := global.GetScreenshotBase64(page)
			if err != nil {
				log.Fatalf("%s", err.Error())
				return nil, err
			}
			var replannerPrompt = prompt.FromMessages(schema.FString,
				schema.SystemMessage(replannerSysPrompt),
				&schema.Message{
					Role: schema.User,
					UserInputMultiContent: []schema.MessageInputPart{
						{
							Type: schema.ChatMessagePartTypeText,
							Text: replannerUserPrompt,
						},
						{
							Type: schema.ChatMessagePartTypeImageURL,
							Image: &schema.MessageInputImage{
								MessagePartCommon: schema.MessagePartCommon{
									Base64Data: of(imgbase64),
									MIMEType:   "image/png",
									Extra: map[string]any{
										"vl_high_resolution_images": true,
									},
								},
								Detail: schema.ImageURLDetailHigh,
							},
						},
					},
				},
			)
			msgs, err := replannerPrompt.Format(ctx, map[string]any{
				"plan":           string(planContent),
				"input":          formatInput(in.UserInput),
				"executed_steps": formatExecutedSteps(in.ExecutedSteps),
				"plan_tool":      planexecute.PlanToolInfo.Name,
				"respond_tool":   planexecute.RespondToolInfo.Name,
			})
			if err != nil {
				return nil, err
			}

			return msgs, nil
		},
	})
}

func of[T any](t T) *T {
	return &t
}
