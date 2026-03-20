package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type AIVisit struct {
	URL       string `json:"url"`
	BotName   string `json:"botName"`
	CreatedAt string `json:"createdAt"`
}

type AIVisitsResponse struct {
	Success    bool      `json:"success"`
	Data       []AIVisit `json:"data"`
	Pagination struct {
		Page       int `json:"page"`
		PageSize   int `json:"pageSize"`
		Total      int `json:"total"`
		TotalPages int `json:"totalPages"`
	} `json:"pagination"`
}

func ListAIVisits(websiteID string, pageSize string, searchQuery string, botName string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if pageSize != "" {
		params["pageSize"] = pageSize
	}
	if searchQuery != "" {
		params["searchQuery"] = searchQuery
	}
	if botName != "" {
		params["botName"] = botName
	}

	var result AIVisitsResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/ai-visits", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	if len(result.Data) == 0 {
		fmt.Println("No AI bot visits found.")
		return
	}

	fmt.Println("AI Bot Visits")
	fmt.Println(strings.Repeat("─", 90))
	fmt.Printf("  %-20s %-50s %s\n", "Bot", "URL", "Date")
	fmt.Println(strings.Repeat("─", 90))
	for _, v := range result.Data {
		url := v.URL
		if len(url) > 48 {
			url = url[:45] + "..."
		}
		date := v.CreatedAt
		if len(date) > 10 {
			date = date[:10]
		}
		fmt.Printf("  %-20s %-50s %s\n", v.BotName, url, date)
	}
	fmt.Printf("\nShowing %d of %d visits\n", len(result.Data), result.Pagination.Total)
}
