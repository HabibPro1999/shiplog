package session

import (
	"sort"
	"strings"
	"time"
)

// ProviderDiscoverer is the interface for session discovery.
// This avoids a direct import of the provider package from session.
type ProviderDiscoverer interface {
	Name() Tool
	Discover(home string, projectFilter string) ([]SessionInfo, error)
}

// FindAllSessions queries all providers for sessions, applying optional filters.
//   - home: user home directory
//   - projectFilter: absolute path to scope to a specific project, or "" for all
//   - toolFilter: tool name to restrict results, or "" for all tools
//   - recentDays: if > 0, only return sessions from the last N days
//   - providers: list of providers to query
//
// Results are sorted by timestamp descending (most recent first).
func FindAllSessions(home, projectFilter string, toolFilter Tool, recentDays int, providers []ProviderDiscoverer) ([]SessionInfo, error) {
	var sessions []SessionInfo

	for _, p := range providers {
		if toolFilter != "" && p.Name() != toolFilter {
			continue
		}
		found, err := p.Discover(home, projectFilter)
		if err != nil {
			continue
		}
		sessions = append(sessions, found...)
	}

	// Apply recent days filter
	if recentDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -recentDays).UTC().Format(time.RFC3339)
		var filtered []SessionInfo
		for _, s := range sessions {
			if s.Timestamp >= cutoff {
				filtered = append(filtered, s)
			} else if s.Timestamp == "" {
				// Keep sessions with no timestamp (can't filter them)
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Timestamp > sessions[j].Timestamp
	})

	return sessions, nil
}

// ParseTool converts a string tool flag value to a Tool constant.
// Returns empty string if the input doesn't match any known tool.
func ParseTool(s string) Tool {
	normalized := Tool(strings.ToLower(strings.TrimSpace(s)))
	switch normalized {
	case ToolClaude, ToolCodex, ToolAider, ToolOpenCode, ToolAmazonQ:
		return normalized
	}
	// Also try matching display names
	switch strings.ToLower(s) {
	case "claude code", "claude":
		return ToolClaude
	case "codex cli", "codex":
		return ToolCodex
	case "aider":
		return ToolAider
	case "opencode":
		return ToolOpenCode
	case "amazon q", "amazonq":
		return ToolAmazonQ
	}
	return ""
}
