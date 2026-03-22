package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type Opportunity struct {
	Query                string           `json:"query"`
	Frequency            int              `json:"frequency"`
	Models               []string         `json:"models"`
	BusinessValue        string           `json:"businessValue"`
	SuggestedContentType string           `json:"suggestedContentType"`
	Competitors          []Competitor     `json:"competitors"`
	CompetitorDomains    []string         `json:"competitorDomains"`
	RankingArticles      []RankingArticle `json:"rankingArticles"`
}

type Competitor struct {
	Company string `json:"company"`
	Count   int    `json:"count"`
}

type RankingArticle struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type OpportunitiesResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Opportunities []Opportunity `json:"opportunities"`
	} `json:"data"`
}

func ListOpportunities(websiteID string, businessValue string, contentType string, limit string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if businessValue != "" {
		params["businessValue"] = businessValue
	}
	if contentType != "" {
		params["contentType"] = contentType
	}
	if limit != "" {
		params["limit"] = limit
	}

	var result OpportunitiesResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/opportunities", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	opps := result.Data.Opportunities

	if isJSON() {
		printJSON(result)
		return
	}

	if len(opps) == 0 {
		fmt.Println("No opportunities found. Run more prompts to discover content gaps.")
		return
	}

	fmt.Printf("Found %d opportunities\n", len(opps))
	fmt.Println(strings.Repeat("─", 80))

	for i, opp := range opps {
		value := opp.BusinessValue
		switch value {
		case "critical":
			value = "CRITICAL"
		case "high":
			value = "HIGH"
		case "medium":
			value = "MEDIUM"
		case "low":
			value = "LOW"
		}

		fmt.Printf("\n%d. \"%s\"\n", i+1, opp.Query)
		fmt.Printf("   Value: %-10s  Type: %-12s  Frequency: %dx  Engines: %s\n",
			value, opp.SuggestedContentType, opp.Frequency, strings.Join(opp.Models, ", "))

		if len(opp.Competitors) > 0 {
			names := make([]string, 0, len(opp.Competitors))
			for _, c := range opp.Competitors {
				if c.Count > 0 {
					names = append(names, fmt.Sprintf("%s (%d)", c.Company, c.Count))
				}
			}
			if len(names) > 0 {
				fmt.Printf("   Competitors: %s\n", strings.Join(names, ", "))
			}
		}

		if len(opp.RankingArticles) > 0 {
			fmt.Printf("   Ranking articles:\n")
			for _, a := range opp.RankingArticles {
				fmt.Printf("     - %s (%d citations)\n", a.URL, a.Count)
			}
		}
	}
}
