package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/canfire/godemo/multiagent/integration-project-manager/agents"
	"github.com/canfire/godemo/multiagent/util/prints"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// 项目开发经理智能体
func main() {
	ctx := context.Background()
	// tcm, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
	// 	BaseURL: "http://localhost:11434",
	// 	Model:   "qwen3:8b",
	// 	Thinking: &api.ThinkValue{
	// 		Value: false,
	// 	},
	// })
	tcm, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek/deepseek-v3",
		BaseURL: "https://api.yygu.cn/v3/llm.chat",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// Init research agent
	researchAgent, err := agents.NewResearchAgent(ctx, tcm)
	if err != nil {
		log.Fatal(err)
	}

	// Init code agent
	codeAgent, err := agents.NewCodeAgent(ctx, tcm)
	if err != nil {
		log.Fatal(err)
	}

	// Init technical agent
	reviewAgent, err := agents.NewReviewAgent(ctx, tcm)
	if err != nil {
		log.Fatal(err)
	}

	// Init project manager agent
	s, err := agents.NewProjectManagerAgent(ctx, tcm)
	if err != nil {
		log.Fatal(err)
	}

	// Combine agents into ADK supervisor pattern
	// Supervisor: project manager
	// Sub-agents: researcher / coder / reviewer
	supervisorAgent, err := supervisor.New(ctx, &supervisor.Config{
		Supervisor: s,
		SubAgents:  []adk.Agent{researchAgent, codeAgent, reviewAgent},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Init Agent runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           supervisorAgent,
		EnableStreaming: true,
		CheckPointStore: newInMemoryStore(),
	})

	//用您自己的查询替换它
	//当使用以下查询时，researchAgent将中断并提示用户通过stdin输入特定的研究主题。
	query := "javascript 报告"
	checkpointID := "1"

	//researchAgent可能会要求用户多次输入信息
	//因此，以下标志“中断”和“完成”用于支持多次中断和恢复。
	interrupted := false
	finished := false

	for !finished {
		var iter *adk.AsyncIterator[*adk.AgentEvent]

		if !interrupted {
			iter = runner.Query(ctx, query, adk.WithCheckPointID(checkpointID))
		} else {
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Print("\ninput additional context for web search: ")
			scanner.Scan()
			fmt.Println()
			nInput := scanner.Text()

			iter, err = runner.Resume(ctx, checkpointID, adk.WithToolOptions([]tool.Option{agents.WithNewInput(nInput)}))
			if err != nil {
				log.Fatal(err)
			}
		}

		interrupted = false

		for {
			event, ok := iter.Next()
			if !ok {
				if !interrupted {
					finished = true
				}
				break
			}
			if event.Err != nil {
				log.Fatal(event.Err)
			}
			if event.Action != nil {
				if event.Action.Interrupted != nil {
					interrupted = true
				}
				if event.Action.Exit {
					finished = true
				}
			}
			prints.Event(event)
		}
	}
}

func newInMemoryStore() compose.CheckPointStore {
	return &inMemoryStore{
		mem: map[string][]byte{},
	}
}

type inMemoryStore struct {
	mem map[string][]byte
}

func (i *inMemoryStore) Set(ctx context.Context, key string, value []byte) error {
	i.mem[key] = value
	return nil
}

func (i *inMemoryStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	v, ok := i.mem[key]
	return v, ok, nil
}
