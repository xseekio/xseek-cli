package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xseekio/xseek-cli/internal/api"
)

// ImageUploadResponse mirrors the V1 API response shape.
type ImageUploadResponse struct {
	Data struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Pathname    string `json:"pathname"`
		ContentType string `json:"contentType"`
		Size        int    `json:"size"`
		Alt         string `json:"alt"`
		Source      string `json:"source"`
	} `json:"data"`
}

// UploadImage pushes a local image file to xSeek's blob storage and prints
// the public URL plus a ready-to-paste markdown snippet so the caller (often
// an LLM driving the CLI) can drop it straight into an article.
func UploadImage(websiteID string, filePath string, alt string, source string) {
	if filePath == "" {
		exitError("--file is required (path to a PNG, JPG, WEBP, or GIF)")
	}
	if alt == "" {
		// Default alt-text from the filename if none was supplied. The skill
		// can override it; this just prevents an embarrassing empty alt.
		base := filepath.Base(filePath)
		ext := filepath.Ext(base)
		alt = strings.TrimSuffix(base, ext)
		alt = strings.ReplaceAll(alt, "-", " ")
		alt = strings.ReplaceAll(alt, "_", " ")
	}

	if _, err := os.Stat(filePath); err != nil {
		exitError(fmt.Sprintf("file not found: %s", filePath))
	}

	client, err := api.NewClient()
	if err != nil {
		exitError(err.Error())
	}
	websiteID = resolveWebsiteID(client, websiteID)

	parts := map[string]string{"alt": alt}
	if source != "" {
		parts["source"] = source
	}

	var result ImageUploadResponse
	err = client.PostMultipart(
		fmt.Sprintf("/websites/%s/images", websiteID),
		parts,
		"file",
		filePath,
		&result,
	)
	if err != nil {
		exitError(err.Error())
	}

	if isJSON() {
		printJSON(result)
		return
	}

	fmt.Printf("Image uploaded\n")
	fmt.Printf("  URL:      %s\n", result.Data.URL)
	fmt.Printf("  Size:     %d bytes\n", result.Data.Size)
	fmt.Printf("  Type:     %s\n", result.Data.ContentType)
	fmt.Printf("  Alt:      %s\n", result.Data.Alt)
	if result.Data.Source != "" {
		fmt.Printf("  Source:   %s\n", result.Data.Source)
	}
	fmt.Printf("\n  Markdown to embed:\n  ![%s](%s)\n", result.Data.Alt, result.Data.URL)
}
