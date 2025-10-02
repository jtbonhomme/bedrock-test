package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
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

type HelloHandler struct{}

type Query struct {
	Topic string `json:"topic"`
}

func HandleBedrockClaude2Chat(w http.ResponseWriter, r *http.Request) {
	const claudePromptFormat = "\n\nHuman: %s\n\nAssistant:"

	var query Query
	var message string

	log.Info().Msg("HandleBedrockClaude2Chat: parsing message from request")

	// parse mesage from request
	error := json.NewDecoder(r.Body).Decode(&query)
	if error != nil {
		message = "how to learn japanese as quick as possible?"
		log.Panic().Msgf("error/ %s", error.Error())
	}

	message = query.Topic
	log.Info().Msgf("message: %s", message)

	prompt := "" + fmt.Sprintf(claudePromptFormat, message)
	log.Info().Msgf("prompt: %s", prompt)

	payload := Request{
		Prompt:            prompt,
		MaxTokensToSample: 2048,
	}

	payloadBytes, error := json.Marshal(payload)
	if error != nil {
		fmt.Fprintf(w, "ERROR")
		// return "", error
	}
	log.Info().Msgf("payload: %s", string(payloadBytes))

	output, error := BedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		&bedrockruntime.InvokeModelWithResponseStreamInput{
			Body:        payloadBytes,
			ModelId:     aws.String("anthropic.claude-v2"),
			ContentType: aws.String("application/json"),
		},
	)
	if error != nil {
		fmt.Fprintf(w, "ERROR")
		// return "", error
	}
	log.Debug().Msgf("output get stream %#v", output.GetStream())

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			//log.Debug().Msg("payload", string(v.Value.Bytes))

			var resp Response
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				fmt.Fprintf(w, "ERROR")
				// return "", err
			}

			log.Debug().Msg(resp.Completion)

			fmt.Fprintf(w, resp.Completion)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			} else {
				log.Warn().Msg("Damn, no flush")
			}

		case *types.UnknownUnionMember:
			log.Warn().Msgf("unknown tag: %s", v.Tag)

		default:
			log.Warn().Msg("union is nil or unknown type")
		}
	}
}

func HandleBedrockClaude3HaikuChat(w http.ResponseWriter, r *http.Request) {
	// list of messages sent from frontend client
	var request FrontEndRequest

	log.Info().Msg("HandleBedrockClaude3HaikuChat: parsing message from request")

	// parse mesage from request
	error := json.NewDecoder(r.Body).Decode(&request)
	if error != nil {
		log.Panic().Msgf("error/ %s", error.Error())
	}

	//messages := request.Messages
	messages := []Message{
		{
			Role: "user",
			Content: []Content{
				{
					Type: "text",
					Text: "Hell Claude",
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
		fmt.Fprintf(w, "ERROR")
		// return "", error
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
		fmt.Fprintf(w, error.Error())
		// return "", error
	}

	log.Info().Msgf("output get stream: %#v", output)

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			//log.Debug().Msg("payload", string(v.Value.Bytes))

			var resp ResponseClaude3
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				fmt.Fprintf(w, "ERROR")
				// return "", err
			}

			log.Debug().Msg(resp.Delta.Text)

			fmt.Fprintf(w, resp.Delta.Text)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			} else {
				log.Debug().Msg("Damn, no flush")
			}

		case *types.UnknownUnionMember:
			log.Debug().Msgf("unknown tag: %v", v.Tag)

		default:
			log.Debug().Msg("union is nil or unknown type")
		}
	}

}

func HandleHaikuImageAnalyzer(w http.ResponseWriter, r *http.Request) {

	// data type request
	type Message struct {
		Role    string        `json:"role"`
		Content []interface{} `json:"content"`
	}

	type Request struct {
		Messages []Message `json:"messages"`
	}

	type RequestBodyClaude3 struct {
		MaxTokensToSample int       `json:"max_tokens"`
		Temperature       float64   `json:"temperature,omitempty"`
		AnthropicVersion  string    `json:"anthropic_version"`
		Messages          []Message `json:"messages"`
	}

	// data type response
	// type ResponseContent struct {
	// 	Type string `json:"type"`
	// 	Text string `json:"text"`
	// }

	// type Response struct {
	// 	Content []ResponseContent `json:"content"`
	// }

	// parse request
	var request Request
	error := json.NewDecoder(r.Body).Decode(&request)

	if error != nil {
		log.Panic().Msgf("error/ %s", error.Error())
	}

	// log.Debug().Msg(request)

	// payload for bedrock claude3 haikue
	messages := request.Messages

	payload := RequestBodyClaude3{
		MaxTokensToSample: 2048,
		AnthropicVersion:  "bedrock-2023-05-31",
		Temperature:       0.9,
		Messages:          messages,
	}

	// convert payload struct to bytes
	payloadBytes, error := json.Marshal(payload)
	if error != nil {
		log.Err(error).Msg("error")
		fmt.Fprintf(w, "ERROR")
		// return "", error
	}

	// log.Debug().Msg("invoke bedrock ...")

	// invoke bedrock claude3 haiku
	output, error := BedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		&bedrockruntime.InvokeModelWithResponseStreamInput{
			Body:        payloadBytes,
			ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
		},
	)

	// response
	// var response Response
	// json.NewDecoder(bytes.NewReader(output.Body)).Decode(&response)
	// log.Debug().Msg(response)

	if error != nil {
		log.Err(error).Msg("error")
		fmt.Fprintf(w, "ERROR")
		// return "", error
	}

	// stream result to client
	for event := range output.GetStream().Events() {

		// log.Debug().Msg(event)

		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			// log.Debug().Msg("payload", string(v.Value.Bytes))

			var resp ResponseClaude3
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				fmt.Fprintf(w, "ERROR")
				// return "", err
			}

			// log.Debug().Msg(resp.Delta.Text)

			fmt.Fprintf(w, resp.Delta.Text)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			} else {
				log.Warn().Msg("Damn, no flush")
			}

		case *types.UnknownUnionMember:
			log.Debug().Msgf("unknown tag: %s", v.Tag)

		default:
			log.Warn().Msg("union is nil or unknown type")
		}
	}
}

func TestHaiku() {

	log.Debug().Msg("Hello")

	payload := RequestBodyClaude3{
		MaxTokensToSample: 2048,
		AnthropicVersion:  "bedrock-2023-05-31",
		Temperature:       0.8,
		Messages: []Message{{
			Role: "user",
			Content: []Content{{
				Type: "text",
				Text: "How to cook chicken soup?",
			}},
		}, {
			Role: "assistant",
			Content: []Content{{
				Type: "text",
				Text: `Here is a basic recipe for cooking chicken soup`}},
		},
			{
				Role: "user",
				Content: []Content{{
					Type: "text",
					Text: "How to customize it for 3 years old girl?",
				}},
			},
		},
	}

	payloadBytes, error := json.Marshal(payload)

	if error != nil {
		log.Err(error).Msg("error")
	}

	output, error := BedrockClient.InvokeModelWithResponseStream(
		context.Background(),
		&bedrockruntime.InvokeModelWithResponseStreamInput{
			Body:        payloadBytes,
			ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
		},
	)

	if error != nil {
		log.Err(error).Msg("error")
	}

	// log.Debug().Msg(output)

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			// log.Debug().Msg("payload", string(v.Value.Bytes))

			// var resp map[string]interface{}
			var resp ResponseClaude3

			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				log.Err(err).Msg("error")
			}

			// log.Debug().Msg(resp.Delta.Text)

		case *types.UnknownUnionMember:
			log.Debug().Msgf("unknown tag:", v.Tag)

		default:
			log.Debug().Msg("union is nil or unknown type")
		}
	}

}
