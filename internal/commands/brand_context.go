package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type ICP struct {
	TargetAudience *string  `json:"targetAudience"`
	Industry       *string  `json:"industry"`
	CompanySize    *string  `json:"companySize"`
	PainPoints     []string `json:"painPoints"`
	UseCase        *string  `json:"useCase"`
	BuyerRole      *string  `json:"buyerRole"`
}

type BrandContextResponse struct {
	Success bool `json:"success"`
	Data    struct {
		CompanyName          *string  `json:"companyName"`
		WebsiteURL           string   `json:"websiteUrl"`
		BrandVoiceGuidelines *string  `json:"brandVoiceGuidelines"`
		BrandTone            *string  `json:"brandTone"`
		ICP                  *ICP     `json:"icp"`
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

	if d.ICP != nil {
		fmt.Println()
		fmt.Println("Ideal Customer Profile (ICP):")
		if d.ICP.TargetAudience != nil {
			fmt.Printf("  Audience:     %s\n", *d.ICP.TargetAudience)
		}
		if d.ICP.Industry != nil {
			fmt.Printf("  Industry:     %s\n", *d.ICP.Industry)
		}
		if d.ICP.CompanySize != nil {
			fmt.Printf("  Company size: %s\n", *d.ICP.CompanySize)
		}
		if d.ICP.BuyerRole != nil {
			fmt.Printf("  Buyer role:   %s\n", *d.ICP.BuyerRole)
		}
		if d.ICP.UseCase != nil {
			fmt.Printf("  Use case:     %s\n", *d.ICP.UseCase)
		}
		if len(d.ICP.PainPoints) > 0 {
			fmt.Println("  Pain points:")
			for _, p := range d.ICP.PainPoints {
				fmt.Printf("    - %s\n", p)
			}
		}
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
