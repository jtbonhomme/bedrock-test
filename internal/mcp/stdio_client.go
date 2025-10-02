package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
)

// MCPContent represents content in an MCP response
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MCPTool represents a tool available in the MCP server
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPToolResponse represents the response from a tool execution
type MCPToolResponse struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// StdioMCPClient represents a client that communicates with MCP server via stdio
type StdioMCPClient struct {
	executable string
	dsn        string
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	mu         sync.Mutex
	requestID  int
}

// MCPRequest represents a request to MCP server
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a response from MCP server
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents an error from MCP server
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolCallParams represents parameters for tool call
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// NewStdioMCPClient creates a new stdio MCP client
func NewStdioMCPClient(executable, dsn string) *StdioMCPClient {
	return &StdioMCPClient{
		executable: executable,
		dsn:        dsn,
	}
}

// Connect starts the MCP server process and establishes communication
func (c *StdioMCPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd != nil {
		return fmt.Errorf("already connected")
	}

	// Start the MCP server process
	c.cmd = exec.Command(c.executable, "--dsn", c.dsn)

	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	log.Printf("Started MCP server process: %s --dsn %s", c.executable, c.dsn)

	// Initialize the connection
	return c.initialize()
}

// initialize sends the initial handshake to the MCP server
func (c *StdioMCPClient) initialize() error {
	// Send initialize request
	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "bedrock-mcp-client",
				"version": "1.0.0",
			},
		},
	}

	resp, err := c.sendRequest(initReq)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialization error: %s", resp.Error.Message)
	}

	log.Println("MCP server initialized successfully")
	return nil
}

// sendRequest sends a request and waits for response
func (c *StdioMCPClient) sendRequest(req MCPRequest) (*MCPResponse, error) {
	// Serialize request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("Sending MCP request: %s", string(reqBytes))

	// Send request
	if _, err := c.stdin.Write(append(reqBytes, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read response
	scanner := bufio.NewScanner(c.stdout)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		return nil, fmt.Errorf("no response received")
	}

	respBytes := scanner.Bytes()
	log.Printf("Received MCP response: %s", string(respBytes))

	// Parse response
	var resp MCPResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// getNextID returns the next request ID
func (c *StdioMCPClient) getNextID() int {
	c.requestID++
	return c.requestID
}

// ListTools lists available tools from the MCP server
func (c *StdioMCPClient) ListTools() ([]MCPTool, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "tools/list",
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list error: %s", resp.Error.Message)
	}

	var result struct {
		Tools []MCPTool `json:"tools"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools result: %w", err)
	}

	return result.Tools, nil
}

// CallTool executes a tool via the MCP server
func (c *StdioMCPClient) CallTool(name string, arguments map[string]interface{}) (ToolResult, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "tools/call",
		Params: ToolCallParams{
			Name:      name,
			Arguments: arguments,
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return ToolResult{}, err
	}

	if resp.Error != nil {
		return ToolResult{}, fmt.Errorf("tools/call error: %s", resp.Error.Message)
	}

	var result ToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return ToolResult{}, fmt.Errorf("failed to unmarshal tool result: %w", err)
	}

	return result, nil
}

// Close closes the connection to the MCP server
func (c *StdioMCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd == nil {
		return nil
	}

	// Close pipes
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}

	// Terminate process
	if c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}

	c.cmd = nil
	return nil
}

// Convenience methods that match the HTTP client interface

// ListDatabases lists available databases
func (c *StdioMCPClient) ListDatabases() (MCPToolResponse, error) {
	result, err := c.CallTool("mcp_postgres_list_database", map[string]interface{}{})
	if err != nil {
		return MCPToolResponse{}, err
	}

	return MCPToolResponse{
		Content: result.Content,
		IsError: result.IsError,
	}, nil
}

// ListTables lists available tables
func (c *StdioMCPClient) ListTables() (MCPToolResponse, error) {
	result, err := c.CallTool("mcp_postgres_list_table", map[string]interface{}{})
	if err != nil {
		return MCPToolResponse{}, err
	}

	return MCPToolResponse{
		Content: result.Content,
		IsError: result.IsError,
	}, nil
}

// ExecuteReadQuery executes a read-only SQL query
func (c *StdioMCPClient) ExecuteReadQuery(query string) (MCPToolResponse, error) {
	result, err := c.CallTool("mcp_postgres_read_query", map[string]interface{}{
		"query": query,
	})
	if err != nil {
		return MCPToolResponse{}, err
	}

	return MCPToolResponse{
		Content: result.Content,
		IsError: result.IsError,
	}, nil
}

// DescribeTable describes the structure of a table
func (c *StdioMCPClient) DescribeTable(tableName string) (MCPToolResponse, error) {
	result, err := c.CallTool("mcp_postgres_desc_table", map[string]interface{}{
		"name": tableName,
	})
	if err != nil {
		return MCPToolResponse{}, err
	}

	return MCPToolResponse{
		Content: result.Content,
		IsError: result.IsError,
	}, nil
}
