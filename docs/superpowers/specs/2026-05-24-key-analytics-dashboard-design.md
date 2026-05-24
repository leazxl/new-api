# Key Analytics Dashboard — Design Spec

## Overview

Add a `/dashboard/keys` page that shows per-API-key analytics for the current user, with snapshot stat cards and trend/distribution/ranking charts. Positioned as a sibling of the existing overview, models, and users dashboard sections.

## Backend

### New Endpoint: `GET /api/log/self/stat/tokens`

Returns per-token aggregate statistics from the `logs` table, scoped to the current user.

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| start_timestamp | int64 | no | Unix timestamp, inclusive |
| end_timestamp | int64 | no | Unix timestamp, inclusive |
| model_name | string | no | Filter by model |

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "token_id": 12,
      "token_name": "my-prod-key",
      "quota": 150000,
      "count": 342,
      "tokens": 1250000
    }
  ]
}
```

**Implementation:**

- `model/log.go`: Add `SumUsedQuotaGroupByToken(userId int, startTime, endTime int64, modelName string) ([]TokenStat, error)`
- `controller/log.go`: Add `GetLogsSelfStatTokens(c *gin.Context)`
- `router/api-router.go`: Register route `GET /api/log/self/stat/tokens`

**SQL (MySQL example — must work on SQLite/PostgreSQL too):**

```sql
SELECT token_id, token_name,
       SUM(quota) AS quota,
       COUNT(*) AS count,
       COALESCE(SUM(prompt_tokens), 0) + COALESCE(SUM(completion_tokens), 0) AS tokens
FROM logs
WHERE user_id = ? AND type = 2
  AND created_at >= ? AND created_at <= ?
  AND token_name != ''
  [AND model_name = ?]
GROUP BY token_id, token_name
ORDER BY quota DESC
```

## Frontend

### Route

`/dashboard/keys` — added to `DASHBOARD_SECTION_IDS` and `SECTION_META` in `section-registry.tsx`.

### KeyStatCards Component

A grid of stat cards, each representing one token:

```
┌──────────────┬──────────────┬──────────────┐
│ my-prod-key  │ dev-key      │ test-key     │
│ 342 reqs     │ 120 reqs     │ 45 reqs      │
│ 1.2M tokens  │ 0.5M tokens  │ 0.1M tokens  │
│ 15K quota    │ 6K quota     │ 1.2K quota   │
└──────────────┴──────────────┴──────────────┘
```

Each card shows: token name, request count, total tokens, total quota consumed. Cards are sorted by quota descending (from the API response).

### KeyCharts Component

Three-tab chart (mirroring ModelCharts):

| Tab | Chart Type | Data |
|-----|-----------|------|
| Trend (趋势) | Multi-line area chart | Quota per token over time (from `quota_data` grouped by token_name) |
| Distribution (分布) | Pie chart | Quota share per token |
| Ranking (排行) | Horizontal bar chart | Tokens ranked by quota/requests |

**Time filter**: Reuses the existing `ModelsFilter` dialog and `DashboardFilters` (start_timestamp, end_timestamp, time_granularity). Same behavior as the models section.

**Data for charts**: The trend chart needs time-series data. For the `/dashboard/models` section, time-series data comes from `GET /api/data/self` (quota_data table). For keys, we need a similar endpoint that groups by (token_name, time_bucket). See below.

### Chart Data Endpoint: Extend `GET /api/data/self`

Extend the existing `GetUserQuotaDates` controller to accept an optional `group_by` parameter:
- `group_by=model_name` (default, existing behavior)
- `group_by=token_name` (new, for key charts)

Alternatively, add a new query param `token_name` to filter quota_data by token. The quota_data table already stores `model_name`; the consume log pipeline records `token_name`. The `CacheQuotaData` map key is currently `(username, model_name, created_at)`. For token-level data, we'd need to also include `token_name` in the aggregation key — but this is a larger change.

**Simpler approach for trend data**: Query the `logs` table directly for time-series data grouped by token_name. Add a new model function `GetLogsGroupByTokenAndDate(userId, startTime, endTime, modelName)` that buckets by date and token_name.

### File Changes

```
web/default/src/features/dashboard/
  api.ts                              — add fetchTokenStats(), fetchTokenChartData()
  types.ts                            — add TokenStat, TokenChartData types
  constants.ts                        — add KEYS_DEFAULT_TAB, KEYS_TAB_OPTIONS
  section-registry.tsx                — register 'keys' section
  index.tsx                           — add Keys section routing + lazy loading
  components/keys/
    key-stat-cards.tsx                — token stat card grid
    key-charts.tsx                    — three-tab chart (trend/distribution/ranking)
```

### i18n

New translation keys for English (and propagated to zh/ja/fr/ru/vi):

- "Key Analytics" — section title
- "View per-key usage analytics and charts" — section description
- "Token Stats" — stat cards header
- "Key Trend" — trend tab label
- "Key Distribution" — pie tab label
- "Key Ranking" — ranking tab label

## Verification

1. Go build: `go build ./...`
2. Frontend build: `cd web/default && bun run build`
3. New API endpoint returns correct data: `curl http://localhost:3000/api/log/self/stat/tokens` with auth cookie
4. Dashboard page loads at `/dashboard/keys` without console errors
5. Time filter works identically to models page
6. Charts render with VChart when data is available
7. Empty state: when user has no keys or no usage data, show appropriate message
