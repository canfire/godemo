package main

import (
	"context"
	"log"
	"time"

	"github.com/canfire/godemo/multiagent/planexecuteagent/agent"
	"github.com/canfire/godemo/multiagent/util/prints"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
)

func main() {
	ctx := context.Background()
	planAgent, err := agent.NewPlanner(ctx)
	if err != nil {
		log.Fatalf("agent.NewPlanner failed, err: %v", err)
	}

	executeAgent, err := agent.NewExecutor(ctx)
	if err != nil {
		log.Fatalf("agent.NewExecutor failed, err: %v", err)
	}

	replanAgent, err := agent.NewReplanAgent(ctx)
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

	// 	query := `计划下个月去北京旅行三天。我需要从纽约出发的航班、酒店推荐和必看景点。
	// 今天是2025-09-09。
	// 给我简单计划即可不需要详细的行程安排。`
	query := `计划下个月去北京的机票。
今天是2025-09-09。`
	// ctx, endSpanFn := startSpanFn(ctx, "plan-execute-replan", query)
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

	// wait for all span to be ended
	time.Sleep(5 * time.Second)

	println(lastMessage.Content)
}
