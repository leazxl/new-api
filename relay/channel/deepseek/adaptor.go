package deepseek

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/claude"
	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/service/openaicompat"
	"github.com/QuantumNous/new-api/setting/reasoning"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

type Adaptor struct {
}

func (a *Adaptor) ConvertGeminiRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeminiChatRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, req *dto.ClaudeRequest) (any, error) {
	adaptor := claude.Adaptor{}
	convertedRequest, err := adaptor.ConvertClaudeRequest(c, info, req)
	if err != nil {
		return nil, err
	}
	claudeRequest, ok := convertedRequest.(*dto.ClaudeRequest)
	if !ok {
		return convertedRequest, nil
	}
	if err := applyDeepSeekV4ClaudeThinkingSuffix(info, claudeRequest); err != nil {
		return nil, err
	}
	return claudeRequest, nil
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	fimBaseUrl := info.ChannelBaseUrl
	switch info.RelayFormat {
	case types.RelayFormatClaude:
		return fmt.Sprintf("%s/anthropic/v1/messages", info.ChannelBaseUrl), nil
	case types.RelayFormatOpenAIResponses, types.RelayFormatOpenAIResponsesCompaction:
		return fmt.Sprintf("%s/v1/chat/completions", info.ChannelBaseUrl), nil
	default:
		if !strings.HasSuffix(info.ChannelBaseUrl, "/beta") {
			fimBaseUrl += "/beta"
		}
		switch info.RelayMode {
		case constant.RelayModeCompletions:
			return fmt.Sprintf("%s/completions", fimBaseUrl), nil
		default:
			return fmt.Sprintf("%s/v1/chat/completions", info.ChannelBaseUrl), nil
		}
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Set("Authorization", "Bearer "+info.ApiKey)
	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	if err := applyDeepSeekV4OpenAIThinkingSuffix(info, request); err != nil {
		return nil, err
	}

	return request, nil
}

func applyDeepSeekV4OpenAIThinkingSuffix(info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) error {
	modelName := request.Model
	if info != nil && info.ChannelMeta != nil && info.UpstreamModelName != "" {
		modelName = info.UpstreamModelName
	}
	baseModel, thinkingType, effort, ok := reasoning.ParseDeepSeekV4ThinkingSuffix(modelName)
	if !ok {
		return nil
	}
	thinking, err := common.Marshal(map[string]string{
		"type": thinkingType,
	})
	if err != nil {
		return fmt.Errorf("error marshalling thinking: %w", err)
	}
	request.Model = baseModel
	request.THINKING = thinking
	request.ReasoningEffort = effort
	if info != nil {
		if info.ChannelMeta != nil {
			info.UpstreamModelName = baseModel
		}
		info.ReasoningEffort = effort
	}
	return nil
}

func applyDeepSeekV4ClaudeThinkingSuffix(info *relaycommon.RelayInfo, request *dto.ClaudeRequest) error {
	modelName := request.Model
	if info != nil && info.ChannelMeta != nil && info.UpstreamModelName != "" {
		modelName = info.UpstreamModelName
	}
	baseModel, thinkingType, effort, ok := reasoning.ParseDeepSeekV4ThinkingSuffix(modelName)
	if !ok {
		return nil
	}
	request.Model = baseModel
	request.Thinking = &dto.Thinking{Type: thinkingType}
	if effort == "" {
		request.OutputConfig = nil
	} else {
		outputConfig, err := common.Marshal(map[string]string{
			"effort": effort,
		})
		if err != nil {
			return fmt.Errorf("error marshalling output_config: %w", err)
		}
		request.OutputConfig = outputConfig
	}
	if info != nil {
		if info.ChannelMeta != nil {
			info.UpstreamModelName = baseModel
		}
		info.ReasoningEffort = effort
	}
	return nil
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	chatReq, err := openaicompat.ResponsesRequestToChatCompletionsRequest(&request)
	if err != nil {
		return nil, err
	}

	if err := applyDeepSeekV4OpenAIThinkingSuffix(info, chatReq); err != nil {
		return nil, err
	}

	// Switch to chat completions mode for the rest of the pipeline.
	info.RelayMode = constant.RelayModeChatCompletions
	info.AppendRequestConversion(types.RelayFormatOpenAI)

	return chatReq, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	switch info.RelayFormat {
	case types.RelayFormatOpenAIResponses, types.RelayFormatOpenAIResponsesCompaction:
		if info.IsStream {
			return responsesViaChatStreamDoResponse(c, resp, info)
		}
		return responsesViaChatDoResponse(c, resp, info)
	case types.RelayFormatClaude:
		adaptor := claude.Adaptor{}
		return adaptor.DoResponse(c, resp, info)
	default:
		adaptor := openai.Adaptor{}
		return adaptor.DoResponse(c, resp, info)
	}
}

func responsesViaChatDoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeReadResponseBodyFailed, http.StatusInternalServerError)
	}

	var chatResp dto.OpenAITextResponse
	if err := common.Unmarshal(responseBody, &chatResp); err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	if oaiErr := chatResp.GetOpenAIError(); oaiErr != nil {
		return nil, types.WithOpenAIError(*oaiErr, resp.StatusCode)
	}

	responsesResp := openaicompat.ChatCompletionsResponseToResponsesResponse(&chatResp, info.UpstreamModelName, nil)
	if responsesResp == nil {
		return nil, types.NewOpenAIError(fmt.Errorf("failed to convert chat response to responses"), types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	jsonData, err := common.Marshal(responsesResp)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeJsonMarshalFailed, http.StatusInternalServerError)
	}

	service.IOCopyBytesGracefully(c, resp, jsonData)

	usage := &dto.Usage{}
	if responsesResp.Usage != nil {
		*usage = *responsesResp.Usage
	}
	return usage, nil
}

func responsesViaChatStreamDoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	responseId := info.RequestId
	model := info.UpstreamModelName
	msgItemId := "msg_" + responseId
	var (
		usage          = &dto.Usage{}
		streamErr      *types.NewAPIError
		sentPreamble   bool
		sentTextItem   bool
		sentReasonItem bool
	)

	sendPreamble := func(sr *helper.StreamResult) {
		if sentPreamble {
			return
		}
		sentPreamble = true

		// response.created
		createdResp := dto.OpenAIResponsesResponse{
			ID:     responseId,
			Object: "response",
			Model:  model,
		}
		jd, _ := common.Marshal(dto.ResponsesStreamResponse{Type: "response.created", Response: &createdResp})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.created"}, string(jd))

		// response.in_progress
		jd, _ = common.Marshal(dto.ResponsesStreamResponse{Type: "response.in_progress", Response: &createdResp})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.in_progress"}, string(jd))
	}

	sendTextPreamble := func(sr *helper.StreamResult) {
		sendPreamble(sr)
		if sentTextItem {
			return
		}
		sentTextItem = true

		// response.output_item.added (message)
		item := dto.ResponsesOutput{Type: "message", ID: msgItemId, Role: "assistant", Status: "in_progress"}
		jd, _ := common.Marshal(dto.ResponsesStreamResponse{Type: "response.output_item.added", Item: &item, OutputIndex: common.GetPointer(0)})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_item.added"}, string(jd))

		// response.content_part.added (output_text)
		part := dto.ResponsesOutputContent{Type: "output_text", Text: ""}
		cpEvt := map[string]any{
			"type":         "response.content_part.added",
			"item_id":      msgItemId,
			"output_index": 0,
			"content_index": 0,
			"part":         part,
		}
		jd, _ = common.Marshal(cpEvt)
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.content_part.added"}, string(jd))
	}

	sendReasonPreamble := func(sr *helper.StreamResult) {
		sendPreamble(sr)
		if sentReasonItem {
			return
		}
		sentReasonItem = true

		// response.output_item.added (reasoning)
		item := dto.ResponsesOutput{Type: "reasoning_summary_text", ID: "rs_" + responseId, Status: "in_progress"}
		jd, _ := common.Marshal(dto.ResponsesStreamResponse{Type: "response.output_item.added", Item: &item, OutputIndex: common.GetPointer(0)})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_item.added"}, string(jd))
	}

	helper.StreamScannerHandler(c, resp, info, func(data string, sr *helper.StreamResult) {
		if streamErr != nil {
			sr.Stop(streamErr)
			return
		}

		var chatChunk dto.ChatCompletionsStreamResponse
		if err := common.UnmarshalJsonStr(data, &chatChunk); err != nil {
			return
		}

		if len(chatChunk.Choices) == 0 {
			return
		}
		choice := chatChunk.Choices[0]

		if chatChunk.Usage != nil {
			if chatChunk.Usage.PromptTokens > 0 {
				usage.PromptTokens = chatChunk.Usage.PromptTokens
			}
			if chatChunk.Usage.CompletionTokens > 0 {
				usage.CompletionTokens = chatChunk.Usage.CompletionTokens
			}
			if chatChunk.Usage.TotalTokens > 0 {
				usage.TotalTokens = chatChunk.Usage.TotalTokens
			}
		}

		delta := choice.Delta

		if delta.Content != nil && *delta.Content != "" {
			sendTextPreamble(sr)
			jd, _ := common.Marshal(dto.ResponsesStreamResponse{Type: "response.output_text.delta", Delta: *delta.Content, ItemID: msgItemId, OutputIndex: common.GetPointer(0), ContentIndex: common.GetPointer(0)})
			helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_text.delta"}, string(jd))
		}

		if delta.ReasoningContent != nil && *delta.ReasoningContent != "" {
			sendReasonPreamble(sr)
			jd, _ := common.Marshal(dto.ResponsesStreamResponse{Type: "response.reasoning_summary_text.delta", Delta: *delta.ReasoningContent})
			helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.reasoning_summary_text.delta"}, string(jd))
		}

		for _, tc := range delta.ToolCalls {
			if tc.Function.Name != "" {
				sendPreamble(sr)
				jd, _ := common.Marshal(dto.ResponsesStreamResponse{
					Type:   "response.output_item.added",
					ItemID: tc.ID,
					Item:   &dto.ResponsesOutput{Type: "function_call", CallId: tc.ID, ID: tc.ID, Name: tc.Function.Name},
				})
				helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_item.added"}, string(jd))
			}
			if tc.Function.Arguments != "" {
				jd, _ := common.Marshal(dto.ResponsesStreamResponse{Type: "response.function_call_arguments.delta", ItemID: tc.ID, Delta: tc.Function.Arguments})
				helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.function_call_arguments.delta"}, string(jd))
			}
		}

		if choice.FinishReason != nil && *choice.FinishReason != "" {
			// Close items before completing
			if sentTextItem {
				jd, _ := common.Marshal(map[string]any{
					"type": "response.content_part.done", "item_id": msgItemId,
					"output_index": 0, "content_index": 0,
					"part": dto.ResponsesOutputContent{Type: "output_text", Text: ""},
				})
				helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.content_part.done"}, string(jd))
				jd, _ = common.Marshal(map[string]any{
					"type": "response.output_item.done", "item": map[string]any{
						"type": "message", "id": msgItemId, "role": "assistant", "status": "completed",
					}, "output_index": 0,
				})
				helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_item.done"}, string(jd))
			}
			if sentReasonItem {
				jd, _ := common.Marshal(map[string]any{
					"type": "response.output_item.done", "item": map[string]any{
						"type": "reasoning_summary_text", "id": "rs_" + responseId, "status": "completed",
					}, "output_index": 0,
				})
				helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_item.done"}, string(jd))
			}
			completedResp := dto.OpenAIResponsesResponse{
				ID:     responseId,
				Object: "response",
				Model:  model,
				Usage:  usage,
			}
			jd, _ := common.Marshal(dto.ResponsesStreamResponse{Type: "response.completed", Response: &completedResp})
			helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.completed"}, string(jd))
		}
	})

	if streamErr != nil {
		return nil, streamErr
	}

	if usage.TotalTokens == 0 {
		usage.PromptTokens = info.GetEstimatePromptTokens()
		usage.TotalTokens = usage.PromptTokens
	}

	return usage, nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
