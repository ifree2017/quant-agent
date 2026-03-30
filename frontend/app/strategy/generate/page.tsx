'use client'
import { useState } from 'react'
import Link from 'next/link'
import { quantApi } from '../../api/client'

const STYLES = ['趋势跟踪', '均值回归', '高频做市', '价值投资', '事件驱动']

// Simple equity chart
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

export default function StrategyGeneratePage() {
  const [symbol, setSymbol] = useState('000001')
  const [style, setStyle] = useState('趋势跟踪')
  const [result, setResult] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleGenerate = async () => {
    if (!symbol.trim()) {
      setError('请输入股票代码')
      return
    }
    setLoading(true)
    setError('')
    setResult(null)
    try {
      const res = await quantApi.generateStrategy({ style }, symbol)
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
            <h1 className="text-sm font-bold text-white">策略生成</h1>
          </div>
          <div className="flex items-center gap-2">
            <Link href="/strategy/list" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略列表</Link>
            <Link href="/backtest" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">回测</Link>
            <Link href="/compare" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略对比</Link>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-6 space-y-4">
        <h2 className="text-lg font-bold text-white mb-4">策略生成</h2>

        {/* 生成参数 */}
        <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
          <h3 className="text-sm font-semibold text-slate-300 mb-3">生成参数</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
            <div>
              <label className="block text-xs text-slate-400 mb-1">股票代码</label>
              <input
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-600 focus:outline-none focus:border-blue-500"
                value={symbol}
                onChange={e => setSymbol(e.target.value)}
                placeholder="例如 000001"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-400 mb-1">交易风格</label>
              <select
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
                value={style}
                onChange={e => setStyle(e.target.value)}
              >
                {STYLES.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
          </div>
          <button
            className="bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm px-4 py-2 rounded-lg transition-colors"
            onClick={handleGenerate}
            disabled={loading}
          >
            {loading ? '生成中...' : '生成策略'}
          </button>
          {error && <p className="text-red-400 text-xs mt-2">{error}</p>}
        </div>

        {/* 生成结果 */}
        {result && (
          <>
            {/* 策略内容 */}
            <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
              <h3 className="text-sm font-semibold text-slate-300 mb-3">生成的策略</h3>
              <div className="bg-slate-900/60 rounded-lg p-3">
                <pre className="text-xs text-slate-300 whitespace-pre-wrap overflow-x-auto">
                  {typeof result.strategy === 'string'
                    ? result.strategy
                    : JSON.stringify(result.strategy, null, 2)}
                </pre>
              </div>
            </div>

            {/* 回测报告 */}
            {result.backtest && (
              <>
                <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
                  <h3 className="text-sm font-semibold text-slate-300 mb-3">回测报告</h3>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                    <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                      <div className="text-[10px] text-slate-500">收益率</div>
                      <div className={`text-sm font-mono font-bold ${(result.backtest.returnRate ?? 0) >= 0 ? 'text-red-400' : 'text-green-400'}`}>
                        {result.backtest.returnRate != null ? `${(result.backtest.returnRate * 100).toFixed(2)}%` : '-'}
                      </div>
                    </div>
                    <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                      <div className="text-[10px] text-slate-500">夏普比率</div>
                      <div className="text-sm font-mono text-slate-200">{result.backtest.sharpeRatio ?? '-'}</div>
                    </div>
                    <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                      <div className="text-[10px] text-slate-500">最大回撤</div>
                      <div className="text-sm font-mono text-red-400">
                        {result.backtest.maxDrawdown != null ? `${(result.backtest.maxDrawdown * 100).toFixed(2)}%` : '-'}
                      </div>
                    </div>
                    <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                      <div className="text-[10px] text-slate-500">胜率</div>
                      <div className="text-sm font-mono text-slate-200">
                        {result.backtest.winRate != null ? `${(result.backtest.winRate * 100).toFixed(2)}%` : '-'}
                      </div>
                    </div>
                  </div>
                </div>

                {result.backtest.equity && (
                  <SimpleEquityChart data={result.backtest.equity} title="资金曲线" />
                )}
              </>
            )}
          </>
        )}
      </main>
    </div>
  )
}
