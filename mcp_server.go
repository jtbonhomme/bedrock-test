package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

// Simple MCP server for testing
func runMCPServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Get available tools
	mux.HandleFunc("/tools", getTools)

	// Execute tool
	mux.HandleFunc("/tools/execute", executeTool)

	fmt.Printf("MCP Server starting on port %s\n", port)
	log.Fatal().Err(http.ListenAndServe(":"+port, mux))
}

func getTools(w http.ResponseWriter, r *http.Request) {
	tools := []MCPTool{
		{
			Name:        "mcp_postgres_list_database",
			Description: "List all databases",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "mcp_postgres_list_table",
			Description: "List all tables",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "mcp_postgres_read_query",
			Description: "Execute read-only query",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL query to execute",
					},
				},
				"required": []string{"query"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

func executeTool(w http.ResponseWriter, r *http.Request) {
	var req MCPToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response MCPToolResponse

	// Mock responses for testing
	switch req.Name {
	case "mcp_postgres_list_database":
		response = MCPToolResponse{
			Content: []MCPContent{
				{Type: "text", Text: "postgres, billing_db, analytics_db"},
			},
		}
	case "mcp_postgres_list_table":
		response = MCPToolResponse{
			Content: []MCPContent{
				{Type: "text", Text: "billing_report_unified_provider, module_mapping, users, projects"},
			},
		}
	case "mcp_postgres_read_query":
		query := req.Arguments["query"].(string)
		response = MCPToolResponse{
			Content: []MCPContent{
				{Type: "text", Text: fmt.Sprintf("Mock result for query: %s\n\nSample data:\nmodule1,team1,500.00\nmodule2,team2,750.00", query)},
			},
		}
	default:
		response = MCPToolResponse{
			Content: []MCPContent{
				{Type: "text", Text: "Tool not implemented"},
			},
			IsError: true,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
