package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type RobotsScanResult struct {
	URL         string      `json:"url"`
	Accessible  bool        `json:"accessible"`
	BotStatuses []BotStatus `json:"botStatuses"`
}

type BotStatus struct {
	Bot     string `json:"bot"`
	Allowed bool   `json:"allowed"`
	Label   string `json:"label"`
}

func ScanRobots(domain string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	// Use the public tools endpoint
	var result RobotsScanResult
	err = client.GetJSON("/tools/robots-txt", map[string]string{"url": domain}, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	fmt.Printf("Robots.txt scan for %s\n", result.URL)
	fmt.Println(strings.Repeat("─", 50))

	if !result.Accessible {
		fmt.Println("⚠  robots.txt not accessible")
		return
	}

	for _, bot := range result.BotStatuses {
		status := "✓ Allowed"
		if !bot.Allowed {
			status = "✗ Blocked"
		}
		fmt.Printf("  %-25s %s\n", bot.Label, status)
	}
}
