package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type PromptRunSource struct {
	URL    string `json:"url"`
	Domain string `json:"domain"`
	Title  string `json:"title"`
}

type PromptRun struct {
	ID       string            `json:"id"`
	Model    string            `json:"model"`
	Mentions interface{}       `json:"mentions"`
	Sources  []PromptRunSource `json:"sources"`
}

type PromptRunsResponse struct {
	Success bool        `json:"success"`
	Data    []PromptRun `json:"data"`
	Meta    struct {
		Total int `json:"total"`
		Page  int `json:"page"`
	} `json:"meta"`
}

func ListPromptRuns(websiteID string, promptID string, pageSize string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if pageSize != "" {
		params["pageSize"] = pageSize
	}

	var result PromptRunsResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/prompts/%s/runs", websiteID, promptID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	if len(result.Data) == 0 {
		fmt.Println("No prompt runs found.")
		return
	}

	fmt.Printf("Prompt Runs (latest %d)\n", len(result.Data))
	fmt.Println(strings.Repeat("─", 80))

	for i, run := range result.Data {
		fmt.Printf("\n%d. Model: %s\n", i+1, run.Model)
		if len(run.Sources) > 0 {
			fmt.Printf("   Sources (%d):\n", len(run.Sources))
			for _, s := range run.Sources {
				title := s.Title
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				fmt.Printf("     - %s (%s)\n", title, s.Domain)
			}
		}
	}
}
