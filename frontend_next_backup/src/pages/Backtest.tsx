import { useState } from 'react'
import { runBacktest } from '../api/client'
import EquityChart from '../components/EquityChart'
import TradeTable from '../components/TradeTable'

export default function Backtest() {
  const [symbol, setSymbol] = useState('000001')
  const [days, setDays] = useState('30')
  const [initialCash, setInitialCash] = useState('100000')
  const [result, setResult] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleRun = async () => {
    if (!symbol.trim()) {
      setError('请输入股票代码')
      return
    }
    setLoading(true)
    setError('')
    setResult(null)
    try {
      const res = await runBacktest({
        symbol,
        days: parseInt(days),
        initialCash: parseFloat(initialCash)
      })
      setResult(res)
    } catch {
      setError('请求失败，请确保后端服务运行中')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <h2 style={{ marginBottom: '16px', fontSize: '20px', fontWeight: 600 }}>回测</h2>

      <div className="card">
        <div className="card-title">回测参数</div>

        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '12px' }}>
          <div className="form-group" style={{ flex: '1', minWidth: '160px', marginBottom: 0 }}>
            <label>股票代码</label>
            <input
              className="form-control"
              value={symbol}
              onChange={e => setSymbol(e.target.value)}
              placeholder="例如 000001"
            />
          </div>
          <div className="form-group" style={{ flex: '1', minWidth: '120px', marginBottom: 0 }}>
            <label>回测天数</label>
            <input
              className="form-control"
              type="number"
              value={days}
              onChange={e => setDays(e.target.value)}
              placeholder="30"
            />
          </div>
          <div className="form-group" style={{ flex: '1', minWidth: '160px', marginBottom: 0 }}>
            <label>初始资金</label>
            <input
              className="form-control"
              type="number"
              value={initialCash}
              onChange={e => setInitialCash(e.target.value)}
              placeholder="100000"
            />
          </div>
        </div>

        <div style={{ marginTop: '12px' }}>
          <button className="btn btn-primary" onClick={handleRun} disabled={loading}>
            {loading ? '回测中...' : '运行回测'}
          </button>
        </div>

        {error && <p className="error-msg">{error}</p>}
      </div>

      {result && (
        <>
          <div className="card">
            <div className="card-title">回测结果</div>
            <div className="result-box">
              <div className="result-item">
                <span className="result-label">收益率</span>
                <span className="result-value">
                  {result.returnRate != null ? `${(result.returnRate * 100).toFixed(2)}%` : '-'}
                </span>
              </div>
              <div className="result-item">
                <span className="result-label">夏普比率</span>
                <span className="result-value">{result.sharpeRatio ?? '-'}</span>
              </div>
              <div className="result-item">
                <span className="result-label">最大回撤</span>
                <span className="result-value">
                  {result.maxDrawdown != null ? `${(result.maxDrawdown * 100).toFixed(2)}%` : '-'}
                </span>
              </div>
              <div className="result-item">
                <span className="result-label">胜率</span>
                <span className="result-value">
                  {result.winRate != null ? `${(result.winRate * 100).toFixed(2)}%` : '-'}
                </span>
              </div>
            </div>
          </div>

          {result.equity && (
            <div className="card">
              <div className="card-title">资金曲线</div>
              <EquityChart data={result.equity} title="回测资金曲线" />
            </div>
          )}

          {result.trades && (
            <div className="card">
              <div className="card-title">交易记录</div>
              <TradeTable trades={result.trades} />
            </div>
          )}
        </>
      )}
    </div>
  )
}
