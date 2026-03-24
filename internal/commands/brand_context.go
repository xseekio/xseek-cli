package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type BrandContextResponse struct {
	Success bool `json:"success"`
	Data    struct {
		CompanyName          *string  `json:"companyName"`
		WebsiteURL           string   `json:"websiteUrl"`
		BrandVoiceGuidelines *string  `json:"brandVoiceGuidelines"`
		BrandTone            *string  `json:"brandTone"`
		KnowledgeChunks      []string `json:"knowledgeChunks"`
		BrandVoiceSamples    []string `json:"brandVoiceSamples"`
	} `json:"data"`
}

func GetBrandContext(websiteID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	var result BrandContextResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/brand-context", websiteID), nil, &result)
	if err != nil {
		exitError(err.Error())
	}

	d := result.Data

	if isJSON() {
		printJSON(d)
		return
	}

	fmt.Println("Brand Context")
	fmt.Println(strings.Repeat("─", 60))

	if d.CompanyName != nil && *d.CompanyName != "" {
		fmt.Printf("  Company:   %s\n", *d.CompanyName)
	}
	fmt.Printf("  Website:   %s\n", d.WebsiteURL)
	if d.BrandTone != nil && *d.BrandTone != "" {
		fmt.Printf("  Tone:      %s\n", *d.BrandTone)
	}

	if d.BrandVoiceGuidelines != nil && *d.BrandVoiceGuidelines != "" {
		fmt.Println()
		fmt.Println("Voice Guidelines:")
		fmt.Println(*d.BrandVoiceGuidelines)
	}

	if len(d.KnowledgeChunks) > 0 {
		fmt.Println()
		fmt.Printf("Knowledge Base: %d chunks\n", len(d.KnowledgeChunks))
		for i, chunk := range d.KnowledgeChunks {
			preview := chunk
			if len(preview) > 100 {
				preview = preview[:97] + "..."
			}
			fmt.Printf("  [%d] %s\n", i+1, preview)
		}
	}

	if len(d.BrandVoiceSamples) > 0 {
		fmt.Println()
		fmt.Printf("Brand Voice Samples: %d\n", len(d.BrandVoiceSamples))
		for i, sample := range d.BrandVoiceSamples {
			preview := sample
			if len(preview) > 100 {
				preview = preview[:97] + "..."
			}
			fmt.Printf("  [%d] %s\n", i+1, preview)
		}
	}
}
