package openaicompat

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

func ResponsesResponseToChatCompletionsResponse(resp *dto.OpenAIResponsesResponse, id string) (*dto.OpenAITextResponse, *dto.Usage, error) {
	if resp == nil {
		return nil, nil, errors.New("response is nil")
	}

	text := ExtractOutputTextFromResponses(resp)

	usage := &dto.Usage{}
	if resp.Usage != nil {
		if resp.Usage.InputTokens != 0 {
			usage.PromptTokens = resp.Usage.InputTokens
			usage.InputTokens = resp.Usage.InputTokens
		}
		if resp.Usage.OutputTokens != 0 {
			usage.CompletionTokens = resp.Usage.OutputTokens
			usage.OutputTokens = resp.Usage.OutputTokens
		}
		if resp.Usage.TotalTokens != 0 {
			usage.TotalTokens = resp.Usage.TotalTokens
		} else {
			usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
		}
		if resp.Usage.InputTokensDetails != nil {
			usage.PromptTokensDetails.CachedTokens = resp.Usage.InputTokensDetails.CachedTokens
			usage.PromptTokensDetails.ImageTokens = resp.Usage.InputTokensDetails.ImageTokens
			usage.PromptTokensDetails.AudioTokens = resp.Usage.InputTokensDetails.AudioTokens
		}
		if resp.Usage.CompletionTokenDetails.ReasoningTokens != 0 {
			usage.CompletionTokenDetails.ReasoningTokens = resp.Usage.CompletionTokenDetails.ReasoningTokens
		}
	}

	created := resp.CreatedAt

	var toolCalls []dto.ToolCallResponse
	if text == "" && len(resp.Output) > 0 {
		for _, out := range resp.Output {
			if out.Type != "function_call" {
				continue
			}
			name := strings.TrimSpace(out.Name)
			if name == "" {
				continue
			}
			callId := strings.TrimSpace(out.CallId)
			if callId == "" {
				callId = strings.TrimSpace(out.ID)
			}
			toolCalls = append(toolCalls, dto.ToolCallResponse{
				ID:   callId,
				Type: "function",
				Function: dto.FunctionResponse{
					Name:      name,
					Arguments: out.ArgumentsString(),
				},
			})
		}
	}

	finishReason := "stop"
	if len(toolCalls) > 0 {
		finishReason = "tool_calls"
	}

	msg := dto.Message{
		Role:    "assistant",
		Content: text,
	}
	if len(toolCalls) > 0 {
		msg.SetToolCalls(toolCalls)
		msg.Content = ""
	}

	out := &dto.OpenAITextResponse{
		Id:      id,
		Object:  "chat.completion",
		Created: created,
		Model:   resp.Model,
		Choices: []dto.OpenAITextResponseChoice{
			{
				Index:        0,
				Message:      msg,
				FinishReason: finishReason,
			},
		},
		Usage: *usage,
	}

	return out, usage, nil
}

func ExtractOutputTextFromResponses(resp *dto.OpenAIResponsesResponse) string {
	if resp == nil || len(resp.Output) == 0 {
		return ""
	}

	var sb strings.Builder

	// Prefer assistant message outputs.
	for _, out := range resp.Output {
		if out.Type != "message" {
			continue
		}
		if out.Role != "" && out.Role != "assistant" {
			continue
		}
		for _, c := range out.Content {
			if c.Type == "output_text" && c.Text != "" {
				sb.WriteString(c.Text)
			}
		}
	}
	if sb.Len() > 0 {
		return sb.String()
	}
	for _, out := range resp.Output {
		for _, c := range out.Content {
			if c.Text != "" {
				sb.WriteString(c.Text)
			}
		}
	}
	return sb.String()
}

// ResponsesRequestToChatCompletionsRequest converts an OpenAI Responses API request
// into a Chat Completions request for endpoints that only support /v1/chat/completions.
func ResponsesRequestToChatCompletionsRequest(req *dto.OpenAIResponsesRequest) (*dto.GeneralOpenAIRequest, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	if req.Model == "" {
		return nil, errors.New("model is required")
	}

	var inputItems []map[string]any
	if len(req.Input) > 0 {
		if err := common.Unmarshal(req.Input, &inputItems); err != nil {
			// input may be a plain string or a single item instead of an array
			var singleItem map[string]any
			var plainStr string
			if common.Unmarshal(req.Input, &singleItem) == nil {
				inputItems = []map[string]any{singleItem}
			} else if common.Unmarshal(req.Input, &plainStr) == nil && strings.TrimSpace(plainStr) != "" {
				inputItems = []map[string]any{
					{"role": "user", "content": plainStr},
				}
			}
		}
	}

	var instructionsParts []string
	if len(req.Instructions) > 0 {
		var s string
		if err := common.Unmarshal(req.Instructions, &s); err == nil && strings.TrimSpace(s) != "" {
			instructionsParts = append(instructionsParts, s)
		}
	}

	messages := make([]dto.Message, 0, len(inputItems)+len(instructionsParts)+1)

	if len(instructionsParts) > 0 {
		messages = append(messages, dto.Message{
			Role:    "system",
			Content: strings.Join(instructionsParts, "\n\n"),
		})
	}

	for _, item := range inputItems {
		itemType := common.Interface2String(item["type"])
		itemRole := common.Interface2String(item["role"])

		switch {
		case itemType == "function_call":
			messages = append(messages, convertFunctionCallItem(item))

		case itemType == "function_call_output":
			messages = append(messages, convertFunctionCallOutputItem(item))

		case itemRole == "system" || itemRole == "developer":
			content := extractItemContent(item)
			if strings.TrimSpace(content) != "" {
				messages = append(messages, dto.Message{
					Role:    "system",
					Content: content,
				})
			}

		case itemRole == "user", itemRole == "assistant":
			messages = append(messages, dto.Message{
				Role:    itemRole,
				Content: buildChatContent(item),
			})

		case itemType == "message":
			msgRole := itemRole
			if msgRole == "" {
				msgRole = "user"
			}
			messages = append(messages, dto.Message{
				Role:    msgRole,
				Content: buildChatContent(item),
			})

		default:
			if b, err := common.Marshal(item); err == nil {
				messages = append(messages, dto.Message{
					Role:    "user",
					Content: string(b),
				})
			}
		}
	}

	// Fallback: if no messages were constructed, add raw input as a user message.
	if len(messages) == 0 && len(req.Input) > 0 {
		messages = append(messages, dto.Message{
			Role:    "user",
			Content: string(req.Input),
		})
	}

	out := &dto.GeneralOpenAIRequest{
		Model:    req.Model,
		Messages: messages,
	}

	if req.MaxOutputTokens != nil {
		out.MaxTokens = req.MaxOutputTokens
	}
	if req.Temperature != nil {
		out.Temperature = req.Temperature
	}
	if req.TopP != nil {
		out.TopP = req.TopP
	}
	if req.Stream != nil {
		out.Stream = req.Stream
	}
	if req.StreamOptions != nil {
		out.StreamOptions = req.StreamOptions
	}
	if req.Reasoning != nil {
		out.ReasoningEffort = req.Reasoning.Effort
	}
	if len(req.Metadata) > 0 {
		out.Metadata = req.Metadata
	}

	// tools
	if len(req.Tools) > 0 {
		out.Tools = convertResponsesToolsToChatTools(req.Tools)
	}

	// tool_choice
	if len(req.ToolChoice) > 0 {
		var tc any
		if err := common.Unmarshal(req.ToolChoice, &tc); err == nil {
			out.ToolChoice = tc
		}
	}

	// text → response_format
	if len(req.Text) > 0 {
		var textCfg map[string]any
		if err := common.Unmarshal(req.Text, &textCfg); err == nil {
			if format, ok := textCfg["format"]; ok {
				var rf dto.ResponseFormat
				if b, err := common.Marshal(format); err == nil {
					_ = common.Unmarshal(b, &rf)
					out.ResponseFormat = &rf
				}
			}
		}
	}

	return out, nil
}

func convertFunctionCallItem(item map[string]any) dto.Message {
	callID := common.Interface2String(item["call_id"])
	name := common.Interface2String(item["name"])
	argsStr := ""
	if argsRaw, err := common.Marshal(item["arguments"]); err == nil {
		argsStr = string(argsRaw)
	}
	toolCalls := []dto.ToolCallRequest{
		{
			ID:   callID,
			Type: "function",
			Function: dto.FunctionRequest{
				Name: name,
			},
		},
	}
	if argsStr != "" {
		toolCalls[0].Function.Parameters = argsStr
	}
	msg := dto.Message{
		Role:    "assistant",
		Content: nil,
	}
	msg.SetToolCalls(toolCalls)
	return msg
}

func convertFunctionCallOutputItem(item map[string]any) dto.Message {
	callID := common.Interface2String(item["call_id"])
	output := item["output"]
	outputStr := ""
	switch v := output.(type) {
	case string:
		outputStr = v
	default:
		if b, err := common.Marshal(v); err == nil {
			outputStr = string(b)
		} else {
			outputStr = fmt.Sprintf("%v", v)
		}
	}
	return dto.Message{
		Role:       "tool",
		Content:    outputStr,
		ToolCallId: callID,
	}
}

func extractItemContent(item map[string]any) string {
	content, ok := item["content"]
	if !ok || content == nil {
		return ""
	}
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var texts []string
		for _, part := range v {
			if partMap, ok := part.(map[string]any); ok {
				if t := common.Interface2String(partMap["text"]); t != "" {
					texts = append(texts, t)
				}
			}
		}
		return strings.Join(texts, "\n")
	default:
		return common.Interface2String(content)
	}
}

func buildChatContent(item map[string]any) any {
	content, ok := item["content"]
	if !ok || content == nil {
		return ""
	}
	switch v := content.(type) {
	case string:
		return v
	case []any:
		contentParts := make([]dto.MediaContent, 0, len(v))
		for _, part := range v {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			partType := common.Interface2String(partMap["type"])
			switch {
			case partType == "input_text" || partType == "output_text":
				contentParts = append(contentParts, dto.MediaContent{
					Type: dto.ContentTypeText,
					Text: common.Interface2String(partMap["text"]),
				})
			case partType == "input_image":
				contentParts = append(contentParts, dto.MediaContent{
					Type: dto.ContentTypeImageURL,
					ImageUrl: dto.MessageImageUrl{
						Url: common.Interface2String(partMap["image_url"]),
					},
				})
			case partType == "input_audio":
				if audioData, ok := partMap["input_audio"]; ok {
					contentParts = append(contentParts, dto.MediaContent{
						Type:       dto.ContentTypeInputAudio,
						InputAudio: audioData,
					})
				}
			case partType == "input_file":
				if fileData, ok := partMap["file"]; ok {
					contentParts = append(contentParts, dto.MediaContent{
						Type: dto.ContentTypeFile,
						File: fileData,
					})
				}
			default:
				contentParts = append(contentParts, dto.MediaContent{
					Type: partType,
				})
			}
		}
		return contentParts
	default:
		return content
	}
}

// ChatCompletionsResponseToResponsesResponse converts a chat completion response
// back into an OpenAI Responses response.
func ChatCompletionsResponseToResponsesResponse(chatResp *dto.OpenAITextResponse, model string, instructions json.RawMessage) *dto.OpenAIResponsesResponse {
	if chatResp == nil {
		return nil
	}

	output := make([]dto.ResponsesOutput, 0)
	if len(chatResp.Choices) > 0 {
		choice := chatResp.Choices[0]
		msg := choice.Message

		var contentParts []dto.ResponsesOutputContent
		if msg.Content != nil {
			switch c := msg.Content.(type) {
			case string:
				if c != "" {
					contentParts = append(contentParts, dto.ResponsesOutputContent{
						Type: "output_text",
						Text: c,
					})
				}
			case []interface{}:
				for _, p := range c {
					if pm, ok := p.(map[string]interface{}); ok {
						cp := dto.ResponsesOutputContent{}
						if t, ok := pm["type"].(string); ok && t != "" {
							cp.Type = t
						}
						if t, ok := pm["text"].(string); ok && t != "" {
							cp.Text = t
						}
						contentParts = append(contentParts, cp)
					}
				}
			}
		}

		output = append(output, dto.ResponsesOutput{
			Type:    "message",
			Role:    "assistant",
			Content: contentParts,
		})

		for _, tc := range msg.ParseToolCalls() {
			argsRaw, _ := common.Marshal(tc.Function.Parameters)
			if argsRaw == nil {
				argsRaw = json.RawMessage(fmt.Sprintf("%v", tc.Function.Parameters))
			}
			output = append(output, dto.ResponsesOutput{
				Type:      "function_call",
				CallId:    tc.ID,
				Name:      tc.Function.Name,
				Arguments: argsRaw,
			})
		}
	}

	modelName := model
	if modelName == "" {
		modelName = chatResp.Model
	}

	resp := &dto.OpenAIResponsesResponse{
		ID:        chatResp.Id,
		Object:    "response",
		CreatedAt: convertCreatedToInt(chatResp.Created),
		Model:     modelName,
		Output:    output,
	}

	ansUsage := chatResp.Usage
	convertedUsage := &dto.Usage{
		PromptTokens:     ansUsage.PromptTokens,
		CompletionTokens: ansUsage.CompletionTokens,
		TotalTokens:      ansUsage.TotalTokens,
	}
	resp.Usage = convertedUsage

	if len(instructions) > 0 {
		resp.Instructions = instructions
	}

	return resp
}

func convertCreatedToInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

// convertResponsesToolsToChatTools converts Responses-format tools (flat structure)
// to Chat Completions format (nested under "function"), filtering out non-function types.
//
// Responses: {"type": "function", "name": "...", "description": "...", "parameters": {...}}
// Chat:      {"type": "function", "function": {"name": "...", "description": "...", "parameters": {...}}}
func convertResponsesToolsToChatTools(toolsRaw json.RawMessage) []dto.ToolCallRequest {
	var raw []map[string]any
	if err := common.Unmarshal(toolsRaw, &raw); err != nil {
		return nil
	}

	result := make([]dto.ToolCallRequest, 0, len(raw))
	for _, t := range raw {
		toolType := common.Interface2String(t["type"])
		if toolType != "" && toolType != "function" {
			continue
		}
		tc := dto.ToolCallRequest{
			Type: "function",
			Function: dto.FunctionRequest{
				Name: common.Interface2String(t["name"]),
			},
		}
		if desc := common.Interface2String(t["description"]); desc != "" {
			tc.Function.Description = desc
		}
		if params, ok := t["parameters"]; ok && params != nil {
			tc.Function.Parameters = params
		}
		result = append(result, tc)
	}
	return result
}
