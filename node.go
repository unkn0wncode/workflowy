// Package workflowy / node.go defines the Node type for the Workflowy API Client.
package workflowy

// Node represents a Workflowy node as defined in the public API.
type Node struct {
	ID          string   `json:"id"`                    // UUID
	Name        string   `json:"name"`                  // Main content of the node
	Note        *string  `json:"note,omitempty"`        // Subtext
	Priority    int      `json:"priority"`              // Sorting order, lower means higher on the list
	Data        NodeData `json:"data"`                  // Display options
	Completed   bool     `json:"completed"`             // Completion status
	CreatedAt   int64    `json:"createdAt"`             // Creation timestamp
	ModifiedAt  int64    `json:"modifiedAt"`            // Last modification timestamp
	CompletedAt *int64   `json:"completedAt,omitempty"` // Completion timestamp, nil if not completed
}

// LayoutMode is the display mode for a node.
type LayoutMode string

// LayoutMode constants. Declared as var for use as pointer values, do not modify on runtime.
var (
	// these modes are documented in the API reference

	LayoutModeBullets  LayoutMode = "bullets"
	LayoutModeTodo     LayoutMode = "todo"
	LayoutModeHeading1 LayoutMode = "h1"
	LayoutModeH2       LayoutMode = "h2"
	LayoutModeH3       LayoutMode = "h3"

	// these modes are not documented in the API reference, but are present in the API responses

	LayoutModeParagraph LayoutMode = "p"
	LayoutModeQuote     LayoutMode = "quote-block"
	LayoutModeBoard     LayoutMode = "board"
	LayoutModeDashboard LayoutMode = "dashboard"
	LayoutModeCode      LayoutMode = "code-block"
	LayoutModeDivider   LayoutMode = "divider"
)

// NodeData holds nested node metadata.
type NodeData struct {
	LayoutMode LayoutMode `json:"layoutMode"`
}
