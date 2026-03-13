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
  scan robots <domain>              Check AI bot access in robots.txt
  generate llms-txt <domain>        Generate an LLMs.txt for a domain
  websites                          List your tracked websites
  prompts <website>                 List prompts for a website
  leaderboard <website>             Brand mention leaderboard
  sources <website>                 AI citation sources
  opportunities <website>           Content gap opportunities
  version                           Print version

Flags:
  --format json                     Output as JSON (works with all commands)

Authentication:
  export XSEEK_API_KEY=your_key     Set via environment variable

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

	// Extract --format flag from anywhere in args
	filtered := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--format" && i+1 < len(args) {
			commands.OutputFormat = args[i+1]
			i++ // skip next
		} else if args[i] == "--format=json" {
			commands.OutputFormat = "json"
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

	case "version":
		fmt.Printf("xseek v%s\n", version)

	case "help", "--help", "-h":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'xseek help' for usage.\n", args[0])
		os.Exit(1)
	}
}
