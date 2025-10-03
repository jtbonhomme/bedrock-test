package mcp

import (
	"fmt"
	"strings"
)

// MCPClientInterface defines the common interface for both HTTP and stdio clients
type MCPClientInterface interface {
	ListDatabases() (MCPToolResponse, error)
	ListTables() (MCPToolResponse, error)
	ExecuteReadQuery(query string) (MCPToolResponse, error)
	DescribeTable(tableName string) (MCPToolResponse, error)
	Close() error
}

// MCPClientConfig holds configuration for creating MCP clients
type MCPClientConfig struct {
	// For HTTP client
	BaseURL string

	// For stdio client
	Executable string
	DSN        string
}

// NewMCPClientAuto creates the appropriate MCP client based on configuration
func NewMCPClientAuto(config MCPClientConfig) (MCPClientInterface, error) {
	// If BaseURL is provided, use HTTP client
	if config.BaseURL != "" {
		return NewMCPHTTPClient(config.BaseURL), nil
	}

	// If executable and DSN are provided, use stdio client
	if config.Executable != "" && config.DSN != "" {
		client := NewStdioMCPClient(config.Executable, config.DSN)
		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect stdio client: %w", err)
		}
		return client, nil
	}

	return nil, fmt.Errorf("invalid configuration: need either BaseURL or (Executable + DSN)")
}

// DetectMCPServerType attempts to detect which type of MCP server to use
func DetectMCPServerType(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return "http"
	}
	if strings.Contains(url, "://") {
		return "stdio" // Assume DSN format like postgresql://...
	}
	if strings.Contains(url, "/") || strings.Contains(url, "\\") {
		return "stdio" // Assume file path
	}
	return "unknown"
}

// ParseMCPConfig parses a configuration string and returns appropriate config
func ParseMCPConfig(configStr string) (MCPClientConfig, error) {
	serverType := DetectMCPServerType(configStr)

	switch serverType {
	case "http":
		return MCPClientConfig{BaseURL: configStr}, nil

	case "stdio":
		// For stdio, we need to determine if it's a DSN or executable path
		if strings.Contains(configStr, "://") {
			// It's a DSN, use default executable
			return MCPClientConfig{
				Executable: "go-mcp-postgres",
				DSN:        configStr,
			}, nil
		} else {
			// It's an executable path, need DSN from environment
			return MCPClientConfig{
				Executable: configStr,
				DSN:        "", // Will need to be provided separately
			}, nil
		}

	default:
		return MCPClientConfig{}, fmt.Errorf("unknown MCP server type for: %s", configStr)
	}
}
