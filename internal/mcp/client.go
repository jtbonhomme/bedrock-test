package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// MCP Client
type MCPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewMCPClient(baseURL string) *MCPClient {
	return &MCPClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get available tools from MCP server
func (c *MCPClient) GetTools() ([]MCPTool, error) {
	log.Debug().Msg("calling GetTools")

	resp, err := c.HTTPClient.Get(c.BaseURL + "/tools")
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tools []MCPTool
	if err := json.Unmarshal(body, &tools); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools: %w", err)
	}

	return tools, nil
}

// Execute a tool via MCP server
func (c *MCPClient) ExecuteTool(request MCPToolRequest) (MCPToolResponse, error) {
	log.Debug().Msgf("Executing MCP tool: %s with args: %v", request.Name, request.Arguments)

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

	log.Debug().Msgf("MCP tool response: %v", toolResp)
	return toolResp, nil
}

// Get PostgreSQL databases
func (c *MCPClient) ListDatabases() (MCPToolResponse, error) {
	return c.ExecuteTool(MCPToolRequest{
		Name:      "mcp_postgres_list_database",
		Arguments: map[string]interface{}{},
	})
}

// Get PostgreSQL tables
func (c *MCPClient) ListTables() (MCPToolResponse, error) {
	return c.ExecuteTool(MCPToolRequest{
		Name:      "mcp_postgres_list_table",
		Arguments: map[string]interface{}{},
	})
}

// Execute PostgreSQL read query
func (c *MCPClient) ExecuteReadQuery(query string) (MCPToolResponse, error) {
	return c.ExecuteTool(MCPToolRequest{
		Name: "mcp_postgres_read_query",
		Arguments: map[string]interface{}{
			"query": query,
		},
	})
}

// Describe table structure
func (c *MCPClient) DescribeTable(tableName string) (MCPToolResponse, error) {
	return c.ExecuteTool(MCPToolRequest{
		Name: "mcp_postgres_desc_table",
		Arguments: map[string]interface{}{
			"name": tableName,
		},
	})
}
