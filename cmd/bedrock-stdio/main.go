package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"

	"github.com/jtbonhomme/bedrock-test/internal/bedrock"
	"github.com/jtbonhomme/bedrock-test/internal/mcp"
)

func checkAWSConfig(cfg aws.Config) {
	log.Debug().Msgf("AWS Region: %s", cfg.Region)
	log.Debug().Msgf("AWS Credentials Provider: %T", cfg.Credentials)

	// Vérifier les variables d'environnement importantes
	if region := os.Getenv("AWS_REGION"); region != "" {
		log.Debug().Msgf("AWS_REGION env var: %s", region)
	}
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		log.Debug().Msgf("AWS_PROFILE env var: %s", profile)
	}
}

func main() {
	var debug bool
	var mcpURL string
	var mcpDSN string
	var query string

	flag.BoolVar(&debug, "d", false, "enable debug mode")
	flag.StringVar(&mcpURL, "mcp", "http://localhost:8080", "MCP server URL (HTTP)")
	flag.StringVar(&mcpDSN, "dsn", "", "MCP DSN for go-mcp-postgres (stdio)")
	flag.StringVar(&query, "q", "List all tables in the database and describe their structure", "Query to send to Claude")
	flag.Parse()

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Info().Msg("run bedrock test program with MCP integration")

	// Create appropriate MCP client based on flags
	var mcpClient mcp.MCPClientInterface
	var err error

	if mcpDSN != "" {
		// Use stdio client with go-mcp-postgres
		log.Info().Msgf("Using stdio MCP client with DSN: %s", maskDSN(mcpDSN))

		config := mcp.MCPClientConfig{
			Executable: "go-mcp-postgres",
			DSN:        mcpDSN,
		}

		mcpClient, err = mcp.NewMCPClientAuto(config)
		if err != nil {
			log.Error().Msgf("Failed to create stdio MCP client: %v", err)
			log.Info().Msg("Make sure go-mcp-postgres is installed and PostgreSQL is accessible")
			os.Exit(1)
		}

	} else {
		// Use HTTP client (fallback or explicit)
		log.Info().Msgf("Using HTTP MCP client at: %s", mcpURL)

		config := mcp.MCPClientConfig{
			BaseURL: mcpURL,
		}

		mcpClient, err = mcp.NewMCPClientAuto(config)
		if err != nil {
			log.Error().Msgf("Failed to create HTTP MCP client: %v", err)
			mcpClient = nil // Continue without MCP
		}
	}

	// Ensure cleanup
	if mcpClient != nil {
		defer func() {
			if err := mcpClient.Close(); err != nil {
				log.Warn().Msgf("Error closing MCP client: %v", err)
			}
		}()
	}

	// Test MCP connection
	//if mcpClient != nil {
	//	log.Info().Msg("Testing MCP connection...")
	//	databases, err := mcpClient.ListDatabases()
	//	if err != nil {
	//		log.Warn().Msgf("Failed to connect to MCP server: %v", err)
	//		log.Info().Msg("Continuing without MCP integration...")
	//		mcpClient = nil
	//	} else {
	//		log.Info().Msgf("MCP server connected successfully. Databases: %v", len(databases.Content))
	//	}
	//}

	// load aws credentials from profile demo using config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("eu-west-1"),
	)
	if err != nil {
		log.Panic().Msgf("failed to load aws config: %v", err)
	}

	log.Debug().Msgf("aws config: %+v\n", awsCfg)
	checkAWSConfig(awsCfg)

	// create bedrock runtime client
	bedrockClient := bedrockruntime.NewFromConfig(awsCfg)

	// Use appropriate function based on MCP availability
	var bedrockAnswer string
	if mcpClient != nil {
		// Use MCP-integrated function
		bedrockAnswer, err = bedrock.CallBedrockClaude3WithMCP(bedrockClient, &mcp.MCPClient{}, query)
		if err != nil {
			log.Err(err).Msg("error calling CallBedrockClaude3WithMCP")

			// Fallback to original function
			log.Info().Msg("Falling back to original Bedrock function...")
			bedrockAnswer, err = bedrock.CallBedrockClaude3HaikuChat(bedrockClient)
			if err != nil {
				log.Err(err).Msg("error calling CallBedrockClaude3HaikuChat")
				return
			}
		}
	} else {
		// Use original function without MCP
		log.Info().Msg("Using Bedrock without MCP integration...")
		bedrockAnswer, err = bedrock.CallBedrockClaude3HaikuChat(bedrockClient)
		if err != nil {
			log.Err(err).Msg("error calling CallBedrockClaude3HaikuChat")
			return
		}
	}

	log.Info().Msgf("answer is \"%s\"", bedrockAnswer)
}

// maskDSN masks the password in a DSN for logging
func maskDSN(dsn string) string {
	// Simple password masking: postgresql://user:password@host:port/db
	// Replace password with ***
	if idx := len("postgresql://"); idx < len(dsn) {
		if colonIdx := findChar(dsn[idx:], ':'); colonIdx != -1 {
			colonIdx += idx
			if atIdx := findChar(dsn[colonIdx:], '@'); atIdx != -1 {
				atIdx += colonIdx
				return dsn[:colonIdx+1] + "***" + dsn[atIdx:]
			}
		}
	}
	return dsn
}

func findChar(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
