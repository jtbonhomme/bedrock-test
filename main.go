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
	var query string

	flag.BoolVar(&debug, "d", false, "enable debug mode")
	flag.StringVar(&mcpURL, "mcp", "http://localhost:8080", "MCP server URL")
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

	// Initialize MCP client
	mcpClient := NewMCPClient(mcpURL)
	log.Info().Msgf("Connecting to MCP server at: %s", mcpURL)

	// Test MCP connection
	databases, err := mcpClient.ListDatabases()
	if err != nil {
		log.Warn().Msgf("Failed to connect to MCP server: %v", err)
		log.Info().Msg("Continuing without MCP integration...")
	} else {
		log.Info().Msgf("MCP server connected successfully. Databases: %v", databases)
	}

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

	// Use new MCP-integrated function
	bedrockAnswer, err := CallBedrockClaude3WithMCP(bedrockClient, mcpClient, query)
	if err != nil {
		log.Err(err).Msg("error calling CallBedrockClaude3WithMCP")

		// Fallback to original function
		log.Info().Msg("Falling back to original Bedrock function...")
		bedrockAnswer, err = CallBedrockClaude3HaikuChat(bedrockClient)
		if err != nil {
			log.Err(err).Msg("error calling CallBedrockClaude3HaikuChat")
			return
		}
	}

	log.Info().Msgf("answer is \"%s\"", bedrockAnswer)
}
