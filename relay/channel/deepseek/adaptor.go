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

	// DeepSeek thinking mode requires reasoning_content to be passed back for
	// multi-turn tool-call conversations. Codex CLI doesn't preserve it, so
	// force disable thinking when the history contains tool_calls without it.
	hasAssistantToolCalls := false
	for _, m := range chatReq.Messages {
		if m.Role == "assistant" && len(m.ParseToolCalls()) > 0 {
			hasAssistantToolCalls = true
			break
		}
	}
	if hasAssistantToolCalls {
		thinkingData, _ := common.Marshal(map[string]string{"type": "disabled"})
		chatReq.THINKING = thinkingData
		chatReq.ReasoningEffort = ""
	}

	// Switch to chat completions mode for the rest of the pipeline.
	info.RelayMode = constant.RelayModeChatCompletions
	info.AppendRequestConversion(types.RelayFormatOpenAI)

	return chatReq, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

type toolCallStreamState struct {
	id     string
	callId string
	name   string
	args   string
	outIdx int
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
	var (
		usage          = &dto.Usage{}
		streamErr      *types.NewAPIError
		sentPreamble   bool
		messageStarted bool
		outputIndex    int
		textOutputIdx  = -1
		toolCalls      = make(map[int]*toolCallStreamState)
	)

	sendPreamble := func() {
		if sentPreamble {
			return
		}
		sentPreamble = true
		baseResp := dto.OpenAIResponsesResponse{ID: responseId, Object: "response", Model: model}
		jd, _ := common.Marshal(dto.ResponsesStreamResponse{Type: "response.created", Response: &baseResp})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.created"}, string(jd))
		jd, _ = common.Marshal(dto.ResponsesStreamResponse{Type: "response.in_progress", Response: &baseResp})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.in_progress"}, string(jd))
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

		// Skip reasoning_content — Codex CLI doesn't handle reasoning events.
		// (Matches codex-bridge behavior.)

		if delta.ToolCalls != nil {
			sendPreamble()
			for _, tc := range delta.ToolCalls {
				idx := 0
				if tc.Index != nil {
					idx = *tc.Index
				}
				state, exists := toolCalls[idx]
				if !exists {
					callId := tc.ID
					fcId := "fc_" + common.GetUUID()
					state = &toolCallStreamState{
						id: fcId, callId: callId, name: tc.Function.Name,
						outIdx: outputIndex,
					}
					toolCalls[idx] = state
					outputIndex++
					jd, _ := common.Marshal(map[string]any{
						"type":         "response.output_item.added",
						"output_index": state.outIdx,
						"item": map[string]any{
							"type": "function_call", "id": fcId, "call_id": callId,
							"name": tc.Function.Name, "arguments": "", "status": "in_progress",
						},
					})
					helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_item.added"}, string(jd))
				}
				if tc.Function.Arguments != "" {
					state.args += tc.Function.Arguments
					jd, _ := common.Marshal(map[string]any{
						"type":         "response.function_call_arguments.delta",
						"output_index": state.outIdx, "call_id": state.callId,
						"delta":        tc.Function.Arguments,
					})
					helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.function_call_arguments.delta"}, string(jd))
				}
			}
			if choice.FinishReason != nil && *choice.FinishReason != "" {
				completeStream(c, responseId, model, usage, &messageStarted, &textOutputIdx, toolCalls, &outputIndex, choice.FinishReason)
			}
			return
		}

		if delta.Content != nil && *delta.Content != "" {
			text := *delta.Content
			if !messageStarted {
				sendPreamble()
				messageStarted = true
				textOutputIdx = outputIndex
				outputIndex++
				jd, _ := common.Marshal(map[string]any{
					"type":         "response.output_item.added",
					"output_index": textOutputIdx,
					"item": map[string]any{
						"type": "message", "id": "msg_" + common.GetUUID(),
						"status": "in_progress", "role": "assistant", "content": []any{},
					},
				})
				helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_item.added"}, string(jd))
				jd, _ = common.Marshal(map[string]any{
					"type": "response.content_part.added", "output_index": textOutputIdx,
					"content_index": 0, "part": map[string]any{"type": "output_text", "text": "", "annotations": []any{}},
				})
				helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.content_part.added"}, string(jd))
			}
			jd, _ := common.Marshal(map[string]any{
				"type": "response.output_text.delta", "output_index": textOutputIdx,
				"content_index": 0, "delta": text,
			})
			helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_text.delta"}, string(jd))
		}

		if choice.FinishReason != nil && *choice.FinishReason != "" {
			completeStream(c, responseId, model, usage, &messageStarted, &textOutputIdx, toolCalls, &outputIndex, choice.FinishReason)
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

func completeStream(c *gin.Context, responseId string, model string, usage *dto.Usage, messageStarted *bool, textOutputIdx *int, toolCalls map[int]*toolCallStreamState, outputIndex *int, finishReason *string) {
	// Finalize tool calls
	for _, state := range toolCalls {
		jd, _ := common.Marshal(map[string]any{
			"type": "response.function_call_arguments.done", "output_index": state.outIdx,
			"call_id": state.callId, "arguments": state.args,
		})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.function_call_arguments.done"}, string(jd))
		jd, _ = common.Marshal(map[string]any{
			"type": "response.output_item.done", "output_index": state.outIdx,
			"item": map[string]any{
				"type": "function_call", "id": state.id, "call_id": state.callId,
				"name": state.name, "arguments": state.args, "status": "completed",
			},
		})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_item.done"}, string(jd))
	}

	// Finalize message
	if *messageStarted && *textOutputIdx >= 0 {
		jd, _ := common.Marshal(map[string]any{
			"type": "response.output_text.done", "output_index": *textOutputIdx, "content_index": 0,
		})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_text.done"}, string(jd))
		jd, _ = common.Marshal(map[string]any{
			"type": "response.content_part.done", "output_index": *textOutputIdx,
			"content_index": 0, "part": map[string]any{"type": "output_text", "text": "", "annotations": []any{}},
		})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.content_part.done"}, string(jd))
		jd, _ = common.Marshal(map[string]any{
			"type": "response.output_item.done", "output_index": *textOutputIdx,
			"item": map[string]any{
				"type": "message", "id": "msg_" + common.GetUUID(),
				"status": "completed", "role": "assistant", "content": []any{},
			},
		})
		helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_item.done"}, string(jd))
	}

	status := "completed"
	finishReasonStr := ""
	if finishReason != nil {
		finishReasonStr = *finishReason
	}
	if finishReasonStr == "length" {
		status = "incomplete"
	}

	completedResp := dto.OpenAIResponsesResponse{
		ID:     responseId,
		Object: "response",
		Model:  model,
		Usage:  usage,
	}
	completedResp.Status, _ = common.Marshal(status)
	jd, _ := common.Marshal(map[string]any{
		"type": "response.completed",
		"response": map[string]any{
			"id":         responseId,
			"object":     "response",
			"status":     status,
			"model":      model,
			"usage":      translateUsageForStream(usage),
		},
	})
	helper.ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.completed"}, string(jd))
}

func translateUsageForStream(u *dto.Usage) map[string]any {
	return map[string]any{
		"input_tokens":  u.PromptTokens,
		"output_tokens": u.CompletionTokens,
		"total_tokens":  u.TotalTokens,
		"input_tokens_details": map[string]any{
			"cached_tokens": u.PromptTokensDetails.CachedTokens,
		},
		"output_tokens_details": map[string]any{
			"reasoning_tokens": u.CompletionTokenDetails.ReasoningTokens,
		},
	}
}
func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
