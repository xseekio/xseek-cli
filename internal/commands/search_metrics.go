package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type SearchMetricsRow struct {
	URL         string  `json:"url"`
	Path        string  `json:"path"`
	Impressions int     `json:"impressions"`
	Clicks      int     `json:"clicks"`
	CTR         float64 `json:"ctr"`
	Position    float64 `json:"position"`
}

type SearchMetricsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Rows       []SearchMetricsRow `json:"rows"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			Total      int `json:"total"`
			TotalPages int `json:"totalPages"`
		} `json:"pagination"`
	} `json:"data"`
}

func ListSearchMetrics(websiteID string, pageSize string, sortBy string, search string) {
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
	if search != "" {
		params["search"] = search
	}

	var result SearchMetricsResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/search-metrics", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	rows := result.Data.Rows
	if len(rows) == 0 {
		fmt.Println("No search metrics found.")
		return
	}

	fmt.Println("Google Search Console — Page Metrics")
	fmt.Println(strings.Repeat("─", 90))
	fmt.Printf("  %-50s %10s %8s %6s %8s\n", "URL", "Impressions", "Clicks", "CTR", "Position")
	fmt.Println(strings.Repeat("─", 90))
	for _, r := range rows {
		url := r.Path
		if len(url) > 48 {
			url = url[:45] + "..."
		}
		fmt.Printf("  %-50s %10d %8d %5.1f%% %8.1f\n", url, r.Impressions, r.Clicks, r.CTR*100, r.Position)
	}
	fmt.Printf("\nShowing %d of %d pages\n", len(rows), result.Data.Pagination.Total)
}
