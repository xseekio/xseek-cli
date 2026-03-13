package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type Source struct {
	URL    string `json:"url"`
	Domain string `json:"domain"`
	Title  string `json:"title"`
	Count  int    `json:"count"`
}

type SourcesResponse struct {
	Sources []Source `json:"sources"`
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

	if isJSON() {
		printJSON(result)
		return
	}

	if len(result.Sources) == 0 {
		fmt.Println("No sources found yet.")
		return
	}

	fmt.Println("AI Citation Sources")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("  %-5s %-30s %s\n", "Count", "Domain", "URL")
	fmt.Println(strings.Repeat("─", 70))
	for _, s := range result.Sources {
		url := s.URL
		if len(url) > 50 {
			url = url[:47] + "..."
		}
		fmt.Printf("  %-5d %-30s %s\n", s.Count, s.Domain, url)
	}
}
