package commands

import (
	"bufio"
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

	// 4. Write .mcp.json — uses xseek channel-server (embedded, no Node.js needed)
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

	cwd, _ := os.Getwd()
	mcpPath := filepath.Join(cwd, ".mcp.json")

	// Always update channel-ui entry
	if existing, err := os.ReadFile(mcpPath); err == nil {
		var existingConfig map[string]interface{}
		if json.Unmarshal(existing, &existingConfig) == nil {
			if servers, ok := existingConfig["mcpServers"].(map[string]interface{}); ok {
				servers["channel-ui"] = mcpConfig["mcpServers"].(map[string]interface{})["channel-ui"]
				mcpConfig = existingConfig
			}
		}
	}

	mcpJSON, _ := json.MarshalIndent(mcpConfig, "", "  ")
	if err := os.WriteFile(mcpPath, mcpJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write .mcp.json: %s\n", err)
	}

	// Also update global ~/.mcp.json if it has a stale channel-ui entry
	home, _ := os.UserHomeDir()
	globalMcpPath := filepath.Join(home, ".mcp.json")
	if existing, err := os.ReadFile(globalMcpPath); err == nil {
		var globalConfig map[string]interface{}
		if json.Unmarshal(existing, &globalConfig) == nil {
			if servers, ok := globalConfig["mcpServers"].(map[string]interface{}); ok {
				if _, has := servers["channel-ui"]; has {
					channelEntry := map[string]interface{}{
						"command": xseekPath,
						"args":    []string{"channel-server", "--port", port},
						"env":     map[string]string{"CHANNEL_UI_PORT": port},
					}
					servers["channel-ui"] = channelEntry
					// Also remove stale "channelui" if present
					delete(servers, "channelui")
					data, _ := json.MarshalIndent(globalConfig, "", "  ")
					os.WriteFile(globalMcpPath, data, 0644)
				}
			}
		}
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

	// 6. Launch Claude Code with channel
	cmd := exec.Command(claudePath,
		"--dangerously-skip-permissions",
		"--chrome",
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
