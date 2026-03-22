package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type LeaderboardEntry struct {
	Company      string `json:"company"`
	Mentions     int    `json:"mentions"`
	Rank         int    `json:"rank"`
	IsCompetitor bool   `json:"isCompetitor"`
}

type LeaderboardResponse struct {
	Success bool               `json:"success"`
	Data    []LeaderboardEntry `json:"data"`
	Meta    struct {
		LastDays      int `json:"lastDays"`
		TotalMentions int `json:"totalMentions"`
	} `json:"meta"`
}

func ReportLeaderboard(websiteID string, days string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if days != "" {
		params["lastDays"] = days
	}

	var result LeaderboardResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/prompts/leaderboard", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	leaderboard := result.Data

	if isJSON() {
		printJSON(leaderboard)
		return
	}

	if len(leaderboard) == 0 {
		fmt.Println("No leaderboard data yet. Run some prompts first.")
		return
	}

	fmt.Println("Brand Mention Leaderboard")
	fmt.Println(strings.Repeat("─", 50))
	for _, entry := range leaderboard {
		bar := strings.Repeat("█", min(entry.Mentions, 40))
		fmt.Printf("  #%-3d %-25s %3d  %s\n", entry.Rank, entry.Company, entry.Mentions, bar)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
