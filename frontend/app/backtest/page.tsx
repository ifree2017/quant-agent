'use client'
import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { quantApi, type BacktestResult } from '../api/client'
import { useAuth } from '../../contexts/AuthContext'

// Simple equity chart component
function SimpleEquityChart({ data, title }: { data: number[]; title: string }) {
  if (!data || data.length === 0) return null
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const points = data.map((v, i) => {
    const x = (i / (data.length - 1)) * 100
    const y = 100 - ((v - min) / range) * 100
    return `${x},${y}`
  }).join(' ')
  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
      <div className="text-xs text-slate-400 mb-2">{title}</div>
      <svg viewBox="0 0 100 60" className="w-full h-32" preserveAspectRatio="none">
        <polyline
          points={points}
          fill="none"
          stroke="#10b981"
          strokeWidth="1.5"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
    </div>
  )
}

// Simple trade table
function SimpleTradeTable({ trades }: { trades: any[] }) {
  if (!trades || trades.length === 0) return null
  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4 overflow-x-auto">
      <div className="text-xs text-slate-400 mb-2">交易记录</div>
      <table className="w-full text-xs">
        <thead>
          <tr className="text-slate-500 border-b border-slate-700">
            <th className="text-left py-1 pr-4">时间</th>
            <th className="text-left py-1 pr-4">股票</th>
            <th className="text-left py-1 pr-4">操作</th>
            <th className="text-right py-1 pr-4">价格</th>
            <th className="text-right py-1">数量</th>
          </tr>
        </thead>
        <tbody>
          {trades.slice(0, 20).map((t: any, i: number) => (
            <tr key={i} className="border-b border-slate-700/50 hover:bg-slate-700/30">
              <td className="py-1.5 pr-4 text-slate-400">{t.date || t.time || '-'}</td>
              <td className="py-1.5 pr-4 text-slate-200">{t.symbol || '-'}</td>
              <td className={`py-1.5 pr-4 font-medium ${t.action === 'BUY' || t.action === '买入' ? 'text-green-400' : 'text-red-400'}`}>
                {t.action === 'BUY' || t.action === '买入' ? '买入' : '卖出'}
              </td>
              <td className="py-1.5 pr-4 text-right font-mono text-slate-200">{t.price?.toFixed(2) || '-'}</td>
              <td className="py-1.5 text-right font-mono text-slate-400">{t.quantity || t.shares || '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default function BacktestPage() {
  const [symbol, setSymbol] = useState('000001')
  const [days, setDays] = useState('30')
  const [initialCash, setInitialCash] = useState('100000')
  const [result, setResult] = useState<BacktestResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const { user } = useAuth()
  const router = useRouter()

  const handleRun = async () => {
    if (!symbol.trim()) {
      setError('请输入股票代码')
      return
    }
    setLoading(true)
    setError('')
    setResult(null)
    try {
      const res = await quantApi.runBacktest({
        symbol,
        days: parseInt(days),
        initialCash: parseFloat(initialCash),
      })
      setResult(res)
    } catch {
      setError('请求失败，请确保后端服务运行中')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-background">
      {/* 顶部导航 */}
      <header className="sticky top-0 z-50 bg-background/90 backdrop-blur border-b border-slate-800 px-4 py-3">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link href="/" className="text-slate-400 hover:text-white text-sm transition-colors">📈 A股监控</Link>
            <span className="text-slate-600">/</span>
            <h1 className="text-sm font-bold text-white">回测</h1>
          </div>
          <div className="flex items-center gap-2">
            <Link href="/strategy/generate" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略生成</Link>
            <Link href="/compare" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略对比</Link>
            <Link href="/style/analyze" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">风格分析</Link>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-6 space-y-4">
        <h2 className="text-lg font-bold text-white mb-4">回测</h2>

        {/* 回测参数 */}
        <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
          <h3 className="text-sm font-semibold text-slate-300 mb-3">回测参数</h3>
          <div className="flex flex-wrap gap-3 mb-3">
            <div className="flex-1 min-w-[160px]">
              <label className="block text-xs text-slate-400 mb-1">股票代码</label>
              <input
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-600 focus:outline-none focus:border-blue-500"
                value={symbol}
                onChange={e => setSymbol(e.target.value)}
                placeholder="例如 000001"
              />
            </div>
            <div className="flex-1 min-w-[120px]">
              <label className="block text-xs text-slate-400 mb-1">回测天数</label>
              <input
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-600 focus:outline-none focus:border-blue-500"
                type="number"
                value={days}
                onChange={e => setDays(e.target.value)}
                placeholder="30"
              />
            </div>
            <div className="flex-1 min-w-[160px]">
              <label className="block text-xs text-slate-400 mb-1">初始资金</label>
              <input
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-600 focus:outline-none focus:border-blue-500"
                type="number"
                value={initialCash}
                onChange={e => setInitialCash(e.target.value)}
                placeholder="100000"
              />
            </div>
          </div>
          <button
            className="bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm px-4 py-2 rounded-lg transition-colors"
            onClick={handleRun}
            disabled={loading}
          >
            {loading ? '回测中...' : '运行回测'}
          </button>
          {error && <p className="text-red-400 text-xs mt-2">{error}</p>}
        </div>

        {/* 回测结果 */}
        {result && (
          <>
            <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
              <h3 className="text-sm font-semibold text-slate-300 mb-3">回测结果</h3>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                  <div className="text-[10px] text-slate-500">收益率</div>
                  <div className={`text-sm font-mono font-bold ${(result.returnRate ?? 0) >= 0 ? 'text-red-400' : 'text-green-400'}`}>
                    {result.returnRate != null ? `${((result.returnRate) * 100).toFixed(2)}%` : '-'}
                  </div>
                </div>
                <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                  <div className="text-[10px] text-slate-500">夏普比率</div>
                  <div className="text-sm font-mono text-slate-200">{result.sharpeRatio ?? '-'}</div>
                </div>
                <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                  <div className="text-[10px] text-slate-500">最大回撤</div>
                  <div className="text-sm font-mono text-red-400">
                    {result.maxDrawdown != null ? `${(result.maxDrawdown * 100).toFixed(2)}%` : '-'}
                  </div>
                </div>
                <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                  <div className="text-[10px] text-slate-500">胜率</div>
                  <div className="text-sm font-mono text-slate-200">
                    {result.winRate != null ? `${(result.winRate * 100).toFixed(2)}%` : '-'}
                  </div>
                </div>
              </div>
            </div>

            {result.equity && (
              <SimpleEquityChart data={result.equity} title="资金曲线" />
            )}

            {result.trades && (
              <SimpleTradeTable trades={result.trades} />
            )}
          </>
        )}
      </main>
    </div>
  )
}
