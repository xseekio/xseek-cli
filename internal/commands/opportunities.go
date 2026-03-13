package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type Opportunity struct {
	Query                string       `json:"query"`
	Frequency            int          `json:"frequency"`
	Models               []string     `json:"models"`
	BusinessValue        string       `json:"businessValue"`
	SuggestedContentType string       `json:"suggestedContentType"`
	Competitors          []Competitor `json:"competitors"`
}

type Competitor struct {
	Company string `json:"company"`
	Count   int    `json:"count"`
}

func ListOpportunities(websiteID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	var opps []Opportunity
	err = client.GetJSON(fmt.Sprintf("/websites/%s/opportunities", websiteID), nil, &opps)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(opps)
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
				names = append(names, fmt.Sprintf("%s (%d)", c.Company, c.Count))
			}
			fmt.Printf("   Competitors: %s\n", strings.Join(names, ", "))
		}
	}
}
