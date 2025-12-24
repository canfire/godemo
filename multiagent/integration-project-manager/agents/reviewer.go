package agents

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

func NewReviewAgent(ctx context.Context, tcm model.ToolCallingChatModel) (adk.Agent, error) {
	// 有多个子agent
	questionAnalysisAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "question_analysis_agent",
		Description: "问题分析 agent",
		Instruction: `您是问题 question_analysis_agent。您的职责包括：

-分析给定的研究或编码结果，以确定关键问题和评估标准。
-将复杂问题分解为清晰、可管理的组件。
-突出潜在问题或关注领域。
-准备一个结构化的框架来指导后续的审查生成。
-在传递内容之前，确保对内容有深入的理解。`,
		Model: tcm,
	})
	if err != nil {
		return nil, err
	}

	generateReviewAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "generate_review_agent",
		Description: "generate review agent",
		Instruction: `您是generate_review_agent。你的职责是：

-根据问题分析进行全面和平衡的评论。
-突出优势、劣势和需要改进的地方。
-提供建设性和可操作的反馈。
-保持评估的客观性和清晰度。
-准备审核内容，以便下一步进行验证。`,
		Model: tcm,
	})
	if err != nil {
		return nil, err
	}

	reviewValidationAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "review_validation_agent",
		Description: "review validation agent",
		Instruction: `你是 Review Validation Agent. 您的任务是：

-验证生成的审查的准确性、连贯性和公平性。
-检查逻辑的一致性和完整性。
-识别任何偏差或错误，并提出纠正建议。
-确认审查与原始分析和项目目标一致。
-批准最终演示文稿的审查，或在必要时要求修改。`,
		Model: tcm,
	})
	if err != nil {
		return nil, err
	}

	return adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
		Name:        "ReviewAgent",
		Description: "The ReviewAgent is responsible for evaluating research and coding results through a sequential workflow. It orchestrates three key steps—question analysis, review generation, and review validation—to provide well-reasoned assessments that support project management decisions.",
		SubAgents:   []adk.Agent{questionAnalysisAgent, generateReviewAgent, reviewValidationAgent},
	})
}
