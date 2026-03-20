package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type Source struct {
	URL              string  `json:"url"`
	Domain           string  `json:"domain"`
	Title            string  `json:"title"`
	CitationCount    int     `json:"citationCount"`
	CitationPercentage float64 `json:"citationPercentage"`
}

type SourcesResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Sources    []Source `json:"sources"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			Total      int `json:"total"`
			TotalPages int `json:"totalPages"`
		} `json:"pagination"`
	} `json:"data"`
}

func ListSources(websiteID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	var result SourcesResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/sources", websiteID), nil, &result)
	if err != nil {
		exitError(err.Error())
	}

	sources := result.Data.Sources

	if isJSON() {
		printJSON(sources)
		return
	}

	if len(sources) == 0 {
		fmt.Println("No sources found yet.")
		return
	}

	fmt.Println("AI Citation Sources")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("  %-5s %-30s %s\n", "Count", "Domain", "URL")
	fmt.Println(strings.Repeat("─", 70))
	for _, s := range sources {
		url := s.URL
		if len(url) > 50 {
			url = url[:47] + "..."
		}
		fmt.Printf("  %-5d %-30s %s\n", s.CitationCount, s.Domain, url)
	}
	fmt.Printf("\nShowing %d of %d sources\n", len(sources), result.Data.Pagination.Total)
}
