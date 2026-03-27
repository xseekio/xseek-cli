package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	skillsRepo    = "xseekio/claude-code-seo-geo-skills"
	skillsBranch  = "main"
	rawGitHubBase = "https://raw.githubusercontent.com"
	gitHubAPIBase = "https://api.github.com"
)

// Skills to install — each becomes ~/.claude/skills/<name>/SKILL.md
var skillDefinitions = []struct {
	Name        string
	File        string
	Description string
	ArgHint     string
}{
	{"add-keywords", "add-keywords.md", "Enrich an existing article with relevant SEO keywords from Google search data. Run this when a user wants to add keywords to an article.", "[url or articleId]"},
	{"apply-comments", "apply-comments.md", "Apply unresolved comments on articles — read comments, apply changes, resolve them. Run this when articles have feedback to address.", "[articleId]"},
	{"aeo-audit", "aeo-audit.md", "Full AI visibility assessment. Run this when a user asks to audit their AI search presence or check AEO performance.", ""},
	{"fact-check", "fact-check.md", "Verify pricing, features, and claims in an article against official sources. Run this to validate competitor data before publishing.", "[url or articleId]"},
	{"find-opportunities", "find-opportunities.md", "Content gap finder for AI search. Run this when a user wants to find topics where competitors get cited by AI but they don't.", ""},
	{"generate-article", "generate-article.md", "Generate an AI-optimized article from content gap data. Run this when a user wants to create new content targeting AI citations.", "[topic]"},
	{"publish-articles", "publish-articles.md", "Publish ready articles from Content Studio to your website. Run this when articles are reviewed and ready to go live.", "[article title]"},
	{"optimize-page", "optimize-page.md", "AI visibility optimization for a specific URL. Run this when a user wants to improve a page's chances of being cited by AI.", "<url>"},
	{"rewrite-page", "rewrite-page.md", "Full AI-optimized content rewrite. Run this when a user wants to rewrite a page to improve AI search citations.", "<url>"},
	{"track-visibility", "track-visibility.md", "AI visibility snapshot. Run this when a user wants a quick overview of their brand's AI search presence.", ""},
	{"weekly-report", "weekly-report.md", "Weekly AI visibility and SEO performance report. Run this when a user asks for a status update or weekly summary.", ""},
	{"writing-rules", "writing-rules.md", "Human-like writing rules for AI content. Referenced by other skills — not invoked directly.", ""},
	{"geo-methods", "geo-methods.md", "Princeton GEO optimization methods with examples and domain tips. Referenced by other skills — not invoked directly.", ""},
}

func Init() {
	home, err := os.UserHomeDir()
	if err != nil {
		exitError(fmt.Sprintf("cannot find home directory: %s", err))
	}

	skillsDir := filepath.Join(home, ".claude", "skills")

	fmt.Println("Installing xSeek skills for Claude Code...")
	fmt.Println()

	installed := 0
	for _, skill := range skillDefinitions {
		dir := filepath.Join(skillsDir, skill.Name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s — failed to create directory: %s\n", skill.Name, err)
			continue
		}

		// Fetch skill content from GitHub
		content, err := fetchSkillContent(skill.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s — failed to download: %s\n", skill.Name, err)
			continue
		}

		// Build SKILL.md with frontmatter
		skillContent := buildSkillFile(skill.Name, skill.Description, skill.ArgHint, content)

		skillPath := filepath.Join(dir, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s — failed to write: %s\n", skill.Name, err)
			continue
		}

		if skill.Name == "writing-rules" || skill.Name == "geo-methods" {
			fmt.Printf("  ✓ %s (reference)\n", skill.Name)
		} else {
			fmt.Printf("  ✓ /%s\n", skill.Name)
		}
		installed++
	}

	// Cleanup old xSeek skills that are no longer in the definitions
	// Only removes skills we previously installed — never touches other skills
	previousSkills := map[string]bool{
		"aeo-audit": true, "add-keywords": true, "apply-comments": true, "fact-check": true,
		"find-opportunities": true, "generate-article": true, "geo-methods": true,
		"optimize-page": true, "publish-articles": true, "rewrite-page": true,
		"track-visibility": true, "weekly-report": true, "writing-rules": true,
		"analyze": true,
	}
	knownSkills := make(map[string]bool)
	for _, skill := range skillDefinitions {
		knownSkills[skill.Name] = true
	}
	for name := range previousSkills {
		if !knownSkills[name] {
			dir := filepath.Join(skillsDir, name)
			if _, err := os.Stat(dir); err == nil {
				os.RemoveAll(dir)
				fmt.Printf("  🗑 %s (removed)\n", name)
			}
		}
	}

	fmt.Println()
	if installed == len(skillDefinitions) {
		fmt.Printf("All %d skills installed.\n", installed)
	} else {
		fmt.Printf("%d/%d skills installed.\n", installed, len(skillDefinitions))
	}


	fmt.Println()
	fmt.Println("Open Claude Code and type:")
	fmt.Println("  /generate-article")
	fmt.Println("  /aeo-audit")
	fmt.Println("  /find-opportunities")
	fmt.Println()
	fmt.Printf("Skills are in %s\n", skillsDir)
}

func fetchSkillContent(filename string) (string, error) {
	url := fmt.Sprintf("%s/%s/%s/%s", rawGitHubBase, skillsRepo, skillsBranch, filename)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}

	return string(body), nil
}

func buildSkillFile(name, description, argHint, content string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", name))
	sb.WriteString(fmt.Sprintf("description: %s\n", description))
	if argHint != "" {
		sb.WriteString(fmt.Sprintf("argument-hint: %s\n", argHint))
	}
	// Reference files are not invoked directly as slash commands
	if name == "writing-rules" || name == "geo-methods" {
		sb.WriteString("disable-model-invocation: true\n")
	}
	sb.WriteString("---\n\n")
	sb.WriteString(content)
	return sb.String()
}

// Update re-downloads and reinstalls all skills
func Update() {
	fmt.Println("Updating xSeek skills...")
	fmt.Println()
	Init()
}

// ListSkills shows installed skills
func ListSkills() {
	home, err := os.UserHomeDir()
	if err != nil {
		exitError(fmt.Sprintf("cannot find home directory: %s", err))
	}

	skillsDir := filepath.Join(home, ".claude", "skills")

	if isJSON() {
		type SkillInfo struct {
			Name      string `json:"name"`
			Installed bool   `json:"installed"`
			Path      string `json:"path"`
		}
		var skills []SkillInfo
		for _, skill := range skillDefinitions {
			path := filepath.Join(skillsDir, skill.Name, "SKILL.md")
			_, err := os.Stat(path)
			skills = append(skills, SkillInfo{
				Name:      skill.Name,
				Installed: err == nil,
				Path:      path,
			})
		}
		printJSON(skills)
		return
	}

	referenceFiles := map[string]bool{"writing-rules": true, "geo-methods": true}

	fmt.Println("xSeek Skills for Claude Code")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
	fmt.Println("  Commands:")
	for _, skill := range skillDefinitions {
		if referenceFiles[skill.Name] {
			continue
		}
		path := filepath.Join(skillsDir, skill.Name, "SKILL.md")
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✓ /%s\n", skill.Name)
		} else {
			fmt.Printf("    ✗ /%s (not installed)\n", skill.Name)
		}
	}
	fmt.Println()
	fmt.Println("  Reference files:")
	for _, skill := range skillDefinitions {
		if !referenceFiles[skill.Name] {
			continue
		}
		path := filepath.Join(skillsDir, skill.Name, "SKILL.md")
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✓ %s\n", skill.Name)
		} else {
			fmt.Printf("    ✗ %s (not installed)\n", skill.Name)
		}
	}
	fmt.Println()
	fmt.Println("Run 'xseek init' to install or update skills.")
}

// CheckForUpdates checks if remote skills have changed
func CheckForUpdates() bool {
	url := fmt.Sprintf("%s/repos/%s/commits/%s", gitHubAPIBase, skillsRepo, skillsBranch)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	var commit struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return false
	}

	// Compare with stored SHA
	home, _ := os.UserHomeDir()
	shaPath := filepath.Join(home, ".claude", "skills", ".xseek-sha")
	stored, _ := os.ReadFile(shaPath)

	if string(stored) == commit.SHA {
		return false
	}

	// Save new SHA
	os.WriteFile(shaPath, []byte(commit.SHA), 0644)
	return true
}
