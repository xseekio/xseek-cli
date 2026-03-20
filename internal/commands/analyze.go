package commands

import (
	"fmt"

	"github.com/xseekio/xseek-cli/internal/api"
)

type AnalyzeResponse struct {
	Success  bool        `json:"success"`
	Analysis interface{} `json:"analysis"`
}

func AnalyzePage(websiteID string, pageURL string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	body := map[string]string{"url": pageURL}

	var result AnalyzeResponse
	err = client.PostJSON(fmt.Sprintf("/websites/%s/aeo-copilot/analyze", websiteID), body, &result)
	if err != nil {
		exitError(err.Error())
	}

	// Always output JSON for analyze since the analysis is a complex object
	printJSON(result)
}
