'use client'
import { useState, useEffect } from 'react'
import Link from 'next/link'
import { quantApi, type Strategy } from '../api/client'

export default function ComparePage() {
  const [selected, setSelected] = useState<string[]>([])
  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [comparisons, setComparisons] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  const [strategiesLoading, setStrategiesLoading] = useState(true)

  useEffect(() => {
    quantApi.listStrategies()
      .then(d => setStrategies(d.strategies || []))
      .catch(() => setStrategies([]))
      .finally(() => setStrategiesLoading(false))
  }, [])

  async function compare() {
    if (selected.length < 2) {
      alert('请至少选择2个策略')
      return
    }
    setLoading(true)
    try {
      const results = await Promise.all(
        selected.map(async (id) => {
          const s = strategies.find((x: any) => x.id === id)
          const bt = await quantApi.getBacktest(id)
          return { strategy: s, backtest: bt?.metrics || bt }
        })
      )
      setComparisons(results)
    } catch {
      setComparisons([])
    } finally {
      setLoading(false)
    }
  }

  function toggleStrategy(id: string) {
    setSelected(prev =>
      prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]
    )
  }

  const metrics = [
    { label: '总收益率', key: 'totalReturn', fmt: (v: number) => v != null ? `${(v * 100).toFixed(2)}%` : '-' },
    { label: '夏普比率', key: 'sharpeRatio', fmt: (v: number) => v != null ? v.toFixed(2) : '-' },
    { label: '最大回撤', key: 'maxDrawdown', fmt: (v: number) => v != null ? `${(v * 100).toFixed(2)}%` : '-' },
    { label: '胜率', key: 'winRate', fmt: (v: number) => v != null ? `${(v * 100).toFixed(1)}%` : '-' },
    { label: '盈亏比', key: 'profitLossRatio', fmt: (v: number) => v != null ? v.toFixed(2) : '-' },
    { label: '总交易次数', key: 'totalTrades', fmt: (v: number) => v != null ? v.toString() : '-' },
  ]

  return (
    <div className="min-h-screen bg-background">
      {/* 顶部导航 */}
      <header className="sticky top-0 z-50 bg-background/90 backdrop-blur border-b border-slate-800 px-4 py-3">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link href="/" className="text-slate-400 hover:text-white text-sm transition-colors">📈 A股监控</Link>
            <span className="text-slate-600">/</span>
            <h1 className="text-sm font-bold text-white">策略对比</h1>
          </div>
          <div className="flex items-center gap-2">
            <Link href="/strategy/list" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略列表</Link>
            <Link href="/strategy/generate" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略生成</Link>
            <Link href="/style/analyze" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">风格分析</Link>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-6 space-y-4">
        <h2 className="text-lg font-bold text-white mb-2">⚖️ 策略对比</h2>
        <p className="text-slate-400 text-sm">选择2个以上策略进行绩效对比</p>

        {/* 策略选择 */}
        <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
          <h3 className="text-sm font-semibold text-slate-300 mb-3">选择策略</h3>
          {strategiesLoading ? (
            <p className="text-slate-500 text-sm">加载中...</p>
          ) : strategies.length === 0 ? (
            <p className="text-slate-500 text-sm">暂无策略，请先 <Link href="/strategy/generate" className="text-blue-400 hover:underline">生成策略</Link></p>
          ) : (
            <div className="flex flex-wrap gap-2 mb-3">
              {strategies.map((s: any) => (
                <button
                  key={s.id}
                  onClick={() => toggleStrategy(s.id)}
                  className={`px-3 py-1.5 rounded-lg text-xs transition-colors border ${
                    selected.includes(s.id)
                      ? 'bg-blue-600/20 border-blue-500/40 text-blue-400'
                      : 'bg-slate-900/60 border-slate-700 text-slate-400 hover:border-slate-600'
                  }`}
                >
                  {selected.includes(s.id) && <span className="mr-1">✓</span>}
                  {s.name} ({s.style})
                </button>
              ))}
            </div>
          )}
          <button
            onClick={compare}
            disabled={loading || selected.length < 2}
            className="bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm px-4 py-2 rounded-lg transition-colors"
          >
            {loading ? '对比中...' : '开始对比'}
          </button>
        </div>

        {/* 对比结果 */}
        {comparisons.length > 0 && (
          <>
            {/* 指标表格 */}
            <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4 overflow-x-auto">
              <h3 className="text-sm font-semibold text-slate-300 mb-3">绩效对比</h3>
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-700">
                    <th className="text-left py-2 pr-4 text-slate-500 font-normal">指标</th>
                    {comparisons.map(c => (
                      <th key={c.strategy?.id} className="text-left py-2 px-2 text-slate-200 font-medium">
                        {c.strategy?.name || '-'}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {metrics.map(m => (
                    <tr key={m.key} className="border-b border-slate-700/50">
                      <td className="py-2 pr-4 text-slate-400">{m.label}</td>
                      {comparisons.map(c => (
                        <td key={c.strategy?.id} className="py-2 px-2 font-mono text-slate-200">
                          {m.fmt(c.backtest?.metrics?.[m.key] ?? c.backtest?.[m.key] ?? null)}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* 收益条形图 */}
            <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
              <h3 className="text-sm font-semibold text-slate-300 mb-3">收益对比</h3>
              <div className="space-y-2">
                {comparisons.map(c => {
                  const ret = c.backtest?.metrics?.totalReturn ?? c.backtest?.totalReturn ?? 0
                  const maxAbs = Math.max(...comparisons.map(x => Math.abs(x.backtest?.metrics?.totalReturn ?? x.backtest?.totalReturn ?? 0)), 0.001)
                  const width = Math.abs(ret) / maxAbs * 100
                  return (
                    <div key={c.strategy?.id} className="flex items-center gap-2">
                      <span className="text-xs text-slate-400 w-24 truncate">{c.strategy?.name || '-'}</span>
                      <div className="flex-1 h-4 bg-slate-900/60 rounded overflow-hidden relative">
                        {ret >= 0 ? (
                          <div className="absolute left-1/2 h-full bg-red-500/30 rounded-r" style={{ width: `${width}%`, marginLeft: 0 }} />
                        ) : (
                          <div className="absolute right-1/2 h-full bg-green-500/30 rounded-l" style={{ width: `${width}%`, marginRight: 0 }} />
                        )}
                      </div>
                      <span className={`text-xs font-mono w-16 text-right ${ret >= 0 ? 'text-red-400' : 'text-green-400'}`}>
                        {(ret * 100).toFixed(2)}%
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          </>
        )}
      </main>
    </div>
  )
}
