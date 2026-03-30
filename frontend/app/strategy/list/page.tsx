'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { quantApi, type Strategy } from '../../api/client'

export default function StrategyListPage() {
  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    load()
  }, [])

  async function load() {
    setLoading(true)
    try {
      const data = await quantApi.listStrategies()
      setStrategies(data.strategies || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('确认删除？')) return
    await quantApi.deleteStrategy(id)
    load()
  }

  return (
    <div className="min-h-screen bg-background">
      {/* 顶部导航 */}
      <header className="sticky top-0 z-50 bg-background/90 backdrop-blur border-b border-slate-800 px-4 py-3">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link href="/" className="text-slate-400 hover:text-white text-sm transition-colors">📈 A股监控</Link>
            <span className="text-slate-600">/</span>
            <h1 className="text-sm font-bold text-white">策略列表</h1>
          </div>
          <div className="flex items-center gap-2">
            <Link href="/strategy/generate" className="bg-blue-600 hover:bg-blue-500 text-white text-xs px-3 py-1.5 rounded-lg transition-colors">+ 生成策略</Link>
            <Link href="/backtest" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">回测</Link>
            <Link href="/compare" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略对比</Link>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-6 space-y-4">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-lg font-bold text-white">📊 策略库</h2>
          <Link href="/strategy/generate" className="bg-blue-600 hover:bg-blue-500 text-white text-xs px-4 py-2 rounded-lg transition-colors">
            + 生成新策略
          </Link>
        </div>

        {loading ? (
          <div className="text-center py-16 text-slate-500 text-sm">加载中...</div>
        ) : strategies.length === 0 ? (
          <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-8 text-center">
            <p className="text-slate-400 text-sm mb-3">暂无策略</p>
            <Link href="/strategy/generate" className="text-blue-400 hover:underline text-sm">生成策略</Link>
          </div>
        ) : (
          <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-700 bg-slate-900/40">
                  <th className="text-left py-3 px-4 text-slate-500 font-normal text-xs">策略名称</th>
                  <th className="text-left py-3 px-4 text-slate-500 font-normal text-xs">风格</th>
                  <th className="text-left py-3 px-4 text-slate-500 font-normal text-xs">版本</th>
                  <th className="text-left py-3 px-4 text-slate-500 font-normal text-xs">创建时间</th>
                  <th className="text-right py-3 px-4 text-slate-500 font-normal text-xs">操作</th>
                </tr>
              </thead>
              <tbody>
                {strategies.map((s) => (
                  <tr key={s.id} className="border-b border-slate-700/50 hover:bg-slate-700/20 transition-colors">
                    <td className="py-3 px-4 text-slate-200 font-medium">{s.name}</td>
                    <td className="py-3 px-4">
                      <span className="text-xs px-2 py-0.5 bg-blue-500/10 text-blue-400 rounded">{s.style}</span>
                    </td>
                    <td className="py-3 px-4 text-slate-400 font-mono text-xs">v{s.version}</td>
                    <td className="py-3 px-4 text-slate-400 text-xs">{new Date(s.created_at).toLocaleDateString('zh-CN')}</td>
                    <td className="py-3 px-4 text-right">
                      <div className="flex justify-end gap-2">
                        <Link
                          href={`/backtest?strategyId=${s.id}`}
                          className="text-xs text-green-400 hover:text-green-300 transition-colors"
                        >
                          回测
                        </Link>
                        <button
                          onClick={() => handleDelete(s.id)}
                          className="text-xs text-red-400 hover:text-red-300 transition-colors"
                        >
                          删除
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>
    </div>
  )
}
