import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Key, Hash, Coins } from 'lucide-react'
import { formatNumber, formatQuota } from '@/lib/format'
import { computeTimeRange } from '@/lib/time'
import { getDefaultDays } from '@/features/dashboard/lib'
import { getTokenStats } from '@/features/dashboard/api'
import type { DashboardFilters, TokenStat } from '@/features/dashboard/types'

interface KeyStatCardsProps {
  filters?: DashboardFilters
  onDataUpdate?: (data: TokenStat[], loading: boolean) => void
}

export function KeyStatCards(props: KeyStatCardsProps) {
  const { t } = useTranslation()
  const { filters, onDataUpdate } = props
  const [stats, setStats] = useState<TokenStat[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const abortController = new AbortController()
    setLoading(true)

    const timeRange = computeTimeRange(
      getDefaultDays(filters?.time_granularity),
      filters?.start_timestamp,
      filters?.end_timestamp
    )

    onDataUpdate?.([], true)

    getTokenStats({
      start_timestamp: timeRange.start_timestamp,
      end_timestamp: timeRange.end_timestamp,
    })
      .then((res) => {
        if (abortController.signal.aborted) return
        const tokenStats = res?.data ?? []
        setStats(tokenStats)
        onDataUpdate?.(tokenStats, false)
      })
      .catch(() => {
        setStats([])
        onDataUpdate?.([], false)
      })
      .finally(() => {
        if (!abortController.signal.aborted) setLoading(false)
      })

    return () => abortController.abort()
  }, [filters?.start_timestamp, filters?.end_timestamp, filters?.time_granularity, onDataUpdate])

  if (loading) {
    return (
      <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3'>
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className='rounded-lg border px-4 py-3'>
            <div className='h-4 w-24 animate-pulse rounded bg-muted' />
            <div className='mt-2 space-y-2'>
              <div className='h-6 w-16 animate-pulse rounded bg-muted' />
              <div className='h-4 w-32 animate-pulse rounded bg-muted' />
              <div className='h-4 w-20 animate-pulse rounded bg-muted' />
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (stats.length === 0) {
    return (
      <div className='flex flex-col items-center justify-center rounded-lg border py-12 text-muted-foreground'>
        <Key className='mb-2 size-8 opacity-40' />
        <p className='text-sm'>{t('No key usage data available')}</p>
      </div>
    )
  }

  return (
    <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3'>
      {stats.map((stat) => (
        <div
          key={stat.token_id}
          className='rounded-lg border px-4 py-3 transition-colors hover:bg-muted/30'
        >
          <div className='flex items-center gap-2'>
            <Key className='text-muted-foreground size-3.5' />
            <span className='text-sm font-medium truncate'>
              {stat.token_name}
            </span>
          </div>
          <div className='mt-3 grid grid-cols-3 gap-2 text-xs'>
            <div>
              <div className='text-muted-foreground'>{t('Requests')}</div>
              <div className='mt-0.5 font-semibold tabular-nums'>
                {formatNumber(stat.count)}
              </div>
            </div>
            <div>
              <div className='text-muted-foreground'>{t('Tokens')}</div>
              <div className='mt-0.5 font-semibold tabular-nums'>
                {formatNumber(stat.tokens)}
              </div>
            </div>
            <div>
              <div className='text-muted-foreground flex items-center gap-1'>
                <Coins className='size-3' />
                {t('Quota')}
              </div>
              <div className='mt-0.5 font-semibold tabular-nums'>
                {formatQuota(stat.quota)}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
