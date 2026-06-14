package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// Claude Code OAuth 常量
const (
	claudeCodeClientID     = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	claudeCodeAuthorizeURL = "https://claude.com/cai/oauth/authorize"
	claudeCodeTokenURL     = "https://platform.claude.com/v1/oauth/token"
	claudeCodeRedirectURI  = "http://localhost:1456/callback"
	claudeCodeScope        = "user:profile user:inference user:sessions:claude_code user:mcp_servers user:file_upload org:create_api_key"
	claudeCodeHTTPTimeout  = 20 * time.Second
)

// ClaudeCodeOAuthTokenResult 存储 OAuth Token 交换结果
type ClaudeCodeOAuthTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	// 从 token 响应中获取的账户信息
	Account struct {
		UUID         string `json:"uuid"`
		EmailAddress string `json:"email_address"`
	} `json:"account"`
	Organization struct {
		UUID string `json:"uuid"`
	} `json:"organization"`
}

// ClaudeCodeOAuthAuthorizationFlow 存储 OAuth 授权流程信息
type ClaudeCodeOAuthAuthorizationFlow struct {
	State        string
	Verifier     string
	Challenge    string
	AuthorizeURL string
}

// CreateClaudeCodeOAuthAuthorizationFlow 创建 OAuth 授权流程
// 生成 PKCE code_verifier 和 code_challenge，构建授权 URL
func CreateClaudeCodeOAuthAuthorizationFlow() (*ClaudeCodeOAuthAuthorizationFlow, error) {
	state, err := createStateHex(16)
	if err != nil {
		return nil, err
	}
	verifier, challenge, err := generatePKCEPair()
	if err != nil {
		return nil, err
	}
	u, err := buildClaudeCodeAuthorizeURL(state, challenge)
	if err != nil {
		return nil, err
	}
	return &ClaudeCodeOAuthAuthorizationFlow{
		State:        state,
		Verifier:     verifier,
		Challenge:    challenge,
		AuthorizeURL: u,
	}, nil
}

// ExchangeClaudeCodeAuthorizationCode 用授权码交换 Token
func ExchangeClaudeCodeAuthorizationCode(ctx context.Context, code string, verifier string) (*ClaudeCodeOAuthTokenResult, error) {
	return ExchangeClaudeCodeAuthorizationCodeWithProxy(ctx, code, verifier, "")
}

// ExchangeClaudeCodeAuthorizationCodeWithProxy 用授权码交换 Token（带代理）
func ExchangeClaudeCodeAuthorizationCodeWithProxy(ctx context.Context, code string, verifier string, proxyURL string) (*ClaudeCodeOAuthTokenResult, error) {
	client, err := getClaudeCodeOAuthHTTPClient(proxyURL)
	if err != nil {
		return nil, err
	}
	return exchangeClaudeCodeAuthorizationCode(ctx, client, claudeCodeTokenURL, claudeCodeClientID, code, verifier, claudeCodeRedirectURI)
}

// RefreshClaudeCodeOAuthToken 刷新 OAuth Token
func RefreshClaudeCodeOAuthToken(ctx context.Context, refreshToken string) (*ClaudeCodeOAuthTokenResult, error) {
	return RefreshClaudeCodeOAuthTokenWithProxy(ctx, refreshToken, "")
}

// RefreshClaudeCodeOAuthTokenWithProxy 刷新 OAuth Token（带代理）
func RefreshClaudeCodeOAuthTokenWithProxy(ctx context.Context, refreshToken string, proxyURL string) (*ClaudeCodeOAuthTokenResult, error) {
	client, err := getClaudeCodeOAuthHTTPClient(proxyURL)
	if err != nil {
		return nil, err
	}
	return refreshClaudeCodeOAuthToken(ctx, client, claudeCodeTokenURL, claudeCodeClientID, refreshToken)
}

// --- 内部实现 ---

func exchangeClaudeCodeAuthorizationCode(
	ctx context.Context,
	client *http.Client,
	tokenURL string,
	clientID string,
	code string,
	verifier string,
	redirectURI string,
) (*ClaudeCodeOAuthTokenResult, error) {
	c := strings.TrimSpace(code)
	v := strings.TrimSpace(verifier)
	if c == "" {
		return nil, errors.New("claude-code oauth: empty authorization code")
	}
	if v == "" {
		return nil, errors.New("claude-code oauth: empty code_verifier")
	}

	// Claude Code 使用 JSON 格式的 token 请求体
	payload := map[string]string{
		"grant_type":   "authorization_code",
		"code":         c,
		"client_id":    clientID,
		"code_verifier": v,
		"redirect_uri": redirectURI,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ClaudeCodeOAuthTokenResult
	if err := common.DecodeJson(resp.Body, &result); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("claude-code oauth code exchange failed: status=%d", resp.StatusCode)
	}
	if strings.TrimSpace(result.AccessToken) == "" {
		return nil, errors.New("claude-code oauth: response missing access_token")
	}

	// 从 JWT 中提取过期时间
	expiresIn := 3600 // 默认 1 小时
	if exp, ok := extractJWTExpiry(result.AccessToken); ok {
		expiresIn = exp
	}
	result.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return &result, nil
}

func refreshClaudeCodeOAuthToken(
	ctx context.Context,
	client *http.Client,
	tokenURL string,
	clientID string,
	refreshToken string,
) (*ClaudeCodeOAuthTokenResult, error) {
	rt := strings.TrimSpace(refreshToken)
	if rt == "" {
		return nil, errors.New("claude-code oauth: empty refresh_token")
	}

	// Claude Code 使用 JSON 格式的刷新请求体
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": rt,
		"client_id":     clientID,
		"scope":         claudeCodeScope,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ClaudeCodeOAuthTokenResult
	if err := common.DecodeJson(resp.Body, &result); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("claude-code oauth refresh failed: status=%d", resp.StatusCode)
	}
	if strings.TrimSpace(result.AccessToken) == "" {
		return nil, errors.New("claude-code oauth: refresh response missing access_token")
	}

	// 从 JWT 中提取过期时间
	expiresIn := 3600
	if exp, ok := extractJWTExpiry(result.AccessToken); ok {
		expiresIn = exp
	}
	result.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return &result, nil
}

func getClaudeCodeOAuthHTTPClient(proxyURL string) (*http.Client, error) {
	baseClient, err := GetHttpClientWithProxy(strings.TrimSpace(proxyURL))
	if err != nil {
		return nil, err
	}
	if baseClient == nil {
		return &http.Client{Timeout: claudeCodeHTTPTimeout}, nil
	}
	clientCopy := *baseClient
	clientCopy.Timeout = claudeCodeHTTPTimeout
	return &clientCopy, nil
}

func buildClaudeCodeAuthorizeURL(state string, challenge string) (string, error) {
	u, err := url.Parse(claudeCodeAuthorizeURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("code", "true")
	q.Set("client_id", claudeCodeClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", claudeCodeRedirectURI)
	q.Set("scope", claudeCodeScope)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// extractJWTExpiry 从 JWT token 中提取过期时间（秒数）
// 不验证签名，仅解码 payload 中的 exp 字段
func extractJWTExpiry(token string) (int, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, false
	}
	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, false
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return 0, false
	}
	exp, ok := claims["exp"].(float64)
	if !ok || exp <= 0 {
		return 0, false
	}
	return int(exp), true
}
