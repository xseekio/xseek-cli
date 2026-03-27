package commands

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	channelBundleURL = "https://www.xseek.io/channel-ui.zip"
	channelRepoZip   = "https://github.com/xseekio/xseek_claude_code_ui_channel/archive/refs/heads/main.zip"
	channelDir       = "channel-ui"
	defaultPort      = "8787"
)

func channelPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".xseek", channelDir)
}

func CloudStart(port string) {
	if port == "" {
		port = defaultPort
	}

	dir := channelPath()

	// 1. Check channel UI is installed
	if !channelInstalled(dir) {
		exitError("Channel UI not installed. Run 'xseek init' first.")
	}

	// 2. Check if port is already in use
	if conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%s", port), time.Second); err == nil {
		conn.Close()
		fmt.Printf("Port %s is already in use (previous xSeek Cloud session?).\n", port)
		fmt.Print("Kill the existing process? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "y" || answer == "yes" {
			// Find and kill the process on the port
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

	// 3. Check claude is available
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		exitError("Claude Code CLI not found. Install it first:\n  npm install -g @anthropic-ai/claude-code")
	}

	serverPath := filepath.Join(dir, "server.cjs")
	if _, err := os.Stat(serverPath); err != nil {
		exitError(fmt.Sprintf("channel server not found at %s\nRun 'xseek init' to reinstall", serverPath))
	}

	// 3. Write .mcp.json — uses node directly, no npm/npx/tsx needed
	mcpCommand := "node"
	mcpArgs := []string{serverPath}

	mcpConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"channel-ui": map[string]interface{}{
				"command": mcpCommand,
				"args":    mcpArgs,
			},
		},
	}

	cwd, _ := os.Getwd()
	mcpPath := filepath.Join(cwd, ".mcp.json")

	// Always update channel-ui entry to ensure correct path
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

	fmt.Println("xSeek Cloud")
	fmt.Println()
	fmt.Printf("  Channel UI:  http://127.0.0.1:%s\n", port)
	fmt.Printf("  Claude Code: %s\n", claudePath)
	fmt.Println()
	fmt.Println("Starting...")
	fmt.Println()

	// 4. Open browser after a short delay
	go func() {
		time.Sleep(3 * time.Second)
		openBrowser(fmt.Sprintf("http://127.0.0.1:%s", port))
	}()

	// 5. Launch Claude Code with channel
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

func channelInstalled(dir string) bool {
	serverFile := filepath.Join(dir, "server.cjs")
	_, err := os.Stat(serverFile)
	return err == nil
}

// installOrUpdateChannel downloads the latest zip from GitHub and extracts it.
// Always re-downloads to ensure the latest version (no git required).
func installOrUpdateChannel(dir string) error {
	// Try bundled zip first (includes node_modules), fall back to GitHub
	resp, err := http.Get(channelBundleURL)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		resp, err = http.Get(channelRepoZip)
	}
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Save to temp file
	tmpFile, err := os.CreateTemp("", "xseek-channel-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to save zip: %w", err)
	}
	tmpFile.Close()

	os.RemoveAll(dir)

	// Extract zip
	if err := extractZip(tmpFile.Name(), dir); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	return nil
}

// extractZip extracts a GitHub archive zip (which has a top-level directory) into dest.
func extractZip(zipPath string, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// GitHub zips have a top-level dir like "repo-main/"
	// We need to strip that prefix
	var prefix string
	for _, f := range r.File {
		parts := strings.SplitN(f.Name, "/", 2)
		if len(parts) > 0 {
			prefix = parts[0] + "/"
			break
		}
	}

	for _, f := range r.File {
		// Strip the top-level directory
		name := strings.TrimPrefix(f.Name, prefix)
		if name == "" {
			continue
		}

		fpath := filepath.Join(dest, name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
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
