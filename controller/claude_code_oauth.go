package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel/claude_code"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type claudeCodeOAuthCompleteRequest struct {
	Input string `json:"input"`
}

func claudeCodeOAuthSessionKey(channelID int, field string) string {
	return fmt.Sprintf("claude_code_oauth_%s_%d", field, channelID)
}

// parseClaudeCodeAuthorizationInput 解析用户输入的授权回调信息
// 支持格式：
//   - "code=xxx&state=xxx" (URL query 格式)
//   - "code=xxx" (仅 code)
//   - 完整 URL
func parseClaudeCodeAuthorizationInput(input string) (code string, state string, err error) {
	v := strings.TrimSpace(input)
	if v == "" {
		return "", "", errors.New("empty input")
	}
	// 尝试从 URL query 参数中提取
	if strings.Contains(v, "code=") {
		u, parseErr := url.Parse(v)
		if parseErr == nil && u.Query().Get("code") != "" {
			q := u.Query()
			code = strings.TrimSpace(q.Get("code"))
			state = strings.TrimSpace(q.Get("state"))
			return code, state, nil
		}
		// 尝试直接解析为 query string
		q, parseErr := url.ParseQuery(v)
		if parseErr == nil && q.Get("code") != "" {
			code = strings.TrimSpace(q.Get("code"))
			state = strings.TrimSpace(q.Get("state"))
			return code, state, nil
		}
	}

	// 直接作为 code 使用
	code = v
	return code, "", nil
}

// StartClaudeCodeOAuth 开始 OAuth 授权流程
func StartClaudeCodeOAuth(c *gin.Context) {
	startClaudeCodeOAuthWithChannelID(c, 0)
}

// StartClaudeCodeOAuthForChannel 为指定通道开始 OAuth 授权流程
func StartClaudeCodeOAuthForChannel(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}
	startClaudeCodeOAuthWithChannelID(c, channelID)
}

func startClaudeCodeOAuthWithChannelID(c *gin.Context, channelID int) {
	if channelID > 0 {
		ch, err := model.GetChannelById(channelID, false)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if ch == nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel not found"})
			return
		}
		if ch.Type != constant.ChannelTypeClaudeCode {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel type is not ClaudeCode"})
			return
		}
	}

	flow, err := service.CreateClaudeCodeOAuthAuthorizationFlow()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	session := sessions.Default(c)
	session.Set(claudeCodeOAuthSessionKey(channelID, "state"), flow.State)
	session.Set(claudeCodeOAuthSessionKey(channelID, "verifier"), flow.Verifier)
	session.Set(claudeCodeOAuthSessionKey(channelID, "created_at"), time.Now().Unix())
	_ = session.Save()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"authorize_url": flow.AuthorizeURL,
		},
	})
}

// CompleteClaudeCodeOAuth 完成 OAuth 授权流程
func CompleteClaudeCodeOAuth(c *gin.Context) {
	completeClaudeCodeOAuthWithChannelID(c, 0)
}

// CompleteClaudeCodeOAuthForChannel 为指定通道完成 OAuth 授权流程
func CompleteClaudeCodeOAuthForChannel(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}
	completeClaudeCodeOAuthWithChannelID(c, channelID)
}

func completeClaudeCodeOAuthWithChannelID(c *gin.Context, channelID int) {
	req := claudeCodeOAuthCompleteRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}

	code, state, err := parseClaudeCodeAuthorizationInput(req.Input)
	if err != nil {
		common.SysError("failed to parse claude-code authorization input: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "解析授权信息失败，请检查输入格式"})
		return
	}
	if strings.TrimSpace(code) == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "missing authorization code"})
		return
	}

	channelProxy := ""
	if channelID > 0 {
		ch, err := model.GetChannelById(channelID, false)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if ch == nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel not found"})
			return
		}
		if ch.Type != constant.ChannelTypeClaudeCode {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel type is not ClaudeCode"})
			return
		}
		channelProxy = ch.GetSetting().Proxy
	}

	session := sessions.Default(c)
	expectedState, _ := session.Get(claudeCodeOAuthSessionKey(channelID, "state")).(string)
	verifier, _ := session.Get(claudeCodeOAuthSessionKey(channelID, "verifier")).(string)
	if strings.TrimSpace(expectedState) == "" || strings.TrimSpace(verifier) == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "oauth flow not started or session expired"})
		return
	}
	// state 校验（如果客户端提供了 state）
	if state != "" && state != expectedState {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "state mismatch"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	tokenRes, err := service.ExchangeClaudeCodeAuthorizationCodeWithProxy(ctx, code, verifier, channelProxy)
	if err != nil {
		common.SysError("failed to exchange claude-code authorization code: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "授权码交换失败，请重试"})
		return
	}

	key := claude_code.OAuthKey{
		AccessToken:      tokenRes.AccessToken,
		RefreshToken:     tokenRes.RefreshToken,
		AccountUUID:      tokenRes.Account.UUID,
		OrganizationUUID: tokenRes.Organization.UUID,
		Email:            tokenRes.Account.EmailAddress,
		LastRefresh:      time.Now().Format(time.RFC3339),
		Expired:          tokenRes.ExpiresAt.Format(time.RFC3339),
		Type:             "claude-code",
	}
	encoded, err := common.Marshal(key)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 清理 session
	session.Delete(claudeCodeOAuthSessionKey(channelID, "state"))
	session.Delete(claudeCodeOAuthSessionKey(channelID, "verifier"))
	session.Delete(claudeCodeOAuthSessionKey(channelID, "created_at"))
	_ = session.Save()

	if channelID > 0 {
		// 直接保存到通道
		if err := model.DB.Model(&model.Channel{}).Where("id = ?", channelID).Update("key", string(encoded)).Error; err != nil {
			common.ApiError(c, err)
			return
		}
		model.InitChannelCache()
		service.ResetProxyClientCache()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "saved",
			"data": gin.H{
				"channel_id":   channelID,
				"account_uuid": key.AccountUUID,
				"email":        key.Email,
				"expires_at":   key.Expired,
				"last_refresh": key.LastRefresh,
			},
		})
		return
	}

	// 返回 key 供用户手动配置
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "generated",
		"data": gin.H{
			"key":          string(encoded),
			"account_uuid": key.AccountUUID,
			"email":        key.Email,
			"expires_at":   key.Expired,
			"last_refresh": key.LastRefresh,
		},
	})
}

// RefreshClaudeCodeChannelCredential 手动刷新 Claude Code 通道凭证
func RefreshClaudeCodeChannelCredential(c *gin.Context) {
	channelId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	oauthKey, ch, err := service.RefreshClaudeCodeChannelCredential(ctx, channelId, service.ClaudeCodeCredentialRefreshOptions{ResetCaches: true})
	if err != nil {
		common.SysError("failed to refresh claude-code channel credential: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "刷新凭证失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "refreshed",
		"data": gin.H{
			"expires_at":   oauthKey.Expired,
			"last_refresh": oauthKey.LastRefresh,
			"account_uuid": oauthKey.AccountUUID,
			"email":        oauthKey.Email,
			"channel_id":   ch.Id,
			"channel_type": ch.Type,
			"channel_name": ch.Name,
		},
	})
}
