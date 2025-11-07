package workflowy

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var testToken string

// TestMain prepares the test environment by reading the API token from the .env file or
// environment variable.
func TestMain(m *testing.M) {
	if data, err := os.ReadFile(".env"); err == nil {
		for line := range strings.SplitSeq(string(data), "\n") {
			if kv := strings.SplitN(line, "=", 2); len(kv) == 2 {
				os.Setenv(kv[0], kv[1])
			}
		}
	}
	if testToken = os.Getenv("WORKFLOWY_API_KEY"); testToken == "" {
		fmt.Fprintln(os.Stderr, "WORKFLOWY_API_KEY not set, skipping integration tests")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// TestClient_SetBaseURL checks that BaseURL can be changed and reset to default.
func TestClient_SetBaseURL(t *testing.T) {
	c := NewClient("test-token")
	custom := "https://example.org/api/" // trailing slash gets trimmed
	c.SetBaseURL(custom)
	require.Equal(t, "https://example.org/api", c.baseURL)

	// empty resets to default
	c.SetBaseURL("")
	require.Equal(t, BaseURL, c.baseURL)
}

// TestClient_SetHTTPClient checks that HTTP client can be changed.
func TestClient_SetHTTPClient(t *testing.T) {
	c := NewClient("test-token")
	hc := &http.Client{Timeout: 42 * time.Second}
	c.SetHTTPClient(hc)
	require.Equal(t, hc.Timeout, c.httpClient.Timeout)
}

// TestFullCycle tests the full cycle of:
//   - creating,
//   - getting,
//   - listing,
//   - updating,
//   - moving,
//   - completing,
//   - uncompleting,
//   - deleting a node.
//
// It creates a new testing node at the root level and deletes it in the end so the account data
// is unchanged if the test succeeds.
func TestFullCycle(t *testing.T) {
	c := NewClient(testToken)

	ctx := t.Context()

	// Create a node
	name := "API Test Node"
	nodeID, err := c.CreateNode(ctx, Create{
		Name:     name,
		Position: &PositionBottom,
	})
	require.NoError(t, err)
	require.NotEmpty(t, nodeID)
	t.Logf("created nodeID: %s", nodeID)

	// Get the node
	node, err := c.GetNode(ctx, nodeID)
	require.NoError(t, err)
	require.NotNil(t, node)
	require.Equal(t, name, node.Name)
	t.Logf("got node: %+v", node)

	// List children (zero)
	children, err := c.ListNodes(ctx, nodeID)
	require.NoError(t, err)
	require.Empty(t, children)
	t.Logf("listed children: %d", len(children))

	// Create a child node
	childName := "Test Child Node"
	childID, err := c.CreateNode(ctx, Create{
		ParentID: nodeID,
		Name:     childName,
	})
	require.NoError(t, err)
	require.NotEmpty(t, childID)
	t.Logf("created childID: %s", childID)

	// List children (one)
	children, err = c.ListNodes(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, 1, len(children))
	require.Equal(t, childID, children[0].ID)
	t.Logf("listed children: %d", len(children))

	// Move the child node to a new parent
	moveTargetName := "API Test Move Target"
	moveTargetID, err := c.CreateNode(ctx, Create{
		Name:     moveTargetName,
		Position: &PositionBottom,
	})
	require.NoError(t, err)
	require.NotEmpty(t, moveTargetID)
	t.Logf("created moveTargetID: %s", moveTargetID)

	err = c.MoveNode(ctx, childID, Move{
		ParentID: moveTargetID,
		Position: &PositionTop,
	})
	require.NoError(t, err)

	children, err = c.ListNodes(ctx, nodeID)
	require.NoError(t, err)
	require.Empty(t, children)
	t.Logf("listed children after move: %d", len(children))

	movedChildren, err := c.ListNodes(ctx, moveTargetID)
	require.NoError(t, err)
	require.Equal(t, 1, len(movedChildren))
	require.Equal(t, childID, movedChildren[0].ID)
	t.Logf("listed moved child: %d", len(movedChildren))

	// Update the node
	name = fmt.Sprintf("%s %s", name, node.ID)
	err = c.UpdateNode(ctx, nodeID, Update{
		Name: &name,
	})
	require.NoError(t, err)
	node, err = c.GetNode(ctx, nodeID)
	require.NoError(t, err)
	require.NotNil(t, node)
	require.Equal(t, name, node.Name)
	t.Logf("updated node: %+v", node)

	// Complete the node
	err = c.CompleteNode(ctx, nodeID)
	require.NoError(t, err)
	node, err = c.GetNode(ctx, nodeID)
	require.NoError(t, err)
	require.NotNil(t, node)
	require.Equal(t, true, node.Completed)
	t.Logf("completed node: %+v", node)

	// Uncomplete the node
	err = c.UncompleteNode(ctx, nodeID)
	require.NoError(t, err)
	node, err = c.GetNode(ctx, nodeID)
	require.NoError(t, err)
	require.NotNil(t, node)
	require.Equal(t, false, node.Completed)
	t.Logf("uncompleted node: %+v", node)

	// Delete the moved child node
	err = c.DeleteNode(ctx, childID)
	require.NoError(t, err)
	deletedChild, err := c.GetNode(ctx, childID)
	require.Error(t, err)
	require.Nil(t, deletedChild)
	t.Logf("deleted child node %s", childID)

	// Delete the node
	err = c.DeleteNode(ctx, nodeID)
	require.NoError(t, err)
	node, err = c.GetNode(ctx, nodeID)
	require.Error(t, err)
	require.Nil(t, node)
	t.Logf("deleted node %s", nodeID)

	// Delete the move target node
	err = c.DeleteNode(ctx, moveTargetID)
	require.NoError(t, err)
	deletedTarget, err := c.GetNode(ctx, moveTargetID)
	require.Error(t, err)
	require.Nil(t, deletedTarget)
	t.Logf("deleted move target node %s", moveTargetID)
}

// TestListTargets tests the ListTargets function.
func TestListTargets(t *testing.T) {
	c := NewClient(testToken)
	ctx := t.Context()
	targets, err := c.ListTargets(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, targets)
	t.Logf("listed targets: %d", len(targets))
}

// TestExportAll tests the ExportAll function.
func TestExportAll(t *testing.T) {
	c := NewClient(testToken)
	ctx := t.Context()
	nodes, err := c.ExportAll(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, nodes)
	t.Logf("exported nodes: %d", len(nodes))
}
