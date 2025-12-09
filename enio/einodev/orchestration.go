package einodev

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

func BuildtestGraph(ctx context.Context) (r compose.Runnable[string, string], err error) {
	const (
		Lambda1          = "Lambda1"
		CustomChatModel2 = "CustomChatModel2"
	)
	g := compose.NewGraph[string, string]()
	_ = g.AddLambdaNode(Lambda1, compose.InvokableLambda(newLambda))
	customChatModel2KeyOfChatModel, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(CustomChatModel2, customChatModel2KeyOfChatModel)
	_ = g.AddEdge(compose.START, Lambda1)
	_ = g.AddEdge(CustomChatModel2, compose.END)
	_ = g.AddEdge(Lambda1, CustomChatModel2)
	r, err = g.Compile(ctx, compose.WithGraphName("testGraph"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
