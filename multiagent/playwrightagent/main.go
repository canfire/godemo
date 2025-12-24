package main

import (
	"context"
	"log"
	"time"

	"github.com/canfire/godemo/multiagent/playwrightagent/agents"
	"github.com/canfire/godemo/multiagent/util/prints"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
)

func main() {
	ctx := context.Background()
	planAgent, err := agents.NewPlanner(ctx)
	if err != nil {
		log.Fatalf("agent.NewPlanner failed, err: %v", err)
	}

	executeAgent, err := agents.NewExecutor(ctx)
	if err != nil {
		log.Fatalf("agent.NewExecutor failed, err: %v", err)
	}

	replanAgent, err := agents.NewReplanAgent(ctx)
	if err != nil {
		log.Fatalf("agent.NewReplanAgent failed, err: %v", err)
	}

	entryAgent, err := planexecute.New(ctx, &planexecute.Config{
		Planner:       planAgent,
		Executor:      executeAgent,
		Replanner:     replanAgent,
		MaxIterations: 20,
	})
	if err != nil {
		log.Fatalf("NewPlanExecuteAgent failed, err: %v", err)
	}

	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           entryAgent,
		EnableStreaming: true,
	})

	query := `帮我搜索查看今日京东股价`
	iter := r.Query(ctx, query)
	var lastMessage adk.Message
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		prints.Event(event)

		if event.Output != nil {
			lastMessage, _, err = adk.GetMessage(event)
		}
	}
	println(lastMessage.Content)
	// wait for all span to be ended
	time.Sleep(5 * time.Minute)

	println(lastMessage.Content)
}
