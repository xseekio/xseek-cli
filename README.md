# xSeek CLI

AI visibility from your terminal. Built for Claude Code.

## Install

```sh
curl -fsSL https://cli.xseek.io/install.sh | sh
```

## Setup

```sh
export XSEEK_API_KEY=your_api_key
```

Get your API key at [xseek.io/dashboard/api-keys](https://www.xseek.io/dashboard/api-keys).

## Commands

```sh
xseek scan robots yoursite.com        # Check AI bot access
xseek generate llms-txt yoursite.com  # Generate LLMs.txt
xseek websites                        # List your websites
xseek leaderboard yoursite.com        # Brand mention leaderboard
xseek sources yoursite.com            # AI citation sources
xseek opportunities yoursite.com      # Content gap opportunities
xseek prompts yoursite.com            # List prompts
```

All commands support `--format json` for scripting.

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
git tag v0.1.0
git push origin v0.1.0
```

GoReleaser builds binaries for macOS and Linux (amd64 + arm64) and publishes them as a GitHub Release.
