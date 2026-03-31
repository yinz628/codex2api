import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts'
import Modal from './Modal'
import { api } from '../api'
import type { AccountRow, AccountUsageDetail } from '../types'
import { getErrorMessage } from '../utils/error'

const COLORS = ['#7c3aed', '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#ec4899', '#8b5cf6', '#06b6d4', '#84cc16', '#f97316']

interface Props {
  account: AccountRow
  onClose: () => void
}

function formatTooltipValue(value: number | string | ReadonlyArray<number | string> | undefined): string {
  if (Array.isArray(value)) {
    return value.join(', ')
  }
  if (value === undefined) {
    return '0'
  }
  return String(value)
}

function StatRow({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border px-3.5 py-2">
      <span className="text-[13px] text-muted-foreground">{label}</span>
      <span className={`tabular-nums font-semibold ${highlight ? 'text-[15px] text-foreground' : 'text-[14px] text-foreground/80'}`}>
        {value}
      </span>
    </div>
  )
}

export default function AccountUsageModal({ account, onClose }: Props) {
  const { t } = useTranslation()
  const [data, setData] = useState<AccountUsageDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await api.getAccountUsage(account.id)
      setData(result)
    } catch (err) {
      setError(getErrorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [account.id])

  useEffect(() => {
    void load()
  }, [load])

  const title = `${t('accounts.usageDetailTitle')} - ${account.email || account.name || `#${account.id}`}`

  return (
    <Modal show title={title} onClose={onClose} contentClassName="sm:max-w-[720px]">
      {loading ? (
        <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">{t('common.loading')}</div>
      ) : error ? (
        <div className="py-8 text-center text-sm text-red-500">{error}</div>
      ) : !data || data.total_requests === 0 ? (
        <div className="py-12 text-center text-sm text-muted-foreground">{t('accounts.noUsageData')}</div>
      ) : (
        <div className="flex gap-6">
          <div className="shrink-0">
            <h4 className="mb-2 text-sm font-semibold">{t('accounts.modelDistribution')}</h4>
            <div className="h-[200px] w-[200px]">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={data.models}
                    dataKey="requests"
                    nameKey="model"
                    cx="50%"
                    cy="50%"
                    innerRadius={45}
                    outerRadius={85}
                    paddingAngle={2}
                    strokeWidth={0}
                  >
                    {data.models.map((_, i) => (
                      <Cell key={i} fill={COLORS[i % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    formatter={(value, name) => [`${formatTooltipValue(value)} 次`, name ?? '']}
                    contentStyle={{ fontSize: 12, borderRadius: 8, border: '1px solid hsl(var(--border))' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="mt-2 space-y-1">
              {data.models.map((modelUsage, i) => (
                <div key={modelUsage.model} className="flex items-center gap-2 text-[12px]">
                  <span className="size-2.5 shrink-0 rounded-full" style={{ background: COLORS[i % COLORS.length] }} />
                  <span className="truncate font-medium text-foreground">{modelUsage.model}</span>
                  <span className="ml-auto shrink-0 tabular-nums text-muted-foreground">{modelUsage.requests.toLocaleString()}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="flex-1 space-y-2.5">
            <StatRow label={t('accounts.totalRequests')} value={data.total_requests.toLocaleString()} highlight />
            <StatRow label={t('accounts.totalTokens')} value={data.total_tokens.toLocaleString()} highlight />
            <div className="h-px bg-border" />
            <StatRow label={t('accounts.inputTokens')} value={data.input_tokens.toLocaleString()} />
            <StatRow label={t('accounts.outputTokens')} value={data.output_tokens.toLocaleString()} />
            <StatRow label={t('accounts.reasoningTokens')} value={data.reasoning_tokens.toLocaleString()} />
            <StatRow label={t('accounts.cachedTokens')} value={data.cached_tokens.toLocaleString()} />
          </div>
        </div>
      )}
    </Modal>
  )
}
