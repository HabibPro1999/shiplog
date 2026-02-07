package provider

import (
	"github.com/HabibPro1999/shiplog/internal/parser"
	"github.com/HabibPro1999/shiplog/internal/session"
)

// Provider defines the interface that each AI coding tool must implement.
type Provider interface {
	// Name returns the tool identifier.
	Name() session.Tool

	// Discover finds sessions on disk. If projectFilter is non-empty, only
	// sessions belonging to that project path are returned.
	Discover(home string, projectFilter string) ([]session.SessionInfo, error)

	// Parse reads a session file and returns its messages and metadata.
	Parse(filePath string) ([]parser.Message, parser.SessionMeta, error)
}

// registry holds all known provider instances.
// Providers are added here as they are implemented.
var registry []Provider

// Register adds a provider to the global registry.
// Called by provider init() functions or explicitly during setup.
func Register(p Provider) {
	registry = append(registry, p)
}

// All returns all registered providers.
func All() []Provider {
	return registry
}

// ForTool returns the provider matching the given tool, or nil if not found.
func ForTool(tool session.Tool) Provider {
	for _, p := range registry {
		if p.Name() == tool {
			return p
		}
	}
	return nil
}
