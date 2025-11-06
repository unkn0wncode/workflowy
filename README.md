# Workflowy Go Client

Minimal Workflowy API wrapper for Golang. Initialize with an API key, then call methods on a `Client` object.

It reflects the official documentation at https://workflowy.com/api-reference/, aside from listing a few newer layout modes that the documentation is behind on.

It has no external dependencies in production and `stretchr/testify` is used for testing.

## Install

```bash
go get github.com/unkn0wncode/workflowy
```

## API Client Object

Create a `Client` object with your API key using `workflowy.NewClient(key)`. You can configure it with the following methods:
- `SetBaseURL` to change the base URL
- `SetHTTPClient` to change the HTTP client

The default HTTP client has a 15 second timeout. It can be created with the `DefaultHTTPClient` function if you want to modify and reuse it.

Methods for API calls:
- `CreateNode` - create a node at parent or root, pass parameters in `Create` struct, returns the new node's ID
- `GetNode` - get a node by ID, returns a `*Node` struct
- `ListNodes` - list nodes under a parent (or root), returns `[]*Node`
- `UpdateNode` - update a node, pass parameters in `Update` struct
- `MoveNode` - move a node to a new parent, pass parameters in `Move` struct
- `DeleteNode` - delete a node by ID
- `CompleteNode` - complete a node by ID
- `UncompleteNode` - uncomplete a node by ID
- `ListTargets` - list all existing targets, returns `[]*Target` (system targets may be returned even if the target node hasn't been created yet)

The `Client` also handles 429 Too Many Requests error, since the API documentation does not provide rate limits. It retries up to 3 times, following directions in "Retry-After" header and respecting `Context` expiration.

## Basic usage example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/unkn0wncode/workflowy"
)

func main() {
    c := workflowy.NewClient(os.Getenv("WORKFLOWY_API_KEY"))
    ctx := context.Background()

    // Create a node at the root
    nodeID, err := c.CreateNode(ctx, workflowy.Create{
        Name: "Hello API",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Get the node
    node, err := c.GetNode(ctx, nodeID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created and fetched node: %+v\n", node)
}
```

## Other operations

- List nodes under a parent (empty parent lists roots)
```go
nodes, err := c.ListNodes(ctx, parentID)
if err != nil {
    log.Fatal(err)
}
for i, node := range nodes {
    fmt.Printf("Node %d: %+v\n", i+1, node)
}
```

- Create a child node at the bottom with specific layout mode
```go
childID, err := c.CreateNode(ctx, workflowy.Create{
    ParentID:   parentID,
    Name:       "Child node",
    Position:   &workflowy.PositionBottom,
    LayoutMode: &workflowy.LayoutModeTodo,
})
```

- Update a node's name (main text content)
```go
newName := "Updated"
err := c.UpdateNode(ctx, nodeID, workflowy.Update{Name: &newName})
```

- Move a node to a new parent
```go
err := c.MoveNode(ctx, nodeID, workflowy.Move{ParentID: parentID, Position: &workflowy.PositionBottom})
```

- Complete and uncomplete
```go
_ = c.CompleteNode(ctx, nodeID)
_ = c.UncompleteNode(ctx, nodeID)
```

- Delete a node
```go
err := c.DeleteNode(ctx, nodeID)
```

- List all existing targets
```go
targets, err := c.ListTargets(ctx)
if err != nil {
    log.Fatal(err)
}
for _, target := range targets {
    fmt.Printf("Target: %+v\n", target)
}
```

## Notes

- Default base URL is `workflowy.BaseURL` (`https://workflowy.com/api/v1`). Use `Client.SetBaseURL` to change it.
- Errors include HTTP status codes via `APIError`.
- To set node type, use the layout mode constants (that are actually vars so you can use them as pointers):
  - `LayoutModeBullets`
  - `LayoutModeTodo`
  - `LayoutModeH1`
  - `LayoutModeH2`
  - `LayoutModeH3`
  - `LayoutModeParagraph`
  - `LayoutModeQuote`
  - `LayoutModeBoard`
  - `LayoutModeDashboard`
  - `LayoutModeCode`
  - `LayoutModeDivider`
- Some other formatting is done with HTML tags, for example `<code>abc123</code>` for inline code.
- IDs of nodes are not readily exposed in the UI. Links contain the last fragment of the ID, but the API won't accept that.

## Testing

Unit tests take the API key from either `WORKFLOWY_API_KEY` environment variable or from an `.env` file, then execute real requests with that key. A node is created for tests at the root level and deleted in the end, so your data should be unchanged, but a small portion of the usage limit is consumed.
