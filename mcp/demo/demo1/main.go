package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

func main1() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Calculator Demo",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	ctx := context.Background()

	calculatorTool := &CalculatorMcp{}

	// Add the calculator handler
	s.AddTool(calculatorTool.Info(ctx), calculatorTool.Handlerfunc)

	// Start the server
	// if err := server.ServeStdio(s); err != nil {
	// 	fmt.Printf("Server error: %v\n", err)
	// }

	// 启动 HTTP Socket 服务
	// addr := "127.0.0.1:8080"

	// fmt.Println("MCP server listening on", addr)

	// Wrap into HTTP server
	httpSrv := server.NewStreamableHTTPServer(s)
	addr := ":8080"
	fmt.Println("MCP HTTP server listening on", addr)
	if err := httpSrv.Start(addr); err != nil {
		fmt.Println("Server error:", err)
	}
}
