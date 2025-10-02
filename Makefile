.PHONY: build run-server run-client test clean

# Build all binaries
build:
	go mod tidy
	go build -o bedrock-client .
	go build -o mcp-server ./cmd/mcp-server

# Run MCP server (in background for testing)
run-server: build
	./mcp-server &

# Run Bedrock client with MCP integration
run-client: build
	./bedrock-client -d -mcp http://localhost:8080 -q "Show me all tables in the database and their structure"

# Run both (server first, then client)
test: build
	./mcp-server &
	sleep 2
	./bedrock-client -d -mcp http://localhost:8080 -q "List all tables and show the cost analysis for eks-zidane module"
	pkill mcp-server || true

# Test without MCP (original functionality)
test-bedrock-only: build
	./bedrock-client -d -q "Explain the French Revolution"

# Clean build artifacts
clean:
	rm -f bedrock-client mcp-server
	pkill mcp-server || true

# Development commands
dev-server:
	go run ./cmd/mcp-server

dev-client:
	go run . -d -mcp http://localhost:8080 -q "Show me the database structure"

# Test MCP server health
test-mcp:
	curl -s http://localhost:8080/health || echo "MCP server not running"
	curl -s http://localhost:8080/tools | jq '.' || echo "Failed to get tools"
