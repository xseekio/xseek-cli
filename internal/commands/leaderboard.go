package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type LeaderboardEntry struct {
	Company  string `json:"company"`
	Mentions int    `json:"mentions"`
	Rank     int    `json:"rank"`
}

type LeaderboardResponse struct {
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
}

func ReportLeaderboard(websiteID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	var result LeaderboardResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/prompts/leaderboard", websiteID), nil, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	if len(result.Leaderboard) == 0 {
		fmt.Println("No leaderboard data yet. Run some prompts first.")
		return
	}

	fmt.Println("Brand Mention Leaderboard")
	fmt.Println(strings.Repeat("─", 50))
	for _, entry := range result.Leaderboard {
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
