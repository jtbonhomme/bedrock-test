package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPMCPClient implements MCPClientInterface for HTTP-based MCP servers
type HTTPMCPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// MCPToolRequest represents a request to execute a tool via HTTP
type MCPToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// NewMCPHTTPClient creates a new HTTP MCP client
func NewMCPHTTPClient(baseURL string) *HTTPMCPClient {
	return &HTTPMCPClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExecuteTool executes a tool via HTTP
func (c *HTTPMCPClient) ExecuteTool(request MCPToolRequest) (MCPToolResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return MCPToolResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/tools/execute",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return MCPToolResponse{}, fmt.Errorf("failed to execute tool: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MCPToolResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	var toolResp MCPToolResponse
	if err := json.Unmarshal(body, &toolResp); err != nil {
		return MCPToolResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return toolResp, nil
}

// ListDatabases lists available databases via HTTP
func (c *HTTPMCPClient) ListDatabases() (MCPToolResponse, error) {
	return c.ExecuteTool(MCPToolRequest{
		Name:      "mcp_postgres_list_database",
		Arguments: map[string]interface{}{},
	})
}

// ListTables lists available tables via HTTP
func (c *HTTPMCPClient) ListTables() (MCPToolResponse, error) {
	return c.ExecuteTool(MCPToolRequest{
		Name:      "mcp_postgres_list_table",
		Arguments: map[string]interface{}{},
	})
}

// ExecuteReadQuery executes a read-only SQL query via HTTP
func (c *HTTPMCPClient) ExecuteReadQuery(query string) (MCPToolResponse, error) {
	return c.ExecuteTool(MCPToolRequest{
		Name: "mcp_postgres_read_query",
		Arguments: map[string]interface{}{
			"query": query,
		},
	})
}

// DescribeTable describes the structure of a table via HTTP
func (c *HTTPMCPClient) DescribeTable(tableName string) (MCPToolResponse, error) {
	return c.ExecuteTool(MCPToolRequest{
		Name: "mcp_postgres_desc_table",
		Arguments: map[string]interface{}{
			"name": tableName,
		},
	})
}

// Close closes the HTTP client (no-op for HTTP)
func (c *HTTPMCPClient) Close() error {
	// HTTP client doesn't need explicit cleanup
	return nil
}
