package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// MCPTool represents a tool available in the MCP server
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPToolRequest represents a request to execute a tool
type MCPToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPToolResponse represents the response from a tool execution
type MCPToolResponse struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent represents content in the response
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Get available tools
	mux.HandleFunc("/tools", getTools)

	// Execute tool
	mux.HandleFunc("/tools/execute", executeTool)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("MCP Server is running"))
	})

	fmt.Printf("🚀 MCP Server starting on port %s\n", port)
	fmt.Printf("📋 Available endpoints:\n")
	fmt.Printf("   GET  /health        - Health check\n")
	fmt.Printf("   GET  /tools         - List available tools\n")
	fmt.Printf("   POST /tools/execute - Execute a tool\n")

	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func getTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := []MCPTool{
		{
			Name:        "mcp_postgres_list_database",
			Description: "List all databases in PostgreSQL",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "mcp_postgres_list_table",
			Description: "List all tables in PostgreSQL",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "mcp_postgres_read_query",
			Description: "Execute a read-only SQL query on PostgreSQL",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The SQL query to execute",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "mcp_postgres_desc_table",
			Description: "Describe the structure of a PostgreSQL table",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the table to describe",
					},
				},
				"required": []string{"name"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

func executeTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Printf("📞 Executing tool: %s with args: %v\n", req.Name, req.Arguments)

	var response MCPToolResponse

	// Mock responses for testing - replace with real PostgreSQL calls
	switch req.Name {
	case "mcp_postgres_list_database":
		response = MCPToolResponse{
			Content: []MCPContent{
				{Type: "text", Text: "Available databases:\n- postgres (default)\n- billing_db \n- analytics_db\n- test_db"},
			},
		}

	case "mcp_postgres_list_table":
		response = MCPToolResponse{
			Content: []MCPContent{
				{Type: "text", Text: "Available tables:\n- billing_report_unified_provider (Cost data by provider and module)\n- module_mapping (Module to team mapping)\n- users (User information)\n- projects (Project details)"},
			},
		}

	case "mcp_postgres_read_query":
		query, ok := req.Arguments["query"].(string)
		if !ok {
			response = MCPToolResponse{
				Content: []MCPContent{
					{Type: "text", Text: "Error: query parameter must be a string"},
				},
				IsError: true,
			}
		} else {
			// Mock query results based on common patterns
			if contains(query, "eks-zidane") {
				response = MCPToolResponse{
					Content: []MCPContent{
						{Type: "text", Text: "Mock results for eks-zidane module:\n\nMonth | Cost ($)\n------|--------\n2025-01 | 650.97\n2025-02 | 659.24\n2025-03 | 660.80\n2025-04 | 674.41\n2025-05 | 693.78\n2025-06 | 728.34\n2025-07 | 731.49\n2025-08 | 676.95\n2025-09 | 577.39\n\nTrend: +13.4% growth year-over-year"},
					},
				}
			} else if contains(query, "list") || contains(query, "table") {
				response = MCPToolResponse{
					Content: []MCPContent{
						{Type: "text", Text: fmt.Sprintf("Query executed: %s\n\nResults:\ntag_module | team | monthly_cost\n-----------|------|-------------\neks-zidane | core-cloud | 677.23\nrds-prod | data-team | 450.67\nec2-web | frontend | 320.45", query)},
					},
				}
			} else {
				response = MCPToolResponse{
					Content: []MCPContent{
						{Type: "text", Text: fmt.Sprintf("Query executed successfully: %s\n\nSample result set returned (3 rows)", query)},
					},
				}
			}
		}

	case "mcp_postgres_desc_table":
		tableName, ok := req.Arguments["name"].(string)
		if !ok {
			response = MCPToolResponse{
				Content: []MCPContent{
					{Type: "text", Text: "Error: name parameter must be a string"},
				},
				IsError: true,
			}
		} else {
			if tableName == "billing_report_unified_provider" {
				response = MCPToolResponse{
					Content: []MCPContent{
						{Type: "text", Text: "Table: billing_report_unified_provider\n\nColumns:\n- provider (VARCHAR) - Cloud provider (AWS, Azure)\n- tag_module (VARCHAR) - Module identifier\n- start_date (DATE) - Billing period start\n- cost (DECIMAL) - Monthly cost in USD\n- created_at (TIMESTAMP) - Record creation time\n\nIndexes:\n- PRIMARY KEY (provider, tag_module, start_date)\n- INDEX idx_module (tag_module)\n- INDEX idx_date (start_date)"},
					},
				}
			} else if tableName == "module_mapping" {
				response = MCPToolResponse{
					Content: []MCPContent{
						{Type: "text", Text: "Table: module_mapping\n\nColumns:\n- cm_tag_module (VARCHAR) - Module tag\n- cm_tag_unit (VARCHAR) - Team/unit name\n- created_at (TIMESTAMP) - Record creation time\n\nIndexes:\n- PRIMARY KEY (cm_tag_module)\n- INDEX idx_unit (cm_tag_unit)"},
					},
				}
			} else {
				response = MCPToolResponse{
					Content: []MCPContent{
						{Type: "text", Text: fmt.Sprintf("Table: %s\n\nColumns:\n- id (INTEGER) - Primary key\n- name (VARCHAR) - Name field\n- created_at (TIMESTAMP) - Creation timestamp", tableName)},
					},
				}
			}
		}

	default:
		response = MCPToolResponse{
			Content: []MCPContent{
				{Type: "text", Text: fmt.Sprintf("Tool '%s' not implemented", req.Name)},
			},
			IsError: true,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to check if a string contains a substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(findSubstring(s, substr) != -1))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}
