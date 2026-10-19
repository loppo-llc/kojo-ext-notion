// kojo-ext-notion wraps Notion's official local MCP server
// (@notionhq/notion-mcp-server). kojo starts this binary as an stdio
// MCP server; it fetches the agent's integration token from kojo's
// extension secrets API and runs the official server with it, so the
// token never appears in the backend CLI's MCP config or command line.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// notionServer is pinned so an upstream release cannot change the
// tools an agent gets without a new release of this package.
const notionServer = "@notionhq/notion-mcp-server@2.5.2"

var version = "dev"

func main() {
	if len(os.Args) < 2 || os.Args[1] != "mcp" {
		fmt.Fprintf(os.Stderr, "kojo-ext-notion %s\nusage: kojo-ext-notion mcp\n", version)
		os.Exit(2)
	}
	code, err := runMCP()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kojo-ext-notion:", err)
	}
	os.Exit(code)
}

func runMCP() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	secrets, err := fetchSecrets(ctx, strings.TrimRight(os.Getenv("KOJO_API_BASE"), "/"), extToken(), os.Getenv("KOJO_EXT_AGENT_ID"))
	cancel()
	if err != nil {
		return 1, fmt.Errorf("fetch secrets: %w", err)
	}
	token := strings.TrimSpace(secrets["token"])
	if token == "" {
		return 1, errors.New("no Notion integration token; set it in the extension settings for this agent")
	}
	npx, err := findNpx(configString("npxPath"))
	if err != nil {
		return 1, err
	}

	cmd := exec.Command(npx, "-y", notionServer)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = childEnv(os.Environ(), token)
	if err := cmd.Start(); err != nil {
		return 1, fmt.Errorf("start %s: %w", notionServer, err)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		for s := range sig {
			_ = cmd.Process.Signal(s)
		}
	}()
	err = cmd.Wait()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if c := ee.ExitCode(); c >= 0 {
			return c, nil
		}
		return 1, nil
	}
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func extToken() string {
	if t := os.Getenv("KOJO_EXT_TOKEN"); t != "" {
		return t
	}
	if f := os.Getenv("KOJO_EXT_TOKEN_FILE"); f != "" {
		if b, err := os.ReadFile(f); err == nil {
			return strings.TrimSpace(string(b))
		}
	}
	return ""
}

func configString(key string) string {
	var m map[string]any
	_ = json.Unmarshal([]byte(os.Getenv("KOJO_EXT_CONFIG")), &m)
	s, _ := m[key].(string)
	return strings.TrimSpace(s)
}

// childEnv drops kojo's extension credentials (the child is third-party
// code that has no use for them) and sets NOTION_TOKEN.
func childEnv(env []string, token string) []string {
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		k, _, _ := strings.Cut(kv, "=")
		switch {
		case strings.HasPrefix(k, "KOJO_EXT_"), k == "NOTION_TOKEN", k == "OPENAPI_MCP_HEADERS":
			continue
		}
		out = append(out, kv)
	}
	return append(out, "NOTION_TOKEN="+token)
}

func findNpx(configured string) (string, error) {
	if configured != "" {
		if _, err := os.Stat(configured); err != nil {
			return "", fmt.Errorf("npxPath %q: %w", configured, err)
		}
		return configured, nil
	}
	if p, err := exec.LookPath("npx"); err == nil {
		return p, nil
	}
	home, _ := os.UserHomeDir()
	for _, p := range []string{"/opt/homebrew/bin/npx", "/usr/local/bin/npx", "/usr/bin/npx", home + "/.n/bin/npx", home + "/.volta/bin/npx"} {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p, nil
		}
	}
	return "", errors.New("npx not found; install Node.js 18+ or set npxPath in the extension settings")
}

func fetchSecrets(ctx context.Context, base, token, agentID string) (map[string]string, error) {
	if base == "" || token == "" {
		return nil, errors.New("KOJO_API_BASE and the extension token are missing; this server is started by kojo")
	}
	u := base + "/api/v1/extension/secrets"
	if agentID != "" {
		u += "?agentId=" + url.QueryEscape(agentID)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kojo returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var out struct {
		Secrets map[string]string `json:"secrets"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out.Secrets, nil
}
