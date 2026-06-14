package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	claudeCodeCredentialRefreshTickInterval = 10 * time.Minute
	claudeCodeCredentialRefreshThreshold    = 24 * time.Hour
	claudeCodeCredentialRefreshBatchSize    = 200
	claudeCodeCredentialRefreshTimeout      = 15 * time.Second
)

var (
	claudeCodeCredentialRefreshOnce    sync.Once
	claudeCodeCredentialRefreshRunning atomic.Bool
)

// ClaudeCodeCredentialRefreshOptions 控制凭证刷新行为的选项
type ClaudeCodeCredentialRefreshOptions struct {
	ResetCaches bool
}

// ClaudeCodeOAuthKey 内部使用的 OAuth Key 结构（与 claude_code.OAuthKey 结构相同）
type ClaudeCodeOAuthKey struct {
	AccessToken      string `json:"access_token,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	AccountUUID      string `json:"account_uuid,omitempty"`
	OrganizationUUID string `json:"organization_uuid,omitempty"`
	Email            string `json:"email,omitempty"`
	LastRefresh      string `json:"last_refresh,omitempty"`
	Expired          string `json:"expired,omitempty"`
	Type             string `json:"type,omitempty"`
}

func parseClaudeCodeOAuthKey(raw string) (*ClaudeCodeOAuthKey, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("claude-code channel: empty oauth key")
	}
	var key ClaudeCodeOAuthKey
	if err := common.Unmarshal([]byte(raw), &key); err != nil {
		return nil, fmt.Errorf("claude-code channel: invalid oauth key json")
	}
	return &key, nil
}

// RefreshClaudeCodeChannelCredential 刷新指定 Claude Code 通道的凭证
func RefreshClaudeCodeChannelCredential(ctx context.Context, channelID int, opts ClaudeCodeCredentialRefreshOptions) (*ClaudeCodeOAuthKey, *model.Channel, error) {
	ch, err := model.GetChannelById(channelID, true)
	if err != nil {
		return nil, nil, err
	}
	if ch == nil {
		return nil, nil, fmt.Errorf("channel not found")
	}
	if ch.Type != constant.ChannelTypeClaudeCode {
		return nil, nil, fmt.Errorf("channel type is not ClaudeCode")
	}

	oauthKey, err := parseClaudeCodeOAuthKey(strings.TrimSpace(ch.Key))
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(oauthKey.RefreshToken) == "" {
		return nil, nil, fmt.Errorf("claude-code channel: refresh_token is required to refresh credential")
	}

	refreshCtx, cancel := context.WithTimeout(ctx, claudeCodeCredentialRefreshTimeout)
	defer cancel()

	res, err := RefreshClaudeCodeOAuthTokenWithProxy(refreshCtx, oauthKey.RefreshToken, ch.GetSetting().Proxy)
	if err != nil {
		return nil, nil, err
	}

	oauthKey.AccessToken = res.AccessToken
	if strings.TrimSpace(res.RefreshToken) != "" {
		oauthKey.RefreshToken = res.RefreshToken
	}
	oauthKey.LastRefresh = time.Now().Format(time.RFC3339)
	oauthKey.Expired = res.ExpiresAt.Format(time.RFC3339)
	if strings.TrimSpace(oauthKey.Type) == "" {
		oauthKey.Type = "claude-code"
	}
	// 更新账户信息
	if strings.TrimSpace(res.Account.UUID) != "" {
		oauthKey.AccountUUID = res.Account.UUID
	}
	if strings.TrimSpace(res.Account.EmailAddress) != "" {
		oauthKey.Email = res.Account.EmailAddress
	}

	encoded, err := common.Marshal(oauthKey)
	if err != nil {
		return nil, nil, err
	}

	if err := model.DB.Model(&model.Channel{}).Where("id = ?", ch.Id).Update("key", string(encoded)).Error; err != nil {
		return nil, nil, err
	}

	if opts.ResetCaches {
		model.InitChannelCache()
		ResetProxyClientCache()
	}

	return oauthKey, ch, nil
}

// StartClaudeCodeCredentialAutoRefreshTask 启动 Claude Code 凭证自动刷新后台任务
func StartClaudeCodeCredentialAutoRefreshTask() {
	claudeCodeCredentialRefreshOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}

		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf("claude-code credential auto-refresh task started: tick=%s threshold=%s", claudeCodeCredentialRefreshTickInterval, claudeCodeCredentialRefreshThreshold))

			ticker := time.NewTicker(claudeCodeCredentialRefreshTickInterval)
			defer ticker.Stop()

			runClaudeCodeCredentialAutoRefreshOnce()
			for range ticker.C {
				runClaudeCodeCredentialAutoRefreshOnce()
			}
		})
	})
}

func runClaudeCodeCredentialAutoRefreshOnce() {
	if !claudeCodeCredentialRefreshRunning.CompareAndSwap(false, true) {
		return
	}
	defer claudeCodeCredentialRefreshRunning.Store(false)

	ctx := context.Background()
	now := time.Now()

	var refreshed int
	var scanned int

	offset := 0
	for {
		var channels []*model.Channel
		err := model.DB.
			Select("id", "name", "key", "status", "channel_info").
			Where("type = ? AND (status = ? OR status = ?)",
				constant.ChannelTypeClaudeCode,
				common.ChannelStatusEnabled,
				common.ChannelStatusAutoDisabled,
			).
			Order("id asc").
			Limit(claudeCodeCredentialRefreshBatchSize).
			Offset(offset).
			Find(&channels).Error
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf("claude-code credential auto-refresh: query channels failed: %v", err))
			return
		}
		if len(channels) == 0 {
			break
		}
		offset += claudeCodeCredentialRefreshBatchSize

		for _, ch := range channels {
			if ch == nil {
				continue
			}
			scanned++
			if ch.ChannelInfo.IsMultiKey {
				continue
			}

			rawKey := strings.TrimSpace(ch.Key)
			if rawKey == "" {
				continue
			}

			oauthKey, err := parseClaudeCodeOAuthKey(rawKey)
			if err != nil {
				continue
			}

			refreshToken := strings.TrimSpace(oauthKey.RefreshToken)
			if refreshToken == "" {
				continue
			}

			expiredAtRaw := strings.TrimSpace(oauthKey.Expired)
			expiredAt, err := time.Parse(time.RFC3339, expiredAtRaw)
			if err == nil && !expiredAt.IsZero() && expiredAt.Sub(now) > claudeCodeCredentialRefreshThreshold {
				continue
			}

			refreshCtx, cancel := context.WithTimeout(ctx, claudeCodeCredentialRefreshTimeout)
			newKey, _, err := RefreshClaudeCodeChannelCredential(refreshCtx, ch.Id, ClaudeCodeCredentialRefreshOptions{ResetCaches: false})
			cancel()
			if err != nil {
				logger.LogWarn(ctx, fmt.Sprintf("claude-code credential auto-refresh: channel_id=%d name=%s refresh failed: %v", ch.Id, ch.Name, err))
				continue
			}

			refreshed++
			logger.LogInfo(ctx, fmt.Sprintf("claude-code credential auto-refresh: channel_id=%d name=%s refreshed, expires_at=%s", ch.Id, ch.Name, newKey.Expired))
		}
	}

	if refreshed > 0 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.LogWarn(ctx, fmt.Sprintf("claude-code credential auto-refresh: InitChannelCache panic: %v", r))
				}
			}()
			model.InitChannelCache()
		}()
		ResetProxyClientCache()
	}

	if common.DebugEnabled {
		logger.LogDebug(ctx, "claude-code credential auto-refresh: scanned=%d refreshed=%d", scanned, refreshed)
	}
}
