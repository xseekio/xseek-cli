package commands

import (
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
	Data    struct {
		ID           string  `json:"id"`
		Title        string  `json:"title"`
		Slug         string  `json:"slug"`
		Status       string  `json:"status"`
		QualityScore *int    `json:"qualityScore"`
		CreatedAt    string  `json:"createdAt"`
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

func PushArticle(websiteID string, title string, filePath string, status string, metaDescription string) {
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

	var result ArticleCreateResponse
	err = client.PostJSON(fmt.Sprintf("/websites/%s/articles", websiteID), body, &result)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result.Data)
		return
	}

	fmt.Printf("Article created\n")
	fmt.Printf("  ID:     %s\n", result.Data.ID)
	fmt.Printf("  Title:  %s\n", result.Data.Title)
	fmt.Printf("  Status: %s\n", result.Data.Status)
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

func UpdateArticle(websiteID string, articleID string, filePath string, title string, status string, metaDescription string) {
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

	if len(body) == 0 {
		exitError("nothing to update — provide content (--file or stdin), --title, --status, or --meta-description")
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
