package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// bedrock client
var BedrockClient *bedrockruntime.Client

func checkAWSConfig(cfg aws.Config) {
	log.Info().Msgf("AWS Region: %s", cfg.Region)
	log.Info().Msgf("AWS Credentials Provider: %T", cfg.Credentials)

	// Vérifier les variables d'environnement importantes
	if region := os.Getenv("AWS_REGION"); region != "" {
		log.Info().Msgf("AWS_REGION env var: %s", region)
	}
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		log.Info().Msgf("AWS_PROFILE env var: %s", profile)
	}
}

func main() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msg("run bedrock test program")

	// load aws credentials from profile demo using config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("eu-west-1"),
	)
	if err != nil {
		log.Panic().Msgf("failed to load aws config: %v", err)
	}

	log.Info().Msgf("aws config: %+v\n", awsCfg)
	checkAWSConfig(awsCfg)

	// create bedrock runtime client
	BedrockClient = bedrockruntime.NewFromConfig(awsCfg)

	CallBedrockClaude3HaikuChat()

	/*
		// create handler multiplexer
		mux := http.NewServeMux()
		// hello message
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// w.Write([]byte("Hello"))
			http.ServeFile(w, r, "./static/claude-haiku.html")
		})

		// uncommen for opensearch client
		// handle query to aoss
		// mux.HandleFunc("/query", HandleAOSSQuery)

		// handle chat with bedrock
		mux.HandleFunc("/bedrock-stream", HandleBedrockClaude2Chat)

		// bedrock chat frontend
		mux.HandleFunc("/claude2", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./static/claude2.html")
		})

		// handle chat with bedrock
		mux.HandleFunc("/bedrock-haiku", HandleBedrockClaude3HaikuChat)

		// bedrock chat frontend
		mux.HandleFunc("/haiku", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./static/claude-haiku.html")
		})

		// create a http server using http
		server := http.Server{
			Addr:           ":3000",
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		log.Info().Msg("serve on :3000")
		server.ListenAndServe()*/
}

func CallBedrockClaude3HaikuChat() (string, error) {
	log.Info().Msg("CallBedrockClaude3HaikuChat")
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
	log.Info().Msgf("messages: %v", messages)

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

	//output, error := BedrockClient.InvokeModel(
	//	context.Background(),
	//	&bedrockruntime.InvokeModelInput{
	//		Body: payloadBytes,
	//		//ModelId: aws.String("anthropic.claude-v2"),
	//		ModelId:     aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
	//		ContentType: aws.String("application/json"),
	//		Accept:      aws.String("application/json"),
	//	},
	//)

	input := &bedrockruntime.InvokeModelWithResponseStreamInput{
		Body:    payloadBytes,
		ModelId: aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
		//ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("*/*"),
	}
	output, error := BedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		input,
	)
	if error != nil {
		log.Err(error).Msg("error invoking InvokeModelWithResponseStream with modelId " + *input.ModelId)
		return "", error
	}

	log.Info().Msgf("output get stream: %#v", output)

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			//log.Debug().Msg("payload", string(v.Value.Bytes))

			var resp ResponseClaude3
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				log.Err(err).Msg("error decoding response")
				return "", err
			}

			log.Debug().Msg(resp.Delta.Text)

		case *types.UnknownUnionMember:
			log.Debug().Msgf("unknown tag: %v", v.Tag)

		default:
			log.Debug().Msg("union is nil or unknown type")
		}
	}

	return "", nil
}
