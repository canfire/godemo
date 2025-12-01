package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/compose"
)

func LambdaDemo() {
	ctx := context.Background()

	chain := compose.NewChain[string, string]()

	chain.AppendLambda(compose.InvokableLambdaWithOption(UpperLambda)).AppendLambda(compose.InvokableLambdaWithOption(AddPrefixLambda))

	r, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译失败：%v", err)
	}

	out, err := r.Invoke(ctx, "hello world")
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}
	fmt.Printf("输出: %s\n", out)
}

// UpperLambda 定义Lambde节点
func UpperLambda(ctx context.Context, input string, opts ...any) (string, error) {
	fmt.Printf("步骤1: 输入 %s\n", input)
	res := strings.ToUpper(input)
	fmt.Printf("步骤1: 输出 %s\n", res)
	return res, nil
}

// AddPrefixLambda 定义Lambde节点
func AddPrefixLambda(ctx context.Context, input string, opts ...any) (string, error) {
	fmt.Printf("步骤2: 输入 %s\n", input)
	res := "处理结果-" + input
	fmt.Printf("步骤2: 输出 %s\n", res)
	return res, nil
}
