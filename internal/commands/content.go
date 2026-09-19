package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type ContentHit struct {
	Kind           string   `json:"kind"`
	SourceID       string   `json:"sourceId"`
	URL            string   `json:"url"`
	Title          string   `json:"title"`
	Excerpt        string   `json:"excerpt"`
	Score          float64  `json:"score"`
	Similarity     float64  `json:"similarity"`
	Language       string   `json:"language"`
	Status         string   `json:"status"`
	KeywordTerms   []string `json:"keywordTerms"`
	TranslationOf  string   `json:"translationOfId"`
}

type ContentIndexed struct {
	Total     int            `json:"total"`
	ByKind    map[string]int `json:"byKind"`
	Languages []string       `json:"languages"`
}

type ContentSearchResponse struct {
	Query           string         `json:"query"`
	Indexed         ContentIndexed `json:"indexed"`
	MinSimilarity   float64        `json:"minSimilarity"`
	PendingIndexing int            `json:"pendingIndexing"`
	Results         []ContentHit   `json:"results"`
	Total           int            `json:"total"`
}

// SearchContent answers "does this site already say X?" against everything
// xSeek holds for it, instead of paging a list of titles.
//
// `articles list` pages at 20 and was never a dedup check: on a site with 399
// articles the first page is 5% of the corpus, and choosing CREATE against
// REWRITE from it is how six articles on one keyword get published.
//
// With no query it returns the SHAPE of the corpus and nothing else: how many
// documents, of which kind, in which languages.
func SearchContent(websiteID string, flags map[string]string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}
	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	for _, k := range []string{"q", "kind", "limit", "minSimilarity", "excludeSourceId"} {
		if v := flags[k]; v != "" {
			params[k] = v
		}
	}
	if v := flags["query"]; v != "" && params["q"] == "" {
		params["q"] = v
	}
	if v := flags["min-similarity"]; v != "" && params["minSimilarity"] == "" {
		params["minSimilarity"] = v
	}

	var result ContentSearchResponse
	if err := client.GetJSON(fmt.Sprintf("/websites/%s/content/search", websiteID), params, &result); err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	// No query: the caller wanted the shape, so give it and stop.
	if params["q"] == "" {
		printIndexed(result.Indexed, result.PendingIndexing)
		return
	}

	if len(result.Results) == 0 {
		printIndexed(result.Indexed, result.PendingIndexing)
		if result.Indexed.Total == 0 {
			fmt.Printf("\nNothing is indexed for this site yet, so this is not evidence the subject is uncovered.\n")
		} else {
			fmt.Printf("\nNo document above %.2f similarity for %q. This site does not cover it.\n", result.MinSimilarity, result.Query)
		}
		return
	}

	if isMarkdown() {
		var b strings.Builder
		fmt.Fprintf(&b, "## Already covered? %q\n\n", result.Query)
		fmt.Fprintf(&b, "%d of %d indexed documents matched above %.2f similarity.\n\n",
			len(result.Results), result.Indexed.Total, result.MinSimilarity)
		for _, h := range result.Results {
			fmt.Fprintf(&b, "- **%.3f** [%s](%s) `%s`", h.Similarity, h.Title, h.URL, h.Kind)
			if h.Status != "" {
				fmt.Fprintf(&b, " `%s`", h.Status)
			}
			if len(h.KeywordTerms) > 0 {
				fmt.Fprintf(&b, " keywords: %s", strings.Join(h.KeywordTerms, ", "))
			}
			fmt.Fprintf(&b, "\n  > %s\n", strings.ReplaceAll(h.Excerpt, "\n", " "))
		}
		fmt.Fprintf(&b, "\nJudge on the OBJECTIVE, not shared words. Same reader, same decision,\nsame competitors judged on the same criteria means a duplicate. Read the\nfull article before calling it one.\n")
		fmt.Print(b.String())
		return
	}

	printIndexed(result.Indexed, result.PendingIndexing)
	fmt.Printf("\n%d match above %.2f similarity for %q\n", len(result.Results), result.MinSimilarity, result.Query)
	fmt.Println(strings.Repeat("─", 72))
	for _, h := range result.Results {
		fmt.Printf("  %.3f  %s\n", h.Similarity, h.Title)
		fmt.Printf("         %s  [%s%s]\n", h.URL, h.Kind, statusSuffix(h.Status))
		if h.Excerpt != "" {
			fmt.Printf("         %s\n", truncate(strings.ReplaceAll(h.Excerpt, "\n", " "), 150))
		}
	}
}

// KeywordConflicts lists articles of one site that were given the SAME primary
// keyword. Exact string equality, no model: the crudest possible check, and it
// still found six-way collisions in production.
func KeywordConflicts(websiteID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}
	websiteID = resolveWebsiteID(client, websiteID)

	var result struct {
		Conflicts []struct {
			Keyword  string `json:"keyword"`
			Articles []struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				URL      string `json:"url"`
				Status   string `json:"status"`
				Language string `json:"language"`
			} `json:"articles"`
		} `json:"conflicts"`
	}
	if err := client.GetJSON(fmt.Sprintf("/websites/%s/content/keyword-conflicts", websiteID), nil, &result); err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}
	if len(result.Conflicts) == 0 {
		fmt.Println("No two articles share a primary keyword. Nothing to consolidate on this test.")
		return
	}
	fmt.Printf("Keywords carrying more than one article: %d\n", len(result.Conflicts))
	fmt.Println(strings.Repeat("─", 72))
	for _, c := range result.Conflicts {
		fmt.Printf("\n  %q  (%d articles)\n", c.Keyword, len(c.Articles))
		for _, a := range c.Articles {
			fmt.Printf("    • [%s] %s\n", a.Status, a.Title)
			if a.URL != "" {
				fmt.Printf("      %s\n", a.URL)
			}
		}
	}
	fmt.Printf("\nTranslations are excluded: a second language is deliberate, not a duplicate.\n")
}

func printIndexed(ix ContentIndexed, pending int) {
	kinds := []string{}
	for k, n := range ix.ByKind {
		kinds = append(kinds, fmt.Sprintf("%s %d", k, n))
	}
	fmt.Printf("Indexed: %d documents", ix.Total)
	if len(kinds) > 0 {
		fmt.Printf(" (%s)", strings.Join(kinds, ", "))
	}
	if len(ix.Languages) > 0 {
		fmt.Printf(" · languages: %s", strings.Join(ix.Languages, ", "))
	}
	fmt.Println()
	if pending > 0 {
		fmt.Printf("%d documents are still being indexed, so a \"not covered\" answer is provisional.\n", pending)
	}
}

func statusSuffix(s string) string {
	if s == "" {
		return ""
	}
	return " " + s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

var _ = strconv.Itoa
