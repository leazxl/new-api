package claude_code

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/claude"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/model_setting"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// Adaptor 实现 channel.Adaptor 接口
// Claude Code 订阅版使用 OAuth Bearer Token 认证，但 API 格式与标准 Claude API 相同
type Adaptor struct {
	// 内嵌 claude.Adaptor 以复用请求/响应处理逻辑
	claude.Adaptor
}

func (a *Adaptor) ConvertGeminiRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeminiChatRequest) (any, error) {
	return nil, errors.New("claude-code channel: endpoint not supported")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	return nil, errors.New("claude-code channel: endpoint not supported")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	return nil, errors.New("claude-code channel: endpoint not supported")
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
}

// GetRequestURL 返回请求 URL
// Claude Code 使用标准 Anthropic Messages API endpoint
func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	requestURL := fmt.Sprintf("%s/v1/messages", info.ChannelBaseUrl)
	if !shouldAppendClaudeCodeBetaQuery(info) {
		return requestURL, nil
	}

	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return "", err
	}
	query := parsedURL.Query()
	query.Set("beta", "true")
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func shouldAppendClaudeCodeBetaQuery(info *relaycommon.RelayInfo) bool {
	if info == nil {
		return false
	}
	if info.IsClaudeBetaQuery {
		return true
	}
	if info.ChannelOtherSettings.ClaudeBetaQuery {
		return true
	}
	return false
}

// SetupRequestHeader 设置请求头
// 关键区别：Claude Code 使用 Authorization: Bearer {oauth_token} 而非 x-api-key
func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)

	// 解析 OAuth Key
	key := strings.TrimSpace(info.ApiKey)
	if !strings.HasPrefix(key, "{") {
		return errors.New("claude-code channel: key must be a JSON object (OAuth key)")
	}

	oauthKey, err := ParseOAuthKey(key)
	if err != nil {
		return err
	}

	accessToken := strings.TrimSpace(oauthKey.AccessToken)
	if accessToken == "" {
		return errors.New("claude-code channel: access_token is required")
	}

	// 使用 Bearer Token 认证（Claude Code OAuth 方式）
	req.Set("Authorization", "Bearer "+accessToken)

	// 设置 Anthropic 版本号
	anthropicVersion := c.Request.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		anthropicVersion = "2023-06-01"
	}
	req.Set("anthropic-version", anthropicVersion)

	// 透传客户端的 anthropic-beta header
	anthropicBeta := c.Request.Header.Get("anthropic-beta")
	if anthropicBeta != "" {
		req.Set("anthropic-beta", anthropicBeta)
	} else {
		// 设置 Claude Code 特有的 beta header
		req.Set("anthropic-beta", "oauth-2025-04-20")
	}

	// 写入 Claude 模型设置相关的 headers
	model_setting.GetClaudeSettings().WriteHeaders(info.OriginModelName, req)

	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	// 复用 claude 包的 OpenAI -> Claude 转换逻辑
	return claude.RequestOpenAI2ClaudeMessage(c, *request)
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return nil, errors.New("claude-code channel: endpoint not supported")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	return nil, errors.New("claude-code channel: endpoint not supported")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	info.FinalRequestRelayFormat = types.RelayFormatClaude
	if info.IsStream {
		return claude.ClaudeStreamHandler(c, resp, info)
	} else {
		return claude.ClaudeHandler(c, resp, info)
	}
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
