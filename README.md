# xSeek CLI

AI visibility from your terminal. Built for Claude Code.

## Install

```sh
curl -fsSL https://cli.xseek.io/install.sh | sh
```

## Setup

```sh
xseek login YOUR_API_KEY
```

Your key is saved to `~/.xseek/config` and used for all future commands. Get your API key at [xseek.io/dashboard/api-keys](https://www.xseek.io/dashboard/api-keys).

To log out:

```sh
xseek logout
```

## Commands

```sh
xseek websites                          # List your websites
xseek leaderboard yoursite.com          # Brand mention leaderboard
xseek sources yoursite.com              # AI citation sources
xseek opportunities yoursite.com        # Content gap opportunities
xseek prompts yoursite.com              # List prompts
xseek prompt-runs yoursite.com <id>     # Latest runs for a prompt
xseek search-metrics yoursite.com       # GSC page-level metrics
xseek search-queries yoursite.com       # GSC search queries
xseek sitemap-pages yoursite.com        # Sitemap pages with AI + GSC data
xseek ai-visits yoursite.com            # AI bot visit logs
xseek web-searches yoursite.com         # LLM web searches
xseek analyze yoursite.com <url>        # AEO Copilot page analysis
xseek scan robots yoursite.com          # Check AI bot access
xseek generate llms-txt yoursite.com    # Generate LLMs.txt
```

All commands support `--format json` for scripting and `--help` for usage details:

```sh
xseek sources --help
```

### Flags

```sh
--help              # Show command usage and available flags
--format json       # JSON output
--pageSize N        # Number of results
--sortBy <field>    # Sort field (impressions, clicks, position)
--days N            # Time range (7, 30, 90)
--value <level>     # Filter by business value (critical, high, medium, low)
--type <type>       # Filter by content type (blog, comparison, howto, faq)
--url <url>         # Filter by URL
--search <term>     # Search/filter term
--bot <name>        # Filter by bot name
--filter attention  # Only pages needing attention
```

## Claude Code

The xSeek CLI is designed to work as a tool that Claude Code can call directly. Pair it with [GEO/SEO Skills](https://www.xseek.io/products/geo-seo-skills) for fully autonomous AI visibility workflows:

```
$ claude
> /aeo-audit
```

Claude Code uses the CLI under the hood to pull leaderboard data, find content gaps, and deliver prioritized action lists.

## Release

Tag a version to trigger a release:

```sh
git tag v0.2.3
git push origin v0.2.3
```

GoReleaser builds binaries for macOS and Linux (amd64 + arm64) and publishes them as a GitHub Release.
