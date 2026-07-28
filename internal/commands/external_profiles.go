package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type ExternalProfile struct {
	Domain string `json:"domain"`
	URL    string `json:"url"`
	Label  string `json:"label"`
}

type ExternalProfilesResponse struct {
	Profiles []ExternalProfile `json:"profiles"`
	Domains  []string          `json:"domains"`
}

// ListExternalProfiles prints the directory/platform profiles the brand already
// maintains (G2, Capterra, …). Used to avoid recommending a listing for a
// directory we're already on.
func ListExternalProfiles(websiteID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	var result ExternalProfilesResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/external-profiles", websiteID), nil, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result.Profiles)
		return
	}

	if len(result.Profiles) == 0 {
		fmt.Println("No external profiles yet.")
		return
	}

	fmt.Println("External Profiles")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("  %-24s %s\n", "Domain", "URL")
	fmt.Println(strings.Repeat("─", 70))
	for _, p := range result.Profiles {
		url := p.URL
		if len(url) > 44 {
			url = url[:41] + "..."
		}
		fmt.Printf("  %-24s %s\n", p.Domain, url)
	}
	fmt.Printf("\n%d profile(s)\n", len(result.Profiles))
}
