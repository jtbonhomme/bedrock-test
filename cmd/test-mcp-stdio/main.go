package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jtbonhomme/bedrock-test/internal/mcp"
)

func main() {
	var (
		executable = flag.String("exe", "go-mcp-postgres", "Path to go-mcp-postgres executable")
		dsn        = flag.String("dsn", "", "PostgreSQL DSN (required)")
		query      = flag.String("query", "", "SQL query to execute")
		debug      = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	if *dsn == "" {
		fmt.Println("❌ DSN is required")
		fmt.Println("Usage examples:")
		fmt.Println("  # Basic connection")
		fmt.Printf("  %s -dsn 'postgresql://user:pass@host:port/db'\n", os.Args[0])
		fmt.Println("  # With SSL")
		fmt.Printf("  %s -dsn 'postgresql://user:pass@host:port/db?sslmode=require'\n", os.Args[0])
		fmt.Println("  # With query")
		fmt.Printf("  %s -dsn 'postgresql://...' -query 'SELECT COUNT(*) FROM billing_report_unified_provider'\n", os.Args[0])
		os.Exit(1)
	}

	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	fmt.Printf("🚀 Connecting to PostgreSQL via go-mcp-postgres\n")
	fmt.Printf("   Executable: %s\n", *executable)
	fmt.Printf("   DSN: %s\n", maskPassword(*dsn))

	// Create stdio MCP client
	client := mcp.NewStdioMCPClient(*executable, *dsn)
	defer client.Close()

	// Connect
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect to MCP server: %v", err)
	}

	fmt.Println("✅ Connected to MCP server")

	// Test basic functionality
	if err := testBasicFunctionality(client); err != nil {
		log.Fatalf("Basic functionality test failed: %v", err)
	}

	// Execute custom query if provided
	if *query != "" {
		if err := executeCustomQuery(client, *query); err != nil {
			log.Printf("Custom query failed: %v", err)
		}
	}

	fmt.Println("🎉 MCP stdio client test completed successfully")
}

func testBasicFunctionality(client *mcp.StdioMCPClient) error {
	fmt.Println("\n📋 Testing basic functionality...")

	// Test 1: List databases
	fmt.Print("   Testing list databases... ")
	databases, err := client.ListDatabases()
	if err != nil {
		return fmt.Errorf("list databases failed: %w", err)
	}
	fmt.Printf("✅ Found %d database(s)\n", len(databases.Content))

	// Test 2: List tables
	fmt.Print("   Testing list tables... ")
	tables, err := client.ListTables()
	if err != nil {
		return fmt.Errorf("list tables failed: %w", err)
	}
	fmt.Printf("✅ Found %d table(s)\n", len(tables.Content))

	// Show some results
	if len(tables.Content) > 0 {
		fmt.Printf("   📊 Tables preview: %s\n",
			truncateString(tables.Content[0].Text, 100))
	}

	return nil
}

func executeCustomQuery(client *mcp.StdioMCPClient, query string) error {
	fmt.Printf("\n🔍 Executing custom query: %s\n", truncateString(query, 80))

	result, err := client.ExecuteReadQuery(query)
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	if result.IsError {
		return fmt.Errorf("query returned error: %v", result.Content)
	}

	fmt.Println("✅ Query executed successfully")
	for i, content := range result.Content {
		fmt.Printf("   Result %d: %s\n", i+1, truncateString(content.Text, 200))
	}

	return nil
}

func maskPassword(dsn string) string {
	// Simple password masking for display
	return dsn // TODO: implement proper password masking
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
