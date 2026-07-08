package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type CannibalUrl struct {
	URL            string `json:"url"`
	Citations      int    `json:"citations"`
	GSCImpressions int    `json:"gscImpressions"`
	AIImpressions  int    `json:"aiImpressions"`
	AITrend        *int   `json:"aiTrend"`
	Verdict        string `json:"verdict"`
}

type CannibalCluster struct {
	Query             string        `json:"query"`
	TrackedCitations  int           `json:"trackedCitations"`
	Coverage          string        `json:"coverage"`
	Competitors       []Competitor  `json:"competitors"`
	Urls              []CannibalUrl `json:"urls"`
	SuggestedCanonical string       `json:"suggestedCanonical"`
	SuggestedFolds    []string      `json:"suggestedFolds"`
}

type CannibalizationResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Clusters []CannibalCluster `json:"clusters"`
		Total    int               `json:"total"`
		Days     int               `json:"days"`
		MinUrls  int               `json:"minUrls"`
	} `json:"data"`
}

// ListCannibalization shows tracked questions where >= minUrls of our own URLs
// are cited (authority split), with a canonical/keep/fold verdict per URL.
func ListCannibalization(websiteID string, minUrls string, days string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if minUrls != "" {
		params["min-urls"] = minUrls
	}
	if days != "" {
		params["days"] = days
	}

	var result CannibalizationResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/cannibalization", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	if len(result.Data.Clusters) == 0 {
		fmt.Println("No cannibalization found. No tracked question has 2+ of your URLs cited.")
		return
	}

	fmt.Println("Cannibalization — questions where several of your URLs compete")
	fmt.Println(strings.Repeat("─", 100))

	for _, c := range result.Data.Clusters {
		q := c.Query
		if len(q) > 80 {
			q = q[:77] + "..."
		}
		fmt.Printf("\n❝ %s ❞\n", q)
		fmt.Printf("   coverage %s · %d citations · %d urls\n", c.Coverage, c.TrackedCitations, len(c.Urls))
		fmt.Printf("   %-56s %6s %8s %6s  %s\n", "url", "cites", "gsc", "ai", "verdict")
		for _, u := range c.Urls {
			path := u.URL
			if i := strings.Index(path, "/blogs/articles/"); i >= 0 {
				path = path[i+len("/blogs/articles/"):]
			}
			if len(path) > 54 {
				path = path[:51] + "..."
			}
			trend := ""
			if u.AITrend != nil {
				trend = fmt.Sprintf(" (%+d%%)", *u.AITrend)
			}
			fmt.Printf("   %-56s %6d %8d %4d%-2s  %s\n", path, u.Citations, u.GSCImpressions, u.AIImpressions, trend, u.Verdict)
		}
		if c.SuggestedCanonical != "" {
			fmt.Printf("   → canonical: %s\n", c.SuggestedCanonical)
		}
	}
	fmt.Printf("\n%d cannibalized clusters (%d-day window)\n", result.Data.Total, result.Data.Days)
}
