package agents

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
)

func NewCodeAgent(ctx context.Context, tcm model.ToolCallingChatModel) (adk.Agent, error) {
	type RAGInput struct {
		Query   string  `json:"query" jsonschema_description:"查询搜索"`
		Context *string `json:"context" jsonschema_description:"用户输入上下文"`
	}
	type RAGOutput struct {
		Documents []string `json:"documents"`
	}
	knowledgeBaseTool, err := utils.InferTool(
		"knowledge_base",
		"知识库，可以回答常见问题，提供答案的具体原因，并提高准确性",
		func(ctx context.Context, input *RAGInput) (output *RAGOutput, err error) {
			// replace it with real knowledge base search
			if input.Query == "" {
				return nil, fmt.Errorf("RAG Input query is required")
			}

			return &RAGOutput{
				[]string{
					"Q: Python中列表和元组有什么区别？\nA：列表是可变的，这意味着您可以在创建后修改其元素，而元组是不可变的，一旦创建就不能更改。列表使用方括号[]，元组使用括号（）。",
					"Q: Java中如何处理异常？\nA：在Java中，您可以使用try-catch块来处理异常。可能引发异常的代码被放置在try块中，catch块处理异常。可选地，可以使用finally块进行清理。",
					"Q: JavaScript中async和wait关键字的用途是什么？\nA:async将函数标记为异步，允许它返回Promise。wait会暂停异步函数的执行，直到Promise解析，从而使异步代码编写更容易。",
					"Q: 如何优化SQL查询以获得更好的性能？\nA：常见的优化包括在频繁查询的列上创建索引、避免SELECT*、高效使用JOIN以及分析查询执行计划以识别瓶颈。",
					"Q: 什么是依赖注入，为什么它有用？\nA：依赖注入是一种设计模式，其中对象从外部源接收其依赖关系，而不是自己创建它们。它促进了松耦合、更容易的测试和更好的代码可维护性。",
				},
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "CodeAgent",
		Description: "CodeAgent专门通过利用知识库作为工具来生成高质量的代码。它回顾了相关知识和最佳实践，以生成针对项目需求量身定制的高效、可维护和准确的代码解决方案。",
		Instruction: `你是CodeAgent。您的职责包括：

-根据项目需求生成高质量、高效和可维护的代码。
-利用知识库工具回忆相关的编码标准、模式和最佳实践。
-确保代码清晰、文档齐全，并符合指定的功能。
-复习相关知识，以提高代码的准确性和质量。
-传达您的编码决策，并在必要时提供解释。
-对用户请求或澄清作出迅速和专业的回应。

工具处理：
当用户的问题模糊或超出您的回答范围时，请使用 knowledge_base 工具从知识库中检索相关结果，并根据结果提供准确的答案。
`,
		Model: tcm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{knowledgeBaseTool},
			},
		},
		MaxIterations: 3,
	})
}
