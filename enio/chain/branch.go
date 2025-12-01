package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/compose"
)

func BranchDemo() {
	ctx := context.Background()

	branchCondition := func(ctx context.Context, input map[string]any) (string, error) {
		lang := input["lang"].(string)
		lang = strings.ToLower(lang)
		fmt.Printf("lang: %s\n", lang)
		switch lang {
		case "go":
			return "go_branch", nil
		case "python":
			return "python_branch", nil
		default:
			return "default_branch", nil
		}
	}

	goBranch := compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		fmt.Println("go branch")
		input["advice"] = "go branch advice"
		input["features"] = []string{"高性能", "高并发", "高可用"}
		return input, nil
	})

	pyBranch := compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		fmt.Println("python branch")
		input["advice"] = "python branch advice"
		input["features"] = []string{"生态丰富", "易上手", "社区活跃"}
		return input, nil
	})

	defaultBranch := compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		fmt.Println("其他语言")
		input["advice"] = "其他语言 AI库"
		input["features"] = []string{"待探索"}
		return input, nil
	})

	chain := compose.NewChain[map[string]any, map[string]any]()

	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		println("开始处理")
		return input, nil
	})).AppendBranch(
		compose.NewChainBranch(branchCondition).
			AddLambda("go_branch", goBranch).
			AddLambda("python_branch", pyBranch).
			AddLambda("default_branch", defaultBranch),
	).AppendLambda(compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		println("处理完成")
		return input, nil
	}))

	r, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译失败：%v", err)
	}

	testCases := []map[string]any{
		{"lang": "go", "test": "开发"},
		{"lang": "python", "test": "测试"},
		{"lang": "java", "test": "运维"},
	}
	for _, testCase := range testCases {
		out, err := r.Invoke(ctx, testCase)
		if err != nil {
			log.Fatalf("执行失败: %v", err)
		}
		fmt.Printf("输出: %v\n", out)
	}
}
