package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type Prompt struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type PromptsResponse struct {
	Prompts []Prompt `json:"prompts"`
}

func ListPrompts(websiteID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	var result PromptsResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/prompts", websiteID), nil, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	if len(result.Prompts) == 0 {
		fmt.Println("No prompts found for this website.")
		return
	}

	fmt.Printf("%-38s  %-30s  %s\n", "ID", "Name", "Status")
	fmt.Println(strings.Repeat("─", 80))
	for _, p := range result.Prompts {
		name := p.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}
		fmt.Printf("%-38s  %-30s  %s\n", p.ID, name, p.Status)
	}
}
