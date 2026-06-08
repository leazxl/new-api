import { useMemo, useState } from 'react'
import { VChart } from '@visactor/react-vchart'
import { BarChart3 as BarChartIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useThemeRadiusPx } from '@/lib/theme-radius'
import { useThemeCustomization } from '@/context/theme-customization-provider'
import { useTheme } from '@/context/theme-provider'
import { KEY_ANALYTICS_CHART_OPTIONS } from '@/features/dashboard/constants'
import type { TokenStat } from '@/features/dashboard/types'
import { formatNumber, formatQuota } from '@/lib/format'

type KeyChartTab = 'trend' | 'proportion' | 'top'

interface KeyChartsProps {
  data: TokenStat[]
  loading?: boolean
  defaultTab?: KeyChartTab
}

function buildChartSpecs(data: TokenStat[], chartRadius: number) {
  const items = [...data].sort((a, b) => b.quota - a.quota)
  const names = items.map((s) => s.token_name)

  return {
    spec_bar: {
      type: 'bar',
      data: { values: items.map((s) => ({ ...s, _label: s.token_name })) },
      xField: '_label',
      yField: 'quota',
      bar: { style: { cornerRadius: chartRadius } },
      axes: [
        {
          orient: 'bottom',
          label: { visible: true, style: { angle: items.length > 5 ? -30 : 0 } },
        },
        { orient: 'left', label: { visible: true } },
      ],
      tooltip: {
        mark: {
          title: { value: (datum: Record<string, unknown>) => datum?._label },
          content: [
            {
              key: 'Quota',
              value: (datum: Record<string, unknown>) =>
                formatQuota((datum?.quota as number) ?? 0),
            },
            {
              key: 'Requests',
              value: (datum: Record<string, unknown>) =>
                formatNumber((datum?.count as number) ?? 0),
            },
            {
              key: 'Tokens',
              value: (datum: Record<string, unknown>) =>
                formatNumber((datum?.tokens as number) ?? 0),
            },
          ],
        },
      },
    } as Record<string, unknown>,

    spec_pie: {
      type: 'pie',
      data: {
        values: items.map((s) => ({
          type: s.token_name,
          value: s.quota,
          count: s.count,
          tokens: s.tokens,
        })),
      },
      valueField: 'value',
      categoryField: 'type',
      outerLabel: { visible: true },
      tooltip: {
        mark: {
          title: { value: (datum: Record<string, unknown>) => datum?.type },
          content: [
            {
              key: 'Quota',
              value: (datum: Record<string, unknown>) =>
                formatQuota((datum?.value as number) ?? 0),
            },
          ],
        },
      },
    } as Record<string, unknown>,

    spec_rank_bar: {
      type: 'bar',
      data: {
        values: items.map((s, i) => ({
          ...s,
          _label: s.token_name,
          _rank: i + 1,
        })),
      },
      xField: 'quota',
      yField: '_label',
      direction: 'horizontal',
      bar: { style: { cornerRadius: chartRadius } },
      axes: [
        { orient: 'left', label: { visible: true } },
        { orient: 'bottom', label: { visible: true } },
      ],
      tooltip: {
        mark: {
          title: { value: (datum: Record<string, unknown>) => datum?._label },
          content: [
            {
              key: 'Quota',
              value: (datum: Record<string, unknown>) =>
                formatQuota((datum?.quota as number) ?? 0),
            },
            {
              key: 'Requests',
              value: (datum: Record<string, unknown>) =>
                formatNumber((datum?.count as number) ?? 0),
            },
          ],
        },
      },
    } as Record<string, unknown>,
  }
}

const CHART_SPEC_KEYS: Record<KeyChartTab, keyof ReturnType<typeof buildChartSpecs>> = {
  trend: 'spec_bar',
  proportion: 'spec_pie',
  top: 'spec_rank_bar',
}

export function KeyCharts(props: KeyChartsProps) {
  const { t } = useTranslation()
  const { resolvedTheme } = useTheme()
  const { customization } = useThemeCustomization()
  const chartRadius = useThemeRadiusPx(
    '--radius-md',
    `${customization.preset}:${customization.radius}`
  )
  const [activeTab, setActiveTab] = useState<KeyChartTab>(
    props.defaultTab ?? 'trend'
  )

  const data: TokenStat[] = props.loading ? [] : (props.data ?? [])

  const chartSpecs = useMemo(
    () => buildChartSpecs(data, chartRadius),
    [data, chartRadius]
  )

  const spec = chartSpecs[CHART_SPEC_KEYS[activeTab]]

  const chartKey = [
    activeTab,
    props.loading ? 'loading' : 'ready',
    data.length,
    resolvedTheme,
    customization.preset,
  ].join('-')

  if (!props.loading && data.length === 0) {
    return null
  }

  return (
    <div className='overflow-hidden rounded-lg border'>
      <div className='flex w-full flex-col gap-1.5 border-b px-3 py-2 sm:gap-3 sm:px-5 sm:py-3 lg:flex-row lg:items-center lg:justify-between'>
        <div className='flex items-center gap-2'>
          <BarChartIcon className='text-muted-foreground/60 size-4' />
          <div className='text-sm font-semibold'>
            {t('Key Analytics')}
          </div>
        </div>

        <div className='bg-muted/60 inline-flex h-7 w-full overflow-x-auto rounded-lg border p-0.5 sm:h-8 sm:w-auto'>
          {KEY_ANALYTICS_CHART_OPTIONS.map((tab) => (
            <button
              key={tab.value}
              type='button'
              onClick={() => setActiveTab(tab.value)}
              className={`shrink-0 rounded-md px-3 text-xs font-medium transition-colors ${
                activeTab === tab.value
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {t(tab.labelKey)}
            </button>
          ))}
        </div>
      </div>

      <div className='h-[300px] p-1.5 sm:h-96 sm:p-2'>
        {props.loading ? (
          <div className='flex h-full items-center justify-center'>
            <div className='text-muted-foreground text-sm'>
              {t('Loading...')}
            </div>
          </div>
        ) : (
          spec && (
            <VChart
              key={chartKey}
              spec={{
                ...spec,
                background: 'transparent',
              }}
            />
          )
        )}
      </div>
    </div>
  )
}
