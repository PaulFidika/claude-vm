package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	repoURL   string
	sessionID string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "claude-vm",
		Short: "Manage remote Claude Code sessions",
		Long:  `claude-vm allows you to run Claude Code in remote VMs with git integration`,
		Run: func(cmd *cobra.Command, args []string) {
			// Default behavior: start new session
			if repoURL != "" {
				startSession(repoURL, true)
			} else {
				startSession(".", false)
			}
		},
	}

	// Start command
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new Claude VM session",
		Run: func(cmd *cobra.Command, args []string) {
			if repoURL != "" {
				startSession(repoURL, true)
			} else {
				startSession(".", false)
			}
		},
	}
	startCmd.Flags().StringVar(&repoURL, "repo", "", "GitHub repository URL")
	rootCmd.AddCommand(startCmd)

	// Connect command
	connectCmd := &cobra.Command{
		Use:   "connect [session-id]",
		Short: "Connect to an existing session",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sid := sessionID
			if len(args) > 0 {
				sid = args[0]
			}
			if sid == "" {
				sid = getMostRecentSession()
			}
			connectToSession(sid)
		},
	}
	rootCmd.AddCommand(connectCmd)

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Run: func(cmd *cobra.Command, args []string) {
			listSessions()
		},
	}
	rootCmd.AddCommand(listCmd)

	// Delete command
	deleteCmd := &cobra.Command{
		Use:   "delete <session-id>",
		Short: "Delete a session",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			deleteSession(args[0])
		},
	}
	rootCmd.AddCommand(deleteCmd)

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status <session-id>",
		Short: "Show session status",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showStatus(args[0])
		},
	}
	rootCmd.AddCommand(statusCmd)

	// Logs command
	logsCmd := &cobra.Command{
		Use:   "logs <session-id>",
		Short: "View session logs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showLogs(args[0])
		},
	}
	rootCmd.AddCommand(logsCmd)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&sessionID, "session", "", "Session ID to use")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startSession(source string, isRepo bool) {
	sm := NewSessionManager()
	if isRepo {
		fmt.Printf("Starting new session with repo: %s\n", source)
		sid := sm.CreateRemoteSession(source)
		connectToSession(sid)
	} else {
		fmt.Println("Starting new session with local directory...")
		sid := sm.CreateLocalSession()
		connectToSession(sid)
	}
}

func connectToSession(sessionID string) {
	fmt.Printf("Connecting to session %s...\n", sessionID)
	sm := NewSessionManager()
	sm.ConnectToSession(sessionID)
}

func listSessions() {
	sm := NewSessionManager()
	sessions := sm.ListSessions()

	fmt.Println("Active sessions:")
	for _, s := range sessions {
		fmt.Printf("  %s (%s) - %s\n", s.ID, s.Status, s.Description)
	}
}

func deleteSession(sessionID string) {
	sm := NewSessionManager()
	if err := sm.DeleteSession(sessionID); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting session: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Session %s deleted\n", sessionID)
}

func showStatus(sessionID string) {
	sm := NewSessionManager()
	status := sm.GetSessionStatus(sessionID)
	fmt.Printf("Session %s:\n", sessionID)
	fmt.Printf("  Status: %s\n", status.State)
	fmt.Printf("  Created: %s\n", status.CreatedAt)
	fmt.Printf("  Last Activity: %s\n", status.LastActivity)
}

func showLogs(sessionID string) {
	sm := NewSessionManager()
	sm.StreamLogs(sessionID)
}

func getMostRecentSession() string {
	sm := NewSessionManager()
	return sm.GetMostRecentSession()
}
