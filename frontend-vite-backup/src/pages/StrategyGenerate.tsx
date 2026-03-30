import { useState } from 'react'
import { generateStrategy } from '../api/client'
import EquityChart from '../components/EquityChart'

const STYLES = ['趋势跟踪', '均值回归', '高频做市', '价值投资', '事件驱动']

export default function StrategyGenerate() {
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
      const res = await generateStrategy({ style }, symbol)
      setResult(res)
    } catch {
      setError('请求失败，请确保后端服务运行中')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <h2 style={{ marginBottom: '16px', fontSize: '20px', fontWeight: 600 }}>策略生成</h2>

      <div className="card">
        <div className="card-title">生成参数</div>

        <div className="form-group">
          <label>股票代码</label>
          <input
            className="form-control"
            value={symbol}
            onChange={e => setSymbol(e.target.value)}
            placeholder="例如 000001"
          />
        </div>

        <div className="form-group">
          <label>交易风格</label>
          <select className="form-control" value={style} onChange={e => setStyle(e.target.value)}>
            {STYLES.map(s => <option key={s} value={s}>{s}</option>)}
          </select>
        </div>

        <button className="btn btn-primary" onClick={handleGenerate} disabled={loading}>
          {loading ? '生成中...' : '生成策略'}
        </button>

        {error && <p className="error-msg">{error}</p>}
      </div>

      {result && (
        <div className="card">
          <div className="card-title">生成的策略</div>
          <div className="result-box">
            {result.strategy && (
              <pre style={{ whiteSpace: 'pre-wrap', fontSize: '13px', color: '#333' }}>
                {typeof result.strategy === 'string'
                  ? result.strategy
                  : JSON.stringify(result.strategy, null, 2)}
              </pre>
            )}
          </div>

          {result.backtest && (
            <>
              <div className="card-title" style={{ marginTop: '16px' }}>回测报告</div>
              <div className="result-box">
                <div className="result-item">
                  <span className="result-label">收益率</span>
                  <span className="result-value">{result.backtest.returnRate != null ? `${(result.backtest.returnRate * 100).toFixed(2)}%` : '-'}</span>
                </div>
                <div className="result-item">
                  <span className="result-label">夏普比率</span>
                  <span className="result-value">{result.backtest.sharpeRatio ?? '-'}</span>
                </div>
                <div className="result-item">
                  <span className="result-label">最大回撤</span>
                  <span className="result-value">{result.backtest.maxDrawdown != null ? `${(result.backtest.maxDrawdown * 100).toFixed(2)}%` : '-'}</span>
                </div>
                <div className="result-item">
                  <span className="result-label">胜率</span>
                  <span className="result-value">{result.backtest.winRate != null ? `${(result.backtest.winRate * 100).toFixed(2)}%` : '-'}</span>
                </div>
              </div>

              {result.backtest.equity && (
                <div style={{ marginTop: '16px' }}>
                  <EquityChart data={result.backtest.equity} title="资金曲线" />
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  )
}
