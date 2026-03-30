import { useState, useEffect } from 'react'
import { listStrategies, getBacktest } from '../api/client'

export default function Compare() {
  const [selected, setSelected] = useState<string[]>([])
  const [strategies, setStrategies] = useState<any[]>([])
  const [comparisons, setComparisons] = useState<any[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    listStrategies().then(d => setStrategies(d.strategies || []))
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
          const bt = await getBacktest(id)
          return { strategy: s, backtest: bt }
        })
      )
      setComparisons(results)
    } finally {
      setLoading(false)
    }
  }

  function toggleStrategy(id: string) {
    setSelected(prev =>
      prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]
    )
  }

  return (
    <div className="container">
      <h2>⚖️ 策略对比</h2>
      <p>选择2个以上策略进行绩效对比：</p>

      <div className="strategy-select">
        {strategies.map((s: any) => (
          <label key={s.id} className={selected.includes(s.id) ? 'selected' : ''}>
            <input
              type="checkbox"
              checked={selected.includes(s.id)}
              onChange={() => toggleStrategy(s.id)}
            />
            {s.name} ({s.style})
          </label>
        ))}
      </div>

      <button onClick={compare} disabled={loading}>
        {loading ? '对比中...' : '开始对比'}
      </button>

      {comparisons.length > 0 && (
        <div className="compare-table">
          <h3>绩效对比</h3>
          <table className="table">
            <thead>
              <tr>
                <th>指标</th>
                {comparisons.map(c => <th key={c.strategy.id}>{c.strategy.name}</th>)}
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>总收益率</td>
                {comparisons.map(c => (
                  <td key={c.strategy.id}>
                    {c.backtest?.metrics?.totalReturn ? (c.backtest.metrics.totalReturn * 100).toFixed(2) + '%' : '-'}
                  </td>
                ))}
              </tr>
              <tr>
                <td>夏普比率</td>
                {comparisons.map(c => (
                  <td key={c.strategy.id}>
                    {c.backtest?.metrics?.sharpeRatio?.toFixed(2) || '-'}
                  </td>
                ))}
              </tr>
              <tr>
                <td>最大回撤</td>
                {comparisons.map(c => (
                  <td key={c.strategy.id}>
                    {c.backtest?.metrics?.maxDrawdown ? (c.backtest.metrics.maxDrawdown * 100).toFixed(2) + '%' : '-'}
                  </td>
                ))}
              </tr>
              <tr>
                <td>胜率</td>
                {comparisons.map(c => (
                  <td key={c.strategy.id}>
                    {c.backtest?.metrics?.winRate ? (c.backtest.metrics.winRate * 100).toFixed(1) + '%' : '-'}
                  </td>
                ))}
              </tr>
              <tr>
                <td>盈亏比</td>
                {comparisons.map(c => (
                  <td key={c.strategy.id}>
                    {c.backtest?.metrics?.profitLossRatio?.toFixed(2) || '-'}
                  </td>
                ))}
              </tr>
              <tr>
                <td>总交易次数</td>
                {comparisons.map(c => (
                  <td key={c.strategy.id}>
                    {c.backtest?.metrics?.totalTrades || 0}
                  </td>
                ))}
              </tr>
            </tbody>
          </table>

          <h3>📊 收益对比</h3>
          <div className="bar-chart">
            {comparisons.map(c => {
              const ret = c.backtest?.metrics?.totalReturn || 0
              const width = Math.abs(ret) * 500
              return (
                <div key={c.strategy.id} className="bar-row">
                  <span className="bar-label">{c.strategy.name}</span>
                  <div className="bar-container">
                    <div
                      className={`bar ${ret >= 0 ? 'bar-pos' : 'bar-neg'}`}
                      style={{ width: `${width}px` }}
                    />
                    <span className="bar-value">{(ret * 100).toFixed(2)}%</span>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
