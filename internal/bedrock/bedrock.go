package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
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
	Type string `json:"type"`
	Text string `json:"text"`
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

func CallBedrockClaude3WithMCP(bedrockClient *bedrockruntime.Client, mcpClient *mcp.MCPClient, userQuery string) (string, error) {
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

	payload := RequestBodyClaude3{
		MaxTokensToSample: 2048,
		AnthropicVersion:  "bedrock-2023-05-31",
		Temperature:       0.1,
		Messages:          messages,
		Tools:             getPostgreSQLTools(),
		ToolChoice:        &ToolChoice{Type: "auto"},
	}

	payloadBytes, error := json.Marshal(payload)
	if error != nil {
		log.Err(error).Msgf("error marshaling payload %#v ", payload)
		return "", error
	}
	log.Debug().Msgf("payload %s", string(payloadBytes))

	input := &bedrockruntime.InvokeModelWithResponseStreamInput{
		Body:        payloadBytes,
		ModelId:     aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
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

	fullAnswer := ""
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var resp ResponseClaude3
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				log.Err(err).Msg("error decoding response")
				return "", err
			}

			// Check if Claude wants to use a tool
			if resp.Delta.Type == "tool_use" {
				// Handle tool use (will be implemented in next step)
				log.Debug().Msg("Claude wants to use a tool")
			} else {
				fullAnswer += resp.Delta.Text
			}

		case *types.UnknownUnionMember:
			log.Debug().Msgf("unknown tag: %v", v.Tag)

		default:
			log.Debug().Msg("union is nil or unknown type")
		}
	}
	s.Stop()

	return fullAnswer, nil
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
