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
	flag.BoolVar(&debug, "d", false, "enable debug mode")
	flag.Parse()

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Info().Msg("run bedrock test program")

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

	bedrockAnswer, err := CallBedrockClaude3HaikuChat(bedrockClient)
	if err != nil {
		log.Err(err).Msg("error calling CallBedrockClaude3HaikuChat")
		return
	}

	log.Info().Msgf("answer is \"%s\"", bedrockAnswer)
}
