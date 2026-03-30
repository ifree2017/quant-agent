'use client'
import { useState, useRef } from 'react'
import Link from 'next/link'
import { quantApi } from '../../api/client'

export default function StyleAnalyzePage() {
  const [file, setFile] = useState<File | null>(null)
  const [records, setRecords] = useState<any[]>([])
  const [result, setResult] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  const handleFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    if (!f) return
    setFile(f)
    setError('')

    const reader = new FileReader()
    reader.onload = (ev) => {
      const text = ev.target?.result as string
      const lines = text.trim().split('\n')
      const headers = lines[0].split(',').map((h: string) => h.trim())
      const data = lines.slice(1).map((line: string) => {
        const vals = line.split(',')
        const obj: any = {}
        headers.forEach((h: string, i: number) => { obj[h] = vals[i] })
        return obj
      })
      setRecords(data)
    }
    reader.readAsText(f)
  }

  const handleAnalyze = async () => {
    if (records.length === 0) {
      setError('请先上传CSV文件')
      return
    }
    setLoading(true)
    setError('')
    setResult(null)
    try {
      const res = await quantApi.analyzeStyle(records)
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
            <h1 className="text-sm font-bold text-white">风格分析</h1>
          </div>
          <div className="flex items-center gap-2">
            <Link href="/strategy/list" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略列表</Link>
            <Link href="/strategy/generate" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略生成</Link>
            <Link href="/compare" className="text-slate-400 hover:text-white text-xs px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors">策略对比</Link>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-6 space-y-4">
        <h2 className="text-lg font-bold text-white mb-4">风格分析</h2>

        {/* 上传交易记录 */}
        <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
          <h3 className="text-sm font-semibold text-slate-300 mb-3">上传交易记录</h3>

          {/* 文件选择 */}
          <div
            className="border-2 border-dashed border-slate-700 rounded-xl p-6 text-center cursor-pointer hover:border-blue-500/50 transition-colors mb-3"
            onClick={() => fileRef.current?.click()}
          >
            <input
              ref={fileRef}
              type="file"
              accept=".csv"
              onChange={handleFile}
              className="hidden"
            />
            <div className="text-sm text-blue-400 mb-1">
              {file ? `已选择: ${file.name}` : '点击选择 CSV 文件'}
            </div>
            <div className="text-xs text-slate-500">
              支持 CSV 格式，包含 date, symbol, action, price, quantity 等字段
            </div>
          </div>

          {records.length > 0 && (
            <p className="text-green-400 text-xs mb-3">
              ✓ 已解析 {records.length} 条交易记录
            </p>
          )}

          <button
            className="bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm px-4 py-2 rounded-lg transition-colors"
            onClick={handleAnalyze}
            disabled={loading}
          >
            {loading ? '分析中...' : '开始分析'}
          </button>

          {error && <p className="text-red-400 text-xs mt-2">{error}</p>}
        </div>

        {/* 分析结果 */}
        {result && (
          <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
            <h3 className="text-sm font-semibold text-slate-300 mb-3">分析结果</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                <div className="text-[10px] text-slate-500 mb-1">交易风格</div>
                <div className="text-sm font-medium text-slate-200">{result.style ?? '-'}</div>
              </div>
              <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                <div className="text-[10px] text-slate-500 mb-1">风险评分</div>
                <div className="text-sm font-mono text-slate-200">{result.riskScore ?? '-'}</div>
              </div>
              <div className="bg-slate-900/60 rounded-lg px-3 py-2">
                <div className="text-[10px] text-slate-500 mb-1">交易频率</div>
                <div className="text-sm text-slate-200">{result.tradeFrequency ?? '-'}</div>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
