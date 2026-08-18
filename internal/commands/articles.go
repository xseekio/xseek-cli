package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

type Article struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Slug            string  `json:"slug"`
	MetaDescription *string `json:"metaDescription"`
	ContentMarkdown *string `json:"contentMarkdown"`
	Status          string  `json:"status"`
	QualityScore    *int    `json:"qualityScore"`
	OpportunityID   *string `json:"opportunityId"`
	PublishedAt     *string `json:"publishedAt"`
	PublishedURL    *string `json:"publishedUrl"`
	CreatedAt       string  `json:"createdAt"`
}

type ArticlesResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Articles   []Article `json:"articles"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			Total      int `json:"total"`
			TotalPages int `json:"totalPages"`
		} `json:"pagination"`
	} `json:"data"`
}

type ArticleResponse struct {
	Success bool    `json:"success"`
	Data    Article `json:"data"`
}

type ArticleCreateResponse struct {
	Success bool `json:"success"`
	Updated bool `json:"updated"`
	Data    struct {
		ID           string  `json:"id"`
		Title        string  `json:"title"`
		Slug         string  `json:"slug"`
		Status       string  `json:"status"`
		QualityScore *int    `json:"qualityScore"`
		CreatedAt    string  `json:"createdAt"`
		// Screenshot coverage as the server recorded it: brands named vs
		// brands carrying a capture. Printed back so the caller can compare
		// it against the article instead of assuming the payload landed.
		VisualCoverage struct {
			Named    int  `json:"named"`
			Captured int  `json:"captured"`
			Missing  int  `json:"missing"`
			Rate     *int `json:"rate"`
		} `json:"visualCoverage"`
		// What the article asserts vs what carries a source. An invented price
		// reads exactly like a checked one, so this is the only place the
		// difference shows.
		ClaimCoverage struct {
			Stated                int `json:"stated"`
			Verified              int `json:"verified"`
			Unverified            int `json:"unverified"`
			UnsourcedAttributions int `json:"unsourcedAttributions"`
		} `json:"claimCoverage"`
	} `json:"data"`
}

func ListArticles(websiteID string, status string, pageSize string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	params := map[string]string{}
	if status != "" {
		params["status"] = status
	}
	if pageSize != "" {
		params["pageSize"] = pageSize
	}

	var result ArticlesResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/articles", websiteID), params, &result)
	if err != nil {
		exitError(err.Error())
	}

	articles := result.Data.Articles

	if isJSON() {
		printJSON(articles)
		return
	}

	if len(articles) == 0 {
		fmt.Println("No articles found.")
		return
	}

	fmt.Println("Content Studio — Articles")
	fmt.Println(strings.Repeat("─", 90))
	fmt.Printf("  %-36s %-30s %-10s %-5s %s\n", "ID", "Title", "Status", "Score", "Date")
	fmt.Println(strings.Repeat("─", 90))
	for _, a := range articles {
		title := a.Title
		if len(title) > 28 {
			title = title[:25] + "..."
		}
		score := "-"
		if a.QualityScore != nil {
			score = fmt.Sprintf("%d", *a.QualityScore)
		}
		date := a.CreatedAt
		if len(date) > 10 {
			date = date[:10]
		}
		fmt.Printf("  %-36s %-30s %-10s %-5s %s\n", a.ID, title, a.Status, score, date)
	}
	fmt.Printf("\nShowing %d of %d articles\n", len(articles), result.Data.Pagination.Total)
}

// loadVisuals reads the screenshot coverage record written by the
// /screenshots skill: one entry per brand the article names, including the
// ones that got no capture and why. Bad JSON is a warning, never a failure.
// The article is the deliverable; this is the audit trail beside it.
// loadJSONArray reads a coverage record the skill wrote: one entry per thing
// the article claims or should have captured. Bad JSON is a warning, never a
// failure. The article is the deliverable; these are the audit trail beside it.
func loadJSONArray(path, flag string) []interface{} {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %s %s could not be read (%s); pushing without it\n", flag, path, err)
		return nil
	}
	var list []interface{}
	if err := json.Unmarshal(data, &list); err != nil {
		fmt.Fprintf(os.Stderr, "warning: %s %s is not a JSON array (%s); pushing without it\n", flag, path, err)
		return nil
	}
	return list
}

func loadVisuals(path string) []interface{} {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: --visuals %s could not be read (%s); pushing without the coverage record\n", path, err)
		return nil
	}
	var list []interface{}
	if err := json.Unmarshal(data, &list); err != nil {
		fmt.Fprintf(os.Stderr, "warning: --visuals %s is not a JSON array (%s); pushing without the coverage record\n", path, err)
		return nil
	}
	return list
}

func PushArticle(websiteID string, title string, filePath string, status string, metaDescription string, keywordTerm string, keywords string, opportunityID string, visualsPath string, description string, metaTitle string, claimsPath string) {
	if title == "" {
		exitError("--title is required")
	}

	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	// Read content from file or stdin
	var content string
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			exitError(fmt.Sprintf("failed to read file: %s", err))
		}
		content = string(data)
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				exitError(fmt.Sprintf("failed to read stdin: %s", err))
			}
			content = string(data)
		}
	}

	body := map[string]interface{}{
		"title": title,
	}
	if content != "" {
		body["contentMarkdown"] = content
	}
	if status != "" {
		body["status"] = status
	} else {
		body["status"] = "draft"
	}
	if metaDescription != "" {
		body["metaDescription"] = metaDescription
	}
	// Four fields, two audiences: title/description are page copy for a reader
	// who already arrived; metaTitle/metaDescription are the tags, read in a
	// search result or by an engine deciding whether to cite.
	if description != "" {
		body["description"] = description
	}
	if metaTitle != "" {
		body["metaTitle"] = metaTitle
	}
	if keywordTerm != "" {
		body["keywordTerm"] = keywordTerm
	}
	// Full target-keyword list (comma-separated). The API stores it as
	// keyword_terms (primary first — keywordTerm is promoted to the front
	// server-side); an article targets a primary + several secondary terms.
	if keywords != "" {
		var list []string
		for _, k := range strings.Split(keywords, ",") {
			if t := strings.TrimSpace(k); t != "" {
				list = append(list, t)
			}
		}
		if len(list) > 0 {
			body["keywordTerms"] = list
		}
	}
	if opportunityID != "" {
		body["opportunityId"] = opportunityID
	}
	if visuals := loadVisuals(visualsPath); len(visuals) > 0 {
		body["visuals"] = visuals
	}
	if claims := loadJSONArray(claimsPath, "--claims"); len(claims) > 0 {
		body["claims"] = claims
	}

	var result ArticleCreateResponse
	err = client.PostJSON(fmt.Sprintf("/websites/%s/articles", websiteID), body, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	if result.Updated {
		fmt.Printf("Article updated (existing match found)\n")
	} else {
		fmt.Printf("Article created\n")
	}
	fmt.Printf("  ID:     %s\n", result.Data.ID)
	fmt.Printf("  Title:  %s\n", result.Data.Title)
	fmt.Printf("  Status: %s\n", result.Data.Status)
	if c := result.Data.ClaimCoverage; c.Stated > 0 {
		fmt.Printf("  Claims:  %d/%d sourced", c.Verified, c.Stated)
		if c.UnsourcedAttributions > 0 {
			fmt.Printf(" · %d credited to a named source with NO link", c.UnsourcedAttributions)
		}
		fmt.Println()
	}
	if c := result.Data.VisualCoverage; c.Named > 0 {
		fmt.Printf("  Visuals: %d/%d brands captured", c.Captured, c.Named)
		if c.Missing > 0 {
			fmt.Printf(" (%d recorded without a screenshot)", c.Missing)
		}
		fmt.Println()
	}
}

func GetArticle(websiteID string, articleID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	var result ArticleResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/articles/%s", websiteID, articleID), nil, &result)
	if err != nil {
		exitError(err.Error())
	}

	a := result.Data

	if isJSON() {
		printJSON(a)
		return
	}

	fmt.Printf("Title:  %s\n", a.Title)
	fmt.Printf("ID:     %s\n", a.ID)
	fmt.Printf("Status: %s\n", a.Status)
	if a.QualityScore != nil {
		fmt.Printf("Score:  %d\n", *a.QualityScore)
	}
	if a.PublishedURL != nil && *a.PublishedURL != "" {
		fmt.Printf("URL:    %s\n", *a.PublishedURL)
	}
	fmt.Println()
	if a.ContentMarkdown != nil {
		fmt.Println(*a.ContentMarkdown)
	}
}

func UpdateArticle(websiteID string, articleID string, filePath string, title string, status string, metaDescription string, opportunityID string, visualsPath string, description string, metaTitle string, claimsPath string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	// Read content from file or stdin
	var content string
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			exitError(fmt.Sprintf("failed to read file: %s", err))
		}
		content = string(data)
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				exitError(fmt.Sprintf("failed to read stdin: %s", err))
			}
			content = string(data)
		}
	}

	body := map[string]interface{}{}
	if content != "" {
		body["contentMarkdown"] = content
	}
	if title != "" {
		body["title"] = title
	}
	if status != "" {
		body["status"] = status
	}
	if metaDescription != "" {
		body["metaDescription"] = metaDescription
	}
	if description != "" {
		body["description"] = description
	}
	if metaTitle != "" {
		body["metaTitle"] = metaTitle
	}
	if opportunityID != "" {
		// Pass empty string ("") via "none" sentinel if you ever need to unlink;
		// for now this only links since "" means "leave unchanged".
		if opportunityID == "none" {
			body["opportunityId"] = nil
		} else {
			body["opportunityId"] = opportunityID
		}
	}

	if visuals := loadVisuals(visualsPath); len(visuals) > 0 {
		body["visuals"] = visuals
	}
	if claims := loadJSONArray(claimsPath, "--claims"); len(claims) > 0 {
		body["claims"] = claims
	}

	if len(body) == 0 {
		exitError("nothing to update — provide content (--file or stdin), --title, --status, --meta-description, --opportunity-id, or --visuals")
	}

	var result ArticleResponse
	err = client.PatchJSON(fmt.Sprintf("/websites/%s/articles/%s", websiteID, articleID), body, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result.Data)
		return
	}

	fmt.Printf("Article updated\n")
	fmt.Printf("  ID:     %s\n", result.Data.ID)
	fmt.Printf("  Title:  %s\n", result.Data.Title)
	fmt.Printf("  Status: %s\n", result.Data.Status)
}

func PublishArticle(websiteID string, articleID string, publishedURL string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	body := map[string]interface{}{
		"publishedUrl": publishedURL,
	}

	var result ArticleResponse
	err = client.PatchJSON(fmt.Sprintf("/websites/%s/articles/%s", websiteID, articleID), body, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result.Data)
		return
	}

	fmt.Printf("Article published\n")
	fmt.Printf("  ID:    %s\n", result.Data.ID)
	fmt.Printf("  Title: %s\n", result.Data.Title)
	fmt.Printf("  URL:   %s\n", publishedURL)
}

type CommentItem struct {
	ID           string  `json:"id"`
	UserName     *string `json:"userName"`
	SelectedText string  `json:"selectedText"`
	Comment      string  `json:"comment"`
	Resolved     bool    `json:"resolved"`
	CreatedAt    string  `json:"createdAt"`
}

type CommentsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Comments []CommentItem `json:"comments"`
	} `json:"data"`
}

func ListComments(websiteID string, articleID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	var result CommentsResponse
	err = client.GetJSON(fmt.Sprintf("/websites/%s/articles/%s/comments", websiteID, articleID), nil, &result)
	if err != nil {
		exitError(err.Error())
	}

	comments := result.Data.Comments

	if isJSON() {
		printJSON(comments)
		return
	}

	if len(comments) == 0 {
		fmt.Println("No comments on this article.")
		return
	}

	fmt.Printf("%d comment(s)\n", len(comments))
	fmt.Println(strings.Repeat("─", 60))
	for _, c := range comments {
		status := "○"
		if c.Resolved {
			status = "✓"
		}
		name := "Unknown"
		if c.UserName != nil {
			name = *c.UserName
		}
		fmt.Printf("  %s [%s] %s\n", status, name, c.CreatedAt[:10])
		fmt.Printf("    \"%s\"\n", c.SelectedText)
		fmt.Printf("    → %s\n\n", c.Comment)
	}
}

func ResolveComment(websiteID string, articleID string, commentID string) {
	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}

	websiteID = resolveWebsiteID(client, websiteID)

	body := map[string]interface{}{
		"commentId": commentID,
		"resolved":  true,
	}

	var result struct {
		Success bool `json:"success"`
	}
	err = client.PatchJSON(fmt.Sprintf("/websites/%s/articles/%s/comments", websiteID, articleID), body, &result)
	if err != nil {
		exitError(err.Error())
	}

	fmt.Println("Comment resolved ✓")
}
