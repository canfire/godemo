package tools

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type AskForClarificationInput struct {
	Question string `json:"question" jsonschema_description:"The specific question you want to ask the user to get the missing information"`
}

func NewAskForClarificationTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"ask_for_clarification",
		"当用户的请求不明确或缺乏继续进行所需的信息时，调用此工具。在你有效使用其他工具之前，用它来问一个后续问题，以获得你需要的细节，比如这本书的类型。",
		func(ctx context.Context, input *AskForClarificationInput, opts ...tool.Option) (output string, err error) {
			fmt.Printf("\nQuestion: %s\n", input.Question)
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Print("\nyour input here: ")
			scanner.Scan()
			fmt.Println()
			nInput := scanner.Text()
			return nInput, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}
