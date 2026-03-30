import { useState, useRef } from 'react'
import { analyzeStyle } from '../api/client'

export default function StyleAnalyze() {
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
      const headers = lines[0].split(',').map(h => h.trim())
      const data = lines.slice(1).map(line => {
        const vals = line.split(',')
        const obj: any = {}
        headers.forEach((h, i) => { obj[h] = vals[i] })
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
      const res = await analyzeStyle(records)
      setResult(res)
    } catch {
      setError('请求失败，请确保后端服务运行中')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <h2 style={{ marginBottom: '16px', fontSize: '20px', fontWeight: 600 }}>风格分析</h2>

      <div className="card">
        <div className="card-title">上传交易记录</div>

        <div
          className="file-upload"
          onClick={() => fileRef.current?.click()}
        >
          <input
            ref={fileRef}
            type="file"
            accept=".csv"
            onChange={handleFile}
          />
          <div style={{ fontSize: '16px', color: '#1890ff' }}>
            {file ? `已选择: ${file.name}` : '点击选择 CSV 文件'}
          </div>
          <div className="file-hint">支持 CSV 格式，包含 date, symbol, action, price, quantity 等字段</div>
        </div>

        {records.length > 0 && (
          <p style={{ fontSize: '13px', color: '#52c41a', marginBottom: '12px' }}>
            已解析 {records.length} 条交易记录
          </p>
        )}

        <button className="btn btn-primary" onClick={handleAnalyze} disabled={loading}>
          {loading ? '分析中...' : '开始分析'}
        </button>

        {error && <p className="error-msg">{error}</p>}

        {result && (
          <div className="result-box">
            <div className="result-item">
              <span className="result-label">交易风格</span>
              <span className="result-value">{result.style ?? '-'}</span>
            </div>
            <div className="result-item">
              <span className="result-label">风险评分</span>
              <span className="result-value">{result.riskScore ?? '-'}</span>
            </div>
            <div className="result-item">
              <span className="result-label">交易频率</span>
              <span className="result-value">{result.tradeFrequency ?? '-'}</span>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
