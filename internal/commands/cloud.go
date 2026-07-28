package commands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const defaultPort = "8787"

func CloudStart(port string, noBrowser bool) {
	if port == "" {
		port = defaultPort
	}

	// 1. Check if port is already in use
	if conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%s", port), time.Second); err == nil {
		conn.Close()
		fmt.Printf("Port %s is already in use (previous xSeek Cloud session?).\n", port)
		fmt.Print("Kill the existing process? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "y" || answer == "yes" {
			var killCmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin", "linux":
				killCmd = exec.Command("sh", "-c", fmt.Sprintf("lsof -ti:%s | xargs kill -9 2>/dev/null", port))
			case "windows":
				killCmd = exec.Command("cmd", "/c", fmt.Sprintf("for /f \"tokens=5\" %%a in ('netstat -aon ^| findstr :%s') do taskkill /F /PID %%a", port))
			}
			if killCmd != nil {
				killCmd.Run()
				time.Sleep(500 * time.Millisecond)
				fmt.Println("  ✓ Previous process killed")
			}
		} else {
			fmt.Println("Use --port to specify a different port: xseek claude --port 9000")
			os.Exit(0)
		}
	}

	// 2. Check claude is available
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		exitError("Claude Code CLI not found. Install it first:\n  npm install -g @anthropic-ai/claude-code")
	}

	// 3. Find xseek binary path for the MCP server command
	xseekPath, err := os.Executable()
	if err != nil {
		xseekPath = "xseek" // fallback
	}

	// 4. Write channel-ui MCP config to an xSeek-owned, Claude-only path.
	//
	// Previously we wrote into <cwd>/.mcp.json AND ~/.mcp.json. That broke
	// Codex users: .mcp.json is the standard MCP config path that Codex
	// (and other MCP clients) ALSO read, so they tried to spawn channel-ui
	// — a server designed for Claude Code's --dangerously-load-development-
	// channels flow — and failed with handshake errors on every launch.
	//
	// Solution: write to ~/.xseek/claude-channel.mcp.json and pass it to
	// Claude Code via --mcp-config. Claude merges this file with its other
	// MCP sources; Codex never sees it.
	home, _ := os.UserHomeDir()
	claudeMcpPath := filepath.Join(home, ".xseek", "claude-channel.mcp.json")
	mcpConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"channel-ui": map[string]interface{}{
				"command": xseekPath,
				"args":    []string{"channel-server", "--port", port},
				"env": map[string]string{
					"CHANNEL_UI_PORT": port,
				},
			},
		},
	}
	mcpJSON, _ := json.MarshalIndent(mcpConfig, "", "  ")
	if err := os.MkdirAll(filepath.Dir(claudeMcpPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create %s: %s\n", filepath.Dir(claudeMcpPath), err)
	}
	if err := os.WriteFile(claudeMcpPath, mcpJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write %s: %s\n", claudeMcpPath, err)
	}

	// Best-effort cleanup of legacy entries that previous CLI versions wrote
	// into Codex-visible paths. Silently scrub channel-ui from <cwd>/.mcp.json
	// and ~/.mcp.json so existing Codex installs stop warning on next launch.
	// We don't delete the files entirely — they may have unrelated entries.
	cwd, _ := os.Getwd()
	for _, p := range []string{
		filepath.Join(cwd, ".mcp.json"),
		filepath.Join(home, ".mcp.json"),
	} {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var cfg map[string]interface{}
		if json.Unmarshal(raw, &cfg) != nil {
			continue
		}
		servers, ok := cfg["mcpServers"].(map[string]interface{})
		if !ok {
			continue
		}
		_, has := servers["channel-ui"]
		_, hasStale := servers["channelui"]
		if !has && !hasStale {
			continue
		}
		delete(servers, "channel-ui")
		delete(servers, "channelui")
		// Only rewrite if there are still other servers to preserve; if the
		// file ends up empty of MCP entries, leave it as a `{ "mcpServers": {} }`
		// shell rather than deleting it, in case other tooling expects it.
		data, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile(p, data, 0644)
	}

	// Also scrub channel-ui from any Codex TOML config we can find. Codex
	// reads <cwd>/.codex/config.toml and ~/.codex/config.toml; if channel-ui
	// is registered there (legacy from prior CLI versions) Codex will try to
	// validate it on every launch and emit a handshake warning. We strip the
	// `[mcp_servers.channel-ui]` and `[mcp_servers.channel-ui.env]` tables
	// in-place — naive line scan, but safe enough since these tables are
	// flat and our own writer never produced anything fancier.
	for _, p := range []string{
		filepath.Join(cwd, ".codex", "config.toml"),
		filepath.Join(home, ".codex", "config.toml"),
	} {
		raw, err := os.ReadFile(p)
		if err != nil || !bytes.Contains(raw, []byte("mcp_servers.channel-ui")) {
			continue
		}
		var out bytes.Buffer
		skip := false
		for _, line := range strings.Split(string(raw), "\n") {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "[mcp_servers.channel-ui]") ||
				strings.HasPrefix(trim, "[mcp_servers.channel-ui.") {
				skip = true
				continue
			}
			if skip && strings.HasPrefix(trim, "[") {
				skip = false
			}
			if !skip {
				out.WriteString(line)
				out.WriteString("\n")
			}
		}
		os.WriteFile(p, bytes.TrimRight(out.Bytes(), "\n"), 0644)
	}

	fmt.Println("xSeek Cloud")
	fmt.Println()
	fmt.Printf("  Channel UI:  http://127.0.0.1:%s\n", port)
	fmt.Printf("  Claude Code: %s\n", claudePath)
	fmt.Println()
	fmt.Println("Starting...")
	fmt.Println()

	// 5. Open browser after a short delay (unless --no-browser)
	if !noBrowser {
		go func() {
			time.Sleep(3 * time.Second)
			openBrowser(fmt.Sprintf("http://127.0.0.1:%s", port))
		}()
	}

	// 6. Launch Claude Code with channel. --mcp-config points at the
	// xSeek-owned Claude-only config so Codex and other MCP clients
	// don't try to spawn channel-ui themselves.
	cmd := exec.Command(claudePath,
		"--dangerously-skip-permissions",
		"--chrome",
		"--mcp-config", claudeMcpPath,
		"--dangerously-load-development-channels",
		"server:channel-ui",
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("CHANNEL_UI_PORT=%s", port))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		exitError(fmt.Sprintf("Claude Code exited: %s", err))
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	cmd.Run()
}
