package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type SearchQuery struct {
	Query       string      `json:"query"`
	Impressions int         `json:"impressions"`
	Position    float64     `json:"position"`
	Pages       interface{} `json:"pages"`
}

type SearchQueriesResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Rows       []SearchQuery `json:"rows"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			Total      int `json:"total"`
			TotalPages int `json:"totalPages"`
		} `json:"pagination"`
	} `json:"data"`
}

func ListSearchQueries(websiteID string, pageSize string, sortBy string, urlFilter string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if pageSize != "" {
		params["pageSize"] = pageSize
	}
	if sortBy != "" {
		params["sortBy"] = sortBy
	}
	if urlFilter != "" {
		params["url"] = urlFilter
	}

	var result SearchQueriesResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/search-queries", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	rows := result.Data.Rows
	if len(rows) == 0 {
		fmt.Println("No search queries found.")
		return
	}

	fmt.Println("Google Search Console — Queries")
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("  %-50s %12s %8s\n", "Query", "Impressions", "Position")
	fmt.Println(strings.Repeat("─", 80))
	for _, r := range rows {
		query := r.Query
		if len(query) > 48 {
			query = query[:45] + "..."
		}
		fmt.Printf("  %-50s %12d %8.1f\n", query, r.Impressions, r.Position)
	}
	fmt.Printf("\nShowing %d of %d queries\n", len(rows), result.Data.Pagination.Total)
}
