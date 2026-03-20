package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

// resolveWebsiteID takes a website ID or URL and returns the website ID.
// If the input looks like a URL/domain, it fetches the website list and matches.
func resolveWebsiteID(client *api.Client, input string) string {
	// If it looks like a UUID, use it directly
	if len(input) == 36 && strings.Count(input, "-") == 4 {
		return input
	}

	// Otherwise, try to match by URL
	var result WebsitesResponse
	err := client.GetJSON("/websites", nil, &result)
	if err != nil {
		exitError(fmt.Sprintf("failed to resolve website: %s", err))
	}
	websites := result.Data

	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimPrefix(input, "www.")
	input = strings.TrimSuffix(input, "/")

	for _, w := range websites {
		wURL := strings.TrimPrefix(w.URL, "https://")
		wURL = strings.TrimPrefix(wURL, "http://")
		wURL = strings.TrimPrefix(wURL, "www.")
		wURL = strings.TrimSuffix(wURL, "/")

		if strings.EqualFold(wURL, input) {
			return w.ID
		}
	}

	// If only one website, use it
	if len(websites) == 1 {
		return websites[0].ID
	}

	exitError(fmt.Sprintf("could not find website matching '%s'\nRun 'xseek websites' to see your websites", input))
	return "" // unreachable
}
