package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

// BrandIdentity mirrors the JSONB column on websites — kept in sync with
// /api/v1/websites/{id}/brand-context. All fields are optional; an unfilled
// brand profile is just `{}`.
type BrandIdentity struct {
	Adjectives      []string `json:"adjectives,omitempty"`
	SignatureWords  []string `json:"signatureWords,omitempty"`
	BannedWords     []string `json:"bannedWords,omitempty"`
	Positions       string   `json:"positions,omitempty"`
	OpeningExamples string   `json:"openingExamples,omitempty"`
	AdmiredBrands   string   `json:"admiredBrands,omitempty"`
	AntiReferences  string   `json:"antiReferences,omitempty"`
	OwnContentUrls  string   `json:"ownContentUrls,omitempty"`
	AlwaysSurface   string   `json:"alwaysSurface,omitempty"`
	NeverSurface    string   `json:"neverSurface,omitempty"`
}

type AudienceTopic struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Audience struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Topics      []AudienceTopic `json:"topics"`
}

type BrandContextResponse struct {
	Success bool `json:"success"`
	Data    struct {
		CompanyName          *string        `json:"companyName"`
		WebsiteURL           string         `json:"websiteUrl"`
		Language             string         `json:"language"`
		BrandVoiceGuidelines *string        `json:"brandVoiceGuidelines"`
		BrandTone            *string        `json:"brandTone"`
		BrandIdentity        *BrandIdentity `json:"brandIdentity"`
		Audiences            []Audience     `json:"audiences"`
		KnowledgeChunks      []string       `json:"knowledgeChunks"`
		BrandVoiceSamples    []string       `json:"brandVoiceSamples"`
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

	if isMarkdown() {
		fmt.Print(renderBrandContextMarkdown(d))
		return
	}

	renderBrandContextHuman(d)
}

func renderBrandContextHuman(d struct {
	CompanyName          *string        `json:"companyName"`
	WebsiteURL           string         `json:"websiteUrl"`
	Language             string         `json:"language"`
	BrandVoiceGuidelines *string        `json:"brandVoiceGuidelines"`
	BrandTone            *string        `json:"brandTone"`
	BrandIdentity        *BrandIdentity `json:"brandIdentity"`
	Audiences            []Audience     `json:"audiences"`
	KnowledgeChunks      []string       `json:"knowledgeChunks"`
	BrandVoiceSamples    []string       `json:"brandVoiceSamples"`
}) {
	fmt.Println("Brand Context")
	fmt.Println(strings.Repeat("─", 60))

	if d.CompanyName != nil && *d.CompanyName != "" {
		fmt.Printf("  Company:   %s\n", *d.CompanyName)
	}
	fmt.Printf("  Website:   %s\n", d.WebsiteURL)
	if d.BrandTone != nil && *d.BrandTone != "" {
		fmt.Printf("  Tone:      %s\n", *d.BrandTone)
	}

	if d.BrandIdentity != nil {
		bi := d.BrandIdentity
		hasAny := len(bi.Adjectives) > 0 || len(bi.SignatureWords) > 0 || len(bi.BannedWords) > 0 ||
			bi.Positions != "" || bi.OpeningExamples != "" || bi.AdmiredBrands != "" ||
			bi.AntiReferences != "" || bi.OwnContentUrls != "" || bi.AlwaysSurface != "" ||
			bi.NeverSurface != ""
		if hasAny {
			fmt.Println()
			fmt.Println("Brand Identity:")
			if len(bi.Adjectives) > 0 {
				fmt.Printf("  Adjectives:       %s\n", strings.Join(bi.Adjectives, ", "))
			}
			if len(bi.SignatureWords) > 0 {
				fmt.Printf("  Signature words:  %s\n", strings.Join(bi.SignatureWords, ", "))
			}
			if len(bi.BannedWords) > 0 {
				fmt.Printf("  Banned words:     %s\n", strings.Join(bi.BannedWords, ", "))
			}
			if bi.Positions != "" {
				fmt.Println("  Positions:")
				fmt.Println(indent(bi.Positions, "    "))
			}
			if bi.OpeningExamples != "" {
				fmt.Println("  Opening examples:")
				fmt.Println(indent(bi.OpeningExamples, "    "))
			}
			if bi.AdmiredBrands != "" {
				fmt.Println("  Admired brands:")
				fmt.Println(indent(bi.AdmiredBrands, "    "))
			}
			if bi.AntiReferences != "" {
				fmt.Println("  Voices to avoid:")
				fmt.Println(indent(bi.AntiReferences, "    "))
			}
			if bi.OwnContentUrls != "" {
				fmt.Println("  Own content URLs:")
				fmt.Println(indent(bi.OwnContentUrls, "    "))
			}
			if bi.AlwaysSurface != "" {
				fmt.Println("  Always surface:")
				fmt.Println(indent(bi.AlwaysSurface, "    "))
			}
			if bi.NeverSurface != "" {
				fmt.Println("  Never surface:")
				fmt.Println(indent(bi.NeverSurface, "    "))
			}
		}
	}

	if d.BrandVoiceGuidelines != nil && *d.BrandVoiceGuidelines != "" {
		fmt.Println()
		fmt.Println("Voice Guidelines:")
		fmt.Println(*d.BrandVoiceGuidelines)
	}

	if len(d.Audiences) > 0 {
		fmt.Println()
		fmt.Printf("Audiences: %d\n", len(d.Audiences))
		for _, a := range d.Audiences {
			fmt.Printf("  • %s", a.Name)
			if a.Description != "" {
				fmt.Printf(" — %s", a.Description)
			}
			fmt.Println()
			for _, t := range a.Topics {
				fmt.Printf("      ◦ %s", t.Name)
				if t.Description != "" {
					fmt.Printf(" — %s", t.Description)
				}
				fmt.Println()
			}
		}
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

// renderBrandContextMarkdown produces a single brand brief that AI agents
// (Claude Code skills, MCP tools) can paste directly into prompts.
func renderBrandContextMarkdown(d struct {
	CompanyName          *string        `json:"companyName"`
	WebsiteURL           string         `json:"websiteUrl"`
	Language             string         `json:"language"`
	BrandVoiceGuidelines *string        `json:"brandVoiceGuidelines"`
	BrandTone            *string        `json:"brandTone"`
	BrandIdentity        *BrandIdentity `json:"brandIdentity"`
	Audiences            []Audience     `json:"audiences"`
	KnowledgeChunks      []string       `json:"knowledgeChunks"`
	BrandVoiceSamples    []string       `json:"brandVoiceSamples"`
}) string {
	var b strings.Builder

	companyLabel := d.WebsiteURL
	if d.CompanyName != nil && *d.CompanyName != "" {
		companyLabel = *d.CompanyName
	}
	fmt.Fprintf(&b, "# Brand brief: %s\n\n", companyLabel)
	fmt.Fprintf(&b, "- **Website:** %s\n", d.WebsiteURL)
	if d.BrandTone != nil && *d.BrandTone != "" {
		fmt.Fprintf(&b, "- **Tone:** %s\n", *d.BrandTone)
	}
	if d.Language != "" {
		fmt.Fprintf(&b, "- **Language:** %s\n", d.Language)
	}
	b.WriteString("\n")

	if d.BrandIdentity != nil {
		bi := d.BrandIdentity
		identityLines := []string{}
		if len(bi.Adjectives) > 0 {
			identityLines = append(identityLines, fmt.Sprintf("- **Adjectives:** %s", strings.Join(bi.Adjectives, ", ")))
		}
		if len(bi.SignatureWords) > 0 {
			identityLines = append(identityLines, fmt.Sprintf("- **Signature words:** %s", strings.Join(bi.SignatureWords, ", ")))
		}
		if len(bi.BannedWords) > 0 {
			identityLines = append(identityLines, fmt.Sprintf("- **Banned words (never use):** %s", strings.Join(bi.BannedWords, ", ")))
		}
		if len(identityLines) > 0 || bi.Positions != "" {
			b.WriteString("## Identity\n\n")
			for _, line := range identityLines {
				b.WriteString(line + "\n")
			}
			if bi.Positions != "" {
				b.WriteString("\n**What we stand for / what we reject:**\n\n")
				b.WriteString(bi.Positions + "\n")
			}
			b.WriteString("\n")
		}

		if d.BrandVoiceGuidelines != nil && *d.BrandVoiceGuidelines != "" || bi.OpeningExamples != "" {
			b.WriteString("## Voice\n\n")
			if d.BrandVoiceGuidelines != nil && *d.BrandVoiceGuidelines != "" {
				b.WriteString(*d.BrandVoiceGuidelines + "\n\n")
			}
			if bi.OpeningExamples != "" {
				b.WriteString("**Opening sentence examples:**\n\n")
				b.WriteString(bi.OpeningExamples + "\n\n")
			}
		}

		if bi.AdmiredBrands != "" || bi.AntiReferences != "" || bi.OwnContentUrls != "" {
			b.WriteString("## Anchors\n\n")
			if bi.AdmiredBrands != "" {
				b.WriteString("**Brands we admire:**\n\n")
				b.WriteString(bi.AdmiredBrands + "\n\n")
			}
			if bi.AntiReferences != "" {
				b.WriteString("**Voices to avoid:**\n\n")
				b.WriteString(bi.AntiReferences + "\n\n")
			}
			if bi.OwnContentUrls != "" {
				b.WriteString("**Our best content (reference URLs):**\n\n")
				b.WriteString(bi.OwnContentUrls + "\n\n")
			}
		}

		if bi.AlwaysSurface != "" || bi.NeverSurface != "" {
			b.WriteString("## Surface rules\n\n")
			if bi.AlwaysSurface != "" {
				b.WriteString("**Always surface:**\n\n")
				b.WriteString(bi.AlwaysSurface + "\n\n")
			}
			if bi.NeverSurface != "" {
				b.WriteString("**Never surface:**\n\n")
				b.WriteString(bi.NeverSurface + "\n\n")
			}
		}
	} else if d.BrandVoiceGuidelines != nil && *d.BrandVoiceGuidelines != "" {
		// No structured identity — still emit voice guidelines.
		b.WriteString("## Voice\n\n")
		b.WriteString(*d.BrandVoiceGuidelines + "\n\n")
	}

	if len(d.Audiences) > 0 {
		b.WriteString("## Audiences\n\n")
		for _, a := range d.Audiences {
			fmt.Fprintf(&b, "### %s", a.Name)
			if a.Description != "" {
				fmt.Fprintf(&b, " — %s", a.Description)
			}
			b.WriteString("\n\n")
			if len(a.Topics) > 0 {
				b.WriteString("Topics:\n\n")
				for _, t := range a.Topics {
					fmt.Fprintf(&b, "- **%s**", t.Name)
					if t.Description != "" {
						fmt.Fprintf(&b, " — %s", t.Description)
					}
					b.WriteString("\n")
				}
				b.WriteString("\n")
			}
		}
	}

	if len(d.KnowledgeChunks) > 0 {
		b.WriteString("## Knowledge entries\n\n")
		for i, chunk := range d.KnowledgeChunks {
			fmt.Fprintf(&b, "### Entry %d\n\n%s\n\n", i+1, chunk)
		}
	}

	if len(d.BrandVoiceSamples) > 0 {
		b.WriteString("## Style references (match this writing)\n\n")
		for i, sample := range d.BrandVoiceSamples {
			fmt.Fprintf(&b, "### Sample %d\n\n%s\n\n", i+1, sample)
		}
	}

	return b.String()
}

func indent(s, prefix string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
