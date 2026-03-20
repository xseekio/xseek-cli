package main

import (
	"fmt"
	"os"

	"github.com/xseekio/xseek-cli/internal/api"
	"github.com/xseekio/xseek-cli/internal/commands"
)

var version = "dev"

func usage() {
	fmt.Printf(`xseek — AI visibility from your terminal (v%s)

Usage:
  xseek <command> [arguments]

Commands:
  login <api-key>                   Save API key to ~/.xseek/config
  logout                            Remove saved API key
  scan robots <domain>              Check AI bot access in robots.txt
  generate llms-txt <domain>        Generate an LLMs.txt for a domain
  websites                          List your tracked websites
  prompts <website>                 List prompts for a website
  prompt-runs <website> <promptId>  Latest runs for a prompt
  leaderboard <website>             Brand mention leaderboard
  sources <website>                 AI citation sources
  opportunities <website>           Content gap opportunities
  search-metrics <website>          GSC page-level metrics
  search-queries <website>          GSC search queries
  sitemap-pages <website>           Sitemap pages with AI + GSC data
  ai-visits <website>               AI bot visit logs
  web-searches <website>            LLM web searches behind prompts
  analyze <website> <url>           AEO Copilot page analysis
  version                           Print version

Flags:
  --format json                     Output as JSON (works with all commands)
  --pageSize N                      Number of results (default varies)
  --sortBy <field>                  Sort field (command-specific)
  --days N                          Time range in days (7, 30, 90)
  --url <url>                       Filter by URL
  --search <term>                   Search/filter term
  --bot <name>                      Filter by bot name
  --filter <type>                   Filter type (e.g. "attention")

Authentication:
  xseek login YOUR_API_KEY          Save key to ~/.xseek/config (recommended)
  export XSEEK_API_KEY=your_key     Or set via environment variable

Website argument can be a website ID or domain (e.g. yoursite.com).
If you only have one website, it's used automatically.

Get your API key at: https://www.xseek.io/dashboard/api-keys
`, version)
}

func main() {
	api.Version = version
	args := os.Args[1:]

	if len(args) == 0 {
		usage()
		os.Exit(0)
	}

	// Extract flags from anywhere in args
	filtered := make([]string, 0, len(args))
	flags := map[string]string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "--format" && i+1 < len(args) {
			commands.OutputFormat = args[i+1]
			i++
		} else if args[i] == "--format=json" {
			commands.OutputFormat = "json"
		} else if len(args[i]) > 2 && args[i][:2] == "--" && i+1 < len(args) {
			flags[args[i][2:]] = args[i+1]
			i++
		} else {
			filtered = append(filtered, args[i])
		}
	}
	args = filtered

	if len(args) == 0 {
		usage()
		os.Exit(0)
	}

	switch args[0] {
	case "login":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek login <api-key>")
			os.Exit(1)
		}
		if err := api.SaveConfig("api_key", args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving API key: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("API key saved to %s\n", api.ConfigPath())
		fmt.Println("You're authenticated. Try: xseek websites")

	case "logout":
		if err := api.SaveConfig("api_key", ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("API key removed. You're logged out.")

	case "scan":
		if len(args) < 3 || args[1] != "robots" {
			fmt.Fprintln(os.Stderr, "Usage: xseek scan robots <domain>")
			os.Exit(1)
		}
		commands.ScanRobots(args[2])

	case "generate":
		if len(args) < 3 || args[1] != "llms-txt" {
			fmt.Fprintln(os.Stderr, "Usage: xseek generate llms-txt <domain>")
			os.Exit(1)
		}
		commands.GenerateLLMsTxt(args[2])

	case "websites":
		commands.ListWebsites()

	case "prompts":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek prompts <website>")
			os.Exit(1)
		}
		commands.ListPrompts(args[1])

	case "prompt-runs":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: xseek prompt-runs <website> <promptId>")
			os.Exit(1)
		}
		commands.ListPromptRuns(args[1], args[2], flags["pageSize"])

	case "leaderboard":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek leaderboard <website>")
			os.Exit(1)
		}
		commands.ReportLeaderboard(args[1])

	case "sources":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek sources <website>")
			os.Exit(1)
		}
		commands.ListSources(args[1])

	case "opportunities":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek opportunities <website>")
			os.Exit(1)
		}
		commands.ListOpportunities(args[1])

	case "search-metrics":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek search-metrics <website>")
			os.Exit(1)
		}
		commands.ListSearchMetrics(args[1], flags["pageSize"], flags["sortBy"], flags["search"])

	case "search-queries":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek search-queries <website>")
			os.Exit(1)
		}
		commands.ListSearchQueries(args[1], flags["pageSize"], flags["sortBy"], flags["url"])

	case "sitemap-pages":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek sitemap-pages <website>")
			os.Exit(1)
		}
		commands.ListSitemapPages(args[1], flags["days"], flags["filter"])

	case "ai-visits":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek ai-visits <website>")
			os.Exit(1)
		}
		commands.ListAIVisits(args[1], flags["pageSize"], flags["search"], flags["bot"])

	case "web-searches":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek web-searches <website>")
			os.Exit(1)
		}
		commands.ListWebSearches(args[1], flags["pageSize"])

	case "analyze":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: xseek analyze <website> <url>")
			os.Exit(1)
		}
		commands.AnalyzePage(args[1], args[2])

	case "version":
		fmt.Printf("xseek v%s\n", version)

	case "help", "--help", "-h":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'xseek help' for usage.\n", args[0])
		os.Exit(1)
	}
}
