package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const (
	channelRepo    = "https://github.com/xseekio/xseek_claude_code_ui_channel.git"
	channelDir     = "channel-ui"
	defaultPort    = "8787"
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

	// 1. Install channel UI if not present
	if !channelInstalled(dir) {
		fmt.Println("Installing xSeek Cloud channel UI...")
		if err := installChannel(dir); err != nil {
			exitError(fmt.Sprintf("failed to install channel UI: %s", err))
		}
		fmt.Println("  ✓ Channel UI installed")
		fmt.Println()
	}

	// 2. Check claude is available
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		exitError("Claude Code CLI not found. Install it first:\n  npm install -g @anthropic-ai/claude-code")
	}

	serverPath := filepath.Join(dir, "channel", "server.ts")
	if _, err := os.Stat(serverPath); err != nil {
		exitError(fmt.Sprintf("channel server not found at %s\nRun 'xseek init' to reinstall", serverPath))
	}

	// 3. Configure MCP for the channel
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

func CloudStop() {
	fmt.Println("xSeek Cloud session ended.")
}

func channelInstalled(dir string) bool {
	nodeModules := filepath.Join(dir, "node_modules")
	_, err := os.Stat(nodeModules)
	return err == nil
}

func installChannel(dir string) error {
	// Remove existing dir if partial install
	os.RemoveAll(dir)

	// Clone
	cmd := exec.Command("git", "clone", "--depth", "1", channelRepo, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// npm install
	npmCmd := exec.Command("npm", "install")
	npmCmd.Dir = dir
	npmCmd.Stdout = os.Stdout
	npmCmd.Stderr = os.Stderr
	if err := npmCmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
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
