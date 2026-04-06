// Package channel implements the xSeek Channel UI MCP server in pure Go.
// No Node.js, npm, or any external runtime needed.
package channel

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed index.html
var embeddedFS embed.FS

// ── JSON-RPC types ──────────────────────────────────────────────

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type jsonrpcNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// ── MCP types ───────────────────────────────────────────────────

type mcpCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Tools        map[string]interface{} `json:"tools,omitempty"`
}

type mcpServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type mcpInitResult struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    mcpCapabilities `json:"capabilities"`
	ServerInfo      mcpServerInfo   `json:"serverInfo"`
	Instructions    string          `json:"instructions,omitempty"`
}

type mcpTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type mcpToolsResult struct {
	Tools []mcpTool `json:"tools"`
}

type mcpCallToolResult struct {
	Content []map[string]string `json:"content"`
}

// ── Chat types ──────────────────────────────────────────────────

type chatMessage struct {
	Text      string `json:"text"`
	Type      string `json:"type"` // "user" or "claude"
	Timestamp int64  `json:"timestamp"`
}

type chatSession struct {
	ChatID      string        `json:"chatId"`
	Messages    []chatMessage `json:"messages"`
	CreatedAt   string        `json:"createdAt"`
	LastMessage string        `json:"lastMessage"`
}

// ── Server state ────────────────────────────────────────────────

type Server struct {
	port       int
	mu         sync.RWMutex
	sessions   map[string]*chatSession
	clients    map[string]*websocket.Conn // chatId -> ws
	nextChatID int
	stdinMu    sync.Mutex
	upgrader   websocket.Upgrader
}

func NewServer(port int) *Server {
	return &Server{
		port:       port,
		sessions:   make(map[string]*chatSession),
		clients:    make(map[string]*websocket.Conn),
		nextChatID: 1,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// ── Session helpers ─────────────────────────────────────────────

func (s *Server) createSession() *chatSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := strconv.Itoa(s.nextChatID)
	s.nextChatID++
	sess := &chatSession{
		ChatID:    id,
		Messages:  []chatMessage{},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	s.sessions[id] = sess
	// log.Printf("[channel-ui] Session created: chat_id=%s", id)
	return sess
}

func (s *Server) getSessionsList() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]map[string]interface{}, 0, len(s.sessions))
	for _, sess := range s.sessions {
		list = append(list, map[string]interface{}{
			"chatId":       sess.ChatID,
			"messageCount": len(sess.Messages),
			"lastMessage":  sess.LastMessage,
			"createdAt":    sess.CreatedAt,
		})
	}
	return list
}

func (s *Server) broadcast(msg interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, _ := json.Marshal(msg)
	for _, ws := range s.clients {
		ws.WriteMessage(websocket.TextMessage, data)
	}
}

func (s *Server) broadcastSessionsList() {
	s.broadcast(map[string]interface{}{
		"type":     "sessions:list",
		"sessions": s.getSessionsList(),
	})
}

// ── MCP stdio handler ───────────────────────────────────────────

func (s *Server) writeStdout(data []byte) {
	s.stdinMu.Lock()
	defer s.stdinMu.Unlock()
	// MCP stdio uses newline-delimited JSON
	os.Stdout.Write(data)
	os.Stdout.Write([]byte("\n"))
}

func (s *Server) sendNotification(method string, params interface{}) {
	msg := jsonrpcNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, _ := json.Marshal(msg)
	s.writeStdout(data)
}

func (s *Server) sendResponse(id interface{}, result interface{}) {
	resp := jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	s.writeStdout(data)
}

func (s *Server) handleMCPStdio() {
	scanner := bufio.NewScanner(os.Stdin)
	// Allow up to 1MB per message
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req jsonrpcRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		switch req.Method {
		case "initialize":
			s.sendResponse(req.ID, mcpInitResult{
				ProtocolVersion: "2024-11-05",
				Capabilities: mcpCapabilities{
					Experimental: map[string]interface{}{"claude/channel": map[string]interface{}{}},
					Tools:        map[string]interface{}{"listChanged": true},
				},
				ServerInfo: mcpServerInfo{Name: "channel-ui", Version: "1.0.0"},
				Instructions: `Messages arrive as <channel source="channel-ui" chat_id="...">.
These are real-time messages from a user through a web chat UI.
You MUST reply to every channel message by calling the "reply" tool with the chat_id from the channel tag and your response text.
Do NOT respond in the terminal — the user cannot see terminal output. The ONLY way they see your response is through the "reply" tool.`,
			})

		case "notifications/initialized":
			// No response needed for notifications
			// log.Printf("[channel-ui] MCP initialized")

		case "tools/list":
			s.sendResponse(req.ID, mcpToolsResult{
				Tools: []mcpTool{
					{
						Name:        "reply",
						Description: "Send a reply message back to the user in the web chat UI.",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"chat_id": map[string]interface{}{
									"type":        "string",
									"description": "The chat_id from the channel message tag",
								},
								"text": map[string]interface{}{
									"type":        "string",
									"description": "The message text to send back to the user",
								},
							},
							"required": []string{"chat_id", "text"},
						},
					},
				},
			})

		case "tools/call":
			var params struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			json.Unmarshal(req.Params, &params)

			if params.Name == "reply" {
				var args struct {
					ChatID string `json:"chat_id"`
					Text   string `json:"text"`
				}
				json.Unmarshal(params.Arguments, &args)

				// Store reply in session
				s.mu.Lock()
				if sess, ok := s.sessions[args.ChatID]; ok {
					sess.Messages = append(sess.Messages, chatMessage{
						Text:      args.Text,
						Type:      "claude",
						Timestamp: time.Now().UnixMilli(),
					})
					if len(args.Text) > 80 {
						sess.LastMessage = args.Text[:80]
					} else {
						sess.LastMessage = args.Text
					}
				}
				// Send to connected client
				if ws, ok := s.clients[args.ChatID]; ok {
					data, _ := json.Marshal(map[string]interface{}{
						"type":      "reply",
						"chatId":    args.ChatID,
						"text":      args.Text,
						"timestamp": time.Now().UnixMilli(),
					})
					ws.WriteMessage(websocket.TextMessage, data)
				}
				s.mu.Unlock()

				s.broadcastSessionsList()

				s.sendResponse(req.ID, mcpCallToolResult{
					Content: []map[string]string{
						{"type": "text", "text": fmt.Sprintf("Reply sent to chat %s", args.ChatID)},
					},
				})
			} else {
				s.sendResponse(req.ID, map[string]interface{}{
					"error": map[string]interface{}{
						"code":    -32601,
						"message": fmt.Sprintf("Unknown tool: %s", params.Name),
					},
				})
			}

		default:
			// Ignore unknown methods (notifications etc.)
			if req.ID != nil {
				s.sendResponse(req.ID, map[string]interface{}{})
			}
		}
	}
}

// ── WebSocket handler ───────────────────────────────────────────

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// log.Printf("[channel-ui] WebSocket upgrade failed: %s", err)
		return
	}

	var boundChatID string

	// Send existing sessions
	sessionsData, _ := json.Marshal(map[string]interface{}{
		"type":     "sessions:list",
		"sessions": s.getSessionsList(),
	})
	ws.WriteMessage(websocket.TextMessage, sessionsData)

	// Send empty loops state
	loopsData, _ := json.Marshal(map[string]interface{}{
		"type":  "loops:state",
		"loops": []interface{}{},
	})
	ws.WriteMessage(websocket.TextMessage, loopsData)

	// log.Printf("[channel-ui] WebSocket connected")

	defer func() {
		if boundChatID != "" {
			s.mu.Lock()
			if current, ok := s.clients[boundChatID]; ok && current == ws {
				delete(s.clients, boundChatID)
			}
			s.mu.Unlock()
			// log.Printf("[channel-ui] Client disconnected from session: chat_id=%s", boundChatID)
		}
		ws.Close()
	}()

	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var data map[string]interface{}
		if err := json.Unmarshal(raw, &data); err != nil {
			continue
		}

		msgType, _ := data["type"].(string)

		switch msgType {
		case "session:create":
			sess := s.createSession()
			boundChatID = sess.ChatID
			s.mu.Lock()
			s.clients[sess.ChatID] = ws
			s.mu.Unlock()

			resp, _ := json.Marshal(map[string]interface{}{
				"type":     "session:created",
				"chatId":   sess.ChatID,
				"messages": []interface{}{},
			})
			ws.WriteMessage(websocket.TextMessage, resp)
			s.broadcastSessionsList()

		case "session:join":
			chatID, _ := data["chatId"].(string)
			s.mu.RLock()
			sess, ok := s.sessions[chatID]
			s.mu.RUnlock()
			if ok {
				boundChatID = sess.ChatID
				s.mu.Lock()
				s.clients[sess.ChatID] = ws
				s.mu.Unlock()

				resp, _ := json.Marshal(map[string]interface{}{
					"type":     "session:joined",
					"chatId":   sess.ChatID,
					"messages": sess.Messages,
				})
				ws.WriteMessage(websocket.TextMessage, resp)
			}

		case "message":
			if boundChatID == "" {
				continue
			}

			text, _ := data["text"].(string)

			// Store user message
			s.mu.Lock()
			if sess, ok := s.sessions[boundChatID]; ok {
				sess.Messages = append(sess.Messages, chatMessage{
					Text:      text,
					Type:      "user",
					Timestamp: time.Now().UnixMilli(),
				})
				if len(text) > 80 {
					sess.LastMessage = text[:80]
				} else {
					sess.LastMessage = text
				}
			}
			s.mu.Unlock()

			// Forward to Claude Code via MCP notification
			s.sendNotification("notifications/claude/channel", map[string]interface{}{
				"content": text,
				"meta":    map[string]string{"chat_id": boundChatID},
			})

			s.broadcastSessionsList()
			// log.Printf("[channel-ui] Message forwarded to Claude: chat_id=%s", boundChatID)

		case "loops:list":
			resp, _ := json.Marshal(map[string]interface{}{
				"type":  "loops:state",
				"loops": []interface{}{},
			})
			ws.WriteMessage(websocket.TextMessage, resp)
		}
	}
}

// ── Run ─────────────────────────────────────────────────────────

func (s *Server) Run() error {
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Serve HTML UI for everything else
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html, err := embeddedFS.ReadFile("index.html")
		if err != nil {
			http.Error(w, "UI not found", 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	})

	// Start MCP protocol handler FIRST (Claude Code sends initialize immediately)
	go s.handleMCPStdio()

	// Start HTTP server
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d already in use: %w", s.port, err)
	}

	// No stderr output — Claude Code interprets any stderr as an error

	// Serve HTTP (blocks)
	return http.Serve(listener, mux)
}
