package commands

import (
	"fmt"

	"github.com/xseekio/xseek-cli/internal/api"
)

type LLMsTxtResult struct {
	URL     string `json:"url"`
	Content string `json:"content"`
}

func GenerateLLMsTxt(domain string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	var result LLMsTxtResult
	err = client.GetJSON("/tools/llms-txt", map[string]string{"url": domain}, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	fmt.Println(result.Content)
}
