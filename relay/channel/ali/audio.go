package ali

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// Ali (DashScope) speech-to-text models such as qwen3-asr-flash do NOT expose a
// Whisper-style /audio/transcriptions endpoint. They are served through the
// OpenAI-compatible chat/completions endpoint, with the audio supplied as an
// input_audio data URI and the recognized text returned in
// choices[0].message.content.
//
// To let clients keep using the standard /v1/audio/transcriptions API, we
// translate the incoming multipart transcription request into a chat request
// here, and translate the chat response back into a transcription response in
// aliSTTHandler.

type aliASRInputAudio struct {
	Data string `json:"data"` // data:<mime>;base64,<...> or a public URL
}

type aliASRContent struct {
	Type       string            `json:"type"`
	InputAudio *aliASRInputAudio `json:"input_audio,omitempty"`
	Text       string            `json:"text,omitempty"`
}

type aliASRMessage struct {
	Role    string          `json:"role"`
	Content []aliASRContent `json:"content"`
}

type aliASROptions struct {
	Language string `json:"language,omitempty"`
}

type aliASRExtraBody struct {
	ASROptions *aliASROptions `json:"asr_options,omitempty"`
}

type aliASRChatRequest struct {
	Model     string           `json:"model"`
	Messages  []aliASRMessage  `json:"messages"`
	Stream    bool             `json:"stream"`
	ExtraBody *aliASRExtraBody `json:"extra_body,omitempty"`
}

func convertAudioToAliChatRequest(c *gin.Context, request dto.AudioRequest) (io.Reader, error) {
	formData, err := common.ParseMultipartFormReusable(c)
	if err != nil {
		return nil, fmt.Errorf("error parsing multipart form: %w", err)
	}

	fileHeaders := formData.File["file"]
	if len(fileHeaders) == 0 {
		return nil, errors.New("file is required")
	}
	fileHeader := fileHeaders[0]

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening audio file: %w", err)
	}
	defer file.Close()

	audioBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading audio file: %w", err)
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fileHeader.Filename))
	}
	if mimeType == "" {
		mimeType = "audio/mpeg"
	}
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(audioBytes))

	chatReq := aliASRChatRequest{
		Model: request.Model,
		Messages: []aliASRMessage{
			{
				Role: "user",
				Content: []aliASRContent{
					{Type: "input_audio", InputAudio: &aliASRInputAudio{Data: dataURI}},
				},
			},
		},
		Stream: false,
	}

	if langs := formData.Value["language"]; len(langs) > 0 {
		if lang := strings.TrimSpace(langs[0]); lang != "" {
			chatReq.ExtraBody = &aliASRExtraBody{ASROptions: &aliASROptions{Language: lang}}
		}
	}

	jsonData, err := common.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("error marshalling ali asr request: %w", err)
	}
	return bytes.NewReader(jsonData), nil
}

// aliSTTHandler converts a DashScope chat/completions response back into the
// OpenAI transcription response expected by /v1/audio/transcriptions clients.
func aliSTTHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, responseFormat string) (*types.NewAPIError, *dto.Usage) {
	defer service.CloseResponseBodyGracefully(resp)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.NewOpenAIError(err, types.ErrorCodeReadResponseBodyFailed, http.StatusInternalServerError), nil
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage *dto.Usage `json:"usage"`
		AliError
	}
	if err := common.Unmarshal(responseBody, &chatResp); err != nil {
		return types.NewError(fmt.Errorf("unmarshal ali asr response failed: %w", err), types.ErrorCodeBadResponseBody), nil
	}
	if chatResp.Code != "" {
		return types.NewOpenAIError(errors.New(chatResp.Message), types.ErrorCodeBadResponse, http.StatusInternalServerError), nil
	}

	text := ""
	if len(chatResp.Choices) > 0 {
		text = chatResp.Choices[0].Message.Content
	}

	var out []byte
	if responseFormat == "text" {
		out = []byte(text)
		resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		// json / verbose_json / others fall back to the minimal {"text": ...} body
		out, err = common.Marshal(dto.AudioResponse{Text: text})
		if err != nil {
			return types.NewError(err, types.ErrorCodeBadResponseBody), nil
		}
		resp.Header.Set("Content-Type", "application/json")
	}

	usage := &dto.Usage{}
	if chatResp.Usage != nil {
		usage = chatResp.Usage
		if usage.PromptTokens == 0 {
			usage.PromptTokens = usage.InputTokens
		}
		if usage.CompletionTokens == 0 {
			usage.CompletionTokens = usage.OutputTokens
		}
	}
	if usage.TotalTokens == 0 {
		usage.PromptTokens = info.GetEstimatePromptTokens()
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	service.IOCopyBytesGracefully(c, resp, out)
	return nil, usage
}
