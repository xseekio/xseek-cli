package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type SitemapPage struct {
	URL            string  `json:"url"`
	Path           string  `json:"path"`
	AIImpressions  int     `json:"aiImpressions"`
	AITrend        *int    `json:"aiTrend"`
	GSCImpressions int     `json:"gscImpressions"`
	GSCClicks      int     `json:"gscClicks"`
	Warning        string  `json:"warning,omitempty"`
}

type SitemapPagesResponse struct {
	Success   bool          `json:"success"`
	Pages     []SitemapPage `json:"pages"`
	PageCount int           `json:"pageCount"`
	Filter    string        `json:"filter,omitempty"`
}

func ListSitemapPages(websiteID string, days string, filter string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if days != "" {
		params["days"] = days
	}
	if filter != "" {
		params["filter"] = filter
	}

	var result SitemapPagesResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/sitemap-pages", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	if len(result.Pages) == 0 {
		fmt.Println("No sitemap pages found.")
		return
	}

	if filter == "attention" {
		fmt.Println("Pages Needing Attention (AI traffic dropped >20%)")
	} else {
		fmt.Println("Sitemap Pages")
	}
	fmt.Println(strings.Repeat("─", 100))
	fmt.Printf("  %-50s %12s %8s %12s %8s\n", "Path", "AI Visits", "Trend", "GSC Impr.", "Clicks")
	fmt.Println(strings.Repeat("─", 100))

	for _, p := range result.Pages {
		path := p.Path
		if len(path) > 48 {
			path = path[:45] + "..."
		}
		trend := "—"
		if p.AITrend != nil {
			trend = fmt.Sprintf("%+d%%", *p.AITrend)
		}
		fmt.Printf("  %-50s %12d %8s %12d %8d\n", path, p.AIImpressions, trend, p.GSCImpressions, p.GSCClicks)
	}
	fmt.Printf("\n%d pages total\n", result.PageCount)
}
