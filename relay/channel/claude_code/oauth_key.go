package claude_code

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
)

// OAuthKey 存储 Claude Code OAuth 认证信息
// 序列化为 JSON 后存储在 Channel.Key 字段中
type OAuthKey struct {
	// OAuth 访问令牌（JWT 格式）
	AccessToken string `json:"access_token,omitempty"`
	// OAuth 刷新令牌
	RefreshToken string `json:"refresh_token,omitempty"`
	// 账户 UUID（从 token 响应中获取）
	AccountUUID string `json:"account_uuid,omitempty"`
	// 组织 UUID（从 token 响应中获取）
	OrganizationUUID string `json:"organization_uuid,omitempty"`
	// 用户邮箱
	Email string `json:"email,omitempty"`
	// 上次刷新时间（RFC3339 格式）
	LastRefresh string `json:"last_refresh,omitempty"`
	// Token 过期时间（RFC3339 格式）
	Expired string `json:"expired,omitempty"`
	// 类型标识
	Type string `json:"type,omitempty"`
}

// ParseOAuthKey 从 JSON 字符串解析 OAuthKey
func ParseOAuthKey(raw string) (*OAuthKey, error) {
	if raw == "" {
		return nil, errors.New("claude-code channel: empty oauth key")
	}
	var key OAuthKey
	if err := common.Unmarshal([]byte(raw), &key); err != nil {
		return nil, errors.New("claude-code channel: invalid oauth key json")
	}
	return &key, nil
}
