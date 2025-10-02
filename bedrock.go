package main

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

type RequestBodyClaude3 struct {
	MaxTokensToSample int       `json:"max_tokens"`
	Temperature       float64   `json:"temperature,omitempty"`
	AnthropicVersion  string    `json:"anthropic_version"`
	Messages          []Message `json:"messages"`
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
