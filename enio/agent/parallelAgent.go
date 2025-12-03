package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/eino-contrib/ollama/api"
)

func newChatModel() model.ToolCallingChatModel {
	cm, err := ollama.NewChatModel(context.Background(), &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "qwen3:8b",
		Thinking: &api.ThinkValue{
			Value: false,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	return cm
}

// 技术分析 Agent
func NewTechnicalAnalystAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "TechnicalAnalyst",
		Description: "从技术角度分析内容",
		Instruction: `你是一个技术专家。请从技术实现、架构设计、性能优化等技术角度分析提供的内容。
重点关注：
1. 技术可行性
2. 架构合理性  
3. 性能考量
4. 技术风险
5. 实现复杂度`,
		Model: newChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

// 商业分析 Agent
func NewBusinessAnalystAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "BusinessAnalyst",
		Description: "从商业角度分析内容",
		Instruction: `你是一个商业分析专家。请从商业价值、市场前景、成本效益等商业角度分析提供的内容。
重点关注：
1. 商业价值
2. 市场需求
3. 竞争优势
4. 成本分析
5. 盈利模式`,
		Model: newChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

// 用户体验分析 Agent
func NewUXAnalystAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "UXAnalyst",
		Description: "从用户体验角度分析内容",
		Instruction: `你是一个用户体验专家。请从用户体验、易用性、用户满意度等角度分析提供的内容。
重点关注：
1. 用户友好性
2. 操作便利性
3. 学习成本
4. 用户满意度
5. 可访问性`,
		Model: newChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

// 安全分析 Agent
func NewSecurityAnalystAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "SecurityAnalyst",
		Description: "从安全角度分析内容",
		Instruction: `你是一个安全专家。请从信息安全、数据保护、隐私合规等安全角度分析提供的内容。
重点关注：
1. 数据安全
2. 隐私保护
3. 访问控制
4. 安全漏洞
5. 合规要求`,
		Model: newChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

func AgentParallel() {
	ctx := context.Background()

	// 创建四个不同角度的分析 Agent
	techAnalyst := NewTechnicalAnalystAgent()
	bizAnalyst := NewBusinessAnalystAgent()
	uxAnalyst := NewUXAnalystAgent()
	secAnalyst := NewSecurityAnalystAgent()

	// 创建 ParallelAgent，同时进行多角度分析
	parallelAgent, err := adk.NewParallelAgent(ctx, &adk.ParallelAgentConfig{
		Name:        "MultiPerspectiveAnalyzer",
		Description: "多角度并行分析：技术 + 商业 + 用户体验 + 安全",
		SubAgents:   []adk.Agent{techAnalyst, bizAnalyst, uxAnalyst, secAnalyst},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: parallelAgent,
	})

	// 要分析的产品方案
	productProposal := `
产品方案：智能客服系统

概述：开发一个基于大语言模型的智能客服系统，能够自动回答用户问题，处理常见业务咨询，并在必要时转接人工客服。

主要功能：
1. 自然语言理解和回复
2. 多轮对话管理
3. 知识库集成
4. 情感分析
5. 人工客服转接
6. 对话历史记录
7. 多渠道接入（网页、微信、APP）

技术架构：
- 前端：React + TypeScript
- 后端：Go + Gin 框架
- 数据库：PostgreSQL + Redis
- AI模型：GPT-4 API
- 部署：Docker + Kubernetes
`

	fmt.Println("开始多角度并行分析...")
	iter := runner.Query(ctx, "请分析以下产品方案：\n"+productProposal)

	// 使用 map 来收集不同分析师的结果
	results := make(map[string]string)
	var mu sync.Mutex

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			log.Printf("分析过程中出现错误: %v", event.Err)
			continue
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			mu.Lock()
			results[event.AgentName] = event.Output.MessageOutput.Message.Content
			mu.Unlock()

			fmt.Printf("\n=== %s 分析完成 ===\n", event.AgentName)
		}
	}

	// 输出所有分析结果
	fmt.Println("\n" + "============================================================")
	fmt.Println("多角度分析结果汇总")
	fmt.Println("============================================================")

	analysisOrder := []string{"TechnicalAnalyst", "BusinessAnalyst", "UXAnalyst", "SecurityAnalyst"}
	analysisNames := map[string]string{
		"TechnicalAnalyst": "技术分析",
		"BusinessAnalyst":  "商业分析",
		"UXAnalyst":        "用户体验分析",
		"SecurityAnalyst":  "安全分析",
	}

	for _, agentName := range analysisOrder {
		if result, exists := results[agentName]; exists {
			fmt.Printf("\n【%s】\n", analysisNames[agentName])
			fmt.Printf("%s\n", result)
			fmt.Println("----------------------------------------")
		}
	}

	fmt.Println("\n多角度并行分析完成！")
	fmt.Printf("共收到 %d 个分析结果\n", len(results))
}
