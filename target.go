// Package workflowy / target.go defined the Target type for the Workflowy API Client.
package workflowy

// Target represents a Workflowy target as defined in the public API.
type Target struct {
	Key  string  `json:"key"`
	Type string  `json:"type"`
	Name *string `json:"name,omitempty"`
}

// TargetType is the type of target as defined in the public API.
type TargetType string

const (
	TargetTypeShortcut TargetType = "shortcut" // User-defined shortcut
	TargetTypeSystem   TargetType = "system"   // System-managed locations like inbox. Always returned, even if the target node hasn't been created yet.
)
