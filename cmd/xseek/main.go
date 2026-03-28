package main

import (
	"fmt"
	"os"

	"github.com/xseekio/xseek-cli/internal/api"
	"github.com/xseekio/xseek-cli/internal/channel"
	"github.com/xseekio/xseek-cli/internal/commands"
)

var version = "dev"

func usage() {
	fmt.Printf(`xseek — AI visibility from your terminal (v%s)

Usage:
  xseek <command> [arguments]

Commands:
  init                              Install xSeek skills for Claude Code
  claude                            Start Claude Code with xSeek Channel UI
  login <api-key>                   Save API key to ~/.xseek/config
  logout                            Remove saved API key
  skills                            List installed Claude Code skills
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
  keywords <website> "<topic>"      Keyword research via DataForSEO
  brand-context <website>           Brand voice, tone, and knowledge base
  articles list <website>           List articles in Content Studio
  articles push <website>           Push a new article (stdin or --file)
  articles get <website> <id>       Get article content
  articles update <website> <id>   Update an existing article's content
  articles publish <website> <id> <url>  Mark article as published
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

var commandHelp = map[string]string{
	"init":           "Usage: xseek init\n\nInstall xSeek skills (slash commands) for Claude Code.\nDownloads the latest skills from GitHub and installs them to ~/.claude/skills/.\n\nAfter running, open Claude Code and type /generate-article to get started.",
	"claude":         "Usage: xseek claude\n\nStart Claude Code with the xSeek Channel UI.\nOpens a web interface at http://127.0.0.1:8787 connected to your Claude Code session.\n\nFlags:\n  --port N    Custom port (default: 8787)\n\nRequires:\n  - Claude Code CLI installed\n  - Node.js installed",
	"skills":         "Usage: xseek skills\n\nList installed xSeek skills for Claude Code.\n\nFlags:\n  --format json    Output as JSON",
	"login":          "Usage: xseek login <api-key>\n\nSave your API key to ~/.xseek/config for future use.",
	"logout":         "Usage: xseek logout\n\nRemove your saved API key.",
	"scan":           "Usage: xseek scan robots <domain>\n\nCheck which AI bots are allowed or blocked in a domain's robots.txt.",
	"generate":       "Usage: xseek generate llms-txt <domain>\n\nGenerate an LLMs.txt file for a domain.",
	"websites":       "Usage: xseek websites\n\nList all tracked websites in your account.\n\nFlags:\n  --format json    Output as JSON",
	"prompts":        "Usage: xseek prompts <website>\n\nList all prompts tracked for a website.\n\nFlags:\n  --format json    Output as JSON",
	"prompt-runs":    "Usage: xseek prompt-runs <website> <promptId>\n\nShow the latest runs for a specific prompt.\n\nFlags:\n  --pageSize N     Number of results (default 10)\n  --format json    Output as JSON",
	"leaderboard":    "Usage: xseek leaderboard <website>\n\nShow the brand mention leaderboard across all prompts.\n\nFlags:\n  --days N         Time window in days (default 30)\n  --format json    Output as JSON",
	"sources":        "Usage: xseek sources <website>\n\nList AI citation sources — the URLs that AI models cite when mentioning your brand.\n\nFlags:\n  --days N         Filter sources from the last N days (e.g. 1, 7, 30)\n  --pageSize N     Number of results (default 20)\n  --search <term>  Filter by URL, domain, or title\n  --format json    Output as JSON",
	"opportunities":  "Usage: xseek opportunities <website>\n\nList content gap opportunities found by AI analysis.\n\nFlags:\n  --value <level>  Filter by business value: critical, high, medium, low\n  --type <type>    Filter by content type: blog, comparison, howto, faq\n  --pageSize N     Max number of results (default 50)\n  --format json    Output as JSON",
	"search-metrics": "Usage: xseek search-metrics <website>\n\nShow GSC page-level metrics (impressions, clicks, position).\n\nFlags:\n  --pageSize N     Number of results (default 20)\n  --sortBy <field> Sort by: impressions, clicks, position, ctr\n  --search <term>  Filter by URL\n  --format json    Output as JSON",
	"search-queries": "Usage: xseek search-queries <website>\n\nShow GSC search queries driving traffic.\n\nFlags:\n  --pageSize N     Number of results (default 20)\n  --sortBy <field> Sort by: impressions, clicks, position, ctr\n  --url <url>      Filter by page URL\n  --format json    Output as JSON",
	"sitemap-pages":  "Usage: xseek sitemap-pages <website>\n\nShow sitemap pages with AI visit and GSC data.\n\nFlags:\n  --days N         Time range in days (default 30)\n  --filter <type>  Filter type: \"attention\" for pages with dropping AI traffic\n  --format json    Output as JSON",
	"ai-visits":      "Usage: xseek ai-visits <website>\n\nShow AI bot visit logs.\n\nFlags:\n  --pageSize N     Number of results (default 20)\n  --search <term>  Filter by URL\n  --bot <name>     Filter by bot name\n  --format json    Output as JSON",
	"web-searches":   "Usage: xseek web-searches <website>\n\nShow LLM web searches triggered by your prompts.\n\nFlags:\n  --pageSize N     Number of results (default 20)\n  --format json    Output as JSON",
	"keywords":       "Usage: xseek keywords <website> \"<topic>\"\n\nResearch keywords for a topic using DataForSEO.\nReturns search volume, keyword difficulty, and related keywords.\nAccepts comma-separated topics (max 10).\n\nExamples:\n  xseek keywords mysite.com \"best crm for small business\"\n  xseek keywords mysite.com \"meilleur crm\" --language fr --location 2124\n  xseek keywords mysite.com \"crm tools\" --format json\n\nFlags:\n  --language <code>  Language code (default: en). Examples: fr, es, de, pt\n  --location <code>  Google location code (default: 2840 = US). Common codes:\n                       2124 = Canada, 2250 = France, 2826 = UK,\n                       2276 = Germany, 2724 = Spain, 2076 = Brazil,\n                       2036 = Australia, 2392 = Japan\n  --format json      Output as JSON",
	"brand-context":  "Usage: xseek brand-context <website>\n\nGet brand voice guidelines, tone, and knowledge base entries for a website.\nUsed by article generation to match your brand's style.\n\nFlags:\n  --format json    Output as JSON",
	"articles":       "Usage: xseek articles <subcommand> <website> [arguments]\n\nSubcommands:\n  list <website>                    List articles in Content Studio\n  push <website>                    Push a new article\n  get <website> <articleId>         Get article content\n  publish <website> <id> <url>      Mark article as published\n\nFlags (list):\n  --status <status>  Filter by status: draft, ready, published\n  --pageSize N       Number of results (default 20)\n  --format json      Output as JSON\n\nFlags (push):\n  --title \"...\"      Article title (required)\n  --file <path>      Read content from file (alternative to stdin)\n  --status <status>  Article status (default: ready)\n  --meta-description \"...\"  Meta description\n  --format json      Output as JSON\n\nExamples:\n  xseek articles list mysite.com\n  cat article.md | xseek articles push mysite.com --title \"My Article\"\n  xseek articles push mysite.com --title \"My Article\" --file article.md\n  xseek articles get mysite.com <id>\n  xseek articles publish mysite.com <id> https://blog.com/article",
}

func printCommandHelp(cmd string) {
	if help, ok := commandHelp[cmd]; ok {
		fmt.Println(help)
	} else {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'xseek help' for usage.\n", cmd)
	}
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
		if args[i] == "--help" || args[i] == "-h" {
			flags["help"] = "true"
		} else if args[i] == "--no-browser" {
			flags["no-browser"] = "true"
		} else if args[i] == "--format" && i+1 < len(args) {
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

	// Per-command --help
	if _, hasHelp := flags["help"]; hasHelp {
		if len(args) > 0 {
			printCommandHelp(args[0])
		} else {
			usage()
		}
		os.Exit(0)
	}

	switch args[0] {
	case "channel-server":
		// Hidden command — runs the MCP channel server (called by Claude Code via .mcp.json)
		p := 8787
		if flags["port"] != "" {
			if v, err := fmt.Sscanf(flags["port"], "%d", &p); err != nil || v == 0 {
				p = 8787
			}
		}
		if envPort := os.Getenv("CHANNEL_UI_PORT"); envPort != "" {
			fmt.Sscanf(envPort, "%d", &p)
		}
		srv := channel.NewServer(p)
		if err := srv.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Channel server error: %s\n", err)
			os.Exit(1)
		}

	case "init":
		commands.Init()

	case "claude":
		commands.CloudStart(flags["port"], flags["no-browser"] == "true")

	case "skills":
		commands.ListSkills()

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
		commands.ReportLeaderboard(args[1], flags["days"])

	case "sources":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek sources <website>")
			os.Exit(1)
		}
		commands.ListSources(args[1], flags["days"], flags["pageSize"], flags["search"])

	case "opportunities":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek opportunities <website>")
			os.Exit(1)
		}
		commands.ListOpportunities(args[1], flags["value"], flags["type"], flags["pageSize"])

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

	case "keywords":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: xseek keywords <website> \"<topic>\"")
			os.Exit(1)
		}
		commands.SearchKeywords(args[1], args[2], flags["language"], flags["location"])

	case "brand-context":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: xseek brand-context <website>")
			os.Exit(1)
		}
		commands.GetBrandContext(args[1])

	case "articles":
		if len(args) < 2 {
			printCommandHelp("articles")
			os.Exit(1)
		}
		switch args[1] {
		case "list":
			if len(args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: xseek articles list <website>")
				os.Exit(1)
			}
			commands.ListArticles(args[2], flags["status"], flags["pageSize"])
		case "push":
			if len(args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: xseek articles push <website> --title \"...\"")
				os.Exit(1)
			}
			commands.PushArticle(args[2], flags["title"], flags["file"], flags["status"], flags["meta-description"])
		case "get":
			if len(args) < 4 {
				fmt.Fprintln(os.Stderr, "Usage: xseek articles get <website> <articleId>")
				os.Exit(1)
			}
			commands.GetArticle(args[2], args[3])
		case "update":
			if len(args) < 4 {
				fmt.Fprintln(os.Stderr, "Usage: xseek articles update <website> <articleId> [--file article.md] [--title \"...\"] [--status draft]")
				os.Exit(1)
			}
			commands.UpdateArticle(args[2], args[3], flags["file"], flags["title"], flags["status"], flags["meta-description"])
		case "publish":
			if len(args) < 5 {
				fmt.Fprintln(os.Stderr, "Usage: xseek articles publish <website> <articleId> <url>")
				os.Exit(1)
			}
			commands.PublishArticle(args[2], args[3], args[4])
		case "comments":
			if len(args) < 4 {
				fmt.Fprintln(os.Stderr, "Usage: xseek articles comments <website> <articleId> [--resolve <commentId>]")
				os.Exit(1)
			}
			if resolveId := flags["resolve"]; resolveId != "" {
				commands.ResolveComment(args[2], args[3], resolveId)
			} else {
				commands.ListComments(args[2], args[3])
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown articles subcommand: %s\nRun 'xseek articles --help' for usage.\n", args[1])
			os.Exit(1)
		}

	case "version":
		fmt.Printf("xseek v%s\n", version)

	case "help", "--help", "-h":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'xseek help' for usage.\n", args[0])
		os.Exit(1)
	}
}
