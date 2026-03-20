package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type WebSearch struct {
	Query     string `json:"query"`
	Model     string `json:"model"`
	CreatedAt string `json:"createdAt"`
}

type WebSearchesResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Rows       []WebSearch `json:"rows"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			Total      int `json:"total"`
			TotalPages int `json:"totalPages"`
		} `json:"pagination"`
	} `json:"data"`
}

func ListWebSearches(websiteID string, pageSize string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if pageSize != "" {
		params["pageSize"] = pageSize
	}

	var result WebSearchesResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/prompt-web-searches", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	rows := result.Data.Rows
	if len(rows) == 0 {
		fmt.Println("No LLM web searches found.")
		return
	}

	fmt.Println("LLM Web Searches")
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("  %-50s %-15s %s\n", "Query", "Model", "Date")
	fmt.Println(strings.Repeat("─", 80))
	for _, r := range rows {
		query := r.Query
		if len(query) > 48 {
			query = query[:45] + "..."
		}
		date := r.CreatedAt
		if len(date) > 10 {
			date = date[:10]
		}
		fmt.Printf("  %-50s %-15s %s\n", query, r.Model, date)
	}
	fmt.Printf("\nShowing %d of %d searches\n", len(rows), result.Data.Pagination.Total)
}
