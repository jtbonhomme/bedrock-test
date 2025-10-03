package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/briandowns/spinner"
	"github.com/rs/zerolog/log"

	"github.com/jtbonhomme/bedrock-test/internal/mcp"
)

// claude3 request data type
type Content struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	Name      string                 `json:"name,omitempty"`        // for tool_use
	ID        string                 `json:"id,omitempty"`          // for tool_use
	Input     map[string]interface{} `json:"input,omitempty"`       // for tool_use
	ToolUseID string                 `json:"tool_use_id,omitempty"` // for tool_result
	Content   string                 `json:"content,omitempty"`     // for tool_result
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// Tool definitions for Claude
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

type InputSchema struct {
	Type       string                        `json:"type"`
	Properties map[string]PropertyDefinition `json:"properties"`
	Required   []string                      `json:"required"`
}

type PropertyDefinition struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type RequestBodyClaude3 struct {
	MaxTokensToSample int         `json:"max_tokens"`
	Temperature       float64     `json:"temperature,omitempty"`
	AnthropicVersion  string      `json:"anthropic_version"`
	Messages          []Message   `json:"messages"`
	Tools             []Tool      `json:"tools,omitempty"`
	ToolChoice        *ToolChoice `json:"tool_choice,omitempty"`
}

type ToolChoice struct {
	Type string `json:"type"`           // "auto", "any", or "tool"
	Name string `json:"name,omitempty"` // if type is "tool"
}

// frontend request data type
type FrontEndRequest struct {
	Messages []Message `json:"messages"`
}

// claude3 response data type
type Delta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ResponseClaude3 struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta Delta  `json:"delta"`
}

// ClaudeResponse represents the response from Claude (non-streaming)
type ClaudeResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Content      []Content `json:"content"`
	Model        string    `json:"model"`
	StopReason   string    `json:"stop_reason"`
	StopSequence string    `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Tool use response types
type ToolUse struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

type ToolResult struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

// claude2 data type
type Request struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

type Response struct {
	Completion string `json:"completion"`
}

type Query struct {
	Topic string `json:"topic"`
}

//output, error := bedrockClient.InvokeModel(
//	context.Background(),
//	&bedrockruntime.InvokeModelInput{
//		Body: payloadBytes,
//		//ModelId: aws.String("anthropic.claude-v2"),
//		ModelId:     aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
//		ContentType: aws.String("application/json"),
//		Accept:      aws.String("application/json"),
//	},
//)

// Define PostgreSQL tools that Claude can use
func getPostgreSQLTools() []Tool {
	log.Debug().Msg("getPostgreSQLTools")

	return []Tool{
		{
			Name:        "list_database",
			Description: "List all databases in the PostgreSQL server",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]PropertyDefinition{},
				Required:   []string{},
			},
		},
		{
			Name:        "list_table",
			Description: "List all tables in the PostgreSQL server. If name is provided, list tables with the specified name, otherwise list all tables",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDefinition{
					"name": {
						Type:        "string",
						Description: "If provided, list tables with the specified name. Otherwise, list all tables",
					},
				},
				Required: []string{},
			},
		},
		{
			Name:        "create_table",
			Description: "Create a new table in the PostgreSQL server. Make sure you have added proper comments for each column and the table itself",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDefinition{
					"query": {
						Type:        "string",
						Description: "The SQL query to create the table",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "alter_table",
			Description: "Alter an existing table in the PostgreSQL server. Make sure you have updated comments for each modified column. DO NOT drop table or existing columns!",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDefinition{
					"query": {
						Type:        "string",
						Description: "The SQL query to alter the table",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "desc_table",
			Description: "Describe table structure",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDefinition{
					"name": {
						Type:        "string",
						Description: "Name of the table to describe",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "read_query",
			Description: "Execute a read-only SQL query. Make sure you have knowledge of the table structure before writing WHERE conditions. Call `desc_table` first if necessary",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDefinition{
					"query": {
						Type:        "string",
						Description: "Execute the SQL query and return the result",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "write_query",
			Description: "Execute a write SQL query. Make sure you have knowledge of the table structure before executing the query. Make sure the data types match the columns' definitions",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDefinition{
					"query": {
						Type:        "string",
						Description: "Execute the SQL query and return the result",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "update_query",
			Description: "Execute an update SQL query. Make sure you have knowledge of the table structure before executing the query. Make sure there is always a WHERE condition. Call `desc_table` first if necessary",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDefinition{
					"query": {
						Type:        "string",
						Description: "Execute the SQL query and return the result",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "delete_query",
			Description: "Execute a delete SQL query. Make sure you have knowledge of the table structure before executing the query. Make sure there is always a WHERE condition. Call `desc_table` first if necessary",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDefinition{
					"query": {
						Type:        "string",
						Description: "Execute the SQL query and return the result",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "count_query",
			Description: "Query the number of rows in a certain table",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertyDefinition{
					"name": {
						Type:        "string",
						Description: "Name of the table to count",
					},
				},
				Required: []string{"name"},
			},
		},
	}
}

func CallBedrockClaude3WithMCP(bedrockClient *bedrockruntime.Client, mcpClient mcp.MCPClientInterface, userQuery string) (string, error) {
	log.Debug().Msg("CallBedrockClaude3WithMCP")

	messages := []Message{
		{
			Role: "user",
			Content: []Content{
				{
					Type: "text",
					Text: userQuery,
				},
			},
		},
	}

	maxIterations := 10 // Prevent infinite loops
	fullAnswer := ""

	for iteration := 0; iteration < maxIterations; iteration++ {
		payload := RequestBodyClaude3{
			MaxTokensToSample: 2048,
			AnthropicVersion:  "bedrock-2023-05-31",
			Temperature:       0.1,
			Messages:          messages,
			Tools:             getPostgreSQLTools(),
			ToolChoice:        &ToolChoice{Type: "auto"},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Err(err).Msgf("error marshaling payload %#v ", payload)
			return "", err
		}
		log.Debug().Msgf("payload %s", string(payloadBytes))

		// Use non-streaming for easier tool handling
		input := &bedrockruntime.InvokeModelInput{
			Body:        payloadBytes,
			ModelId:     aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
		}

		output, err := bedrockClient.InvokeModel(context.Background(), input)
		if err != nil {
			log.Err(err).Msg("error invoking InvokeModel with modelId " + *input.ModelId)
			return "", err
		}

		// Parse Claude's response
		var claudeResponse ClaudeResponse
		err = json.Unmarshal(output.Body, &claudeResponse)
		if err != nil {
			log.Err(err).Msg("error unmarshaling Claude response")
			return "", err
		}

		// Process Claude's response content
		toolUsed := false
		var assistantContent []Content

		for _, content := range claudeResponse.Content {
			if content.Type == "text" {
				fullAnswer += content.Text
				assistantContent = append(assistantContent, content)
			} else if content.Type == "tool_use" {
				toolUsed = true
				log.Debug().Msgf("Claude wants to use tool: %s", content.Name)

				// Execute the tool via MCP
				toolResult, err := executeMCPTool(mcpClient, content.Name, content.Input)
				if err != nil {
					log.Err(err).Msgf("error executing tool %s", content.Name)
					toolResult = fmt.Sprintf("Error executing tool: %v", err)
				}

				// Add tool use to assistant content
				assistantContent = append(assistantContent, content)

				// Add tool result to messages
				messages = append(messages, Message{
					Role:    "assistant",
					Content: assistantContent,
				})

				messages = append(messages, Message{
					Role: "user",
					Content: []Content{
						{
							Type:      "tool_result",
							ToolUseID: content.ID,
							Content:   toolResult,
						},
					},
				})
			}
		}

		// If no tools were used, we're done
		if !toolUsed {
			// Add final assistant message
			if len(assistantContent) > 0 {
				messages = append(messages, Message{
					Role:    "assistant",
					Content: assistantContent,
				})
			}
			break
		}
	}

	return fullAnswer, nil
}

// executeMCPTool executes a tool via the MCP client
func executeMCPTool(mcpClient mcp.MCPClientInterface, toolName string, arguments map[string]interface{}) (string, error) {
	log.Debug().Msgf("Executing MCP tool: %s with args: %v", toolName, arguments)

	switch toolName {
	case "list_database":
		result, err := mcpClient.ListDatabases()
		if err != nil {
			return "", err
		}
		return formatMCPResult(result), nil

	case "list_table":
		result, err := mcpClient.ListTables()
		if err != nil {
			return "", err
		}
		return formatMCPResult(result), nil

	case "desc_table":
		tableName, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("desc_table requires 'name' parameter")
		}
		result, err := mcpClient.DescribeTable(tableName)
		if err != nil {
			return "", err
		}
		return formatMCPResult(result), nil

	case "read_query":
		query, ok := arguments["query"].(string)
		if !ok {
			return "", fmt.Errorf("read_query requires 'query' parameter")
		}
		result, err := mcpClient.ExecuteReadQuery(query)
		if err != nil {
			return "", err
		}
		return formatMCPResult(result), nil

	case "count_query":
		tableName, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("count_query requires 'name' parameter")
		}
		// Use ExecuteReadQuery for count
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		result, err := mcpClient.ExecuteReadQuery(query)
		if err != nil {
			return "", err
		}
		return formatMCPResult(result), nil

	default:
		return "", fmt.Errorf("unsupported tool: %s", toolName)
	}
}

// formatMCPResult formats MCP tool response for Claude
func formatMCPResult(response mcp.MCPToolResponse) string {
	if response.IsError {
		return fmt.Sprintf("Error: %v", response.Content)
	}

	var result string
	for _, content := range response.Content {
		if content.Type == "text" {
			result += content.Text + "\n"
		}
	}
	return result
}

func CallBedrockClaude3HaikuChat(bedrockClient *bedrockruntime.Client) (string, error) {
	log.Debug().Msg("CallBedrockClaude3HaikuChat")
	messages := []Message{
		{
			Role: "user",
			Content: []Content{
				{
					Type: "text",
					Text: "Hello Claude, can you explain me hhw French revolution happened?",
				},
			},
		},
	}
	log.Debug().Msgf("messages: %v", messages)

	payload := RequestBodyClaude3{
		MaxTokensToSample: 2048,
		AnthropicVersion:  "bedrock-2023-05-31",
		Temperature:       0.9,
		Messages:          messages,
	}

	payloadBytes, error := json.Marshal(payload)
	if error != nil {
		log.Err(error).Msgf("error marshaling payload %#v ", payload)
		return "", error
	}
	log.Debug().Msgf("payload %s", string(payloadBytes))

	input := &bedrockruntime.InvokeModelWithResponseStreamInput{
		Body:    payloadBytes,
		ModelId: aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
		//ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("*/*"),
	}
	output, error := bedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		input,
	)
	if error != nil {
		log.Err(error).Msg("error invoking InvokeModelWithResponseStream with modelId " + *input.ModelId)
		return "", error
	}

	log.Debug().Msgf("output get stream: %#v", output)

	fullAnswer := ""
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond) // Build our new spinner
	s.Start()                                                   // Start the spinner

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			var resp ResponseClaude3
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				log.Err(err).Msg("error decoding response")
				return "", err
			}

			fullAnswer += resp.Delta.Text

		case *types.UnknownUnionMember:
			log.Debug().Msgf("unknown tag: %v", v.Tag)

		default:
			log.Debug().Msg("union is nil or unknown type")
		}
	}
	s.Stop()

	return fullAnswer, nil
}
