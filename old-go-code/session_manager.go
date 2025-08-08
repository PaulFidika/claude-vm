package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type SessionManager struct {
	apiURL string
	client *http.Client
}

type Session struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type SessionStatus struct {
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

func NewSessionManager() *SessionManager {
	apiURL := os.Getenv("CLAUDE_VM_API")
	if apiURL == "" {
		apiURL = "https://api.claude-vm.com"
	}

	return &SessionManager{
		apiURL: apiURL,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (sm *SessionManager) CreateRemoteSession(repoURL string) string {
	// TODO: Implement API call to create session
	fmt.Printf("Creating session with repo: %s\n", repoURL)
	return "session-" + generateID()
}

func (sm *SessionManager) CreateLocalSession() string {
	// TODO: Implement local directory upload and session creation
	fmt.Println("Uploading local directory...")
	return "session-" + generateID()
}

func (sm *SessionManager) ConnectToSession(sessionID string) error {
	wsURL := fmt.Sprintf("wss://%s/sessions/%s/connect", sm.apiURL, sessionID)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to session %s\n", sessionID)
	fmt.Println("Type 'exit' to disconnect")
	fmt.Println()

	// Handle bidirectional communication
	done := make(chan struct{})

	// Read from WebSocket and write to stdout
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("Connection error: %v\n", err)
				}
				return
			}
			fmt.Print(string(message))
		}
	}()

	// Read from stdin and send to WebSocket
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if text == "exit" {
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, []byte(text)); err != nil {
				fmt.Printf("Write error: %v\n", err)
				return
			}
		}
	}()

	<-done
	fmt.Println("\nDisconnected from session")
	return nil
}

func (sm *SessionManager) ListSessions() []Session {
	// TODO: Implement API call
	// Mock data for now
	return []Session{
		{
			ID:          "session-abc123",
			Status:      "running",
			Description: "github.com/user/repo",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "session-def456",
			Status:      "stopped",
			Description: "local project",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
	}
}

func (sm *SessionManager) DeleteSession(sessionID string) error {
	// TODO: Implement API call
	fmt.Printf("Deleting session: %s\n", sessionID)
	return nil
}

func (sm *SessionManager) GetSessionStatus(sessionID string) SessionStatus {
	// TODO: Implement API call
	return SessionStatus{
		State:        "running",
		CreatedAt:    time.Now().Add(-1 * time.Hour),
		LastActivity: time.Now().Add(-5 * time.Minute),
	}
}

func (sm *SessionManager) StreamLogs(sessionID string) {
	// TODO: Implement WebSocket streaming of logs
	fmt.Printf("Streaming logs for session %s...\n", sessionID)
	fmt.Println("2025-01-15 10:30:00 | Claude: Starting work on TODO list")
	fmt.Println("2025-01-15 10:30:15 | Claude: Analyzing codebase structure")
	fmt.Println("2025-01-15 10:31:00 | Claude: Implementing authentication feature")
}

func (sm *SessionManager) GetMostRecentSession() string {
	sessions := sm.ListSessions()
	if len(sessions) > 0 {
		return sessions[0].ID
	}
	return ""
}

func generateID() string {
	// Simple ID generation - in production use UUID
	return fmt.Sprintf("%d", time.Now().Unix())
}
