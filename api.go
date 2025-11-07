// Package workflowy / api.go defines API calls for the Workflowy API Client.
package workflowy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Create is the input to CreateNode.
type Create struct {
	ParentID   string      `json:"parent_id,omitempty"` // Leave empty to create at the root level
	Name       string      `json:"name"`
	Note       *string     `json:"note,omitempty"`
	LayoutMode *LayoutMode `json:"layoutMode,omitempty"`
	Position   *Position   `json:"position,omitempty"`
}

// Position is the position of a node in the parent's list of children nodes ("top" or "bottom").
type Position string

// Position constants. Declared as var for use as pointer values, do not modify on runtime.
var (
	PositionTop    Position = "top"
	PositionBottom Position = "bottom"
)

// Update is the input to UpdateNode.
type Update struct {
	Name       *string     `json:"name,omitempty"`
	Note       *string     `json:"note,omitempty"`
	LayoutMode *LayoutMode `json:"layoutMode,omitempty"`
}

// Move is the input to MoveNode.
type Move struct {
	ParentID string    `json:"parent_id"`          // node ID or target key
	Position *Position `json:"position,omitempty"` // defaults to "top" if not provided
}

type statusResponse struct {
	Status string `json:"status"`
}

// GetNode fetches a single node by its ID.
func (c *Client) GetNode(ctx context.Context, nodeID string) (*Node, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("nodeID is required")
	}
	req, err := c.newRequest(ctx, http.MethodGet, "/nodes/"+url.PathEscape(nodeID), nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Node Node `json:"node"`
	}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return &out.Node, nil
}

// ListNodes returns children of a given parent (unordered; sort by priority client-side).
func (c *Client) ListNodes(ctx context.Context, parentID string) ([]*Node, error) {
	path := "/nodes"
	if parentID != "" {
		path += "?parent_id=" + url.QueryEscape(parentID)
	}
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Nodes []*Node `json:"nodes"`
	}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Nodes, nil
}

// CreateNode creates a child node under a given parent. Returns the new item ID.
func (c *Client) CreateNode(ctx context.Context, in Create) (string, error) {
	if in.Name == "" {
		return "", fmt.Errorf("name is required")
	}
	req, err := c.newRequest(ctx, http.MethodPost, "/nodes", in)
	if err != nil {
		return "", err
	}
	var resp struct {
		ItemID string `json:"item_id"`
	}
	if err := c.do(req, &resp); err != nil {
		return "", err
	}
	if resp.ItemID == "" {
		return "", fmt.Errorf("empty item_id in response")
	}
	return resp.ItemID, nil
}

// UpdateNode updates a node's fields.
func (c *Client) UpdateNode(ctx context.Context, nodeID string, in Update) error {
	if nodeID == "" {
		return fmt.Errorf("nodeID is required")
	}
	req, err := c.newRequest(ctx, http.MethodPost, "/nodes/"+url.PathEscape(nodeID), in)
	if err != nil {
		return err
	}
	var status statusResponse
	if err := c.do(req, &status); err != nil {
		return err
	}
	if status.Status != "ok" {
		return fmt.Errorf("unexpected status: %s", status.Status)
	}
	return nil
}

// MoveNode moves a node to a new parent/target.
func (c *Client) MoveNode(ctx context.Context, nodeID string, in Move) error {
	if nodeID == "" {
		return fmt.Errorf("nodeID is required")
	}
	if in.ParentID == "" {
		return fmt.Errorf("parentID is required")
	}
	req, err := c.newRequest(ctx, http.MethodPost, "/nodes/"+url.PathEscape(nodeID)+"/move", in)
	if err != nil {
		return err
	}
	var status statusResponse
	if err := c.do(req, &status); err != nil {
		return err
	}
	if status.Status != "ok" {
		return fmt.Errorf("unexpected status: %s", status.Status)
	}
	return nil
}

// DeleteNode deletes a node by ID.
func (c *Client) DeleteNode(ctx context.Context, nodeID string) error {
	if nodeID == "" {
		return fmt.Errorf("nodeID is required")
	}
	req, err := c.newRequest(ctx, http.MethodDelete, "/nodes/"+url.PathEscape(nodeID), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// CompleteNode marks a node as completed.
func (c *Client) CompleteNode(ctx context.Context, nodeID string) error {
	if nodeID == "" {
		return fmt.Errorf("nodeID is required")
	}
	req, err := c.newRequest(ctx, http.MethodPost, "/nodes/"+url.PathEscape(nodeID)+"/complete", nil)
	if err != nil {
		return err
	}
	var status statusResponse
	if err := c.do(req, &status); err != nil {
		return err
	}
	if status.Status != "ok" {
		return fmt.Errorf("unexpected status: %s", status.Status)
	}
	return nil
}

// UncompleteNode marks a node as not completed.
func (c *Client) UncompleteNode(ctx context.Context, nodeID string) error {
	if nodeID == "" {
		return fmt.Errorf("nodeID is required")
	}
	req, err := c.newRequest(ctx, http.MethodPost, "/nodes/"+url.PathEscape(nodeID)+"/uncomplete", nil)
	if err != nil {
		return err
	}
	var status statusResponse
	if err := c.do(req, &status); err != nil {
		return err
	}
	if status.Status != "ok" {
		return fmt.Errorf("unexpected status: %s", status.Status)
	}
	return nil
}

// ListTargets returns a list of targets.
func (c *Client) ListTargets(ctx context.Context) ([]*Target, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/targets", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Targets []*Target `json:"targets"`
	}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Targets, nil
}

// ExportAll returns all nodes as a flat list.
// This request is rate limited to 1 per minute.
func (c *Client) ExportAll(ctx context.Context) ([]*Node, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/nodes-export", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Nodes []*Node `json:"nodes"`
	}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Nodes, nil
}
