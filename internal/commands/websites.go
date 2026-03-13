package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type Website struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	CompanyName string `json:"companyName"`
}

func ListWebsites() {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	var websites []Website
	err = client.GetJSON("/websites", nil, &websites)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(websites)
		return
	}

	if len(websites) == 0 {
		fmt.Println("No websites found. Add one at https://www.xseek.io/dashboard")
		return
	}

	fmt.Printf("%-38s  %-30s  %s\n", "ID", "URL", "Company")
	fmt.Println(strings.Repeat("─", 90))
	for _, w := range websites {
		fmt.Printf("%-38s  %-30s  %s\n", w.ID, w.URL, w.CompanyName)
	}
}
