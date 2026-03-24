package commands

import (
	"fmt"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type RelatedKeyword struct {
	Keyword           string  `json:"keyword"`
	SearchVolume      *int    `json:"searchVolume"`
	KeywordDifficulty *int    `json:"keywordDifficulty"`
}

type KeywordResult struct {
	Query             string           `json:"query"`
	MatchedKeyword    *string          `json:"matchedKeyword"`
	SearchVolume      *int             `json:"searchVolume"`
	KeywordDifficulty *int             `json:"keywordDifficulty"`
	CPC               *float64         `json:"cpc"`
	Competition       *float64         `json:"competition"`
	RelatedKeywords   []RelatedKeyword `json:"relatedKeywords"`
}

type KeywordsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Results []KeywordResult `json:"results"`
	} `json:"data"`
}

func SearchKeywords(websiteID string, query string, language string, location string) {
	if query == "" {
		exitError("query is required: xseek keywords <website> \"<topic>\"")
	}

	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{
		"q": query,
	}
	if language != "" {
		params["language"] = language
	}
	if location != "" {
		params["location"] = location
	}

	var result KeywordsResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/keywords", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result.Data.Results)
		return
	}

	for _, r := range result.Data.Results {
		fmt.Printf("Query: %s\n", r.Query)
		if r.MatchedKeyword != nil {
			fmt.Printf("Best keyword: %s\n", *r.MatchedKeyword)
		}
		if r.SearchVolume != nil {
			fmt.Printf("Search volume: %d/mo\n", *r.SearchVolume)
		}
		if r.KeywordDifficulty != nil {
			fmt.Printf("Keyword difficulty: %d/100\n", *r.KeywordDifficulty)
		}

		if len(r.RelatedKeywords) > 0 {
			fmt.Println()
			fmt.Println("Related Keywords")
			fmt.Println(strings.Repeat("─", 60))
			fmt.Printf("  %-35s %8s %5s\n", "Keyword", "Vol/mo", "KD")
			fmt.Println(strings.Repeat("─", 60))
			for _, rk := range r.RelatedKeywords {
				vol := "-"
				if rk.SearchVolume != nil {
					vol = fmt.Sprintf("%d", *rk.SearchVolume)
				}
				kd := "-"
				if rk.KeywordDifficulty != nil {
					kd = fmt.Sprintf("%d", *rk.KeywordDifficulty)
				}
				keyword := rk.Keyword
				if len(keyword) > 33 {
					keyword = keyword[:30] + "..."
				}
				fmt.Printf("  %-35s %8s %5s\n", keyword, vol, kd)
			}
			fmt.Printf("\n%d related keywords found\n", len(r.RelatedKeywords))
		}
		fmt.Println()
	}
}
